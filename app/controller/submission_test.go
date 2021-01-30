package controller_test

import (
	"fmt"
	"github.com/leoleoasd/EduOJBackend/app/request"
	"github.com/leoleoasd/EduOJBackend/app/response"
	"github.com/leoleoasd/EduOJBackend/app/response/resource"
	"github.com/leoleoasd/EduOJBackend/base"
	"github.com/leoleoasd/EduOJBackend/database/models"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestCreateSubmission(t *testing.T) {
	// publicFalseProblem means a problem which "public" field is false
	publicFalseProblem, _ := createProblemForTest(t, "test_create_submission_public_false", 0, nil)
	assert.Nil(t, base.DB.Model(&publicFalseProblem).Update("public", false).Error)
	assert.Nil(t, base.DB.Model(&publicFalseProblem).Update("language_allowed", "test_language,golang").Error)
	failTests := []failTest{
		{
			// testCreateSubmissionNonExistingProblem
			name:   "NonExistingProblem",
			method: "POST",
			path:   base.Echo.Reverse("submission.createSubmission", -1),
			req: addFieldContentSlice([]reqContent{
				newFileContent("code", "code_file_name", b64Encode("test code content")),
			}, map[string]string{"language": "test_language"}),
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			// testCreateSubmissionPublicFalseProblem
			name:   "PublicFalseProblem",
			method: "POST",
			path:   base.Echo.Reverse("submission.createSubmission", publicFalseProblem.ID),
			req: addFieldContentSlice([]reqContent{
				newFileContent("code", "code_file_name", b64Encode("test code content")),
			}, map[string]string{"language": "test_language"}),
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			// testCreateSubmissionWithoutCode
			name:   "WithoutCode",
			method: "POST",
			path:   base.Echo.Reverse("submission.createSubmission", publicFalseProblem.ID),
			req: request.CreateSubmissionRequest{
				Language: "test_language",
			},
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusBadRequest,
			resp:       response.ErrorResp("INVALID_FILE", nil),
		},
		{
			// testCreateSubmissionInvalidLanguage
			name:   "InvalidLanguage",
			method: "POST",
			path:   base.Echo.Reverse("submission.createSubmission", publicFalseProblem.ID),
			req: addFieldContentSlice([]reqContent{
				newFileContent("code", "code_file_name", b64Encode("test code content")),
			}, map[string]string{"language": "invalid_language"}),
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusBadRequest,
			resp:       response.ErrorResp("INVALID_LANGUAGE", nil),
		},
	}

	// testCreateSubmissionFail
	runFailTests(t, failTests, "CreateSubmission")

	const (
		normalUser = iota
		problemCreator
		adminUser
	)

	successfulTests := []struct {
		name          string
		testCaseCount int
		problemPublic bool
		requestUser   int // 0->normalUser / 1->problemCreator / 2->adminUser
		response      resource.SubmissionDetail
	}{
		// testCreateSubmissionWithoutTestCases
		{
			name:          "WithoutTestCases",
			testCaseCount: 0,
			problemPublic: true,
			requestUser:   normalUser,
		},
		// testCreateSubmissionPublicProblem
		{
			name:          "PublicProblem",
			testCaseCount: 1,
			problemPublic: true,
			requestUser:   normalUser,
		},
		// testCreateSubmissionCreator
		{
			name:          "Creator",
			testCaseCount: 2,
			problemPublic: true,
			requestUser:   problemCreator,
		},
		// testCreateSubmissionAdmin
		{
			name:          "Admin",
			testCaseCount: 5,
			problemPublic: true,
			requestUser:   adminUser,
		},
	}
	t.Run("testCreateSubmissionSuccess", func(t *testing.T) {
		for i, test := range successfulTests {
			i := i
			test := test
			t.Run("testCreateSubmission"+test.name, func(t *testing.T) {
				problem, creator := createProblemForTest(t, "test_create_submission", i, nil)
				assert.Nil(t, base.DB.Model(&problem).Update("language_allowed", "test_language,golang").Error)
				for j := 0; j < test.testCaseCount; j++ {
					createTestCaseForTest(t, problem, testCaseData{
						Score:  0,
						Sample: true,
						InputFile: newFileContent("input", "input_file",
							b64Encode(fmt.Sprintf("test_create_submission_%d_test_case_%d_input_content", i, j))),
						OutputFile: newFileContent("output", "output_file",
							b64Encode(fmt.Sprintf("test_create_submission_%d_test_case_%d_output_content", i, j))),
					})
				}
				problem.LoadTestCases()
				var applyUser reqOption
				switch test.requestUser {
				case normalUser:
					applyUser = applyNormalUser
				case problemCreator:
					applyUser = headerOption{
						"Set-User-For-Test": {fmt.Sprintf("%d", creator.ID)},
					}
				case adminUser:
					applyUser = applyAdminUser
				default:
					t.Fail()
				}
				req := makeReq(t, "POST", base.Echo.Reverse("submission.createSubmission", problem.ID),
					addFieldContentSlice([]reqContent{
						newFileContent("code", "code_file_name", b64Encode(fmt.Sprintf("test_create_submission_%d_code", i))),
					}, map[string]string{
						"language": "test_language",
					}), applyUser)
				httpResp := makeResp(req)
				resp := response.CreateSubmissionResponse{}
				mustJsonDecode(httpResp, &resp)
				responseSubmission := *resp.Data.SubmissionDetail
				databaseSubmission := models.Submission{}
				reqUserID, err := strconv.ParseUint(req.Header.Get("Set-User-For-Test"), 10, 64)
				assert.Nil(t, err)
				assert.Nil(t, base.DB.Preload("Runs").First(&databaseSubmission, "problem_id = ? and user_id = ?", problem.ID, reqUserID).Error)
				databaseSubmissionDetail := resource.GetSubmissionDetail(&databaseSubmission)
				databaseRunData := map[uint]struct {
					ID        uint
					CreatedAt time.Time
				}{}
				for _, run := range databaseSubmission.Runs {
					databaseRunData[run.TestCaseID] = struct {
						ID        uint
						CreatedAt time.Time
					}{
						ID:        run.ID,
						CreatedAt: run.CreatedAt,
					}
				}
				expectedRunSlice := make([]resource.Run, test.testCaseCount)
				for i, testCase := range problem.TestCases {
					expectedRunSlice[i] = resource.Run{
						ID:           databaseRunData[testCase.ID].ID,
						UserID:       uint(reqUserID),
						ProblemID:    problem.ID,
						ProblemSetId: 0,
						TestCaseID:   testCase.ID,
						Sample:       testCase.Sample,
						SubmissionID: databaseSubmission.ID,
						Priority:     127,
						Judged:       false,
						Status:       "PENDING",
						MemoryUsed:   0,
						TimeUsed:     0,
						CreatedAt:    databaseRunData[testCase.ID].CreatedAt,
					}
				}
				expectedSubmission := resource.SubmissionDetail{
					ID:           databaseSubmissionDetail.ID,
					UserID:       uint(reqUserID),
					ProblemID:    problem.ID,
					ProblemSetId: 0,
					Language:     "test_language",
					FileName:     "code_file_name",
					Priority:     127,
					Judged:       false,
					Score:        0,
					Status:       "PENDING",
					Runs:         expectedRunSlice,
					CreatedAt:    databaseSubmission.CreatedAt,
				}
				assert.Equal(t, &expectedSubmission, databaseSubmissionDetail)
				assert.Equal(t, expectedSubmission, responseSubmission)
			})
		}

	})

}
