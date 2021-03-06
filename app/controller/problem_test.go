package controller_test

import (
	"context"
	"fmt"
	"github.com/EduOJ/backend/app/request"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/app/response/resource"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/utils"
	"github.com/EduOJ/backend/database"
	"github.com/EduOJ/backend/database/models"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

var inputTextBase64 = "aW5wdXQgdGV4dAo="
var outputTextBase64 = "b3V0cHV0IHRleHQK"
var newInputTextBase64 = "bmV3IGlucHV0IHRleHQK"
var newOutputTextBase64 = "bmV3IG91dHB1dCB0ZXh0"
var attachmentFileBase64 = "YXR0YWNobWVudCBmaWxlIGZvciB0ZXN0"
var newAttachmentFileBase64 = "bmV3IGF0dGFjaG1lbnQgZmlsZSBmb3IgdGVzdAo="

type testCaseData struct {
	Score      uint
	Sample     bool
	InputFile  *fileContent
	OutputFile *fileContent
}

func createProblemForTest(t *testing.T, name string, id int, attachmentFile *fileContent, creator models.User) (problem models.Problem) {
	problem = models.Problem{
		Name:               fmt.Sprintf("problem_for_testing_%s_%d", name, id),
		Description:        fmt.Sprintf("a problem used to test API: %s(%d)", name, id),
		AttachmentFileName: "",
		Public:             true,
		Privacy:            false,
		MemoryLimit:        1024,
		TimeLimit:          1000,
		LanguageAllowed:    []string{"test_language"},
		BuildArg:           fmt.Sprintf("test_%s_build_arg_%d", name, id),
		CompareScriptName:  "cmp1",
	}
	if attachmentFile != nil {
		problem.AttachmentFileName = attachmentFile.fileName
	}
	assert.NoError(t, base.DB.Create(&problem).Error)
	creator.GrantRole("problem_creator", problem)
	if attachmentFile != nil {
		_, err := base.Storage.PutObject(context.Background(), "problems", fmt.Sprintf("%d/attachment", problem.ID), attachmentFile.reader, attachmentFile.size, minio.PutObjectOptions{})
		assert.NoError(t, err)
		_, err = attachmentFile.reader.Seek(0, io.SeekStart)
		assert.NoError(t, err)
	}
	return
}

func createTestCaseForTest(t *testing.T, problem models.Problem, data testCaseData) (testCase models.TestCase) {
	var inputFileName, outputFileName string

	if data.InputFile != nil {
		inputFileName = data.InputFile.fileName
	}
	if data.OutputFile != nil {
		outputFileName = data.OutputFile.fileName
	}

	testCase = models.TestCase{
		Score:          data.Score,
		Sample:         data.Sample,
		InputFileName:  inputFileName,
		OutputFileName: outputFileName,
	}
	assert.NoError(t, base.DB.Model(&problem).Association("TestCases").Append(&testCase))

	if data.InputFile != nil {
		_, err := base.Storage.PutObject(context.Background(), "problems", fmt.Sprintf("%d/input/%d.in", problem.ID, testCase.ID), data.InputFile.reader, data.InputFile.size, minio.PutObjectOptions{})
		assert.NoError(t, err)
		_, err = data.InputFile.reader.Seek(0, io.SeekStart)
		assert.NoError(t, err)
	}
	if data.OutputFile != nil {
		_, err := base.Storage.PutObject(context.Background(), "problems", fmt.Sprintf("%d/output/%d.out", problem.ID, testCase.ID), data.OutputFile.reader, data.OutputFile.size, minio.PutObjectOptions{})
		assert.NoError(t, err)
		_, err = data.OutputFile.reader.Seek(0, io.SeekStart)
		assert.NoError(t, err)
	}

	return
}

func TestGetProblem(t *testing.T) {
	t.Parallel()

	// publicFalseProblem means a problem which "public" field is false
	publicFalseProblem := models.Problem{
		Name:               "test_get_problem_public_false",
		AttachmentFileName: "test_get_problem_public_false_attachment_file_name",
		LanguageAllowed:    []string{"test_get_problem_public_false_language_allowed"},
	}
	assert.NoError(t, base.DB.Create(&publicFalseProblem).Error)

	failTests := []failTest{
		{
			name:   "NonExistId",
			method: "GET",
			path:   base.Echo.Reverse("problem.getProblem", -1),
			req:    request.GetProblemRequest{},
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "PublicFalseFail",
			method: "GET",
			path:   base.Echo.Reverse("problem.getProblem", publicFalseProblem.ID),
			req:    request.GetProblemRequest{},
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
	}

	runFailTests(t, failTests, "GetProblem")

	successTests := []struct {
		name      string
		path      string
		req       request.GetProblemRequest
		problem   models.Problem
		isAdmin   bool
		testCases []models.TestCase
	}{
		{
			name: "AdminUserWithoutTestCases",
			path: "id",
			req:  request.GetProblemRequest{},
			problem: models.Problem{
				Name:               "test_get_problem_1",
				AttachmentFileName: "test_get_problem_1_attachment_file_name",
				LanguageAllowed:    []string{"test_get_problem_1_language_allowed"},
				Public:             true,
			},
			isAdmin:   true,
			testCases: nil,
		},
		{
			name: "NormalUserWithoutTestCases",
			path: "id",
			req:  request.GetProblemRequest{},
			problem: models.Problem{
				Name:               "test_get_problem_2",
				AttachmentFileName: "test_get_problem_2_attachment_file_name",
				LanguageAllowed:    []string{"test_get_problem_2_language_allowed"},
				Public:             true,
			},
			isAdmin:   false,
			testCases: nil,
		},
		{
			name: "PublicFalseSuccess",
			path: "id",
			req:  request.GetProblemRequest{},
			problem: models.Problem{
				Name:               "test_get_problem_3",
				AttachmentFileName: "test_get_problem_3_attachment_file_name",
				LanguageAllowed:    []string{"test_get_problem_3_language_allowed"},
				Public:             false,
			},
			isAdmin:   true,
			testCases: nil,
		},
		{
			name: "AdminUserWithTestCases",
			path: "id",
			req:  request.GetProblemRequest{},
			problem: models.Problem{
				Name:               "test_admin_get_problem_4",
				AttachmentFileName: "test_admin_get_problem_4_attachment_file_name",
				LanguageAllowed:    []string{"test_get_problem_4_language_allowed"},
				Public:             true,
			},
			isAdmin: true,
			testCases: []models.TestCase{
				{
					Score:          100,
					InputFileName:  "test_admin_get_problem_4_test_case_1_input_file_name",
					OutputFileName: "test_admin_get_problem_4_test_case_1_output_file_name",
				},
				{
					Score:          100,
					InputFileName:  "test_admin_get_problem_4_test_case_2_input_file_name",
					OutputFileName: "test_admin_get_problem_4_test_case_2_output_file_name",
				},
			},
		},
		{
			name: "NormalUserWithTestCases",
			path: "id",
			req:  request.GetProblemRequest{},
			problem: models.Problem{
				Name:               "test_admin_get_problem_5",
				AttachmentFileName: "test_admin_get_problem_5_attachment_file_name",
				LanguageAllowed:    []string{"test_get_problem_5_language_allowed"},
				Public:             true,
			},
			isAdmin: false,
			testCases: []models.TestCase{
				{
					Score:          100,
					Sample:         true,
					InputFileName:  "test_admin_get_problem_5_test_case_1_input_file_name",
					OutputFileName: "test_admin_get_problem_5_test_case_1_output_file_name",
				},
				{
					Score:          100,
					Sample:         false,
					InputFileName:  "test_admin_get_problem_5_test_case_2_input_file_name",
					OutputFileName: "test_admin_get_problem_5_test_case_2_output_file_name",
				},
			},
		},
	}

	t.Run("testGetProblemSuccess", func(t *testing.T) {
		t.Parallel()
		for i, test := range successTests {
			i := i
			test := test
			t.Run("testGetProblem"+test.name, func(t *testing.T) {
				t.Parallel()
				assert.NoError(t, base.DB.Create(&test.problem).Error)
				for j := range test.testCases {
					assert.NoError(t, base.DB.Model(&test.problem).Association("TestCases").Append(&test.testCases[j]))
				}
				user := createUserForTest(t, "get_problem", i)
				if test.isAdmin {
					user.GrantRole("problem_creator", test.problem)
				}
				httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("problem.getProblem", test.problem.ID), request.GetUserRequest{}, headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", user.ID)},
				}))
				assert.Equal(t, http.StatusOK, httpResp.StatusCode)
				if test.isAdmin {
					resp := response.GetProblemResponseForAdmin{}
					expectResp := response.GetProblemResponseForAdmin{
						Message: "SUCCESS",
						Error:   nil,
						Data: struct {
							*resource.ProblemForAdmin `json:"problem"`
						}{
							resource.GetProblemForAdmin(&test.problem),
						},
					}
					mustJsonDecode(httpResp, &resp)
					assert.Equal(t, expectResp, resp)
				} else {
					resp := response.GetProblemResponse{}
					expectResp := response.GetProblemResponse{
						Message: "SUCCESS",
						Error:   nil,
						Data: struct {
							*resource.Problem `json:"problem"`
						}{
							resource.GetProblem(&test.problem),
						},
					}
					mustJsonDecode(httpResp, &resp)
					assert.Equal(t, expectResp, resp)
				}
			})
		}
	})
}

func TestGetProblems(t *testing.T) {
	t.Parallel()

	problem1 := models.Problem{
		Name:               "test_get_problems_1",
		Description:        "test_get_problems_1_description",
		AttachmentFileName: "test_get_problems_1_attachment_file_name",
		LanguageAllowed:    []string{"test_get_problems_1_language_allowed"},
		Public:             true,
	}
	problem2 := models.Problem{
		Name:               "test_get_problems_2",
		Description:        "test_get_problems_2_description",
		AttachmentFileName: "test_get_problems_2_attachment_file_name",
		LanguageAllowed:    []string{"test_get_problems_2_language_allowed"},
		Public:             true,
	}
	problem3 := models.Problem{
		Name:               "test_get_problems_3",
		Description:        "test_get_problems_3_description",
		AttachmentFileName: "test_get_problems_3_attachment_file_name",
		LanguageAllowed:    []string{"test_get_problems_3_language_allowed"},
		Public:             true,
	}
	problem4 := models.Problem{
		Name:               "test_get_problems_4",
		Description:        "test_get_problems_4_description",
		AttachmentFileName: "test_get_problems_4_attachment_file_name",
		LanguageAllowed:    []string{"test_get_problems_4_language_allowed"},
		Public:             false,
	}
	assert.NoError(t, base.DB.Create(&problem1).Error)
	assert.NoError(t, base.DB.Create(&problem2).Error)
	assert.NoError(t, base.DB.Create(&problem3).Error)
	assert.NoError(t, base.DB.Create(&problem4).Error)

	user := createUserForTest(t, "get_problems_submitter", 0)
	otherUser := createUserForTest(t, "get_problems_submitter", 1)
	submissionPassed1 := createSubmissionForTest(t, "get_problems_1_passed", 1, &problem1, &user, nil, 0)
	submissionPassed2 := createSubmissionForTest(t, "get_problems_1_passed", 2, &problem1, &otherUser, nil, 0)
	submissionPassed3 := createSubmissionForTest(t, "get_problems_2_passed", 3, &problem2, &user, nil, 0)
	submissionPassed4 := createSubmissionForTest(t, "get_problems_2_passed", 4, &problem2, &user, nil, 0)
	submissionPassed1.Status = "ACCEPTED"
	submissionPassed2.Status = "ACCEPTED"
	submissionPassed3.Status = "ACCEPTED"
	submissionPassed4.Status = "ACCEPTED"
	assert.NoError(t, base.DB.Save(&submissionPassed1).Error)
	assert.NoError(t, base.DB.Save(&submissionPassed2).Error)
	assert.NoError(t, base.DB.Save(&submissionPassed3).Error)
	assert.NoError(t, base.DB.Save(&submissionPassed4).Error)

	submissionFailed1 := createSubmissionForTest(t, "get_problems_1_failed", 1, &problem1, &otherUser, nil, 0)
	submissionFailed2 := createSubmissionForTest(t, "get_problems_2_failed", 2, &problem2, &user, nil, 0)
	submissionFailed3 := createSubmissionForTest(t, "get_problems_3_failed", 3, &problem3, &user, nil, 0)
	submissionFailed4 := createSubmissionForTest(t, "get_problems_3_failed", 4, &problem3, &user, nil, 0)
	submissionFailed1.Status = "WRONG_ANSWER"
	submissionFailed2.Status = "TIME_LIMIT_EXCEEDED"
	submissionFailed3.Status = "RUNTIME_ERROR"
	submissionFailed4.Status = "PENDING"
	assert.NoError(t, base.DB.Save(&submissionFailed1).Error)
	assert.NoError(t, base.DB.Save(&submissionFailed2).Error)
	assert.NoError(t, base.DB.Save(&submissionFailed3).Error)
	assert.NoError(t, base.DB.Save(&submissionFailed4).Error)

	failTests := []failTest{
		{
			name:   "InvalidStatus",
			method: "GET",
			path:   base.Echo.Reverse("problem.getProblems"),
			req: request.GetProblemsRequest{
				Search: "",
				UserID: user.ID, // non-existing user id
				Limit:  0,
				Offset: 0,
				Tried:  true,
				Passed: true,
			},
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusBadRequest,
			resp:       response.ErrorResp("INVALID_STATUS", nil),
		},
		{
			name:   "WithoutUserId",
			method: "GET",
			path:   base.Echo.Reverse("problem.getProblems"),
			req: request.GetProblemsRequest{
				Search: "",
				UserID: 0,
				Limit:  0,
				Offset: 0,
				Tried:  true,
				Passed: false,
			},
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusBadRequest,
			resp: response.ErrorResp("VALIDATION_ERROR", []interface{}{
				map[string]interface{}{
					"field":       "UserID",
					"reason":      "required_with",
					"translation": "当选取尝试过题目或选取通过题目不为空时，用户ID为必填字段",
				},
			}),
		},
	}

	runFailTests(t, failTests, "")

	problem1.Description = ""
	problem2.Description = ""
	problem3.Description = ""
	problem4.Description = ""

	type respData struct {
		Problems []*models.Problem `json:"problems"`
		Total    int               `json:"total"`
		Count    int               `json:"count"`
		Offset   int               `json:"offset"`
		Prev     *string           `json:"prev"`
		Next     *string           `json:"next"`
	}

	successTests := []struct {
		name     string
		req      request.GetProblemsRequest
		respData respData
		isAdmin  bool
	}{
		{
			name: "All",
			req: request.GetProblemsRequest{
				Search: "test_get_problems",
				UserID: 0,
				Limit:  0,
				Offset: 0,
				Tried:  false,
				Passed: false,
			},
			respData: respData{
				Problems: []*models.Problem{
					&problem1,
					&problem2,
					&problem3,
				},
				Total:  3,
				Count:  3,
				Offset: 0,
				Prev:   nil,
				Next:   nil,
			},
			isAdmin: false,
		},
		{
			name: "AllWithAdminPermission",
			req: request.GetProblemsRequest{
				Search: "test_get_problems",
				UserID: 0,
				Limit:  0,
				Offset: 0,
				Tried:  false,
				Passed: false,
			},
			respData: respData{
				Problems: []*models.Problem{
					&problem1,
					&problem2,
					&problem3,
					&problem4,
				},
				Total:  4,
				Count:  4,
				Offset: 0,
				Prev:   nil,
				Next:   nil,
			},
			isAdmin: true,
		},
		{
			name: "NonExist",
			req: request.GetProblemsRequest{
				Search: "test_get_problems_non_exist",
				UserID: 0,
				Limit:  0,
				Offset: 0,
				Tried:  false,
				Passed: false,
			},
			respData: respData{
				Problems: []*models.Problem{},
				Total:    0,
				Count:    0,
				Offset:   0,
				Prev:     nil,
				Next:     nil,
			},
			isAdmin: false,
		},
		{
			name: "Search",
			req: request.GetProblemsRequest{
				Search: "test_get_problems_2",
				UserID: 0,
				Limit:  0,
				Offset: 0,
				Tried:  false,
				Passed: false,
			},
			respData: respData{
				Problems: []*models.Problem{
					&problem2,
				},
				Total:  1,
				Count:  1,
				Offset: 0,
				Prev:   nil,
				Next:   nil,
			},
			isAdmin: false,
		},
		{
			name: "Paginator",
			req: request.GetProblemsRequest{
				Search: "test_get_problems",
				UserID: 0,
				Limit:  2,
				Offset: 0,
				Tried:  false,
				Passed: false,
			},
			respData: respData{
				Problems: []*models.Problem{
					&problem1,
					&problem2,
				},
				Total:  3,
				Count:  2,
				Offset: 0,
				Prev:   nil,
				Next: getUrlStringPointer("problem.getProblems", map[string]string{
					"limit":  "2",
					"offset": "2",
				}),
			},
			isAdmin: false,
		},
		{
			name: "Passed",
			req: request.GetProblemsRequest{
				Search: "test_get_problems",
				UserID: user.ID,
				Limit:  0,
				Offset: 0,
				Tried:  false,
				Passed: true,
			},
			respData: respData{
				Problems: []*models.Problem{
					&problem1,
					&problem2,
				},
				Total:  2,
				Count:  2,
				Offset: 0,
				Prev:   nil,
				Next:   nil,
			},
			isAdmin: false,
		},
		{
			name: "Tried",
			req: request.GetProblemsRequest{
				Search: "test_get_problems",
				UserID: user.ID,
				Limit:  0,
				Offset: 0,
				Tried:  true,
				Passed: false,
			},
			respData: respData{
				Problems: []*models.Problem{
					&problem3,
				},
				Total:  1,
				Count:  1,
				Offset: 0,
				Prev:   nil,
				Next:   nil,
			},
			isAdmin: false,
		},
	}

	t.Run("testGetProblemsSuccess", func(t *testing.T) {
		t.Parallel()
		for i, test := range successTests {
			i := i
			test := test
			t.Run("testGetProblems"+test.name, func(t *testing.T) {
				t.Parallel()
				user := createUserForTest(t, "get_problems", i)
				// assert.False(t,user.Can("manage_problem"))
				if test.isAdmin {
					user.GrantRole("admin")
				}
				httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("problem.getProblems"), test.req, headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", user.ID)},
				}))
				assert.Equal(t, http.StatusOK, httpResp.StatusCode)

				if test.isAdmin {
					resp := response.GetProblemsResponseForAdmin{}
					mustJsonDecode(httpResp, &resp)
					expectResp := response.GetProblemsResponseForAdmin{
						Message: "SUCCESS",
						Error:   nil,
						Data: struct {
							Problems []resource.ProblemForAdmin `json:"problems"`
							Total    int                        `json:"total"`
							Count    int                        `json:"count"`
							Offset   int                        `json:"offset"`
							Prev     *string                    `json:"prev"`
							Next     *string                    `json:"next"`
						}{
							Problems: resource.GetProblemForAdminSlice(test.respData.Problems),
							Total:    test.respData.Total,
							Count:    test.respData.Count,
							Offset:   test.respData.Offset,
							Prev:     test.respData.Prev,
							Next:     test.respData.Next,
						},
					}
					assert.Equal(t, expectResp, resp)
				} else {
					resp := response.GetProblemsResponse{}
					mustJsonDecode(httpResp, &resp)
					expectResp := response.GetProblemsResponse{
						Message: "SUCCESS",
						Error:   nil,
						Data: struct {
							Problems []resource.Problem `json:"problems"`
							Total    int                `json:"total"`
							Count    int                `json:"count"`
							Offset   int                `json:"offset"`
							Prev     *string            `json:"prev"`
							Next     *string            `json:"next"`
						}{
							Problems: resource.GetProblemSlice(test.respData.Problems),
							Total:    test.respData.Total,
							Count:    test.respData.Count,
							Offset:   test.respData.Offset,
							Prev:     test.respData.Prev,
							Next:     test.respData.Next,
						},
					}
					assert.Equal(t, expectResp, resp)
				}

			})
		}
	})
}

func TestGetProblemAttachmentFile(t *testing.T) {
	t.Parallel()
	problemWithoutAttachmentFile := models.Problem{
		Name:               "test_get_problem_attachment_file_0",
		Description:        "test_get_problem_attachment_file_0_desc",
		AttachmentFileName: "",
		Public:             true,
		Privacy:            false,
		MemoryLimit:        1024,
		TimeLimit:          1000,
		LanguageAllowed:    []string{"test_get_problem_attachment_file_0_language_allowed"},
		BuildArg:           "test_get_problem_attachment_file_0_build_arg",
		CompareScriptName:  "cmp1",
	}
	// publicFalseProblem means a problem which "public" field is false
	publicFalseProblem := models.Problem{
		Name:               "test_get_problem_attachment_file_1",
		Description:        "test_get_problem_attachment_file_1_desc",
		AttachmentFileName: "",
		Public:             false,
		Privacy:            false,
		MemoryLimit:        1024,
		TimeLimit:          1000,
		LanguageAllowed:    []string{"test_get_problem_attachment_file_1_language_allowed"},
		BuildArg:           "test_get_problem_attachment_file_1_build_arg",
		CompareScriptName:  "cmp1",
	}
	assert.NoError(t, base.DB.Create(&problemWithoutAttachmentFile).Error)
	assert.NoError(t, base.DB.Create(&publicFalseProblem).Error)

	failTests := []failTest{
		{
			name:   "NonExistingProblem",
			method: "GET",
			path:   base.Echo.Reverse("problem.getProblemAttachmentFile", -1),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "ProblemWithoutAttachmentFile",
			method: "GET",
			path:   base.Echo.Reverse("problem.getProblemAttachmentFile", problemWithoutAttachmentFile.ID),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("ATTACHMENT_NOT_FOUND", nil),
		},
		{
			name:   "PublicFalse",
			method: "GET",
			path:   base.Echo.Reverse("problem.getProblemAttachmentFile", publicFalseProblem.ID),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
	}

	runFailTests(t, failTests, "GetProblemAttachmentFile")

	successTests := []struct {
		name string
		file *fileContent
	}{
		{
			name: "PDFFile",
			file: newFileContent("", "test_get_problem_attachment.pdf", "cGRmIGNvbnRlbnQK"),
		},
		{
			name: "NonPDFFile",
			file: newFileContent("", "test_get_problem_attachment.txt", "dHh0IGNvbnRlbnQK"),
		},
	}

	t.Run("testGetProblemAttachmentFileSuccess", func(t *testing.T) {
		t.Parallel()
		for i, test := range successTests {
			i := i
			test := test
			t.Run("testGetProblemAttachmentFile"+test.name, func(t *testing.T) {
				t.Parallel()
				user := createUserForTest(t, "test_get_problem_attachment_file", i+2)
				problem := createProblemForTest(t, "test_get_problem_attachment_file", i+2, test.file, user)
				httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("problem.getProblemAttachmentFile", problem.ID), nil, applyNormalUser))
				fileBytes, err := ioutil.ReadAll(test.file.reader)
				assert.NoError(t, err)
				assert.Equal(t, string(fileBytes), getPresignedURLContent(t, httpResp.Header.Get("Location")))
			})
		}
	})

}

func TestCreateProblem(t *testing.T) {
	t.Parallel()
	FailTests := []failTest{
		{
			// testCreateProblemWithoutParams
			name:   "WithoutParams",
			method: "POST",
			path:   base.Echo.Reverse("problem.createProblem"),
			req: request.CreateProblemRequest{
				Name:              "",
				Description:       "",
				Public:            nil,
				Privacy:           nil,
				MemoryLimit:       0,
				TimeLimit:         0,
				LanguageAllowed:   "",
				BuildArg:          "",
				CompareScriptName: "",
			},
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusBadRequest,
			resp: response.ErrorResp("VALIDATION_ERROR", []interface{}{
				map[string]interface{}{
					"field":       "Name",
					"reason":      "required",
					"translation": "名称为必填字段",
				},
				map[string]interface{}{
					"field":       "Description",
					"reason":      "required",
					"translation": "描述为必填字段",
				},
				map[string]interface{}{
					"field":       "Public",
					"reason":      "required",
					"translation": "是否公开为必填字段",
				},
				map[string]interface{}{
					"field":       "Privacy",
					"reason":      "required",
					"translation": "是否可以看到详细评测结果为必填字段",
				},
				map[string]interface{}{
					"field":       "MemoryLimit",
					"reason":      "required",
					"translation": "内存限制为必填字段",
				},
				map[string]interface{}{
					"field":       "TimeLimit",
					"reason":      "required",
					"translation": "运行时间限制为必填字段",
				},
				map[string]interface{}{
					"field":       "LanguageAllowed",
					"reason":      "required",
					"translation": "可用语言为必填字段",
				},
				map[string]interface{}{
					"field":       "CompareScriptName",
					"reason":      "required",
					"translation": "评测脚本为必填字段",
				},
			}),
		},
		{
			// testCreateProblemPermissionDenied
			name:   "PermissionDenied",
			method: "POST",
			path:   base.Echo.Reverse("problem.createProblem"),
			req: request.CreateProblemRequest{
				Name: "test_create_problem_perm",
			},
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, FailTests, "CreateProblem")

	boolTrue := true
	boolFalse := false

	successTests := []struct {
		name       string
		req        request.CreateProblemRequest
		attachment *fileContent
	}{
		{
			name: "SuccessWithoutAttachment",
			req: request.CreateProblemRequest{
				Name:              "test_create_problem_1",
				Description:       "test_create_problem_1_desc",
				MemoryLimit:       4294967296,
				TimeLimit:         1000,
				LanguageAllowed:   "test_create_problem_1_language_allowed",
				CompareScriptName: "cmp1",
				Public:            &boolFalse,
				Privacy:           &boolTrue,
			},
			attachment: nil,
		},
		{
			name: "SuccessWithAttachment",
			req: request.CreateProblemRequest{
				Name:              "test_create_problem_2",
				Description:       "test_create_problem_2_desc",
				MemoryLimit:       4294967296,
				TimeLimit:         1000,
				LanguageAllowed:   "test_create_problem_2_language_allowed",
				CompareScriptName: "cmp2",
				Public:            &boolTrue,
				Privacy:           &boolFalse,
			},
			attachment: newFileContent("attachment_file", "test_create_problem_attachment_file", attachmentFileBase64),
		},
	}
	t.Run("TestCreateProblemSuccess", func(t *testing.T) {
		t.Parallel()
		for i, test := range successTests {
			i := i
			test := test
			t.Run("TestCreateProblem"+test.name, func(t *testing.T) {
				t.Parallel()
				user := createUserForTest(t, "create_problem", i)
				user.GrantRole("admin")
				var data interface{}
				if test.attachment != nil {
					data = addFieldContentSlice([]reqContent{
						test.attachment,
					}, map[string]string{
						"name":                test.req.Name,
						"description":         test.req.Description,
						"memory_limit":        fmt.Sprint(test.req.MemoryLimit),
						"time_limit":          fmt.Sprint(test.req.TimeLimit),
						"language_allowed":    test.req.LanguageAllowed,
						"compare_script_name": fmt.Sprint(test.req.CompareScriptName),
						"public":              fmt.Sprint(*test.req.Public),
						"privacy":             fmt.Sprint(*test.req.Privacy),
					})
				} else {
					data = test.req
				}
				httpReq := makeReq(t, "POST", base.Echo.Reverse("problem.createProblem"), data, headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", user.ID)},
				})
				httpResp := makeResp(httpReq)
				assert.Equal(t, http.StatusCreated, httpResp.StatusCode)
				databaseProblem := models.Problem{}
				assert.NoError(t, base.DB.Where("name = ?", test.req.Name).First(&databaseProblem).Error)
				// request == database
				assert.Equal(t, test.req.Name, databaseProblem.Name)
				assert.Equal(t, test.req.Description, databaseProblem.Description)
				assert.Equal(t, test.req.MemoryLimit, databaseProblem.MemoryLimit)
				assert.Equal(t, test.req.TimeLimit, databaseProblem.TimeLimit)
				assert.Equal(t, strings.Split(test.req.LanguageAllowed, ","), []string(databaseProblem.LanguageAllowed))
				assert.Equal(t, test.req.CompareScriptName, databaseProblem.CompareScriptName)
				assert.Equal(t, *test.req.Public, databaseProblem.Public)
				assert.Equal(t, *test.req.Privacy, databaseProblem.Privacy)
				// response == database
				resp := response.CreateProblemResponse{}
				mustJsonDecode(httpResp, &resp)
				jsonEQ(t, response.UpdateProblemResponse{
					Message: "SUCCESS",
					Error:   nil,
					Data: struct {
						*resource.ProblemForAdmin `json:"problem"`
					}{
						resource.GetProblemForAdmin(&databaseProblem),
					},
				}, resp)
				assert.True(t, user.HasRole("problem_creator", databaseProblem))
				if test.attachment != nil {
					storageContent := getObjectContent(t, "problems", fmt.Sprintf("%d/attachment", databaseProblem.ID))
					expectedContent, err := ioutil.ReadAll(test.attachment.reader)
					assert.NoError(t, err)
					assert.Equal(t, expectedContent, storageContent)
					assert.Equal(t, test.attachment.fileName, databaseProblem.AttachmentFileName)
				} else {
					assert.Equal(t, "", databaseProblem.AttachmentFileName)
				}
			})
		}
	})
}

func TestUpdateProblem(t *testing.T) {
	t.Parallel()
	boolTrue := true
	boolFalse := false
	problem1 := models.Problem{
		Name:               "test_update_problem_1",
		AttachmentFileName: "test_update_problem_1_attachment_file_name",
		LanguageAllowed:    []string{"test_update_problem_1_language_allowed"},
	}
	assert.NoError(t, base.DB.Create(&problem1).Error)

	userWithProblem1Perm := models.User{
		Username: "test_update_problem_user_p1",
		Nickname: "test_update_problem_user_p1_nick",
		Email:    "test_update_problem_user_p1@e.e",
		Password: utils.HashPassword("test_update_problem_user_p1"),
	}
	assert.NoError(t, base.DB.Create(&userWithProblem1Perm).Error)
	userWithProblem1Perm.GrantRole("problem_creator", problem1)

	FailTests := []failTest{
		{
			name:   "WithoutParams",
			method: "PUT",
			path:   base.Echo.Reverse("problem.updateProblem", problem1.ID),
			req: request.UpdateProblemRequest{
				Name:              "",
				Description:       "",
				Public:            nil,
				Privacy:           nil,
				MemoryLimit:       0,
				TimeLimit:         0,
				LanguageAllowed:   "",
				BuildArg:          "",
				CompareScriptName: "",
			},
			reqOptions: []reqOption{
				headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", userWithProblem1Perm.ID)},
				},
			},
			statusCode: http.StatusBadRequest,
			resp: response.ErrorResp("VALIDATION_ERROR", []interface{}{
				map[string]interface{}{
					"field":       "Name",
					"reason":      "required",
					"translation": "名称为必填字段",
				},
				map[string]interface{}{
					"field":       "Description",
					"reason":      "required",
					"translation": "描述为必填字段",
				},
				map[string]interface{}{
					"field":       "Public",
					"reason":      "required",
					"translation": "是否公开为必填字段",
				},
				map[string]interface{}{
					"field":       "Privacy",
					"reason":      "required",
					"translation": "是否可以看到详细评测结果为必填字段",
				},
				map[string]interface{}{
					"field":       "MemoryLimit",
					"reason":      "required",
					"translation": "内存限制为必填字段",
				},
				map[string]interface{}{
					"field":       "TimeLimit",
					"reason":      "required",
					"translation": "运行时间限制为必填字段",
				},
				map[string]interface{}{
					"field":       "LanguageAllowed",
					"reason":      "required",
					"translation": "可用语言为必填字段",
				},
				map[string]interface{}{
					"field":       "CompareScriptName",
					"reason":      "required",
					"translation": "评测脚本为必填字段",
				},
			}),
		},
		{
			name:   "NonExistId",
			method: "PUT",
			path:   base.Echo.Reverse("problem.updateProblem", -1),
			req: request.UpdateProblemRequest{
				Name:              "test_update_problem_non_exist",
				Description:       "test_update_problem_non_exist_desc",
				Public:            &boolFalse,
				Privacy:           &boolFalse,
				MemoryLimit:       1024,
				TimeLimit:         1000,
				LanguageAllowed:   "test_update_problem_non_exist_language_allowed",
				BuildArg:          "test_update_problem_non_exist_build_arg",
				CompareScriptName: "cmp1",
			},
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "PermissionDenied",
			method: "PUT",
			path:   base.Echo.Reverse("problem.updateProblem", problem1.ID),
			req: request.UpdateProblemRequest{
				Name:            "test_update_problem_prem",
				LanguageAllowed: "test_update_problem_perm_language_allowed",
			},
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, FailTests, "UpdateProblem")

	successTests := []struct {
		name               string
		path               string
		originalProblem    models.Problem
		expectedProblem    models.Problem
		req                request.UpdateProblemRequest
		updatedAttachment  *fileContent
		originalAttachment *fileContent
		testCases          []models.TestCase
	}{
		{
			name: "WithoutAttachmentAndTestCase",
			path: "id",
			originalProblem: models.Problem{
				Name:              "test_update_problem_3",
				Description:       "test_update_problem_3_desc",
				LanguageAllowed:   []string{"test_update_problem_3_language_allowed"},
				Public:            false,
				Privacy:           true,
				MemoryLimit:       1024,
				TimeLimit:         1000,
				CompareScriptName: "cmp1",
			},
			expectedProblem: models.Problem{
				Name:              "test_update_problem_30",
				Description:       "test_update_problem_30_desc",
				LanguageAllowed:   []string{"test_update_problem_30_language_allowed"},
				Public:            true,
				Privacy:           false,
				MemoryLimit:       2048,
				TimeLimit:         2000,
				CompareScriptName: "cmp2",
			},
			req: request.UpdateProblemRequest{
				Name:              "test_update_problem_30",
				Description:       "test_update_problem_30_desc",
				LanguageAllowed:   "test_update_problem_30_language_allowed",
				Public:            &boolTrue,
				Privacy:           &boolFalse,
				MemoryLimit:       2048,
				TimeLimit:         2000,
				CompareScriptName: "cmp2",
			},
			updatedAttachment: nil,
			testCases:         nil,
		},
		{
			name: "WithAddingAttachment",
			path: "id",
			originalProblem: models.Problem{
				Name:               "test_update_problem_4",
				Description:        "test_update_problem_4_desc",
				LanguageAllowed:    []string{"test_update_problem_4_language_allowed"},
				Public:             true,
				Privacy:            true,
				MemoryLimit:        1024,
				TimeLimit:          1000,
				CompareScriptName:  "cmp1",
				AttachmentFileName: "",
			},
			expectedProblem: models.Problem{
				Name:               "test_update_problem_40",
				Description:        "test_update_problem_40_desc",
				LanguageAllowed:    []string{"test_update_problem_40_language_allowed"},
				Public:             false,
				Privacy:            false,
				MemoryLimit:        2048,
				TimeLimit:          2000,
				CompareScriptName:  "cmp2",
				AttachmentFileName: "test_update_problem_attachment_40",
			},
			req: request.UpdateProblemRequest{
				Name:              "test_update_problem_40",
				Description:       "test_update_problem_40_desc",
				LanguageAllowed:   "test_update_problem_40_language_allowed",
				Public:            &boolFalse,
				Privacy:           &boolFalse,
				MemoryLimit:       2048,
				TimeLimit:         2000,
				CompareScriptName: "cmp2",
			},
			updatedAttachment: newFileContent("attachment_file", "test_update_problem_attachment_40", newAttachmentFileBase64),
			testCases:         nil,
		},
		{
			name: "WithChangingAttachment",
			path: "id",
			originalProblem: models.Problem{
				Name:               "test_update_problem_5",
				Description:        "test_update_problem_5_desc",
				LanguageAllowed:    []string{"test_update_problem_5_language_allowed"},
				Public:             true,
				Privacy:            true,
				MemoryLimit:        1024,
				TimeLimit:          1000,
				CompareScriptName:  "cmp1",
				AttachmentFileName: "test_update_problem_attachment_5",
			},
			expectedProblem: models.Problem{
				Name:               "test_update_problem_50",
				Description:        "test_update_problem_50_desc",
				LanguageAllowed:    []string{"test_update_problem_50_language_allowed"},
				Public:             false,
				Privacy:            false,
				MemoryLimit:        2048,
				TimeLimit:          2000,
				CompareScriptName:  "cmp2",
				AttachmentFileName: "test_update_problem_attachment_50",
			},
			req: request.UpdateProblemRequest{
				Name:              "test_update_problem_50",
				Description:       "test_update_problem_50_desc",
				LanguageAllowed:   "test_update_problem_50_language_allowed",
				Public:            &boolFalse,
				Privacy:           &boolFalse,
				MemoryLimit:       2048,
				TimeLimit:         2000,
				CompareScriptName: "cmp2",
			},
			originalAttachment: newFileContent("attachment_file", "test_update_problem_attachment_5", attachmentFileBase64),
			updatedAttachment:  newFileContent("attachment_file", "test_update_problem_attachment_50", newAttachmentFileBase64),
			testCases:          nil,
		},
		{
			name: "WithoutChangingAttachment",
			path: "id",
			originalProblem: models.Problem{
				Name:               "test_update_problem_6",
				Description:        "test_update_problem_6_desc",
				LanguageAllowed:    []string{"test_update_problem_6_language_allowed"},
				Public:             true,
				Privacy:            true,
				MemoryLimit:        1024,
				TimeLimit:          1000,
				CompareScriptName:  "cmp1",
				AttachmentFileName: "test_update_problem_attachment_6",
			},
			expectedProblem: models.Problem{
				Name:               "test_update_problem_60",
				Description:        "test_update_problem_60_desc",
				LanguageAllowed:    []string{"test_update_problem_60_language_allowed"},
				Public:             false,
				Privacy:            false,
				MemoryLimit:        2048,
				TimeLimit:          2000,
				CompareScriptName:  "cmp2",
				AttachmentFileName: "test_update_problem_attachment_6",
			},
			req: request.UpdateProblemRequest{
				Name:              "test_update_problem_60",
				Description:       "test_update_problem_60_desc",
				LanguageAllowed:   "test_update_problem_60_language_allowed",
				Public:            &boolFalse,
				Privacy:           &boolFalse,
				MemoryLimit:       2048,
				TimeLimit:         2000,
				CompareScriptName: "cmp2",
			},
			originalAttachment: newFileContent("attachment_file", "test_update_problem_attachment_6", attachmentFileBase64),
			updatedAttachment:  nil,
			testCases:          nil,
		},
		{
			name: "WithTestCase",
			path: "id",
			originalProblem: models.Problem{
				Name:              "test_update_problem_7",
				Description:       "test_update_problem_7_desc",
				LanguageAllowed:   []string{"test_update_problem_7_language_allowed"},
				Public:            false,
				Privacy:           true,
				MemoryLimit:       1024,
				TimeLimit:         1000,
				CompareScriptName: "cmp1",
			},
			expectedProblem: models.Problem{
				Name:              "test_update_problem_70",
				Description:       "test_update_problem_70_desc",
				LanguageAllowed:   []string{"test_update_problem_70_language_allowed"},
				Public:            true,
				Privacy:           false,
				MemoryLimit:       2048,
				TimeLimit:         2000,
				CompareScriptName: "cmp2",
			},
			req: request.UpdateProblemRequest{
				Name:              "test_update_problem_70",
				Description:       "test_update_problem_70_desc",
				LanguageAllowed:   "test_update_problem_70_language_allowed",
				Public:            &boolTrue,
				Privacy:           &boolFalse,
				MemoryLimit:       2048,
				TimeLimit:         2000,
				CompareScriptName: "cmp2",
			},
			updatedAttachment: nil,
			testCases: []models.TestCase{
				{
					Score:          100,
					Sample:         true,
					InputFileName:  "test_update_problem_7_test_case_1_input_file_name",
					OutputFileName: "test_update_problem_7_test_case_1_output_file_name",
				},
				{
					Score:          100,
					Sample:         false,
					InputFileName:  "test_update_problem_7_test_case_2_input_file_name",
					OutputFileName: "test_update_problem_7_test_case_2_output_file_name",
				},
			},
		},
	}

	t.Run("TestUpdateProblemSuccess", func(t *testing.T) {
		t.Parallel()
		for i, test := range successTests {
			i := i
			test := test
			t.Run("TestUpdateProblem"+test.name, func(t *testing.T) {
				t.Parallel()
				assert.NoError(t, base.DB.Create(&test.originalProblem).Error)
				for j := range test.testCases {
					assert.NoError(t, base.DB.Model(&test.originalProblem).Association("TestCases").Append(&test.testCases[j]))
				}
				if test.originalAttachment != nil {
					_, err := base.Storage.PutObject(context.Background(), "problems", fmt.Sprintf("%d/attachment", test.originalProblem.ID), test.originalAttachment.reader, test.originalAttachment.size, minio.PutObjectOptions{})
					assert.NoError(t, err)
					_, err = test.originalAttachment.reader.Seek(0, io.SeekStart)
					assert.NoError(t, err)
				}
				path := base.Echo.Reverse("problem.updateProblem", test.originalProblem.ID)
				user := createUserForTest(t, "update_problem", i)
				user.GrantRole("problem_creator", test.originalProblem)
				var data interface{}
				if test.updatedAttachment != nil {
					data = addFieldContentSlice([]reqContent{
						test.updatedAttachment,
					}, map[string]string{
						"name":                test.req.Name,
						"description":         test.req.Description,
						"memory_limit":        fmt.Sprint(test.req.MemoryLimit),
						"time_limit":          fmt.Sprint(test.req.TimeLimit),
						"language_allowed":    test.req.LanguageAllowed,
						"compare_script_name": fmt.Sprint(test.req.CompareScriptName),
						"public":              fmt.Sprint(*test.req.Public),
						"privacy":             fmt.Sprint(*test.req.Privacy),
					})
				} else {
					data = test.req
				}
				httpResp := makeResp(makeReq(t, "PUT", path, data, headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", user.ID)},
				}))
				databaseProblem := models.Problem{}
				assert.NoError(t, base.DB.First(&databaseProblem, test.originalProblem.ID).Error)
				// ignore other fields
				test.expectedProblem.ID = databaseProblem.ID
				test.expectedProblem.CreatedAt = databaseProblem.CreatedAt
				test.expectedProblem.UpdatedAt = databaseProblem.UpdatedAt
				test.expectedProblem.DeletedAt = databaseProblem.DeletedAt
				assert.Equal(t, test.expectedProblem, databaseProblem)
				assert.NoError(t, base.DB.Set("gorm:auto_preload", true).Model(databaseProblem).Association("TestCases").Find(&databaseProblem.TestCases))
				if test.testCases != nil {
					jsonEQ(t, test.testCases, databaseProblem.TestCases)
				} else {
					assert.Equal(t, []models.TestCase{}, databaseProblem.TestCases)
				}
				assert.Equal(t, http.StatusOK, httpResp.StatusCode)
				jsonEQ(t, response.UpdateProblemResponse{
					Message: "SUCCESS",
					Error:   nil,
					Data: struct {
						*resource.ProblemForAdmin `json:"problem"`
					}{
						resource.GetProblemForAdmin(&databaseProblem),
					},
				}, httpResp)
				if test.updatedAttachment != nil || test.originalAttachment != nil {
					storageContent := getObjectContent(t, "problems", fmt.Sprintf("%d/attachment", databaseProblem.ID))
					var err error
					var expectedContent []byte
					if test.updatedAttachment != nil {
						expectedContent, err = ioutil.ReadAll(test.updatedAttachment.reader)
					} else {
						expectedContent, err = ioutil.ReadAll(test.originalAttachment.reader)
					}
					assert.NoError(t, err)
					assert.Equal(t, expectedContent, storageContent)
				}
			})
		}
	})
}

func TestDeleteProblem(t *testing.T) {
	t.Parallel()

	failTests := []failTest{
		{
			name:   "NonExistId",
			method: "DELETE",
			path:   base.Echo.Reverse("problem.deleteProblem", -1),
			req:    request.DeleteProblemRequest{},
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "PermissionDenied",
			method: "DELETE",
			path:   base.Echo.Reverse("problem.deleteProblem", -1),
			req:    request.DeleteProblemRequest{},
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "DeleteProblem")

	successTests := []struct {
		name               string
		problem            models.Problem
		originalAttachment *fileContent
		testCases          []struct {
			testcase   models.TestCase
			inputFile  *fileContent
			outputFile *fileContent
		}
	}{
		{
			name: "SuccessWithoutAttachmentAndTestCases",
			problem: models.Problem{
				Name:               "test_delete_problem_1",
				AttachmentFileName: "",
				LanguageAllowed:    []string{"test_delete_problem_1_language_allowed"},
			},
			originalAttachment: nil,
			testCases:          nil,
		},
		{
			name: "SuccessWithAttachment",
			problem: models.Problem{
				Name:               "test_delete_problem_2",
				AttachmentFileName: "test_delete_problem_attachment_2",
				LanguageAllowed:    []string{"test_delete_problem_2_language_allowed"},
			},
			originalAttachment: newFileContent("attachment_file", "test_delete_problem_attachment_2", attachmentFileBase64),
			testCases:          nil,
		},
		{
			name: "SuccessWithTestCases",
			problem: models.Problem{
				Name:               "test_delete_problem_3",
				AttachmentFileName: "",
				LanguageAllowed:    []string{"test_delete_problem_3_language_allowed"},
			},
			originalAttachment: nil,
			testCases: []struct {
				testcase   models.TestCase
				inputFile  *fileContent
				outputFile *fileContent
			}{
				{
					testcase: models.TestCase{
						Score:          100,
						Sample:         true,
						InputFileName:  "test_delete_problem_3_test_case_1_input_file_name",
						OutputFileName: "test_delete_problem_3_test_case_1_output_file_name",
					},
					inputFile:  newFileContent("input_file", "test_delete_problem_3_test_case_1.in", inputTextBase64),
					outputFile: newFileContent("output_file", "test_delete_problem_3_test_case_1.out", outputTextBase64),
				},
				{
					testcase: models.TestCase{
						Score:          100,
						Sample:         false,
						InputFileName:  "test_delete_problem_3_test_case_2_input_file_name",
						OutputFileName: "test_delete_problem_3_test_case_2_output_file_name",
					},
					inputFile:  newFileContent("input_file", "test_delete_problem_3_test_case_2.in", inputTextBase64),
					outputFile: newFileContent("output_file", "test_delete_problem_3_test_case_2.out", outputTextBase64),
				},
			},
		},
		{
			name: "SuccessWithAttachmentAndTestCases",
			problem: models.Problem{
				Name:               "test_delete_problem_4",
				AttachmentFileName: "test_delete_problem_attachment_4",
				LanguageAllowed:    []string{"test_delete_problem_4_language_allowed"},
			},
			originalAttachment: newFileContent("attachment_file", "test_delete_problem_attachment_4", attachmentFileBase64),
			testCases: []struct {
				testcase   models.TestCase
				inputFile  *fileContent
				outputFile *fileContent
			}{
				{
					testcase: models.TestCase{
						Score:          100,
						Sample:         true,
						InputFileName:  "test_delete_problem_4_test_case_1_input_file_name",
						OutputFileName: "test_delete_problem_4_test_case_1_output_file_name",
					},
					inputFile:  newFileContent("input_file", "test_delete_problem_4_test_case_1.in", inputTextBase64),
					outputFile: newFileContent("output_file", "test_delete_problem_4_test_case_1.out", outputTextBase64),
				},
				{
					testcase: models.TestCase{
						Score:          100,
						Sample:         false,
						InputFileName:  "test_delete_problem_4_test_case_2_input_file_name",
						OutputFileName: "test_delete_problem_4_test_case_2_output_file_name",
					},
					inputFile:  newFileContent("input_file", "test_delete_problem_4_test_case_2.in", inputTextBase64),
					outputFile: newFileContent("output_file", "test_delete_problem_4_test_case_2.out", outputTextBase64),
				},
			},
		},
	}

	t.Run("TestDeleteProblemSuccess", func(t *testing.T) {
		t.Parallel()
		for i, test := range successTests {
			i := i
			test := test
			t.Run("TestDeleteProblem"+test.name, func(t *testing.T) {
				t.Parallel()
				assert.NoError(t, base.DB.Create(&test.problem).Error)
				for j := range test.testCases {
					createTestCaseForTest(t, test.problem, testCaseData{
						Score:      test.testCases[j].testcase.Score,
						Sample:     test.testCases[j].testcase.Sample,
						InputFile:  test.testCases[j].inputFile,
						OutputFile: test.testCases[j].outputFile,
					})
				}
				if test.originalAttachment != nil {
					_, err := base.Storage.PutObject(context.Background(), "problems", fmt.Sprintf("%d/attachment", test.problem.ID), test.originalAttachment.reader, test.originalAttachment.size, minio.PutObjectOptions{})
					assert.NoError(t, err)
					_, err = test.originalAttachment.reader.Seek(0, io.SeekStart)
					assert.NoError(t, err)
				}
				user := createUserForTest(t, "delete_problem", i)
				user.GrantRole("problem_creator", test.problem)
				httpResp := makeResp(makeReq(t, "DELETE", base.Echo.Reverse("problem.deleteProblem", test.problem.ID), request.DeleteProblemRequest{}, headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", user.ID)},
				}))
				resp := response.Response{}
				mustJsonDecode(httpResp, &resp)
				assert.Equal(t, http.StatusOK, httpResp.StatusCode)
				assert.Equal(t, response.Response{
					Message: "SUCCESS",
					Error:   nil,
					Data:    nil,
				}, resp)
				assert.Equal(t, gorm.ErrRecordNotFound, base.DB.First(models.Problem{}, test.problem.ID).Error)
				assert.False(t, user.HasRole("problem_creator", test.problem))
				for j := range test.testCases {
					assert.Equal(t, gorm.ErrRecordNotFound, base.DB.First(models.TestCase{}, test.testCases[j].testcase.ID).Error)
				}
			})
		}
	})
}

func TestCreateTestCase(t *testing.T) {
	t.Parallel()
	user := createUserForTest(t, "create_test_case", 0)
	problem := createProblemForTest(t, "create_test_case", 0, nil, user)

	failTests := []failTest{
		{
			name:   "NonExistingProblem",
			method: "POST",
			path:   base.Echo.Reverse("problem.createTestCase", -1),
			req: addFieldContentSlice([]reqContent{
				newFileContent("input_file", "test_create_test_case_non_existing_problem.in", inputTextBase64),
				newFileContent("output_file", "test_create_test_case_non_existing_problem.out", outputTextBase64),
			}, map[string]string{
				"score":  "100",
				"sample": "true",
			}),
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "LackInputFile",
			method: "POST",
			path:   base.Echo.Reverse("problem.createTestCase", problem.ID),
			req: addFieldContentSlice([]reqContent{
				newFileContent("output_file", "test_create_test_case_lack_input_file.out", outputTextBase64),
			}, map[string]string{
				"score":  "100",
				"sample": "true",
			}),
			reqOptions: []reqOption{
				headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", user.ID)},
				},
			},
			statusCode: http.StatusBadRequest,
			resp:       response.ErrorResp("INVALID_FILE", nil),
		},
		{
			name:   "LackOutputFile",
			method: "POST",
			path:   base.Echo.Reverse("problem.createTestCase", problem.ID),
			req: addFieldContentSlice([]reqContent{
				newFileContent("input_file", "test_create_test_case_lack_output_file.in", inputTextBase64),
			}, map[string]string{
				"score":  "100",
				"sample": "true",
			}),
			reqOptions: []reqOption{
				headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", user.ID)},
				},
			},
			statusCode: http.StatusBadRequest,
			resp:       response.ErrorResp("INVALID_FILE", nil),
		},
		{
			name:   "LackBothFile",
			method: "POST",
			path:   base.Echo.Reverse("problem.createTestCase", problem.ID),
			req: addFieldContentSlice([]reqContent{}, map[string]string{
				"score":  "100",
				"sample": "true",
			}),
			reqOptions: []reqOption{
				headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", user.ID)},
				},
			},
			statusCode: http.StatusBadRequest,
			resp:       response.ErrorResp("INVALID_FILE", nil),
		},
		{
			name:   "PermissionDenied",
			method: "POST",
			path:   base.Echo.Reverse("problem.createTestCase", problem.ID),
			req: addFieldContentSlice([]reqContent{
				newFileContent("input_file", "test_create_test_case_permission_denied.in", inputTextBase64),
				newFileContent("output_file", "test_create_test_case_permission_denied.out", outputTextBase64),
			}, map[string]string{
				"score":  "100",
				"sample": "true",
			}),
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "CreateTestCase")

	t.Run("TestCreateTestCaseSuccess", func(t *testing.T) {
		t.Parallel()
		req := makeReq(t, "POST", base.Echo.Reverse("problem.createTestCase", problem.ID), addFieldContentSlice([]reqContent{
			newFileContent("input_file", "test_create_test_case_success.in", inputTextBase64),
			newFileContent("output_file", "test_create_test_case_success.out", outputTextBase64),
		}, map[string]string{
			"score":  "100",
			"sample": "true",
		}), headerOption{
			"Set-User-For-Test": {fmt.Sprintf("%d", user.ID)},
		})
		//req.Header.Set("Set-User-For-Test", fmt.Sprintf("%d", user.ID))
		httpResp := makeResp(req)
		expectedTestCase := models.TestCase{
			ProblemID:      problem.ID,
			Score:          100,
			Sample:         true,
			InputFileName:  "test_create_test_case_success.in",
			OutputFileName: "test_create_test_case_success.out",
		}
		databaseTestCase := models.TestCase{}
		assert.NoError(t, base.DB.Where("problem_id = ? and input_file_name = ?", problem.ID, "test_create_test_case_success.in").First(&databaseTestCase).Error)
		assert.Equal(t, expectedTestCase.ProblemID, databaseTestCase.ProblemID)
		assert.Equal(t, expectedTestCase.InputFileName, databaseTestCase.InputFileName)
		assert.Equal(t, expectedTestCase.OutputFileName, databaseTestCase.OutputFileName)
		assert.Equal(t, expectedTestCase.Score, databaseTestCase.Score)
		assert.Equal(t, expectedTestCase.Sample, databaseTestCase.Sample)
		resp := response.CreateTestCaseResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, http.StatusCreated, httpResp.StatusCode)
		assert.Equal(t, response.CreateTestCaseResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.TestCaseForAdmin `json:"test_case"`
			}{
				resource.GetTestCaseForAdmin(&databaseTestCase),
			},
		}, resp)
		assert.Equal(t, []byte("input text\n"), getObjectContent(t, "problems", fmt.Sprintf("%d/input/%d.in", problem.ID, databaseTestCase.ID)))
		assert.Equal(t, []byte("output text\n"), getObjectContent(t, "problems", fmt.Sprintf("%d/output/%d.out", problem.ID, databaseTestCase.ID)))
	})
}

func TestGetTestCaseInputFile(t *testing.T) {
	t.Parallel()
	user := createUserForTest(t, "get_test_case_input_file", 0)
	problem := createProblemForTest(t, "get_test_case_input_file", 0, nil, user)
	testCase := createTestCaseForTest(t, problem, testCaseData{
		Score:      0,
		Sample:     false,
		InputFile:  nil,
		OutputFile: nil,
	})
	failTests := []failTest{
		{
			name:   "NonExistingProblem",
			method: "GET",
			path:   base.Echo.Reverse("problem.getTestCaseInputFile", -1, testCase.ID),
			req:    request.GetTestCaseInputFileRequest{},
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "NonExistingTestCase",
			method: "GET",
			path:   base.Echo.Reverse("problem.getTestCaseInputFile", problem.ID, -1),
			req:    request.GetTestCaseInputFileRequest{},
			reqOptions: []reqOption{
				headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", user.ID)},
				},
			},
			statusCode: http.StatusNotFound,
			resp: response.ErrorResp("TEST_CASE_NOT_FOUND", map[string]interface{}{
				"Err":  map[string]interface{}{},
				"Func": "ParseUint",
				"Num":  "-1",
			}),
		},
		{
			name:   "PermissionDenied",
			method: "GET",
			path:   base.Echo.Reverse("problem.getTestCaseInputFile", problem.ID, testCase.ID),
			req:    request.GetTestCaseInputFileRequest{},
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "GetTestCaseInputFile")

	successTests := []struct {
		name    string
		data    testCaseData
		reqUser reqOption
	}{
		{
			// testGetTestCaseInputSample
			name: "Sample",
			data: testCaseData{
				Score:      0,
				Sample:     true,
				InputFile:  newFileContent("", "test_get_test_case_input_file_success.in", inputTextBase64),
				OutputFile: nil,
			},
			reqUser: applyNormalUser,
		},
		{
			// testGetTestCaseInputAdmin
			name: "Admin",
			data: testCaseData{
				Score:      0,
				Sample:     false,
				InputFile:  newFileContent("", "test_get_test_case_input_file_success.in", inputTextBase64),
				OutputFile: nil,
			},
			reqUser: applyAdminUser,
		},
	}

	t.Run("testGetTestCaseInputSuccess", func(t *testing.T) {
		t.Parallel()
		for _, test := range successTests {
			test := test
			t.Run("testGetTestCase"+test.name, func(t *testing.T) {
				t.Parallel()
				testCase := createTestCaseForTest(t, problem, test.data)
				req := makeReq(t, "GET", base.Echo.Reverse("problem.getTestCaseInputFile", problem.ID, testCase.ID), request.GetTestCaseInputFileRequest{},
					test.reqUser,
				)
				httpResp := makeResp(req)
				assert.Equal(t, "input text\n", getPresignedURLContent(t, httpResp.Header.Get("Location")))
			})
		}
	})
}

func TestGetTestCaseOutputFile(t *testing.T) {
	t.Parallel()
	user := createUserForTest(t, "get_test_case_output_file", 0)
	problem := createProblemForTest(t, "get_test_case_output_file", 0, nil, user)
	testCase := createTestCaseForTest(t, problem, testCaseData{
		Score:      0,
		Sample:     false,
		InputFile:  nil,
		OutputFile: nil,
	})
	failTests := []failTest{
		{
			name:   "NonExistingProblem",
			method: "GET",
			path:   base.Echo.Reverse("problem.getTestCaseOutputFile", -1, testCase.ID),
			req:    request.GetTestCaseOutputFileRequest{},
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "NonExistingTestCase",
			method: "GET",
			path:   base.Echo.Reverse("problem.getTestCaseOutputFile", problem.ID, -1),
			req:    request.GetTestCaseOutputFileRequest{},
			reqOptions: []reqOption{
				headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", user.ID)},
				},
			},
			statusCode: http.StatusNotFound,
			resp: response.ErrorResp("TEST_CASE_NOT_FOUND", map[string]interface{}{
				"Err":  map[string]interface{}{},
				"Func": "ParseUint",
				"Num":  "-1",
			}),
		},
		{
			name:   "PermissionDenied",
			method: "GET",
			path:   base.Echo.Reverse("problem.getTestCaseOutputFile", problem.ID, testCase.ID),
			req:    request.GetTestCaseOutputFileRequest{},
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "GetTestCaseOutputFile")

	successTests := []struct {
		name    string
		data    testCaseData
		reqUser reqOption
	}{
		{
			// testGetTestCaseOutputSample
			name: "Sample",
			data: testCaseData{
				Score:      0,
				Sample:     true,
				InputFile:  nil,
				OutputFile: newFileContent("", "test_get_test_case_output_file_success.out", outputTextBase64),
			},
			reqUser: applyNormalUser,
		},
		{
			// testGetTestCaseOutputAdmin
			name: "Admin",
			data: testCaseData{
				Score:      0,
				Sample:     false,
				InputFile:  nil,
				OutputFile: newFileContent("", "test_get_test_case_output_file_success.out", outputTextBase64),
			},
			reqUser: applyAdminUser,
		},
	}

	t.Run("testGetTestCaseOutputSuccess", func(t *testing.T) {
		t.Parallel()
		for _, test := range successTests {
			test := test
			t.Run("testGetTestCase"+test.name, func(t *testing.T) {
				t.Parallel()
				testCase := createTestCaseForTest(t, problem, test.data)
				req := makeReq(t, "GET", base.Echo.Reverse("problem.getTestCaseOutputFile", problem.ID, testCase.ID), request.GetTestCaseOutputFileRequest{},
					test.reqUser,
				)
				httpResp := makeResp(req)
				assert.Equal(t, "output text\n", getPresignedURLContent(t, httpResp.Header.Get("Location")))
			})
		}
	})
}

func TestUpdateTestCase(t *testing.T) {
	t.Parallel()
	user := createUserForTest(t, "update_test_case", 0)
	problem := createProblemForTest(t, "update_test_case", 0, nil, user)
	boolTrue := true

	failTests := []failTest{
		{
			name:   "NonExistingProblem",
			method: "PUT",
			path:   base.Echo.Reverse("problem.updateTestCase", -1, 1),
			req: request.UpdateTestCaseRequest{
				Score:  100,
				Sample: &boolTrue,
			},
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "NonExistingTestCase",
			method: "PUT",
			path:   base.Echo.Reverse("problem.updateTestCase", problem.ID, -1),
			req: request.UpdateTestCaseRequest{
				Score:  100,
				Sample: &boolTrue,
			},
			reqOptions: []reqOption{
				headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", user.ID)},
				},
			},
			statusCode: http.StatusNotFound,
			resp: response.ErrorResp("TEST_CASE_NOT_FOUND", map[string]interface{}{
				"Err":  map[string]interface{}{},
				"Func": "ParseUint",
				"Num":  "-1",
			}),
		},
		{
			name:   "PermissionDenied",
			method: "PUT",
			path:   base.Echo.Reverse("problem.updateTestCase", problem.ID, 1),
			req: request.UpdateTestCaseRequest{
				Score:  100,
				Sample: &boolTrue,
			},
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "UpdateTestCase")

	successTests := []struct {
		name             string
		originalData     testCaseData
		updatedData      testCaseData
		expectedTestCase models.TestCase
	}{
		{
			name: "SuccessWithoutUpdatingFile",
			originalData: testCaseData{
				Score:      0,
				Sample:     false,
				InputFile:  newFileContent("input_file", "test_update_test_case_1.in", inputTextBase64),
				OutputFile: newFileContent("output_file", "test_update_test_case_1.out", outputTextBase64),
			},
			updatedData: testCaseData{
				Score:      100,
				Sample:     true,
				InputFile:  nil,
				OutputFile: nil,
			},
			expectedTestCase: models.TestCase{
				ProblemID:      problem.ID,
				Score:          100,
				Sample:         true,
				InputFileName:  "test_update_test_case_1.in",
				OutputFileName: "test_update_test_case_1.out",
			},
		},
		{
			name: "SuccessWithUpdatingInputFile",
			originalData: testCaseData{
				Score:      0,
				Sample:     true,
				InputFile:  newFileContent("input_file", "test_update_test_case_2.in", inputTextBase64),
				OutputFile: newFileContent("output_file", "test_update_test_case_2.out", outputTextBase64),
			},
			updatedData: testCaseData{
				Score:      100,
				Sample:     false,
				InputFile:  newFileContent("input_file", "test_update_test_case_20.in", newInputTextBase64),
				OutputFile: nil,
			},
			expectedTestCase: models.TestCase{
				ProblemID:      problem.ID,
				Score:          100,
				Sample:         false,
				InputFileName:  "test_update_test_case_20.in",
				OutputFileName: "test_update_test_case_2.out",
			},
		},
		{
			name: "SuccessWithUpdatingOutputFile",
			originalData: testCaseData{
				Score:      0,
				Sample:     true,
				InputFile:  newFileContent("input_file", "test_update_test_case_3.in", inputTextBase64),
				OutputFile: newFileContent("output_file", "test_update_test_case_3.out", outputTextBase64),
			},
			updatedData: testCaseData{
				Score:      100,
				Sample:     true,
				InputFile:  nil,
				OutputFile: newFileContent("output_file", "test_update_test_case_30.out", "bmV3IG91dHB1dCB0ZXh0"),
			},
			expectedTestCase: models.TestCase{
				ProblemID:      problem.ID,
				Score:          100,
				Sample:         true,
				InputFileName:  "test_update_test_case_3.in",
				OutputFileName: "test_update_test_case_30.out",
			},
		},
		{
			name: "SuccessWithUpdatingBothFile",
			originalData: testCaseData{
				Score:      0,
				Sample:     false,
				InputFile:  newFileContent("input_file", "test_update_test_case_4.in", inputTextBase64),
				OutputFile: newFileContent("output_file", "test_update_test_case_4.out", outputTextBase64),
			},
			updatedData: testCaseData{
				Score:      100,
				Sample:     false,
				InputFile:  newFileContent("input_file", "test_update_test_case_40.in", newInputTextBase64),
				OutputFile: newFileContent("output_file", "test_update_test_case_40.out", newOutputTextBase64),
			},
			expectedTestCase: models.TestCase{
				ProblemID:      problem.ID,
				Score:          100,
				Sample:         false,
				InputFileName:  "test_update_test_case_40.in",
				OutputFileName: "test_update_test_case_40.out",
			},
		},
	}

	t.Run("TestUpdateTestCaseSuccess", func(t *testing.T) {
		t.Parallel()
		for _, test := range successTests {
			test := test
			t.Run("TestUpdateTestCase"+test.name, func(t *testing.T) {
				t.Parallel()
				testCase := createTestCaseForTest(t, problem, test.originalData)
				var reqContentSlice []reqContent
				if test.updatedData.InputFile != nil {
					reqContentSlice = append(reqContentSlice, test.updatedData.InputFile)
				}
				if test.updatedData.OutputFile != nil {
					reqContentSlice = append(reqContentSlice, test.updatedData.OutputFile)
				}
				req := makeReq(t, "PUT", base.Echo.Reverse("problem.updateTestCase", problem.ID, testCase.ID), addFieldContentSlice(
					reqContentSlice, map[string]string{
						"score":  fmt.Sprintf("%d", test.updatedData.Score),
						"sample": fmt.Sprintf("%t", test.updatedData.Sample),
					}), headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", user.ID)},
				})
				httpResp := makeResp(req)
				databaseTestCase := models.TestCase{}
				base.DB.First(&databaseTestCase, testCase.ID)
				assert.Equal(t, test.expectedTestCase.ProblemID, databaseTestCase.ProblemID)
				assert.Equal(t, test.expectedTestCase.Score, databaseTestCase.Score)
				assert.Equal(t, test.expectedTestCase.Sample, databaseTestCase.Sample)
				assert.Equal(t, test.expectedTestCase.InputFileName, databaseTestCase.InputFileName)
				assert.Equal(t, test.expectedTestCase.OutputFileName, databaseTestCase.OutputFileName)

				var expectedInputFileReader io.Reader
				if test.updatedData.InputFile != nil {
					expectedInputFileReader = test.updatedData.InputFile.reader
				} else {
					expectedInputFileReader = test.originalData.InputFile.reader
				}
				expectedInputContent, err := ioutil.ReadAll(expectedInputFileReader)
				assert.NoError(t, err)
				assert.Equal(t, (expectedInputContent), (getObjectContent(t, "problems", fmt.Sprintf("%d/input/%d.in", problem.ID, databaseTestCase.ID))))

				var expectedOutputFileReader io.Reader
				if test.updatedData.OutputFile != nil {
					expectedOutputFileReader = test.updatedData.OutputFile.reader
				} else {
					expectedOutputFileReader = test.originalData.OutputFile.reader
				}
				expectedOutputContent, err := ioutil.ReadAll(expectedOutputFileReader)
				assert.NoError(t, err)
				assert.Equal(t, (expectedOutputContent), (getObjectContent(t, "problems", fmt.Sprintf("%d/output/%d.out", problem.ID, databaseTestCase.ID))))

				resp := response.UpdateTestCaseResponse{}
				mustJsonDecode(httpResp, &resp)
				assert.Equal(t, response.UpdateTestCaseResponse{
					Message: "SUCCESS",
					Error:   nil,
					Data: struct {
						*resource.TestCaseForAdmin `json:"test_case"`
					}{
						resource.GetTestCaseForAdmin(&databaseTestCase),
					},
				}, resp)
			})
		}
	})
}

func TestDeleteTestCase(t *testing.T) {
	t.Parallel()
	user := createUserForTest(t, "delete_test_case", 0)
	problem := createProblemForTest(t, "delete_test_case", 0, nil, user)

	failTests := []failTest{
		{
			name:   "NonExistingProblem",
			method: "DELETE",
			path:   base.Echo.Reverse("problem.deleteTestCase", -1, 1),
			req:    request.DeleteProblemRequest{},
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "NonExistingTestCase",
			method: "DELETE",
			path:   base.Echo.Reverse("problem.deleteTestCase", problem.ID, -1),
			req:    request.DeleteProblemRequest{},
			reqOptions: []reqOption{
				headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", user.ID)},
				},
			},
			statusCode: http.StatusNotFound,
			resp: response.ErrorResp("TEST_CASE_NOT_FOUND", map[string]interface{}{
				"Err":  map[string]interface{}{},
				"Func": "ParseUint",
				"Num":  "-1",
			}),
		},
		{
			name:   "PermissionDenied",
			method: "DELETE",
			path:   base.Echo.Reverse("problem.deleteTestCase", problem.ID, 1),
			req:    request.DeleteProblemRequest{},
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "DeleteTestCase")

	t.Run("TestDeleteTestCaseSuccess", func(t *testing.T) {
		t.Parallel()
		testCase := createTestCaseForTest(t, problem, testCaseData{
			Score:      0,
			Sample:     true,
			InputFile:  newFileContent("input_file", "test_delete_test_case_0.in", inputTextBase64),
			OutputFile: newFileContent("output_file", "test_delete_test_case_0.out", outputTextBase64),
		})

		req := makeReq(t, "DELETE", base.Echo.Reverse("problem.deleteTestCase", problem.ID, testCase.ID), request.DeleteTestCaseRequest{}, headerOption{
			"Set-User-For-Test": {fmt.Sprintf("%d", user.ID)},
		})
		httpResp := makeResp(req)
		resp := response.Response{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		assert.Equal(t, response.Response{
			Message: "SUCCESS",
			Error:   nil,
			Data:    nil,
		}, resp)
		databaseTestcase := models.TestCase{}
		err := base.DB.First(&databaseTestcase, testCase.ID).Error
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})
}

func TestDeleteTestCases(t *testing.T) {
	t.Parallel()
	user := createUserForTest(t, "delete_test_cases", 0)
	problem := createProblemForTest(t, "delete_test_cases", 0, nil, user)

	failTests := []failTest{
		{
			name:   "NonExistingProblem",
			method: "DELETE",
			path:   base.Echo.Reverse("problem.deleteTestCases", -1),
			req:    request.DeleteTestCasesRequest{},
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "PermissionDenied",
			method: "DELETE",
			path:   base.Echo.Reverse("problem.deleteTestCases", problem.ID),
			req:    request.DeleteTestCasesRequest{},
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "DeleteTestCases")

	t.Run("TestDeleteTestCasesSuccess", func(t *testing.T) {
		t.Parallel()
		for i := 0; i < 5; i++ {
			createTestCaseForTest(t, problem, testCaseData{
				Score:      0,
				Sample:     false,
				InputFile:  newFileContent("input_file", fmt.Sprintf("test_delete_test_cases_%d.in", i), inputTextBase64),
				OutputFile: newFileContent("output_file", fmt.Sprintf("test_delete_test_cases_%d.out", i), outputTextBase64),
			})
		}
		req := makeReq(t, "DELETE", base.Echo.Reverse("problem.deleteTestCases", problem.ID), request.DeleteTestCasesRequest{}, headerOption{
			"Set-User-For-Test": {fmt.Sprintf("%d", user.ID)},
		})
		httpResp := makeResp(req)
		resp := response.Response{}
		mustJsonDecode(httpResp, &resp)
		var databaseTestCases []models.TestCase
		assert.NoError(t, base.DB.Find(&databaseTestCases, "problem_id = ?", problem.ID).Error)
		assert.Equal(t, 0, len(databaseTestCases))
	})
}

func TestGetRandomProblem(t *testing.T) {
	t.Cleanup(database.SetupDatabaseForTest())

	t.Run("Empty", func(t *testing.T) {
		httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("problem.getRandomProblem"),
			request.GetRandomProblem{}, applyNormalUser))
		assert.Equal(t, http.StatusNotFound, httpResp.StatusCode)
		resp := response.Response{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.ErrorResp("NOT_FOUND", nil), resp)
	})

	user := createUserForTest(t, "get_random_problem", 0)
	problems := make(map[uint]*models.Problem)
	for i := 0; i < 3; i++ {
		p := createProblemForTest(t, "get_random_problem", 0, nil, user)
		problems[p.ID] = &p
	}

	t.Run("Success", func(t *testing.T) {
		httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("problem.getRandomProblem"),
			request.GetRandomProblem{}, applyNormalUser))
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		resp := response.GetRandomProblemResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.GetRandomProblemResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.Problem `json:"problem"`
			}{
				resource.GetProblem(problems[resp.Data.ID]),
			},
		}, resp)
	})
}

func TestGetUserProblemInfo(t *testing.T) {
	t.Parallel()

	user := createUserForTest(t, "get_user_problem_info", 1)
	otherUser := createUserForTest(t, "get_user_problem_info", 2)
	problem1 := createProblemForTest(t, "get_user_problem_info", 1, nil, user)
	problem2 := createProblemForTest(t, "get_user_problem_info", 2, nil, user)
	problem3 := createProblemForTest(t, "get_user_problem_info", 3, nil, user)

	submissionPassed1 := createSubmissionForTest(t, "get_user_problem_info_1_passed", 1, &problem1, &user, nil, 0)
	submissionPassed2 := createSubmissionForTest(t, "get_user_problem_info_1_passed", 2, &problem1, &otherUser, nil, 0)
	submissionPassed3 := createSubmissionForTest(t, "get_user_problem_info_2_passed", 3, &problem2, &user, nil, 0)
	submissionPassed4 := createSubmissionForTest(t, "get_user_problem_info_2_passed", 4, &problem2, &user, nil, 0)
	submissionPassed1.Status = "ACCEPTED"
	submissionPassed2.Status = "ACCEPTED"
	submissionPassed3.Status = "ACCEPTED"
	submissionPassed4.Status = "ACCEPTED"
	assert.NoError(t, base.DB.Save(&submissionPassed1).Error)
	assert.NoError(t, base.DB.Save(&submissionPassed2).Error)
	assert.NoError(t, base.DB.Save(&submissionPassed3).Error)
	assert.NoError(t, base.DB.Save(&submissionPassed4).Error)

	submissionFailed1 := createSubmissionForTest(t, "get_user_problem_info_1_failed", 1, &problem1, &otherUser, nil, 0)
	submissionFailed2 := createSubmissionForTest(t, "get_user_problem_info_2_failed", 2, &problem2, &user, nil, 0)
	submissionFailed3 := createSubmissionForTest(t, "get_user_problem_info_3_failed", 3, &problem3, &user, nil, 0)
	submissionFailed4 := createSubmissionForTest(t, "get_user_problem_info_3_failed", 4, &problem3, &user, nil, 0)
	submissionFailed1.Status = "WRONG_ANSWER"
	submissionFailed2.Status = "TIME_LIMIT_EXCEEDED"
	submissionFailed3.Status = "RUNTIME_ERROR"
	submissionFailed4.Status = "PENDING"
	assert.NoError(t, base.DB.Save(&submissionFailed1).Error)
	assert.NoError(t, base.DB.Save(&submissionFailed2).Error)
	assert.NoError(t, base.DB.Save(&submissionFailed3).Error)
	assert.NoError(t, base.DB.Save(&submissionFailed4).Error)

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("problem.getUserProblemInfo", user.ID),
			request.GetUserProblemInfo{}, applyNormalUser))
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		resp := response.GetUserProblemInfoResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.GetUserProblemInfoResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				TriedCount  int `json:"tried_count"`
				PassedCount int `json:"passed_count"`
				Rank        int `json:"rank"`
			}{
				TriedCount:  1,
				PassedCount: 2,
				Rank:        0,
			},
		}, resp)
	})

	t.Run("NonExistingUser", func(t *testing.T) {
		t.Parallel()
		httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("problem.getUserProblemInfo", -1),
			request.GetUserProblemInfo{}, applyNormalUser))
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		resp := response.GetUserProblemInfoResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.GetUserProblemInfoResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				TriedCount  int `json:"tried_count"`
				PassedCount int `json:"passed_count"`
				Rank        int `json:"rank"`
			}{
				TriedCount:  0,
				PassedCount: 0,
				Rank:        0,
			},
		}, resp)
	})
}
