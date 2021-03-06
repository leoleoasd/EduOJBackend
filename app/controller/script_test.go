package controller_test

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"github.com/EduOJ/backend/app/request"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/app/response/resource"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/database"
	"github.com/EduOJ/backend/database/models"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestJudgerGetScript(t *testing.T) {
	script := models.Script{
		Name:     "test_judger_get_script",
		Filename: "test_judger_get_script.zip",
	}
	assert.NoError(t, base.DB.Create(&script).Error)
	file := newFileContent("test_judger_get_script", "test_judger_get_script.zip", b64Encode("test_judger_get_script_zip_content"))
	_, err := base.Storage.PutObject(context.Background(), "scripts", script.Name, file.reader, file.size, minio.PutObjectOptions{})
	assert.NoError(t, err)
	_, err = file.reader.Seek(0, io.SeekStart)
	assert.NoError(t, err)

	noFileScript := models.Script{
		Name:     "test_no_file_script",
		Filename: "test_no_file_script",
	}
	assert.NoError(t, base.DB.Create(&noFileScript).Error)

	t.Run("Success", func(t *testing.T) {
		req := makeReq(t, "GET", base.Echo.Reverse("judger.getScript", script.Name), "", judgerAuthorize)
		resp := makeResp(req)
		assert.Equal(t, http.StatusFound, resp.StatusCode)
		content := getPresignedURLContent(t, resp.Header.Get("Location"))
		assert.Equal(t, "test_judger_get_script_zip_content", content)
	})

	t.Run("MissingName", func(t *testing.T) {
		req := makeReq(t, "GET", base.Echo.Reverse("judger.getScript"), "", judgerAuthorize)
		resp := makeResp(req)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		jsonEQ(t, response.ErrorResp("NOT_FOUND", nil), resp)
	})

	t.Run("MissingFile", func(t *testing.T) {
		req := makeReq(t, "GET", base.Echo.Reverse("judger.getScript", noFileScript.Name), "", judgerAuthorize)
		resp := makeResp(req)
		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.NotEmpty(t, resp.Header.Get("Location"))
		resp, err := http.Get(resp.Header.Get("Location"))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		bodyBuf := bytes.Buffer{}
		_, err = bodyBuf.ReadFrom(resp.Body)
		assert.NoError(t, err)
		var xmlresp struct {
			Name     xml.Name `xml:"Error"`
			Code     string   `xml:"Code"`
			Resource string   `xml:"Resource"`
		}
		assert.NoError(t, xml.Unmarshal(bodyBuf.Bytes(), &xmlresp))
		assert.Equal(t, struct {
			Name     xml.Name `xml:"Error"`
			Code     string   `xml:"Code"`
			Resource string   `xml:"Resource"`
		}{
			Code:     "NoSuchKey",
			Resource: "test_no_file_script",
		}, xmlresp)
	})
}

func TestCreateScript(t *testing.T) {
	t.Parallel()

	failTests := []failTest{
		{
			name:   "WithoutParams",
			method: "POST",
			path:   base.Echo.Reverse("script.createScript"),
			req: addFieldContentSlice([]reqContent{
				newFileContent("file", "script_file_name", b64Encode("test script file content")),
			}, nil),
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusBadRequest,
			resp: response.ErrorResp("VALIDATION_ERROR", []interface{}{
				map[string]interface{}{
					"field":       "Name",
					"reason":      "required",
					"translation": "名称为必填字段",
				},
			}),
		},
		{
			name:   "InvalidFile",
			method: "POST",
			path:   base.Echo.Reverse("script.createScript"),
			req: request.CreateScriptRequest{
				Name: "test_create_script_invalid_file",
			},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusBadRequest,
			resp:       response.ErrorResp("INVALID_FILE", nil),
		},
		{
			name:   "PermissionDenied",
			method: "POST",
			path:   base.Echo.Reverse("script.createScript"),
			req: addFieldContentSlice([]reqContent{
				newFileContent("file", "script_file_name", b64Encode("test script file content")),
			}, map[string]string{
				"Name": "test_create_script_permission_denied",
			}),
			reqOptions: []reqOption{applyNormalUser},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "")

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		httpResp := makeResp(makeReq(t, "POST", base.Echo.Reverse("script.createScript"), addFieldContentSlice([]reqContent{
			newFileContent("file", "script_file_name_success", b64Encode("test script file content")),
		}, map[string]string{
			"Name": "test_create_script_success",
		}), applyAdminUser))
		assert.Equal(t, http.StatusCreated, httpResp.StatusCode)

		databaseScript := models.Script{}
		assert.NoError(t, base.DB.First(&databaseScript, "name = ?", "test_create_script_success").Error)
		expectedScript := models.Script{
			Name:      "test_create_script_success",
			Filename:  "script_file_name_success",
			CreatedAt: databaseScript.CreatedAt,
			UpdatedAt: databaseScript.UpdatedAt,
		}
		assert.Equal(t, expectedScript, databaseScript)

		resp := response.CreateScriptResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.CreateScriptResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.Script `json:"script"`
			}{
				resource.GetScript(&expectedScript),
			},
		}, resp)

		assert.Equal(t, "test script file content", string(getObjectContent(t, "scripts", databaseScript.Name)))
	})
}

func createScriptForTest(t *testing.T, name string, id int) *models.Script {
	script := models.Script{
		Name:      fmt.Sprintf("test_%s_%d_name", name, id),
		Filename:  fmt.Sprintf("test_%s_%d_filename", name, id),
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}
	assert.NoError(t, base.DB.Create(&script).Error)
	content := fmt.Sprintf("test_%s_%d_content", name, id)
	_, err := base.Storage.PutObject(context.Background(), "scripts", script.Name, strings.NewReader(content),
		int64(len(content)), minio.PutObjectOptions{})
	assert.NoError(t, err)
	return &script
}

func TestGetScript(t *testing.T) {
	t.Parallel()

	script := createScriptForTest(t, "get_script", 0)

	failTests := []failTest{
		{
			name:       "NotFound",
			method:     "GET",
			path:       base.Echo.Reverse("script.getScript", "non_existing_script_name"),
			req:        request.GetScriptRequest{},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:       "PermissionDenied",
			method:     "GET",
			path:       base.Echo.Reverse("script.getScript", script.Name),
			req:        request.GetScriptRequest{},
			reqOptions: []reqOption{applyNormalUser},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "")

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("script.getScript", script.Name),
			request.GetScriptRequest{}, applyAdminUser))
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)

		resp := response.GetScriptResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.GetScriptResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.Script `json:"script"`
			}{
				resource.GetScript(script),
			},
		}, resp)
	})
}

func TestGetScriptFile(t *testing.T) {
	t.Parallel()

	script := createScriptForTest(t, "get_script_file", 0)
	failTests := []failTest{
		{
			name:       "NotFound",
			method:     "GET",
			path:       base.Echo.Reverse("script.getScriptFile", "non_existing_script_name"),
			req:        request.GetScriptRequest{},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:       "PermissionDenied",
			method:     "GET",
			path:       base.Echo.Reverse("script.getScriptFile", script.Name),
			req:        request.GetScriptRequest{},
			reqOptions: []reqOption{applyNormalUser},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "")

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("script.getScriptFile", script.Name),
			request.GetScriptRequest{}, applyAdminUser))
		assert.Equal(t, http.StatusFound, httpResp.StatusCode)
		assert.Equal(t, "test_get_script_file_0_content", getPresignedURLContent(t, httpResp.Header.Get("Location")))
	})
}

func TestGetScripts(t *testing.T) {
	// Not Parallel
	t.Cleanup(database.SetupDatabaseForTest())
	initGeneralTestingUsers()

	failTests := []failTest{
		{
			name:       "PermissionDenied",
			method:     "GET",
			path:       base.Echo.Reverse("script.getScripts"),
			req:        request.GetScriptRequest{},
			reqOptions: []reqOption{applyNormalUser},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "")

	script1 := createScriptForTest(t, "get_scripts", 1)
	script2 := createScriptForTest(t, "get_scripts", 2)
	script3 := createScriptForTest(t, "get_scripts", 3)
	script4 := createScriptForTest(t, "get_scripts", 4)
	script1.CreatedAt = script1.CreatedAt.Add(10 * time.Minute)
	script2.CreatedAt = script2.CreatedAt.Add(time.Hour)
	script3.CreatedAt = script3.CreatedAt.Add(20 * time.Minute)
	script4.CreatedAt = script4.CreatedAt.Add(30 * time.Minute)
	assert.NoError(t, base.DB.Save(&script1).Error)
	assert.NoError(t, base.DB.Save(&script2).Error)
	assert.NoError(t, base.DB.Save(&script3).Error)
	assert.NoError(t, base.DB.Save(&script4).Error)

	type respData struct {
		Scripts []*resource.Script `json:"scripts"`
		Total   int                `json:"total"`
		Count   int                `json:"count"`
		Offset  int                `json:"offset"`
		Prev    *string            `json:"prev"`
		Next    *string            `json:"next"`
	}

	successTests := []struct {
		name string
		req  request.GetScriptsRequest
		resp respData
	}{
		{
			name: "All",
			req: request.GetScriptsRequest{
				Limit:  4,
				Offset: 0,
			},
			resp: respData{
				Scripts: resource.GetScriptSlice([]*models.Script{
					script2,
					script4,
					script3,
					script1,
				}),
				Total:  6,
				Count:  4,
				Offset: 0,
				Prev:   nil,
				Next: getUrlStringPointer("script.getScripts", map[string]string{
					"offset": "4",
					"limit":  "4",
				}),
			},
		},
		{
			name: "Paginator",
			req: request.GetScriptsRequest{
				Limit:  2,
				Offset: 1,
			},
			resp: respData{
				Scripts: resource.GetScriptSlice([]*models.Script{
					script4,
					script3,
				}),
				Total:  6,
				Count:  2,
				Offset: 1,
				Prev:   nil,
				Next: getUrlStringPointer("script.getScripts", map[string]string{
					"offset": "3",
					"limit":  "2",
				}),
			},
		},
	}

	for _, test := range successTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("script.getScripts"),
				test.req, applyAdminUser))
			assert.Equal(t, http.StatusOK, httpResp.StatusCode)

			resp := response.GetScriptsResponse{}
			mustJsonDecode(httpResp, &resp)
			assert.Equal(t, response.GetScriptsResponse{
				Message: "SUCCESS",
				Error:   nil,
				Data:    test.resp,
			}, resp)

			var ss []models.Script
			assert.NoError(t, base.DB.Find(&ss).Error)
			t.Log("script slice:")
			for _, s := range ss {
				t.Logf("%+v\n", s)
			}

			t.Log("expected:")
			for _, s := range test.resp.Scripts {
				t.Logf("%+v\n", *s)
			}

			t.Log("actual:")
			for _, s := range resp.Data.Scripts {
				t.Logf("%+v\n", *s)
			}
		})
	}
}

func TestUpdateScript(t *testing.T) {
	t.Parallel()
	script := createScriptForTest(t, "update_script", 0)
	failTests := []failTest{
		{
			name:   "WithoutParams",
			method: "PUT",
			path:   base.Echo.Reverse("script.updateScript", script.Name),
			req: addFieldContentSlice([]reqContent{
				newFileContent("file", "script_file_name", b64Encode("test script file content")),
			}, nil),
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusBadRequest,
			resp: response.ErrorResp("VALIDATION_ERROR", []interface{}{
				map[string]interface{}{
					"field":       "Name",
					"reason":      "required",
					"translation": "名称为必填字段",
				},
			}),
		},
		{
			name:   "NonExistingScript",
			method: "PUT",
			path:   base.Echo.Reverse("script.updateScript", "non_existing_script_name"),
			req: request.UpdateScriptRequest{
				Name: "test_update_script_non_existing_script",
			},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "PermissionDenied",
			method: "PUT",
			path:   base.Echo.Reverse("script.updateScript", script.Name),
			req: addFieldContentSlice([]reqContent{
				newFileContent("file", "script_file_name", b64Encode("test script file content")),
			}, map[string]string{
				"Name": "test_update_script_permission_denied",
			}),
			reqOptions: []reqOption{applyNormalUser},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "")

	script1 := createScriptForTest(t, "update_script", 1)
	script2 := createScriptForTest(t, "update_script", 2)

	t.Run("SuccessWithFile", func(t *testing.T) {
		t.Parallel()

		httpResp := makeResp(makeReq(t, "PUT", base.Echo.Reverse("script.updateScript", script1.Name), addFieldContentSlice([]reqContent{
			newFileContent("file", "script_file_name_updated", b64Encode("test script file updated content")),
		}, map[string]string{
			"Name": "test_update_script_success_with_file",
		}), applyAdminUser))
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)

		databaseScript := models.Script{}
		assert.NoError(t, base.DB.First(&databaseScript, "name = ?", "test_update_script_success_with_file").Error)
		expectedScript := models.Script{
			Name:      "test_update_script_success_with_file",
			Filename:  "script_file_name_updated",
			CreatedAt: databaseScript.CreatedAt,
			UpdatedAt: databaseScript.UpdatedAt,
		}
		assert.Equal(t, expectedScript, databaseScript)

		resp := response.UpdateScriptResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.UpdateScriptResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.Script `json:"script"`
			}{
				resource.GetScript(&expectedScript),
			},
		}, resp)

		assert.Equal(t, "test script file updated content", string(getObjectContent(t, "scripts", databaseScript.Name)))
		checkObjectNonExist(t, "scripts", script1.Name)
	})
	t.Run("SuccessWithoutFile", func(t *testing.T) {
		t.Parallel()

		httpResp := makeResp(makeReq(t, "PUT", base.Echo.Reverse("script.updateScript", script2.Name), request.UpdateScriptRequest{
			Name: "test_update_script_success_without_file",
		}, applyAdminUser))
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)

		databaseScript := models.Script{}
		assert.NoError(t, base.DB.First(&databaseScript, "name = ?", "test_update_script_success_without_file").Error)
		expectedScript := models.Script{
			Name:      "test_update_script_success_without_file",
			Filename:  "test_update_script_2_filename",
			CreatedAt: databaseScript.CreatedAt,
			UpdatedAt: databaseScript.UpdatedAt,
		}
		assert.Equal(t, expectedScript, databaseScript)

		resp := response.UpdateScriptResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.UpdateScriptResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.Script `json:"script"`
			}{
				resource.GetScript(&expectedScript),
			},
		}, resp)

		assert.Equal(t, "test_update_script_2_content", string(getObjectContent(t, "scripts", databaseScript.Name)))
	})
}

func TestDeleteScript(t *testing.T) {
	t.Parallel()

	script := createScriptForTest(t, "delete_script_fail", 0)
	scriptInUseLanguage := createScriptForTest(t, "delete_script_in_use", 1)
	scriptInUseProblem := createScriptForTest(t, "delete_script_in_use", 2)
	language := models.Language{
		Name:             "test_delete_script_in_use",
		ExtensionAllowed: nil,
		BuildScriptName:  scriptInUseLanguage.Name,
		RunScriptName:    scriptInUseLanguage.Name,
	}
	assert.NoError(t, base.DB.Create(&language).Error)
	problem := models.Problem{
		Name:          "test_delete_script_in_use",
		CompareScript: *scriptInUseProblem,
	}
	assert.NoError(t, base.DB.Create(&problem).Error)
	failTests := []failTest{
		{
			name:       "PermissionDenied",
			method:     "DELETE",
			path:       base.Echo.Reverse("script.deleteScript", script.Name),
			req:        request.DeleteScriptRequest{},
			reqOptions: []reqOption{applyNormalUser},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			name:       "InUseLanguage",
			method:     "DELETE",
			path:       base.Echo.Reverse("script.deleteScript", scriptInUseLanguage.Name),
			req:        request.DeleteScriptRequest{},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusBadRequest,
			resp:       response.ErrorResp("SCRIPT_IN_USE", nil),
		},
		{
			name:       "InUseProblem",
			method:     "DELETE",
			path:       base.Echo.Reverse("script.deleteScript", scriptInUseProblem.Name),
			req:        request.DeleteScriptRequest{},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusBadRequest,
			resp:       response.ErrorResp("SCRIPT_IN_USE", nil),
		},
	}

	runFailTests(t, failTests, "")

	t.Run("Success", func(t *testing.T) {
		script := createScriptForTest(t, "delete_script_success", 0)
		httpResp := makeResp(makeReq(t, "DELETE", base.Echo.Reverse("script.deleteScript", script.Name),
			request.DeleteScriptRequest{}, applyAdminUser))
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		resp := response.Response{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.Response{
			Message: "SUCCESS",
			Error:   nil,
			Data:    nil,
		}, resp)
		assert.ErrorIs(t, gorm.ErrRecordNotFound, base.DB.First(&models.Script{}, "name = ?", script.Name).Error)
	})
	t.Run("NonExist", func(t *testing.T) {
		name := "non_existing_script_name"
		httpResp := makeResp(makeReq(t, "DELETE", base.Echo.Reverse("script.deleteScript", name),
			request.DeleteScriptRequest{}, applyAdminUser))
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		resp := response.Response{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.Response{
			Message: "SUCCESS",
			Error:   nil,
			Data:    nil,
		}, resp)
	})
}
