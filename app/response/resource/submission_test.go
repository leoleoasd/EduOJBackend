package resource_test

import (
	"fmt"
	"github.com/EduOJ/backend/app/response/resource"
	"github.com/EduOJ/backend/database/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func createRunForTest(name string, submissionId uint, id uint) models.Run {
	return models.Run{
		ID:                 id,
		UserID:             id,
		User:               nil,
		ProblemID:          id,
		Problem:            nil,
		ProblemSetId:       id,
		TestCaseID:         id,
		TestCase:           nil,
		Sample:             true,
		SubmissionID:       submissionId,
		Submission:         nil,
		Priority:           127,
		Judged:             true,
		Status:             fmt.Sprintf("test_%s_submission_%d_run_%d_status", name, submissionId, id),
		MemoryUsed:         1024,
		TimeUsed:           1000,
		OutputStrippedHash: fmt.Sprintf("test_%s_submission_%d_run_%d_output_stripped_hash", name, submissionId, id),
		CreatedAt:          time.Date(int(id), 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),
		UpdatedAt:          time.Date(int(id), 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
	}
}

func createSubmissionForTest(name string, id uint, runCount uint) (submission models.Submission) {
	submission = models.Submission{
		ID:           id,
		UserID:       id,
		User:         nil,
		ProblemID:    id,
		Problem:      nil,
		ProblemSetId: id,
		LanguageName: fmt.Sprintf("test_%s_submission_%d_language", name, id),
		FileName:     fmt.Sprintf("test_%s_submission_%d_file_name", name, id),
		Priority:     127,
		Judged:       false,
		Score:        id,
		Status:       fmt.Sprintf("test_%s_submission_%d_status", name, id),
		Runs:         make([]models.Run, runCount),
		CreatedAt:    time.Date(int(id), 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),
		UpdatedAt:    time.Date(int(id), 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
	}
	for i := range submission.Runs {
		submission.Runs[i] = createRunForTest(name, id, uint(i))
	}
	return
}

func TestGetRunAndGetRunSlice(t *testing.T) {
	run1 := createRunForTest("get_run", 0, 1)
	run2 := createRunForTest("get_run", 2, 3)
	t.Run("testGetRun", func(t *testing.T) {
		actualRun := resource.GetRun(&run1)
		expectedRun := resource.Run{
			ID:           1,
			UserID:       1,
			ProblemID:    1,
			ProblemSetId: 1,
			TestCaseID:   1,
			Sample:       true,
			SubmissionID: 0,
			Priority:     127,
			Judged:       true,
			Status:       "test_get_run_submission_0_run_1_status",
			MemoryUsed:   1024,
			TimeUsed:     1000,
			CreatedAt:    time.Date(1, 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),
			UpdatedAt:    time.Date(1, 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
		}
		assert.Equal(t, &expectedRun, actualRun)
	})
	t.Run("testGetRunSlice", func(t *testing.T) {
		actualRunSlice := resource.GetRunSlice([]models.Run{run1, run2})
		expectedRunSlice := []resource.Run{
			{
				ID:           1,
				UserID:       1,
				ProblemID:    1,
				ProblemSetId: 1,
				TestCaseID:   1,
				Sample:       true,
				SubmissionID: 0,
				Priority:     127,
				Judged:       true,
				Status:       "test_get_run_submission_0_run_1_status",
				MemoryUsed:   1024,
				TimeUsed:     1000,
				CreatedAt:    time.Date(1, 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),
				UpdatedAt:    time.Date(1, 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
			}, {
				ID:           3,
				UserID:       3,
				ProblemID:    3,
				ProblemSetId: 3,
				TestCaseID:   3,
				Sample:       true,
				SubmissionID: 2,
				Priority:     127,
				Judged:       true,
				Status:       "test_get_run_submission_2_run_3_status",
				MemoryUsed:   1024,
				TimeUsed:     1000,
				CreatedAt:    time.Date(3, 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),
				UpdatedAt:    time.Date(3, 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
			},
		}
		assert.Equal(t, expectedRunSlice, actualRunSlice)
	})
}

func TestGetSubmissionAndGetSubmissionDetail(t *testing.T) {
	user := createUserForTest("get_submission", 1)
	problem := createProblemForTest("get_submission", 1, 2)
	submission := createSubmissionForTest("get_submission", 1, 2)
	submission.User = &user
	submission.Problem = &problem
	t.Run("testGetSubmission", func(t *testing.T) {
		actualSubmission := resource.GetSubmission(&submission)
		expectedSubmission := resource.Submission{
			ID:           1,
			UserID:       1,
			User:         resource.GetUser(&user),
			ProblemID:    1,
			ProblemName:  "test_get_submission_problem_1",
			ProblemSetId: 1,
			Language:     "test_get_submission_submission_1_language",
			Judged:       false,
			Score:        1,
			Status:       "test_get_submission_submission_1_status",
			CreatedAt:    time.Date(1, 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),

			UpdatedAt: time.Date(1, 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
		}
		assert.Equal(t, expectedSubmission, *actualSubmission)
	})
	t.Run("testGetSubmissionDetail", func(t *testing.T) {
		actualSubmission := resource.GetSubmissionDetail(&submission)
		expectedSubmission := resource.SubmissionDetail{
			ID:           1,
			UserID:       1,
			User:         resource.GetUser(&user),
			ProblemID:    1,
			ProblemName:  "test_get_submission_problem_1",
			ProblemSetId: 1,
			Language:     "test_get_submission_submission_1_language",
			FileName:     "test_get_submission_submission_1_file_name",
			Priority:     127,
			Judged:       false,
			Score:        1,
			Status:       "test_get_submission_submission_1_status",
			Runs: []resource.Run{
				{
					ID:           0,
					UserID:       0,
					ProblemID:    0,
					ProblemSetId: 0,
					TestCaseID:   0,
					Sample:       true,
					SubmissionID: 1,
					Priority:     127,
					Judged:       true,
					Status:       "test_get_submission_submission_1_run_0_status",
					MemoryUsed:   1024,
					TimeUsed:     1000,
					CreatedAt:    time.Date(0, 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),
					UpdatedAt:    time.Date(0, 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
				},
				{
					ID:           1,
					UserID:       1,
					ProblemID:    1,
					ProblemSetId: 1,
					TestCaseID:   1,
					Sample:       true,
					SubmissionID: 1,
					Priority:     127,
					Judged:       true,
					Status:       "test_get_submission_submission_1_run_1_status",
					MemoryUsed:   1024,
					TimeUsed:     1000,
					CreatedAt:    time.Date(1, 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),
					UpdatedAt:    time.Date(1, 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
				},
			},
			CreatedAt: time.Date(1, 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),
			UpdatedAt: time.Date(1, 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
		}
		assert.Equal(t, &expectedSubmission, actualSubmission)
	})
}

func TestGetSubmissionSliceAndGetSubmissionDetailSlice(t *testing.T) {
	user1 := createUserForTest("get_submission", 1)
	problem1 := createProblemForTest("get_submission", 1, 2)
	submission1 := createSubmissionForTest("get_submission", 1, 2)
	submission1.User = &user1
	submission1.Problem = &problem1
	user2 := createUserForTest("get_submission", 2)
	problem2 := createProblemForTest("get_submission", 2, 1)
	submission2 := createSubmissionForTest("get_submission", 2, 1)
	submission2.User = &user2
	submission2.Problem = &problem2
	t.Run("testGetSubmissionSlice", func(t *testing.T) {
		actualSubmissionSlice := resource.GetSubmissionSlice([]models.Submission{submission1, submission2})
		expectedSubmissionSlice := []resource.Submission{
			{
				ID:           1,
				UserID:       1,
				User:         resource.GetUser(&user1),
				ProblemID:    1,
				ProblemName:  "test_get_submission_problem_1",
				ProblemSetId: 1,
				Language:     "test_get_submission_submission_1_language",
				Judged:       false,
				Score:        1,
				Status:       "test_get_submission_submission_1_status",
				CreatedAt:    time.Date(1, 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),

				UpdatedAt: time.Date(1, 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
			}, {
				ID:           2,
				UserID:       2,
				User:         resource.GetUser(&user2),
				ProblemID:    2,
				ProblemName:  "test_get_submission_problem_2",
				ProblemSetId: 2,
				Language:     "test_get_submission_submission_2_language",
				Judged:       false,
				Score:        2,
				Status:       "test_get_submission_submission_2_status",
				CreatedAt:    time.Date(2, 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),
				UpdatedAt:    time.Date(2, 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
			},
		}
		assert.Equal(t, expectedSubmissionSlice, actualSubmissionSlice)
	})
	t.Run("testGetSubmissionDetailSlice", func(t *testing.T) {
		actualSubmissionSlice := resource.GetSubmissionDetailSlice([]models.Submission{submission1, submission2})
		expectedSubmissionSlice := []resource.SubmissionDetail{
			{
				ID:           1,
				UserID:       1,
				User:         resource.GetUser(&user1),
				ProblemID:    1,
				ProblemName:  "test_get_submission_problem_1",
				ProblemSetId: 1,
				Language:     "test_get_submission_submission_1_language",
				FileName:     "test_get_submission_submission_1_file_name",
				Priority:     127,
				Judged:       false,
				Score:        1,
				Status:       "test_get_submission_submission_1_status",
				Runs: []resource.Run{
					{
						ID:           0,
						UserID:       0,
						ProblemID:    0,
						ProblemSetId: 0,
						TestCaseID:   0,
						Sample:       true,
						SubmissionID: 1,
						Priority:     127,
						Judged:       true,
						Status:       "test_get_submission_submission_1_run_0_status",
						MemoryUsed:   1024,
						TimeUsed:     1000,
						CreatedAt:    time.Date(0, 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),
						UpdatedAt:    time.Date(0, 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
					},
					{
						ID:           1,
						UserID:       1,
						ProblemID:    1,
						ProblemSetId: 1,
						TestCaseID:   1,
						Sample:       true,
						SubmissionID: 1,
						Priority:     127,
						Judged:       true,
						Status:       "test_get_submission_submission_1_run_1_status",
						MemoryUsed:   1024,
						TimeUsed:     1000,
						CreatedAt:    time.Date(1, 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),
						UpdatedAt:    time.Date(1, 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
					},
				},
				CreatedAt: time.Date(1, 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),
				UpdatedAt: time.Date(1, 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
			}, {
				ID:           2,
				UserID:       2,
				User:         resource.GetUser(&user2),
				ProblemID:    2,
				ProblemName:  "test_get_submission_problem_2",
				ProblemSetId: 2,
				Language:     "test_get_submission_submission_2_language",
				FileName:     "test_get_submission_submission_2_file_name",
				Priority:     127,
				Judged:       false,
				Score:        2,
				Status:       "test_get_submission_submission_2_status",
				Runs: []resource.Run{
					{
						ID:           0,
						UserID:       0,
						ProblemID:    0,
						ProblemSetId: 0,
						TestCaseID:   0,
						Sample:       true,
						SubmissionID: 2,
						Priority:     127,
						Judged:       true,
						Status:       "test_get_submission_submission_2_run_0_status",
						MemoryUsed:   1024,
						TimeUsed:     1000,
						CreatedAt:    time.Date(0, 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),
						UpdatedAt:    time.Date(0, 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
					},
				},
				CreatedAt: time.Date(2, 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),
				UpdatedAt: time.Date(2, 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
			},
		}
		assert.Equal(t, expectedSubmissionSlice, actualSubmissionSlice)
	})
}
