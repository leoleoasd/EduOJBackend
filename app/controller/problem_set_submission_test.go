package controller_test

import (
	"context"
	"fmt"
	"github.com/EduOJ/backend/app/request"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/app/response/resource"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/database/models"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestProblemSetCreateSubmission(t *testing.T) {
	t.Parallel()
	user := createUserForTest(t, "test_problem_set_create_submission_fail", 0)
	problem := createProblemForTest(t, "test_problem_set_create_submission_fail", 0, nil, user)
	problem.Public = false
	problem.LanguageAllowed = []string{"test_language", "golang"}
	assert.NoError(t, base.DB.Save(&problem).Error)
	class := createClassForTest(t, "test_problem_set_create_submission_fail", 0, nil, nil)
	problemSet := createProblemSetForTest(t, "test_problem_set_create_submission_fail", 0, &class, []models.Problem{problem})
	problemSet.StartTime = time.Now().Add(-1 * time.Hour)
	problemSet.EndTime = time.Now().Add(time.Hour)
	assert.NoError(t, base.DB.Save(&problemSet).Error)
	problemSetNotInOpenTime := createProblemSetForTest(t, "test_problem_set_create_submission_not_in_open_time", 0, &class, []models.Problem{problem})
	problemSetNotInOpenTime.StartTime = time.Now().Add(-1 * time.Hour)
	problemSetNotInOpenTime.EndTime = time.Now().Add(time.Hour)
	assert.NoError(t, base.DB.Save(&problemSetNotInOpenTime).Error)

	failTests := []failTest{
		{
			name:   "WithoutParas",
			method: "POST",
			path:   base.Echo.Reverse("problemSet.createSubmission", problemSet.ID, problem.ID),
			req:    request.ProblemSetGetSubmissionRequest{},
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusBadRequest,
			resp: response.ErrorResp("VALIDATION_ERROR", []interface{}{
				map[string]interface{}{
					"field":       "Language",
					"reason":      "required",
					"translation": "语言为必填字段",
				},
			}),
		},
		{
			name:   "NonExistingProblemSet",
			method: "POST",
			path:   base.Echo.Reverse("problemSet.createSubmission", -1, problem.ID),
			req: addFieldContentSlice([]reqContent{
				newFileContent("code", "code_file_name", b64Encode("test code content")),
			}, map[string]string{"language": "test_language"}),
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("PROBLEM_SET_NOT_FOUND", nil),
		},
		{
			name:   "NonExistingProblem",
			method: "POST",
			path:   base.Echo.Reverse("problemSet.createSubmission", problemSet.ID, -1),
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
			name:   "WithoutCode",
			method: "POST",
			path:   base.Echo.Reverse("problemSet.createSubmission", problemSet.ID, problem.ID),
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
			name:   "InvalidLanguage",
			method: "POST",
			path:   base.Echo.Reverse("problemSet.createSubmission", problemSet.ID, problem.ID),
			req: addFieldContentSlice([]reqContent{
				newFileContent("code", "code_file_name", b64Encode("test code content")),
			}, map[string]string{"language": "invalid_language"}),
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusBadRequest,
			resp:       response.ErrorResp("INVALID_LANGUAGE", nil),
		},
		{
			name:   "NotInOpenTime",
			method: "POST",
			path:   base.Echo.Reverse("problemSet.createSubmission", problemSetNotInOpenTime.ID, problem.ID),
			req: addFieldContentSlice([]reqContent{
				newFileContent("code", "code_file_name", b64Encode("test code content")),
			}, map[string]string{"language": "test_language"}),
			reqOptions: []reqOption{
				applyUser(user),
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			name:   "PermissionDenied",
			method: "POST",
			path:   base.Echo.Reverse("problemSet.createSubmission", problemSet.ID, problem.ID),
			req: addFieldContentSlice([]reqContent{
				newFileContent("code", "code_file_name", b64Encode("test code content")),
			}, map[string]string{"language": "test_language"}),
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "")

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		user := createUserForTest(t, "test_problem_set_create_submission_success", 0)
		student := createUserForTest(t, "test_problem_set_create_submission_success_student", 0)
		problem := createProblemForTest(t, "test_problem_set_create_submission_success", 0, nil, user)
		problem.Public = false
		problem.LanguageAllowed = []string{"test_language", "golang"}
		assert.NoError(t, base.DB.Save(&problem).Error)
		testCase1 := createTestCaseForTest(t, problem, testCaseData{
			Score:      10,
			Sample:     true,
			InputFile:  nil,
			OutputFile: nil,
		})
		testCase2 := createTestCaseForTest(t, problem, testCaseData{
			Score:      20,
			Sample:     false,
			InputFile:  nil,
			OutputFile: nil,
		})
		class := createClassForTest(t, "test_problem_set_create_submission_success", 0, nil, []*models.User{&student})
		problemSet := createProblemSetForTest(t, "test_problem_set_create_submission_success", 0, &class, []models.Problem{problem})
		problemSet.StartTime = time.Now().Add(-1 * time.Hour).UTC()
		problemSet.EndTime = time.Now().Add(time.Hour).UTC()
		assert.NoError(t, base.DB.Save(&problemSet).Error)
		assert.NoError(t, base.DB.First(&problem, problem.ID).Error)
		assert.NoError(t, base.DB.Preload("Problems").First(&problemSet, problemSet.ID).Error)

		httpResp := makeResp(makeReq(t, "POST", base.Echo.Reverse("problemSet.createSubmission", problemSet.ID, problem.ID),
			addFieldContentSlice([]reqContent{
				newFileContent("code", "code_file_name.test_language", b64Encode("problem_set_create_submission_code_success")),
			}, map[string]string{
				"language": "test_language",
			}), applyUser(student)))
		assert.Equal(t, http.StatusCreated, httpResp.StatusCode)

		language := models.Language{}
		assert.NoError(t, base.DB.First(&language, "name = ?", "test_language").Error)

		databaseSubmission := models.Submission{}
		assert.NoError(t, base.DB.Preload("Runs").Preload("User").Preload("Problem").
			Preload("Language").Preload("ProblemSet.Problems").Preload("ProblemSet.Grades").
			First(&databaseSubmission, "problem_id = ? and user_id = ?", problem.ID, student.ID).Error)
		expectedSubmission := models.Submission{
			ID:           databaseSubmission.ID,
			UserID:       student.ID,
			User:         &student,
			ProblemID:    problem.ID,
			Problem:      &problem,
			ProblemSetId: problemSet.ID,
			ProblemSet:   problemSet,
			LanguageName: "test_language",
			Language:     &language,
			FileName:     "code_file_name.test_language",
			Priority:     models.PriorityDefault + 8,
			Judged:       false,
			Score:        0,
			Status:       "PENDING",
			Runs: []models.Run{
				{
					ID:                 databaseSubmission.Runs[0].ID,
					UserID:             student.ID,
					ProblemID:          problem.ID,
					ProblemSetId:       problemSet.ID,
					TestCaseID:         testCase1.ID,
					Sample:             true,
					SubmissionID:       databaseSubmission.ID,
					Priority:           models.PriorityDefault + 8,
					Judged:             false,
					JudgerName:         "",
					JudgerMessage:      "",
					Status:             "PENDING",
					MemoryUsed:         0,
					TimeUsed:           0,
					OutputStrippedHash: "",
					CreatedAt:          databaseSubmission.Runs[0].CreatedAt,
					UpdatedAt:          databaseSubmission.Runs[0].UpdatedAt,
					DeletedAt:          gorm.DeletedAt{},
				},
				{
					ID:                 databaseSubmission.Runs[1].ID,
					UserID:             student.ID,
					ProblemID:          problem.ID,
					ProblemSetId:       problemSet.ID,
					TestCaseID:         testCase2.ID,
					Sample:             false,
					SubmissionID:       databaseSubmission.ID,
					Priority:           models.PriorityDefault + 8,
					Judged:             false,
					JudgerName:         "",
					JudgerMessage:      "",
					Status:             "PENDING",
					MemoryUsed:         0,
					TimeUsed:           0,
					OutputStrippedHash: "",
					CreatedAt:          databaseSubmission.Runs[1].CreatedAt,
					UpdatedAt:          databaseSubmission.Runs[1].UpdatedAt,
					DeletedAt:          gorm.DeletedAt{},
				},
			},
			CreatedAt: databaseSubmission.CreatedAt,
			UpdatedAt: databaseSubmission.UpdatedAt,
			DeletedAt: gorm.DeletedAt{},
		}
		assert.Equal(t, expectedSubmission, databaseSubmission)
		resp := response.ProblemSetCreateSubmissionResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.ProblemSetCreateSubmissionResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.SubmissionDetail `json:"submission"`
			}{
				resource.GetSubmissionDetail(&expectedSubmission),
			},
		}, resp)
		assert.Equal(t, "problem_set_create_submission_code_success",
			string(getObjectContent(t, "submissions", fmt.Sprintf("%d/code", databaseSubmission.ID))))
	})
}

func TestProblemSetGetSubmission(t *testing.T) {
	t.Parallel()

	user := createUserForTest(t, "problem_set_get_submission_fail", 0)
	problem := createProblemForTest(t, "problem_set_get_submission_fail", 0, nil, user)
	class := createClassForTest(t, "test_problem_set_get_submission_fail", 0, nil, nil)
	problemSet := createProblemSetForTest(t, "problem_set_get_submission_fail", 0, &class, []models.Problem{problem})
	submission := createSubmissionForTest(t, "problem_set_get_submission_fail", 0, &problem, &user,
		newFileContent("code", "code_file_name", b64Encode("problem_set_get_submission_fail_0")), 2)
	submission.ProblemSetId = problemSet.ID
	assert.NoError(t, base.DB.Save(&submission).Error)

	failTests := []failTest{
		{
			name:   "NonExistingProblemSet",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getSubmission", -1, submission.ID),
			req:    request.ProblemSetGetSubmissionRequest{},
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "NonExistingSubmission",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getSubmission", problemSet.ID, -1),
			req:    request.ProblemSetGetSubmissionRequest{},
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "PermissionDenied",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getSubmission", problemSet.ID, submission.ID),
			req:    request.ProblemSetGetSubmissionRequest{},
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "")

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		student := createUserForTest(t, "problem_set_get_submission_success", 0)
		problem := createProblemForTest(t, "problem_set_get_submission_success", 0, nil, student)
		class := createClassForTest(t, "test_problem_set_get_submission_success", 0, nil, nil)
		problemSet := createProblemSetForTest(t, "problem_set_get_submission_success", 0, &class, []models.Problem{problem})
		submission := createSubmissionForTest(t, "problem_set_get_submission_success", 0, &problem, &student,
			newFileContent("code", "code_file_name", b64Encode("problem_set_get_submission_success_0")), 2)
		submission.ProblemSetId = problemSet.ID
		assert.NoError(t, base.DB.Save(&submission).Error)
		httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("problemSet.getSubmission", problemSet.ID, submission.ID),
			request.ProblemSetGetSubmissionRequest{}, applyUser(student)))
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		resp := response.ProblemSetGetSubmissionResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.ProblemSetGetSubmissionResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.SubmissionDetail `json:"submission"`
			}{
				resource.GetSubmissionDetail(&submission),
			},
		}, resp)
	})
}

func TestProblemSetGetSubmissions(t *testing.T) {
	t.Parallel()

	failClass := createClassForTest(t, "test_problem_set_get_submissions_fail", 0, nil, nil)
	failProblemSet := createProblemSetForTest(t, "problem_problem_set_get_submissions_fail", 0, &failClass, nil)
	failTests := []failTest{
		{
			name:   "WithoutParas",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getSubmissions", failProblemSet.ID),
			req: request.ProblemSetGetSubmissionsRequest{
				Limit: -1,
			},
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusBadRequest,
			resp: response.ErrorResp("VALIDATION_ERROR", []interface{}{
				map[string]interface{}{
					"field":       "Limit",
					"reason":      "min",
					"translation": "单页个数最小只能为0",
				},
			}),
		},
	}
	runFailTests(t, failTests, "")

	problemCreator1 := createUserForTest(t, "problem_set_get_submissions", 1)
	problemCreator2 := createUserForTest(t, "problem_set_get_submissions", 2)
	problemCreator3 := createUserForTest(t, "problem_set_get_submissions", 3)
	problem1 := createProblemForTest(t, "problem_set_get_submissions", 1, nil, problemCreator1)
	problem2 := createProblemForTest(t, "problem_set_get_submissions", 2, nil, problemCreator2)
	problem3 := createProblemForTest(t, "problem_set_get_submissions", 3, nil, problemCreator3)
	student := createUserForTest(t, "problem_set_get_submissions", 0)
	class := createClassForTest(t, "problem_set_get_submissions", 0, nil, []*models.User{&student})
	problemSet := createProblemSetForTest(t, "problem_set_get_submissions", 0, &class, []models.Problem{problem1, problem2, problem3})
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
	}
	submissions := make([]models.Submission, len(submissionRelations))

	for i := range submissions {
		submissions[i] = createSubmissionForTest(t, "problem_set_get_submissions", i, submissionRelations[i].problem, submissionRelations[i].submitter,
			newFileContent("code", "code_file_name", b64Encodef("test_problem_set_get_submissions_code_%d", i)), 0)
		submissions[i].ProblemSetId = problemSet.ID
		assert.NoError(t, base.DB.Save(&submissions[i]).Error)
	}

	successTests := []struct {
		name        string
		req         request.ProblemSetGetSubmissionsRequest
		submissions []models.Submission
		Total       int
		Offset      int
		Prev        *string
		Next        *string
	}{
		{
			// testProblemSetGetSubmissionsAll
			name: "All",
			req: request.ProblemSetGetSubmissionsRequest{
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
			// testProblemSetGetSubmissionsSelectUser
			name: "SelectUser",
			req: request.ProblemSetGetSubmissionsRequest{
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
			// testProblemSetGetSubmissionsSelectProblem
			name: "SelectProblem",
			req: request.ProblemSetGetSubmissionsRequest{
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
			// testProblemSetGetSubmissionsSelectUserAndProblem
			name: "SelectUserAndProblem",
			req: request.ProblemSetGetSubmissionsRequest{
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
			// testProblemSetGetSubmissionsPaginator
			name: "Paginator",
			req: request.ProblemSetGetSubmissionsRequest{
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
			Next: getUrlStringPointer("problemSet.getSubmissions", map[string]string{
				"limit":  "3",
				"offset": "4",
			}, problemSet.ID),
		},
	}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		for _, test := range successTests {
			test := test
			t.Run("testProblemSetGetSubmissions"+test.name, func(t *testing.T) {
				t.Parallel()
				httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("problemSet.getSubmissions", problemSet.ID), test.req, applyNormalUser))
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

func TestProblemSetGetSubmissionCode(t *testing.T) {
	t.Parallel()

	user := createUserForTest(t, "problem_set_get_submission_code", 0)
	problem := createProblemForTest(t, "problem_set_get_submission_code", 0, nil, user)
	class := createClassForTest(t, "test_problem_set_get_submission_code", 0, nil, nil)
	problemSet := createProblemSetForTest(t, "problem_set_get_submission_code", 0, &class, []models.Problem{problem})
	submission := createSubmissionForTest(t, "problem_set_get_submission_code", 0, &problem, &user,
		newFileContent("code", "code_file_name", b64Encode("problem_set_get_submission_code_0")), 2)
	submission.ProblemSetId = problemSet.ID
	assert.NoError(t, base.DB.Save(&submission).Error)

	failTests := []failTest{
		{
			name:   "NonExistingProblemSet",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getSubmissionCode", -1, submission.ID),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "NonExistingSubmission",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getSubmissionCode", problemSet.ID, -1),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "PermissionDenied",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getSubmissionCode", problemSet.ID, submission.ID),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "")

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("problemSet.getSubmissionCode", problemSet.ID, submission.ID),
			nil, applyAdminUser))
		assert.Equal(t, http.StatusFound, httpResp.StatusCode)
		assert.Equal(t, "problem_set_get_submission_code_0", getPresignedURLContent(t, httpResp.Header.Get("Location")))
	})
}

func TestProblemSetGetRunCompilerOutput(t *testing.T) {
	t.Parallel()

	user := createUserForTest(t, "problem_set_get_run_compiler_output", 0)
	problem := createProblemForTest(t, "problem_set_get_run_compiler_output", 0, nil, user)
	class := createClassForTest(t, "test_problem_set_get_run_compiler_output", 0, nil, nil)
	problemSet := createProblemSetForTest(t, "problem_set_get_run_compiler_output", 0, &class, []models.Problem{problem})
	submission := createSubmissionForTest(t, "problem_set_get_run_compiler_output", 0, &problem, &user,
		newFileContent("compiler_output", "compiler_output_file_name",
			b64Encode("problem_set_get_run_compiler_output_0")), 2)
	submission.ProblemSetId = problemSet.ID
	for i := range submission.Runs {
		submission.Runs[i].ProblemSetId = problemSet.ID
		assert.NoError(t, base.DB.Save(&submission.Runs[i]).Error)
		content := fmt.Sprintf("problem_set_get_run_compiler_output_%d", i)
		var _, err = base.Storage.PutObject(context.Background(), "submissions",
			fmt.Sprintf("%d/run/%d/compiler_output", submission.ID, submission.Runs[i].ID),
			strings.NewReader(content), int64(len(content)), minio.PutObjectOptions{})
		assert.NoError(t, err)
	}
	assert.NoError(t, base.DB.Save(&submission).Error)

	failTests := []failTest{
		{
			name:   "NonExistingProblemSet",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunCompilerOutput", -1, submission.ID, submission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "NonExistingSubmission",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunCompilerOutput", problemSet.ID, -1, submission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "NonExistingRun",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunCompilerOutput", problemSet.ID, submission.ID, -1),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "PermissionDenied",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunCompilerOutput", problemSet.ID, submission.ID, submission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "")

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		httpResp := makeResp(makeReq(t, "GET",
			base.Echo.Reverse("problemSet.getRunCompilerOutput", problemSet.ID, submission.ID, submission.Runs[0].ID), nil, applyUser(user)))
		assert.Equal(t, http.StatusFound, httpResp.StatusCode)
		assert.Equal(t, "problem_set_get_run_compiler_output_0", getPresignedURLContent(t, httpResp.Header.Get("Location")))
	})
}

func TestProblemSetGetRunOutput(t *testing.T) {
	t.Parallel()

	user := createUserForTest(t, "problem_set_get_submission_run_output", 0)
	problem := createProblemForTest(t, "problem_set_get_submission_run_output", 0, nil, user)
	class := createClassForTest(t, "test_problem_set_get_submission_run_output", 0, nil, nil)
	problemSet := createProblemSetForTest(t, "problem_set_get_submission_run_output", 0, &class, []models.Problem{problem})
	submission := createSubmissionForTest(t, "problem_set_get_submission_run_output", 0, &problem, &user,
		newFileContent("output", "output_file_name",
			b64Encode("problem_set_get_submission_run_output_0")), 2)
	submission.ProblemSetId = problemSet.ID
	submission.Runs[0].Sample = true
	submission.Runs[1].Sample = false
	for i := range submission.Runs {
		submission.Runs[i].ProblemSetId = problemSet.ID
		assert.NoError(t, base.DB.Save(&submission.Runs[i]).Error)
		content := fmt.Sprintf("problem_set_get_submission_run_output_%d", i)
		var _, err = base.Storage.PutObject(context.Background(), "submissions",
			fmt.Sprintf("%d/run/%d/output", submission.ID, submission.Runs[i].ID),
			strings.NewReader(content), int64(len(content)), minio.PutObjectOptions{})
		assert.NoError(t, err)
	}
	assert.NoError(t, base.DB.Save(&submission).Error)

	failTests := []failTest{
		{
			name:   "NonExistingProblemSet",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunOutput", -1, submission.ID, submission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "NonExistingSubmission",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunOutput", problemSet.ID, -1, submission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "NonExistingRun",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunOutput", problemSet.ID, submission.ID, -1),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "PermissionDenied",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunOutput", problemSet.ID, submission.ID, submission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			name:   "NotSample",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunOutput", problemSet.ID, submission.ID, submission.Runs[1].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyUser(user),
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "")

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		httpResp := makeResp(makeReq(t, "GET",
			base.Echo.Reverse("problemSet.getRunOutput", problemSet.ID, submission.ID, submission.Runs[0].ID), nil, applyUser(user)))
		assert.Equal(t, http.StatusFound, httpResp.StatusCode)
		assert.Equal(t, "problem_set_get_submission_run_output_0", getPresignedURLContent(t, httpResp.Header.Get("Location")))
	})
}

func TestProblemSetGetRunInput(t *testing.T) {
	t.Parallel()

	user := createUserForTest(t, "problem_set_get_submission_run_input", 0)
	problem := createProblemForTest(t, "problem_set_get_submission_run_input", 0, nil, user)
	class := createClassForTest(t, "test_problem_set_get_submission_run_input", 0, nil, nil)
	problemSet := createProblemSetForTest(t, "problem_set_get_submission_run_input", 0, &class, []models.Problem{problem})
	submission := createSubmissionForTest(t, "problem_set_get_submission_run_input", 0, &problem, &user,
		newFileContent("input", "input_file_name",
			b64Encode("problem_set_get_submission_run_input_0")), 2)
	submission.ProblemSetId = problemSet.ID
	submission.Runs[0].Sample = true
	submission.Runs[1].Sample = false
	for i := range submission.Runs {
		submission.Runs[i].ProblemSetId = problemSet.ID
		assert.NoError(t, base.DB.Save(&submission.Runs[i]).Error)
		content := fmt.Sprintf("problem_set_get_submission_run_input_%d", i)
		var _, err = base.Storage.PutObject(context.Background(), "submissions",
			fmt.Sprintf("%d/run/%d/input", submission.ID, submission.Runs[i].ID),
			strings.NewReader(content), int64(len(content)), minio.PutObjectOptions{})
		assert.NoError(t, err)
	}
	assert.NoError(t, base.DB.Save(&submission).Error)

	failTests := []failTest{
		{
			name:   "NonExistingProblemSet",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunInput", -1, submission.ID, submission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "NonExistingSubmission",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunInput", problemSet.ID, -1, submission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "NonExistingRun",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunInput", problemSet.ID, submission.ID, -1),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "PermissionDenied",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunInput", problemSet.ID, submission.ID, submission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			name:   "NotSample",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunInput", problemSet.ID, submission.ID, submission.Runs[1].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyUser(user),
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "")

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		httpResp := makeResp(makeReq(t, "GET",
			base.Echo.Reverse("problemSet.getRunInput", problemSet.ID, submission.ID, submission.Runs[0].ID), nil, applyUser(user)))
		assert.Equal(t, http.StatusFound, httpResp.StatusCode)
		assert.Equal(t, "problem_set_get_submission_run_input_0", getPresignedURLContent(t, httpResp.Header.Get("Location")))
	})
}

func TestProblemSetGetRunComparerOutput(t *testing.T) {
	t.Parallel()

	user := createUserForTest(t, "problem_set_get_submission_run_comparer_output", 0)
	problem := createProblemForTest(t, "problem_set_get_submission_run_comparer_output", 0, nil, user)
	class := createClassForTest(t, "test_problem_set_get_submission_run_comparer_output", 0, nil, nil)
	problemSet := createProblemSetForTest(t, "problem_set_get_submission_run_comparer_output", 0, &class, []models.Problem{problem})
	submission := createSubmissionForTest(t, "problem_set_get_submission_run_comparer_output", 0, &problem, &user,
		newFileContent("comparer_output", "comparer_output_file_name",
			b64Encode("problem_set_get_submission_run_comparer_output_0")), 2)
	submission.ProblemSetId = problemSet.ID
	submission.Runs[0].Sample = true
	submission.Runs[1].Sample = false
	for i := range submission.Runs {
		submission.Runs[i].ProblemSetId = problemSet.ID
		assert.NoError(t, base.DB.Save(&submission.Runs[i]).Error)
		content := fmt.Sprintf("problem_set_get_submission_run_comparer_output_%d", i)
		var _, err = base.Storage.PutObject(context.Background(), "submissions",
			fmt.Sprintf("%d/run/%d/comparer_output", submission.ID, submission.Runs[i].ID),
			strings.NewReader(content), int64(len(content)), minio.PutObjectOptions{})
		assert.NoError(t, err)
	}
	assert.NoError(t, base.DB.Save(&submission).Error)

	failTests := []failTest{
		{
			name:   "NonExistingProblemSet",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunComparerOutput", -1, submission.ID, submission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "NonExistingSubmission",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunComparerOutput", problemSet.ID, -1, submission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "NonExistingRun",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunComparerOutput", problemSet.ID, submission.ID, -1),
			req:    nil,
			reqOptions: []reqOption{
				applyAdminUser,
			},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "PermissionDenied",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunComparerOutput", problemSet.ID, submission.ID, submission.Runs[0].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyNormalUser,
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
		{
			name:   "NotSample",
			method: "GET",
			path:   base.Echo.Reverse("problemSet.getRunComparerOutput", problemSet.ID, submission.ID, submission.Runs[1].ID),
			req:    nil,
			reqOptions: []reqOption{
				applyUser(user),
			},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "")

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		httpResp := makeResp(makeReq(t, "GET",
			base.Echo.Reverse("problemSet.getRunComparerOutput", problemSet.ID, submission.ID, submission.Runs[0].ID), nil, applyUser(user)))
		assert.Equal(t, http.StatusFound, httpResp.StatusCode)
		assert.Equal(t, "problem_set_get_submission_run_comparer_output_0", getPresignedURLContent(t, httpResp.Header.Get("Location")))
	})
}
