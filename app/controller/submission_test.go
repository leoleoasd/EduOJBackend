package controller_test

import (
	"context"
	"fmt"
	"github.com/EduOJ/backend/app/request"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/app/response/resource"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/database"
	"github.com/EduOJ/backend/database/models"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"
)

const (
	normalUser = iota
	problemCreator
	adminUser
	submitter
)

func createSubmissionForTest(t *testing.T, name string, id int, problem *models.Problem, user *models.User, code *fileContent, testCaseCount int) models.Submission {
	for i := 0; i < testCaseCount; i++ {
		createTestCaseForTest(t, *problem, testCaseData{
			Score:      0,
			Sample:     false,
			InputFile:  newFileContent("input", "input_file", b64Encodef("problem_%d_test_case_%d_input", problem.ID, i)),
			OutputFile: newFileContent("output", "output_file", b64Encodef("problem_%d_test_case_%d_output", problem.ID, i)),
		})
	}
	problem.LoadTestCases()
	submission := models.Submission{
		UserID:       user.ID,
		User:         user,
		ProblemID:    problem.ID,
		Problem:      problem,
		ProblemSetId: 0,
		LanguageName: "test_language",
		Language:     nil,
		FileName:     fmt.Sprintf("test_%s_code_file_name_%d.test_language", name, id),
		Priority:     models.PriorityDefault,
		Judged:       false,
		Score:        0,
		Status:       "PENDING",
		Runs:         make([]models.Run, len(problem.TestCases)),
		CreatedAt:    time.Time{},
		UpdatedAt:    time.Time{},
	}
	for i, testCase := range problem.TestCases {
		submission.Runs[i] = models.Run{
			UserID:             user.ID,
			ProblemID:          problem.ID,
			ProblemSetId:       0,
			TestCaseID:         testCase.ID,
			Sample:             testCase.Sample,
			Priority:           models.PriorityDefault,
			Judged:             false,
			Status:             "PENDING",
			MemoryUsed:         0,
			TimeUsed:           0,
			OutputStrippedHash: "",
		}
	}
	assert.NoError(t, base.DB.Create(&submission).Error)
	if code != nil {
		_, err := base.Storage.PutObject(context.Background(), "submissions", fmt.Sprintf("%d/code", submission.ID), code.reader, code.size, minio.PutObjectOptions{})
		assert.NoError(t, err)
		_, err = code.reader.Seek(0, io.SeekStart)
		assert.NoError(t, err)
	}
	return submission
}

func TestCreateSubmission(t *testing.T) {
	t.Parallel()
	user := createUserForTest(t, "test_create_submission_public_false", 0)
	// publicFalseProblem means a problem which "public" field is false
	publicFalseProblem := createProblemForTest(t, "test_create_submission_public_false", 0, nil, user)
	publicFalseProblem.Public = false
	publicFalseProblem.LanguageAllowed = []string{"test_language", "golang"}
	assert.NoError(t, base.DB.Save(&publicFalseProblem).Error)

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
		t.Parallel()
		for i, test := range successfulTests {
			i := i
			test := test
			t.Run("testCreateSubmission"+test.name, func(t *testing.T) {
				t.Parallel()
				creator := createUserForTest(t, "test_create_submission", i)
				problem := createProblemForTest(t, "test_create_submission", i, nil, creator)
				problem.LanguageAllowed = []string{"test_language", "golang"}
				assert.NoError(t, base.DB.Save(&problem).Error)
				for j := 0; j < test.testCaseCount; j++ {
					createTestCaseForTest(t, problem, testCaseData{
						Score:  0,
						Sample: true,
						InputFile: newFileContent("input", "input_file",
							b64Encodef("test_create_submission_%d_test_case_%d_input_content", i, j)),
						OutputFile: newFileContent("output", "output_file",
							b64Encodef("test_create_submission_%d_test_case_%d_output_content", i, j)),
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
						newFileContent("code", "code_file_name.test_language", b64Encodef("test_create_submission_%d_code", i)),
					}, map[string]string{
						"language": "test_language",
					}), applyUser)
				httpResp := makeResp(req)
				resp := response.CreateSubmissionResponse{}
				assert.Equal(t, http.StatusCreated, httpResp.StatusCode)
				mustJsonDecode(httpResp, &resp)

				responseSubmission := *resp.Data.SubmissionDetail
				databaseSubmission := models.Submission{}
				reqUserID, err := strconv.ParseUint(req.Header.Get("Set-User-For-Test"), 10, 64)
				assert.NoError(t, err)
				assert.NoError(t, base.DB.Preload("Runs").Preload("User").Preload("Problem").
					First(&databaseSubmission, "problem_id = ? and user_id = ?", problem.ID, reqUserID).Error)
				databaseSubmissionDetail := resource.GetSubmissionDetail(&databaseSubmission)
				databaseRunData := map[uint]models.Run{}
				for _, run := range databaseSubmission.Runs {
					databaseRunData[run.TestCaseID] = run
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
						UpdatedAt:    databaseRunData[testCase.ID].UpdatedAt,
					}
				}
				reqUser := models.User{}
				assert.NoError(t, base.DB.First(&reqUser, reqUserID).Error)
				expectedSubmission := resource.SubmissionDetail{
					ID:           databaseSubmissionDetail.ID,
					UserID:       uint(reqUserID),
					User:         resource.GetUser(&reqUser),
					ProblemID:    problem.ID,
					ProblemName:  problem.Name,
					ProblemSetId: 0,
					Language:     "test_language",
					FileName:     "code_file_name.test_language",
					Priority:     127,
					Judged:       false,
					Score:        0,
					Status:       "PENDING",
					Runs:         expectedRunSlice,
					CreatedAt:    databaseSubmission.CreatedAt,
					UpdatedAt:    databaseSubmission.UpdatedAt,
				}
				assert.Equal(t, &expectedSubmission, databaseSubmissionDetail)
				assert.Equal(t, expectedSubmission, responseSubmission)

				storageContent := getObjectContent(t, "submissions", fmt.Sprintf("%d/code", databaseSubmissionDetail.ID))
				expectedContent := fmt.Sprintf("test_create_submission_%d_code", i)
				assert.Equal(t, []byte(expectedContent), storageContent)
			})
		}

	})
}

func TestGetSubmission(t *testing.T) {
	t.Parallel()

	// notPublicProblem means a problem which "public" field is false
	notPublicProblemCreator := createUserForTest(t, "get_submission_fail", 0)
	notPublicProblem := createProblemForTest(t, "get_submission_fail", 0, nil, notPublicProblemCreator)
	assert.NoError(t, base.DB.Model(&notPublicProblem).Update("public", false).Error)
	publicFalseSubmission := createSubmissionForTest(t, "get_submission_fail", 0, &notPublicProblem, &notPublicProblemCreator,
		newFileContent("code", "code_file_name", b64Encode("test_get_submission_fail_0")), 2)
	publicProblemCreator := createUserForTest(t, "get_submission_fail", 1)
	publicProblem := createProblemForTest(t, "get_submission_fail", 1, nil, publicProblemCreator)
	publicSubmission := createSubmissionForTest(t, "get_submission_fail", 1, &publicProblem, &publicProblemCreator,
		newFileContent("code", "code_file_name", b64Encode("test_get_submission_fail_1")), 2)

	failTests := []failTest{
		{
			// testGetSubmissionNormalUserNonExisting
			name:   "NormalUserNonExisting",
			method: "GET",
			path:   base.Echo.Reverse("submission.getSubmission", -1),
			req:    request.GetSubmissionRequest{},
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			// testGetSubmissionAdminUserNonExisting
			name:   "AdminUserNonExisting",
			method: "GET",
			path:   base.Echo.Reverse("submission.getSubmission", -1),
			req:    request.GetSubmissionRequest{},
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			// testGetSubmissionPublicFalse
			name:   "PublicFalse",
			method: "GET",
			path:   base.Echo.Reverse("submission.getSubmission", publicFalseSubmission.ID),
			req:    request.GetSubmissionRequest{},
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			// testGetSubmissionSubmittedByOthers
			name:   "SubmittedByOthers",
			method: "GET",
			path:   base.Echo.Reverse("submission.getSubmission", publicSubmission.ID),
			req:    request.GetSubmissionRequest{},
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	// testGetSubmissionFail
	runFailTests(t, failTests, "GetSubmission")

	successTests := []struct {
		name          string
		code          *fileContent
		testCaseCount int
		requestUser   uint
	}{
		{
			// testGetSubmissionWithoutTestCases
			name:          "WithoutTestCases",
			code:          newFileContent("code", "code_file_name", b64Encode("test_get_submission_code_0")),
			testCaseCount: 0,
			requestUser:   adminUser,
		},
		{
			// testGetSubmissionAdminUser
			name:          "AdminUser",
			code:          newFileContent("code", "code_file_name", b64Encode("test_get_submission_code_1")),
			testCaseCount: 2,
			requestUser:   adminUser,
		},
		{
			// testGetSubmissionSubmitter
			name:          "Submitter",
			code:          newFileContent("code", "code_file_name", b64Encode("test_get_submission_code_2")),
			testCaseCount: 2,
			requestUser:   submitter,
		},
	}

	t.Run("testGetSubmissionSuccess", func(t *testing.T) {
		t.Parallel()
		for i, test := range successTests {
			i := i
			test := test
			t.Run("testGetSubmission"+test.name, func(t *testing.T) {
				t.Parallel()
				user := createUserForTest(t, "get_submission", i)
				problem := createProblemForTest(t, "get_submission", i, nil, user)
				base.DB.Model(&problem).Update("public", false)
				submission := createSubmissionForTest(t, "get_submission", i, &problem, &user, test.code, test.testCaseCount)
				var applyUser reqOption
				switch test.requestUser {
				case adminUser:
					applyUser = applyAdminUser
				case submitter:
					applyUser = headerOption{
						"Set-User-For-Test": {fmt.Sprintf("%d", user.ID)},
					}
				default:
					t.Fail()
				}
				httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("submission.getSubmission", submission.ID),
					request.GetSubmissionRequest{}, applyUser))
				resp := response.GetSubmissionResponse{}
				mustJsonDecode(httpResp, &resp)
				assert.Equal(t, http.StatusOK, httpResp.StatusCode)
				expectedSubmissionDetail := resource.GetSubmissionDetail(&submission)
				expectedSubmissionDetail.CreatedAt.UTC()
				assert.Equal(t, response.GetSubmissionResponse{
					Message: "SUCCESS",
					Error:   nil,
					Data: struct {
						*resource.SubmissionDetail `json:"submission"`
					}{
						expectedSubmissionDetail,
					},
				}, resp)
			})
		}
	})
}

func TestGetSubmissions(t *testing.T) {
	// Not Parallel
	t.Cleanup(database.SetupDatabaseForTest())

	problemCreator1 := createUserForTest(t, "get_submissions", 1)
	problemCreator2 := createUserForTest(t, "get_submissions", 2)
	problemCreator3 := createUserForTest(t, "get_submissions", 3)
	problem1 := createProblemForTest(t, "get_submissions", 1, nil, problemCreator1)
	problem2 := createProblemForTest(t, "get_submissions", 2, nil, problemCreator2)
	problem3 := createProblemForTest(t, "get_submissions", 3, nil, problemCreator3)
	base.DB.Model(&problem1).Update("public", false)
	submissionRelations := []struct {
		problem   *models.Problem
		submitter *models.User
	}{
		0: {
			problem:   &problem1,
			submitter: &problemCreator1,
		},
		1: {
			problem:   &problem2,
			submitter: &problemCreator1,
		},
		2: {
			problem:   &problem2,
			submitter: &problemCreator2,
		},
		3: {
			problem:   &problem2,
			submitter: &problemCreator3,
		},
		4: {
			problem:   &problem2,
			submitter: &problemCreator2,
		},
		5: {
			problem:   &problem3,
			submitter: &problemCreator3,
		},
		// TODO: test submission in problem sets
	}
	submissions := make([]models.Submission, len(submissionRelations))

	for i := range submissions {
		submissions[i] = createSubmissionForTest(t, "get_submissions", i, submissionRelations[i].problem, submissionRelations[i].submitter,
			newFileContent("code", "code_file_name", b64Encodef("test_get_submissions_code_%d", i)), 0)
	}

	successTests := []struct {
		name        string
		req         request.GetSubmissionsRequest
		submissions []models.Submission
		Total       int
		Offset      int
		Prev        *string
		Next        *string
	}{
		{
			// testGetSubmissionsAll
			name: "All",
			req: request.GetSubmissionsRequest{
				ProblemId: 0,
				UserId:    0,
				Limit:     0,
				Offset:    0,
			},
			submissions: []models.Submission{
				submissions[5],
				submissions[4],
				submissions[3],
				submissions[2],
				submissions[1],
				submissions[0],
			},
			Total:  6,
			Offset: 0,
			Prev:   nil,
			Next:   nil,
		},
		{
			// testGetSubmissionsSelectUser
			name: "SelectUser",
			req: request.GetSubmissionsRequest{
				ProblemId: 0,
				UserId:    problemCreator3.ID,
				Limit:     0,
				Offset:    0,
			},
			submissions: []models.Submission{
				submissions[5],
				submissions[3],
			},
			Total:  2,
			Offset: 0,
			Prev:   nil,
			Next:   nil,
		},
		{
			// testGetSubmissionsSelectProblem
			name: "SelectProblem",
			req: request.GetSubmissionsRequest{
				ProblemId: problem2.ID,
				UserId:    0,
				Limit:     0,
				Offset:    0,
			},
			submissions: []models.Submission{
				submissions[4],
				submissions[3],
				submissions[2],
				submissions[1],
			},
			Total:  4,
			Offset: 0,
			Prev:   nil,
			Next:   nil,
		},
		{
			// testGetSubmissionsSelectUserAndProblem
			name: "SelectUserAndProblem",
			req: request.GetSubmissionsRequest{
				ProblemId: problem2.ID,
				UserId:    problemCreator2.ID,
				Limit:     0,
				Offset:    0,
			},
			submissions: []models.Submission{
				submissions[4],
				submissions[2],
			},
			Total:  2,
			Offset: 0,
			Prev:   nil,
			Next:   nil,
		},
		{
			// testGetSubmissionsPaginator
			name: "Paginator",
			req: request.GetSubmissionsRequest{
				ProblemId: 0,
				UserId:    0,
				Limit:     3,
				Offset:    1,
			},
			submissions: []models.Submission{
				submissions[4],
				submissions[3],
				submissions[2],
			},
			Total:  6,
			Offset: 1,
			Prev:   nil,
			Next: getUrlStringPointer("submission.getSubmissions", map[string]string{
				"limit":  "3",
				"offset": "4",
			}),
		},
	}

	t.Run("testGetSubmissionsSuccess", func(t *testing.T) {
		t.Parallel()
		for _, test := range successTests {
			test := test
			t.Run("testGetSubmissions"+test.name, func(t *testing.T) {
				t.Parallel()
				httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("submission.getSubmissions"), test.req, applyNormalUser))
				resp := response.GetSubmissionsResponse{}
				mustJsonDecode(httpResp, &resp)
				assert.Equal(t, http.StatusOK, httpResp.StatusCode)
				assert.Equal(t, response.GetSubmissionsResponse{
					Message: "SUCCESS",
					Error:   nil,
					Data: struct {
						Submissions []resource.Submission `json:"submissions"`
						Total       int                   `json:"total"`
						Count       int                   `json:"count"`
						Offset      int                   `json:"offset"`
						Prev        *string               `json:"prev"`
						Next        *string               `json:"next"`
					}{
						Submissions: resource.GetSubmissionSlice(test.submissions),
						Total:       test.Total,
						Count:       len(test.submissions),
						Offset:      test.Offset,
						Prev:        test.Prev,
						Next:        test.Next,
					},
				}, resp)
			})
		}
	})
}

func TestGetSubmissionCode(t *testing.T) {
	t.Parallel()

	// notPublicProblem means a problem which "public" field is false
	notPublicProblemCreator := createUserForTest(t, "get_submission_code_fail", 0)
	notPublicProblem := createProblemForTest(t, "get_submission_code_fail", 0, nil, notPublicProblemCreator)
	assert.NoError(t, base.DB.Model(&notPublicProblem).Update("public", false).Error)
	notPublicSubmission := createSubmissionForTest(t, "get_submission_code_fail", 0, &notPublicProblem, &notPublicProblemCreator,
		newFileContent("code", "code_file_name", b64Encode("test_get_submission_code_fail_0")), 2)
	publicProblemCreator := createUserForTest(t, "get_submission_code_fail", 1)
	publicProblem := createProblemForTest(t, "get_submission_code_fail", 1, nil, publicProblemCreator)
	publicSubmission := createSubmissionForTest(t, "get_submission_code_fail", 1, &publicProblem, &publicProblemCreator,
		newFileContent("code", "code_file_name", b64Encode("test_get_submission_code_fail_1")), 2)

	failTests := []failTest{
		{
			// testGetSubmissionCodeNormalUserNonExisting
			name:   "NormalUserNonExisting",
			method: "GET",
			path:   base.Echo.Reverse("submission.getSubmissionCode", -1),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			// testGetSubmissionCodeAdminUserNonExisting
			name:   "AdminUserNonExisting",
			method: "GET",
			path:   base.Echo.Reverse("submission.getSubmissionCode", -1),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			// testGetSubmissionCodePublicFalse
			name:   "PublicFalse",
			method: "GET",
			path:   base.Echo.Reverse("submission.getSubmissionCode", notPublicSubmission.ID),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			// testGetSubmissionCodeSubmittedByOthers
			name:   "SubmittedByOthers",
			method: "GET",
			path:   base.Echo.Reverse("submission.getSubmissionCode", publicSubmission.ID),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	// testGetSubmissionCodeFail
	runFailTests(t, failTests, "GetSubmissionCode")

	t.Run("testGetSubmissionCodeSuccess", func(t *testing.T) {
		t.Parallel()
		content := "test_get_submission_code_content"
		user := createUserForTest(t, "get_submission_code", 0)
		problem := createProblemForTest(t, "get_submission_code", 0, nil, user)
		base.DB.Model(&problem).Update("public", false)
		submission := createSubmissionForTest(t, "get_submission_code", 0, &problem, &user,
			newFileContent("code", "code_file_name", b64Encode(content)), 2)
		httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("submission.getSubmissionCode", submission.ID),
			nil, applyAdminUser))
		assert.Equal(t, http.StatusFound, httpResp.StatusCode)
		assert.Equal(t, content, getPresignedURLContent(t, httpResp.Header.Get("Location")))
	})
}

func TestGetRunCompilerOutput(t *testing.T) {
	t.Parallel()

	// notPublicProblem means a problem which "public" field is false
	notPublicProblemCreator := createUserForTest(t, "get_run_compiler_output_fail", 0)
	notPublicProblem := createProblemForTest(t, "get_run_compiler_output_fail", 0, nil, notPublicProblemCreator)
	assert.NoError(t, base.DB.Model(&notPublicProblem).Update("public", false).Error)
	notPublicSubmission := createSubmissionForTest(t, "get_run_compiler_output_fail", 0, &notPublicProblem, &notPublicProblemCreator,
		newFileContent("code", "code_file_name", b64Encode("test_get_run_compiler_output_fail_0")), 2)

	publicProblemCreator := createUserForTest(t, "get_run_compiler_output_fail", 1)
	publicProblem := createProblemForTest(t, "get_run_compiler_output_fail", 1, nil, publicProblemCreator)
	publicSubmission := createSubmissionForTest(t, "get_run_compiler_output_fail", 1, &publicProblem, &publicProblemCreator,
		newFileContent("code", "code_file_name", b64Encode("test_get_run_compiler_output_fail_1")), 2)

	failTests := []failTest{
		{
			// testGetRunCompilerOutputNormalUserNonExistingSubmission
			name:   "NormalUserNonExistingSubmission",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunCompilerOutput", -1, -1),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			// testGetSubmissionCodeAdminUserNonExisting
			name:   "AdminUserNonExistingSubmission",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunCompilerOutput", -1, -1),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("SUBMISSION_NOT_FOUND", nil),
		},
		{
			// testGetRunCompilerOutputNonExistingRun
			name:   "NonExistingRun",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunCompilerOutput", publicSubmission.ID, -1),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			// testGetRunCompilerOutputPublicFalse
			name:   "PublicFalse",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunCompilerOutput", notPublicSubmission.ID, notPublicSubmission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			// testGetRunCompilerOutputSubmittedByOthers
			name:   "SubmittedByOthers",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunCompilerOutput", publicSubmission.ID, publicSubmission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	// testGetRunCompilerOutputFail
	runFailTests(t, failTests, "GetRunCompilerOutput")

	t.Run("testGetRunCompilerOutputSuccess", func(t *testing.T) {
		t.Parallel()
		content := "test_get_run_compiler_output_content"
		user := createUserForTest(t, "get_run_compiler_output", 0)
		problem := createProblemForTest(t, "get_run_compiler_output", 0, nil, user)
		base.DB.Model(&problem).Update("public", false)
		submission := createSubmissionForTest(t, "get_run_compiler_output", 0, &problem, &user,
			newFileContent("code", "code_file_name", b64Encode("code_content")), 2)
		file := newFileContent("compiler_output", "compiler.out", b64Encode(content))
		_, err := base.Storage.PutObject(context.Background(), "submissions", fmt.Sprintf("%d/run/%d/compiler_output", submission.ID, submission.Runs[0].ID), file.reader, file.size, minio.PutObjectOptions{})
		assert.NoError(t, err)
		httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("submission.getRunCompilerOutput", submission.ID, submission.Runs[0].ID),
			nil, applyAdminUser))
		assert.Equal(t, http.StatusFound, httpResp.StatusCode)
		assert.Equal(t, content, getPresignedURLContent(t, httpResp.Header.Get("Location")))
	})
}

func TestGetRunOutput(t *testing.T) {
	t.Parallel()

	// notPublicProblem means a problem which "public" field is false
	notPublicProblemCreator := createUserForTest(t, "get_run_output_fail", 0)
	notPublicProblem := createProblemForTest(t, "get_run_output_fail", 0, nil, notPublicProblemCreator)
	assert.NoError(t, base.DB.Model(&notPublicProblem).Update("public", false).Error)
	notPublicProblemSubmitter := createUserForTest(t, "get_run_output_fail_submit", 0)
	notPublicSubmission := createSubmissionForTest(t, "get_run_output_fail", 0, &notPublicProblem, &notPublicProblemSubmitter,
		newFileContent("code", "code_file_name", b64Encode("test_get_run_output_fail_0")), 2)

	publicProblemCreator := createUserForTest(t, "get_run_output_fail", 1)
	publicProblem := createProblemForTest(t, "get_run_output_fail", 1, nil, publicProblemCreator)
	publicProblemSubmitter := createUserForTest(t, "get_run_output_fail_submit", 1)
	publicSubmission := createSubmissionForTest(t, "get_run_output_fail", 1, &publicProblem, &publicProblemSubmitter,
		newFileContent("code", "code_file_name", b64Encode("test_get_run_output_fail_1")), 2)

	failTests := []failTest{
		{
			// testGetRunOutputNormalUserNonExistingSubmission
			name:   "NormalUserNonExistingSubmission",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunOutput", -1, -1),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			// testGetSubmissionCodeAdminUserNonExisting
			name:   "AdminUserNonExistingSubmission",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunOutput", -1, -1),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("SUBMISSION_NOT_FOUND", nil),
		},
		{
			// testGetRunOutputSubmitterNonExistingRun
			name:   "SubmitterNonExistingRun",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunOutput", publicSubmission.ID, -1),
			req:    nil,
			reqOptions: []reqOption{
				headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", publicProblemSubmitter.ID)},
				},
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			// testGetRunOutputAdminUserNonExistingRun
			name:   "AdminUserNonExistingRun",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunOutput", publicSubmission.ID, -1),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			// testGetRunOutputPublicFalse
			name:   "PublicFalse",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunOutput", notPublicSubmission.ID, notPublicSubmission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			// testGetRunOutputSubmittedByOthers
			name:   "SubmittedByOthers",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunOutput", publicSubmission.ID, publicSubmission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	// testGetRunOutputFail
	runFailTests(t, failTests, "GetRunOutput")

	successTests := []struct {
		name        string
		requestUser uint
		sample      bool
	}{
		{
			// testGetRunOutputAdminUser
			name:        "AdminUser",
			requestUser: adminUser,
			sample:      false,
		},
		{
			// testGetRunSubmitterSample
			name:        "SubmitterSample",
			requestUser: submitter,
			sample:      true,
		},
	}

	t.Run("testGetRunOutputSuccess", func(t *testing.T) {
		for i, test := range successTests {
			i := i
			test := test
			t.Run("testGetRunOutput"+test.name, func(t *testing.T) {
				t.Parallel()
				content := "test_get_run_output_content"
				user := createUserForTest(t, "get_run_output", i)
				problem := createProblemForTest(t, "get_run_output", i, nil, user)
				base.DB.Model(&problem).Update("public", false)
				submitterUser := createUserForTest(t, "get_run_output_submit", i)
				submission := createSubmissionForTest(t, "get_run_output", i, &problem, &submitterUser,
					newFileContent("code", "code_file_name", b64Encode("code_content")), 2)
				if test.sample {
					assert.NoError(t, base.DB.Model(&submission.Runs[0]).Update("sample", true).Error)
				}
				file := newFileContent("output", fmt.Sprintf("%d.out", i), b64Encode(content))
				_, err := base.Storage.PutObject(context.Background(), "submissions", fmt.Sprintf("%d/run/%d/output", submission.ID, submission.Runs[0].ID), file.reader, file.size, minio.PutObjectOptions{})
				assert.NoError(t, err)
				var applyUser reqOption
				switch test.requestUser {
				case submitter:
					applyUser = headerOption{
						"Set-User-For-Test": {fmt.Sprintf("%d", submitterUser.ID)},
					}
				case adminUser:
					applyUser = applyAdminUser
				default:
					t.Fail()
				}
				httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("submission.getRunOutput", submission.ID, submission.Runs[0].ID),
					nil, applyUser))
				assert.Equal(t, http.StatusFound, httpResp.StatusCode)
				assert.Equal(t, content, getPresignedURLContent(t, httpResp.Header.Get("Location")))
			})
		}
	})
}

func TestGetRunComparerOutput(t *testing.T) {
	t.Parallel()

	// notPublicProblem means a problem which "public" field is false
	notPublicProblemCreator := createUserForTest(t, "get_run_comparer_output_fail", 0)
	notPublicProblem := createProblemForTest(t, "get_run_comparer_output_fail", 0, nil, notPublicProblemCreator)
	assert.NoError(t, base.DB.Model(&notPublicProblem).Update("public", false).Error)
	notPublicProblemSubmitter := createUserForTest(t, "get_run_comparer_output_fail_submit", 0)
	notPublicSubmission := createSubmissionForTest(t, "get_run_comparer_output_fail", 0, &notPublicProblem, &notPublicProblemSubmitter,
		newFileContent("code", "code_file_name", b64Encode("test_get_run_comparer_output_fail_0")), 2)

	publicProblemCreator := createUserForTest(t, "get_run_comparer_output_fail", 1)
	publicProblem := createProblemForTest(t, "get_run_comparer_output_fail", 1, nil, publicProblemCreator)
	publicProblemSubmitter := createUserForTest(t, "get_run_comparer_output_fail_submit", 1)
	publicSubmission := createSubmissionForTest(t, "get_run_comparer_output_fail", 1, &publicProblem, &publicProblemSubmitter,
		newFileContent("code", "code_file_name", b64Encode("test_get_run_comparer_output_fail_1")), 2)

	failTests := []failTest{
		{
			// testGetRunComparerOutputNormalUserNonExistingSubmission
			name:   "NormalUserNonExistingSubmission",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunComparerOutput", -1, -1),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			// testGetSubmissionCodeAdminUserNonExisting
			name:   "AdminUserNonExistingSubmission",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunComparerOutput", -1, -1),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("SUBMISSION_NOT_FOUND", nil),
		},
		{
			// testGetRunComparerOutputSubmitterNonExistingRun
			name:   "SubmitterNonExistingRun",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunComparerOutput", publicSubmission.ID, -1),
			req:    nil,
			reqOptions: []reqOption{
				headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", publicProblemSubmitter.ID)},
				},
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			// testGetRunComparerOutputAdminUserNonExistingRun
			name:   "AdminUserNonExistingRun",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunComparerOutput", publicSubmission.ID, -1),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			// testGetRunComparerOutputNotSample
			name:   "SubmitterInvalidRunId",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunComparerOutput", publicSubmission.ID, publicSubmission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", publicProblemSubmitter.ID)},
				},
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			// testGetRunComparerOutputPublicFalse
			name:   "PublicFalse",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunComparerOutput", notPublicSubmission.ID, notPublicSubmission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			// testGetRunComparerOutputSubmittedByOthers
			name:   "SubmittedByOthers",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunComparerOutput", publicSubmission.ID, publicSubmission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	// testGetRunComparerOutputFail
	runFailTests(t, failTests, "GetRunComparerOutput")

	successTests := []struct {
		name        string
		requestUser uint
		sample      bool
	}{
		{
			// testGetRunComparerOutputAdminUser
			name:        "AdminUser",
			requestUser: adminUser,
			sample:      false,
		},
		{
			// testGetRunSubmitterSample
			name:        "SubmitterSample",
			requestUser: submitter,
			sample:      true,
		},
	}

	t.Run("testGetRunComparerOutputSuccess", func(t *testing.T) {
		for i, test := range successTests {
			i := i
			test := test
			t.Run("testGetRunComparerOutput"+test.name, func(t *testing.T) {
				t.Parallel()
				content := "test_get_run_comparer_output_content"
				user := createUserForTest(t, "get_run_comparer_output", i)
				problem := createProblemForTest(t, "get_run_comparer_output", i, nil, user)
				base.DB.Model(&problem).Update("public", false)
				submitterUser := createUserForTest(t, "get_run_comparer_output_submit", i)
				submission := createSubmissionForTest(t, "get_run_comparer_output", i, &problem, &submitterUser,
					newFileContent("code", "code_file_name", b64Encode("code_content")), 2)
				if test.sample {
					assert.NoError(t, base.DB.Model(&submission.Runs[0]).Update("sample", true).Error)
				}
				file := newFileContent("comparer_output", "comparer.out", b64Encode(content))
				_, err := base.Storage.PutObject(context.Background(), "submissions", fmt.Sprintf("%d/run/%d/comparer_output", submission.ID, submission.Runs[0].ID), file.reader, file.size, minio.PutObjectOptions{})
				assert.NoError(t, err)
				var applyUser reqOption
				switch test.requestUser {
				case submitter:
					applyUser = headerOption{
						"Set-User-For-Test": {fmt.Sprintf("%d", submitterUser.ID)},
					}
				case adminUser:
					applyUser = applyAdminUser
				default:
					t.Fail()
				}
				httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("submission.getRunComparerOutput", submission.ID, submission.Runs[0].ID),
					nil, applyUser))
				assert.Equal(t, http.StatusFound, httpResp.StatusCode)
				assert.Equal(t, content, getPresignedURLContent(t, httpResp.Header.Get("Location")))
			})
		}
	})
}

func TestGetRunInput(t *testing.T) {
	t.Parallel()

	user := createUserForTest(t, "get_run_input_fail", 1)
	problem := createProblemForTest(t, "get_run_input_fail", 1, nil, user)
	problemSubmitter := createUserForTest(t, "get_run_input_fail_submit", 1)
	submission := createSubmissionForTest(t, "get_run_input_fail", 1, &problem, &problemSubmitter,
		newFileContent("code", "code_file_name", b64Encode("test_get_run_input_fail_1")), 2)

	failTests := []failTest{
		{
			// testGetRunComparerOutputNormalUserNonExistingSubmission
			name:   "NormalUserNonExistingSubmission",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunInput", -1, -1),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			// testGetSubmissionCodeAdminUserNonExisting
			name:   "AdminUserNonExistingSubmission",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunInput", -1, -1),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("SUBMISSION_NOT_FOUND", nil),
		},
		{
			// testGetRunComparerOutputSubmitterNonExistingRun
			name:   "SubmitterNonExistingRun",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunInput", submission.ID, -1),
			req:    nil,
			reqOptions: []reqOption{
				headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", problemSubmitter.ID)},
				},
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			// testGetRunComparerOutputAdminUserNonExistingRun
			name:   "AdminUserNonExistingRun",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunInput", submission.ID, -1),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			// testGetRunComparerOutputNotSample
			name:   "SubmitterInvalidRunId",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunInput", submission.ID, submission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				headerOption{
					"Set-User-For-Test": {fmt.Sprintf("%d", problemSubmitter.ID)},
				},
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			// testGetRunComparerOutputSubmittedByOthers
			name:   "SubmittedByOthers",
			method: "GET",
			path:   base.Echo.Reverse("submission.getRunInput", submission.ID, submission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	// testGetRunComparerOutputFail
	runFailTests(t, failTests, "GetRunInput")

	successTests := []struct {
		name        string
		requestUser uint
		sample      bool
	}{
		{
			// testGetRunComparerOutputAdminUser
			name:        "AdminUser",
			requestUser: adminUser,
			sample:      false,
		},
		{
			// testGetRunSubmitterSample
			name:        "SubmitterSample",
			requestUser: submitter,
			sample:      true,
		},
	}

	t.Run("testGetRunInputSuccess", func(t *testing.T) {
		for i, test := range successTests {
			i := i
			test := test
			t.Run("testGetRunComparerOutput"+test.name, func(t *testing.T) {
				t.Parallel()
				user := createUserForTest(t, "get_run_input", i)
				problem := createProblemForTest(t, "get_run_input", i, nil, user)
				base.DB.Model(&problem).Update("public", false)
				submitterUser := createUserForTest(t, "get_run_input_submit", i)
				submission := createSubmissionForTest(t, "get_run_input", i, &problem, &submitterUser,
					newFileContent("code", "code_file_name", b64Encode("code_content")), 1)
				if test.sample {
					assert.NoError(t, base.DB.Model(&submission.Runs[0]).Update("sample", true).Error)
				}
				var applyUser reqOption
				switch test.requestUser {
				case submitter:
					applyUser = headerOption{
						"Set-User-For-Test": {fmt.Sprintf("%d", submitterUser.ID)},
					}
				case adminUser:
					applyUser = applyAdminUser
				default:
					t.Fail()
				}
				httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("submission.getRunInput", submission.ID, submission.Runs[0].ID),
					nil, applyUser))
				assert.Equal(t, http.StatusFound, httpResp.StatusCode)
				assert.Equal(t, fmt.Sprintf("problem_%d_test_case_%d_input", problem.ID, 0), getPresignedURLContent(t, httpResp.Header.Get("Location")))
			})
		}
	})
}
