package controller_test

import (
	"fmt"
	"github.com/EduOJ/backend/app/request"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/log"
	"github.com/EduOJ/backend/database/models"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestAdminGetLogs(t *testing.T) {

	failTests := []failTest{
		{
			name:   "LevelIsNotANumber",
			method: "GET",
			path:   base.Echo.Reverse("admin.getLogs"),
			req: request.AdminGetLogsRequest{
				Levels: "1,2,NotANumber",
				Limit:  0,
				Offset: 0,
			},
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusBadRequest,
			resp:       response.ErrorResp("INVALID_LEVEL", nil),
		},
		{
			name:   "LevelExceeded",
			method: "GET",
			path:   base.Echo.Reverse("admin.getLogs"),
			req: request.AdminGetLogsRequest{
				Levels: "1,2,5",
				Limit:  0,
				Offset: 0,
			},
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusBadRequest,
			resp:       response.ErrorResp("INVALID_LEVEL", nil),
		},
		{
			name:   "PermissionDenied",
			method: "GET",
			path:   base.Echo.Reverse("admin.getLogs"),
			req: request.AdminGetLogsRequest{
				Levels: "1,2,3",
				Limit:  0,
				Offset: 0,
			},
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "AdminGetLogs")

	logs := make([]models.Log, 5)
	var ll log.Level
	for ll = 0; ll < 5; ll++ {
		llInt := int(ll)
		l := models.Log{
			Level:   &llInt,
			Message: fmt.Sprintf("test_admin_get_logs_%s_message", ll.String()),
			Caller:  fmt.Sprintf("test_admin_get_logs_%s_caller", ll.String()),
		}
		assert.NoError(t, base.DB.Create(&l).Error)
		logs[ll] = l
	}

	successTests := []struct {
		name string
		req  request.AdminGetLogsRequest
		resp response.AdminGetLogsResponse
	}{
		{
			name: "SingleLevel",
			req: request.AdminGetLogsRequest{
				Levels: "2",
				Limit:  0,
				Offset: 0,
			},
			resp: response.AdminGetLogsResponse{
				Message: "SUCCESS",
				Error:   nil,
				Data: struct {
					Logs   []models.Log `json:"logs"`
					Total  int          `json:"total"`
					Count  int          `json:"count"`
					Offset int          `json:"offset"`
					Prev   *string      `json:"prev"`
					Next   *string      `json:"next"`
				}{
					Logs: []models.Log{
						logs[2],
					},
					Total:  1,
					Count:  1,
					Offset: 0,
					Prev:   nil,
					Next:   nil,
				},
			},
		},
		{
			name: "MultipleLevel",
			req: request.AdminGetLogsRequest{
				Levels: "0,1,2,4",
				Limit:  0,
				Offset: 0,
			},
			resp: response.AdminGetLogsResponse{
				Message: "SUCCESS",
				Error:   nil,
				Data: struct {
					Logs   []models.Log `json:"logs"`
					Total  int          `json:"total"`
					Count  int          `json:"count"`
					Offset int          `json:"offset"`
					Prev   *string      `json:"prev"`
					Next   *string      `json:"next"`
				}{
					Logs: []models.Log{
						logs[4], logs[2], logs[1], logs[0],
					},
					Total:  4,
					Count:  4,
					Offset: 0,
					Prev:   nil,
					Next:   nil,
				},
			},
		},
		{
			name: "LimitAndOffset",
			req: request.AdminGetLogsRequest{
				Levels: "0,1,2,4",
				Limit:  2,
				Offset: 1,
			},
			resp: response.AdminGetLogsResponse{
				Message: "SUCCESS",
				Error:   nil,
				Data: struct {
					Logs   []models.Log `json:"logs"`
					Total  int          `json:"total"`
					Count  int          `json:"count"`
					Offset int          `json:"offset"`
					Prev   *string      `json:"prev"`
					Next   *string      `json:"next"`
				}{
					Logs: []models.Log{
						logs[2], logs[1],
					},
					Total:  4,
					Count:  2,
					Offset: 1,
					Prev:   nil,
					Next: getUrlStringPointer("admin.getLogs", map[string]string{
						"limit":  "2",
						"offset": "3",
					}),
				},
			},
		},
	}

	t.Run("testAdminGetLogsSuccess", func(t *testing.T) {
		t.Parallel()
		for _, test := range successTests {
			test := test
			t.Run("testAdminGetLogs"+test.name, func(t *testing.T) {
				t.Parallel()
				httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("admin.getLogs"), test.req, applyAdminUser))
				assert.Equal(t, http.StatusOK, httpResp.StatusCode)
				resp := response.Response{}
				mustJsonDecode(httpResp, &resp)
				jsonEQ(t, test.resp, resp)
			})
		}
	})

}
