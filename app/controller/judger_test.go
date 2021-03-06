package controller_test

import (
	"bytes"
	"fmt"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/database"
	"github.com/EduOJ/backend/database/models"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestGetTask(t *testing.T) {
	// Not parallel
	t.Cleanup(database.SetupDatabaseForTest())
	initGeneralTestingUsers()

	user := createUserForTest(t, "get_task", 1)
	problem := createProblemForTest(t, "get_task", 1, nil, user)
	submission := createSubmissionForTest(t, "test_task", 1, &problem, &user, newFileContent(
		"", "code.test_language", b64Encode("balh"),
	), 1)
	var language models.Language
	var compareScript = models.Script{
		Name:     "test_get_task",
		Filename: "test",
	}
	assert.NoError(t, base.DB.Model(&problem).Association("CompareScript").Append(&compareScript))
	assert.NoError(t, base.DB.Model(&submission).Preload("RunScript").Preload("BuildScript").Association("Language").Find(&language))
	req := makeReq(t, "GET", base.Echo.Reverse("judger.getTask"), "", judgerAuthorize)
	httpResp := makeResp(req)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	buf := bytes.Buffer{}
	_, _ = buf.ReadFrom(httpResp.Body)
	var resp response.GetTaskResponse
	mustJsonDecode(buf.String(), &resp)
	t.Log(buf.String())
	jsonEQ(t, response.GetTaskResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			RunID             uint            `json:"run_id"`
			Language          models.Language `json:"language"`
			TestCaseID        uint            `json:"test_case_id"`
			InputFile         string          `json:"input_file"`
			OutputFile        string          `json:"output_file"`
			CodeFile          string          `json:"code_file"`
			TestCaseUpdatedAt time.Time       `json:"test_case_updated_at"`
			MemoryLimit       uint64          `json:"memory_limit"`
			TimeLimit         uint            `json:"time_limit"`
			BuildArg          string          `json:"build_arg"`
			CompareScript     models.Script   `json:"compare_script"`
		}{
			submission.Runs[0].ID,
			language,
			submission.Runs[0].TestCaseID,
			resp.Data.InputFile,
			resp.Data.OutputFile,
			resp.Data.CodeFile,
			problem.TestCases[0].UpdatedAt,
			problem.MemoryLimit,
			problem.TimeLimit,
			problem.BuildArg,
			compareScript,
		},
	}, resp)
	req = makeReq(t, "GET", base.Echo.Reverse("judger.getTask"), "", judgerAuthorize)
	httpResp = makeResp(req)
	assert.Equal(t, http.StatusNotFound, httpResp.StatusCode)
	jsonEQ(t, response.Response{
		Message: "NOT_FOUND",
		Error:   nil,
		Data:    nil,
	}, httpResp)
}

func TestUpdateRun(t *testing.T) {

	var compareScript = models.Script{
		Name:     "update_run",
		Filename: "test",
	}
	assert.NoError(t, base.DB.Create(&compareScript).Error)
	t.Run("SuccessWithNotAcceptedAnswers", func(t *testing.T) {
		t.Parallel()
		compareScript := compareScript
		user := createUserForTest(t, "update_run", 1)
		problem := createProblemForTest(t, "update_run", 1, nil, user)
		submission := createSubmissionForTest(t, "update_run", 1, &problem, &user, newFileContent(
			"", "code.test_language", b64Encode("balh"),
		), 3)
		var language models.Language
		assert.NoError(t, base.DB.Model(&problem).Association("CompareScript").Append(&compareScript))
		assert.NoError(t, base.DB.Model(&submission).Preload("RunScript").Preload("BuildScript").Association("Language").Find(&language))
		submission.Runs[0].Status = "JUDGING"
		submission.Runs[0].JudgerName = "test_judger"
		submission.Runs[1].Status = "JUDGING"
		submission.Runs[1].JudgerName = "test_judger"
		submission.Runs[2].Status = "JUDGING"
		submission.Runs[2].JudgerName = "test_judger"
		assert.NoError(t, base.DB.Save(&submission.Runs[0]).Error)
		assert.NoError(t, base.DB.Save(&submission.Runs[1]).Error)
		assert.NoError(t, base.DB.Save(&submission.Runs[2]).Error)
		runIDs := []uint{
			submission.Runs[0].ID,
			submission.Runs[1].ID,
			submission.Runs[2].ID,
		}
		_ = runIDs

		output := newFileContent("output_file", "c", b64Encode("output"))
		comparerOutput := newFileContent("comparer_output_file", "c", b64Encode("comparer_output"))
		compilerOutput := newFileContent("compiler_output_file", "c", b64Encode("compiler_output"))

		req := makeReq(t, "PUT", base.Echo.Reverse("judger.updateRun", submission.Runs[0].ID), []reqContent{
			output,
			comparerOutput,
			compilerOutput,
			&fieldContent{
				key:   "status",
				value: "ACCEPTED",
			},
			&fieldContent{
				key:   "memory_used",
				value: "123123",
			},
			&fieldContent{
				key:   "time_used",
				value: "1234",
			},
			&fieldContent{
				key:   "output_stripped_hash",
				value: "2333",
			},
		}, judgerAuthorize)
		httpResp := makeResp(req)
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		jsonEQ(t, response.Response{
			Message: "SUCCESS",
		}, httpResp)
		assert.NoError(t, base.DB.Preload("Runs").First(&submission, submission.ID).Error)
		assert.Equal(t, uint(123123), submission.Runs[0].MemoryUsed)
		assert.Equal(t, uint(1234), submission.Runs[0].TimeUsed)
		assert.Equal(t, "2333", submission.Runs[0].OutputStrippedHash)
		assert.Equal(t, "ACCEPTED", submission.Runs[0].Status)
		assert.Equal(t, true, submission.Runs[0].Judged)
		assert.Equal(t, uint(33), submission.Score)
		assert.Equal(t, false, submission.Judged)
		assert.Equal(t, "PENDING", submission.Status)

		req = makeReq(t, "PUT", base.Echo.Reverse("judger.updateRun", submission.Runs[1].ID), []reqContent{
			output,
			comparerOutput,
			compilerOutput,
			&fieldContent{
				key:   "status",
				value: "WRONG_ANSWER",
			},
			&fieldContent{
				key:   "memory_used",
				value: "1231235",
			},
			&fieldContent{
				key:   "time_used",
				value: "1234",
			},
			&fieldContent{
				key:   "output_stripped_hash",
				value: "2333",
			},
		}, judgerAuthorize)
		httpResp = makeResp(req)
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		jsonEQ(t, response.Response{
			Message: "SUCCESS",
		}, httpResp)
		assert.NoError(t, base.DB.Preload("Runs").First(&submission, submission.ID).Error)
		assert.Equal(t, uint(1231235), submission.Runs[1].MemoryUsed)
		assert.Equal(t, uint(1234), submission.Runs[1].TimeUsed)
		assert.Equal(t, "2333", submission.Runs[1].OutputStrippedHash)
		assert.Equal(t, "WRONG_ANSWER", submission.Runs[1].Status)
		assert.Equal(t, true, submission.Runs[1].Judged)
		assert.Equal(t, uint(33), submission.Score)
		assert.Equal(t, false, submission.Judged)
		assert.Equal(t, "PENDING", submission.Status)

		req = makeReq(t, "PUT", base.Echo.Reverse("judger.updateRun", submission.Runs[2].ID), []reqContent{
			output,
			comparerOutput,
			compilerOutput,
			&fieldContent{
				key:   "status",
				value: "ACCEPTED",
			},
			&fieldContent{
				key:   "memory_used",
				value: "1231233",
			},
			&fieldContent{
				key:   "time_used",
				value: "1234",
			},
			&fieldContent{
				key:   "output_stripped_hash",
				value: "2333",
			},
		}, judgerAuthorize)
		httpResp = makeResp(req)
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		jsonEQ(t, response.Response{
			Message: "SUCCESS",
		}, httpResp)
		assert.NoError(t, base.DB.Preload("Runs").First(&submission, submission.ID).Error)
		assert.Equal(t, uint(1231233), submission.Runs[2].MemoryUsed)
		assert.Equal(t, uint(1234), submission.Runs[2].TimeUsed)
		assert.Equal(t, "2333", submission.Runs[2].OutputStrippedHash)
		assert.Equal(t, "ACCEPTED", submission.Runs[2].Status)
		assert.Equal(t, true, submission.Runs[2].Judged)
		assert.Equal(t, uint(66), submission.Score)
		assert.Equal(t, true, submission.Judged)
		assert.Equal(t, "WRONG_ANSWER", submission.Status)

		storageContent := getObjectContent(t, "submissions", fmt.Sprintf("%d/run/%d/output", submission.ID, submission.Runs[0].ID))
		assert.Equal(t, []byte("output"), storageContent)
		storageContent = getObjectContent(t, "submissions", fmt.Sprintf("%d/run/%d/comparer_output", submission.ID, submission.Runs[0].ID))
		assert.Equal(t, []byte("comparer_output"), storageContent)
		storageContent = getObjectContent(t, "submissions", fmt.Sprintf("%d/run/%d/compiler_output", submission.ID, submission.Runs[0].ID))
		assert.Equal(t, []byte("compiler_output"), storageContent)

	})

	t.Run("SuccessAccepted", func(t *testing.T) {
		t.Parallel()
		compareScript := compareScript
		user := createUserForTest(t, "update_run", 2)
		problem := createProblemForTest(t, "update_run", 2, nil, user)
		submission := createSubmissionForTest(t, "update_run", 2, &problem, &user, newFileContent(
			"", "code.test_language", b64Encode("balh"),
		), 3)
		var language models.Language
		assert.NoError(t, base.DB.Model(&problem).Association("CompareScript").Append(&compareScript))
		assert.NoError(t, base.DB.Model(&submission).Preload("RunScript").Preload("BuildScript").Association("Language").Find(&language))
		submission.Runs[0].Status = "JUDGING"
		submission.Runs[0].JudgerName = "test_judger"
		submission.Runs[1].Status = "JUDGING"
		submission.Runs[1].JudgerName = "test_judger"
		submission.Runs[2].Status = "JUDGING"
		submission.Runs[2].JudgerName = "test_judger"
		assert.NoError(t, base.DB.Save(&submission.Runs[0]).Error)
		assert.NoError(t, base.DB.Save(&submission.Runs[1]).Error)
		assert.NoError(t, base.DB.Save(&submission.Runs[2]).Error)
		runIDs := []uint{
			submission.Runs[0].ID,
			submission.Runs[1].ID,
			submission.Runs[2].ID,
		}
		_ = runIDs

		output := newFileContent("output_file", "c", b64Encode("output"))
		comparerOutput := newFileContent("comparer_output_file", "c", b64Encode("comparer_output"))
		compilerOutput := newFileContent("compiler_output_file", "c", b64Encode("compiler_output"))

		req := makeReq(t, "PUT", base.Echo.Reverse("judger.updateRun", submission.Runs[0].ID), []reqContent{
			output,
			comparerOutput,
			compilerOutput,
			&fieldContent{
				key:   "status",
				value: "ACCEPTED",
			},
			&fieldContent{
				key:   "memory_used",
				value: "123123",
			},
			&fieldContent{
				key:   "time_used",
				value: "1234",
			},
			&fieldContent{
				key:   "output_stripped_hash",
				value: "2333",
			},
		}, judgerAuthorize)
		httpResp := makeResp(req)
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		jsonEQ(t, response.Response{
			Message: "SUCCESS",
		}, httpResp)
		assert.NoError(t, base.DB.Preload("Runs").First(&submission, submission.ID).Error)
		assert.Equal(t, uint(123123), submission.Runs[0].MemoryUsed)
		assert.Equal(t, uint(1234), submission.Runs[0].TimeUsed)
		assert.Equal(t, "2333", submission.Runs[0].OutputStrippedHash)
		assert.Equal(t, "ACCEPTED", submission.Runs[0].Status)
		assert.Equal(t, true, submission.Runs[0].Judged)
		assert.Equal(t, uint(33), submission.Score)
		assert.Equal(t, false, submission.Judged)
		assert.Equal(t, "PENDING", submission.Status)

		req = makeReq(t, "PUT", base.Echo.Reverse("judger.updateRun", submission.Runs[1].ID), []reqContent{
			output,
			comparerOutput,
			compilerOutput,
			&fieldContent{
				key:   "status",
				value: "ACCEPTED",
			},
			&fieldContent{
				key:   "memory_used",
				value: "1231235",
			},
			&fieldContent{
				key:   "time_used",
				value: "1234",
			},
			&fieldContent{
				key:   "output_stripped_hash",
				value: "2333",
			},
		}, judgerAuthorize)
		httpResp = makeResp(req)
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		jsonEQ(t, response.Response{
			Message: "SUCCESS",
		}, httpResp)
		assert.NoError(t, base.DB.Preload("Runs").First(&submission, submission.ID).Error)
		assert.Equal(t, uint(1231235), submission.Runs[1].MemoryUsed)
		assert.Equal(t, uint(1234), submission.Runs[1].TimeUsed)
		assert.Equal(t, "2333", submission.Runs[1].OutputStrippedHash)
		assert.Equal(t, "ACCEPTED", submission.Runs[1].Status)
		assert.Equal(t, true, submission.Runs[1].Judged)
		assert.Equal(t, uint(66), submission.Score)
		assert.Equal(t, false, submission.Judged)
		assert.Equal(t, "PENDING", submission.Status)

		req = makeReq(t, "PUT", base.Echo.Reverse("judger.updateRun", submission.Runs[2].ID), []reqContent{
			output,
			comparerOutput,
			compilerOutput,
			&fieldContent{
				key:   "status",
				value: "ACCEPTED",
			},
			&fieldContent{
				key:   "memory_used",
				value: "1231233",
			},
			&fieldContent{
				key:   "time_used",
				value: "1234",
			},
			&fieldContent{
				key:   "output_stripped_hash",
				value: "2333",
			},
		}, judgerAuthorize)
		httpResp = makeResp(req)
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		jsonEQ(t, response.Response{
			Message: "SUCCESS",
		}, httpResp)
		assert.NoError(t, base.DB.Preload("Runs").First(&submission, submission.ID).Error)
		assert.Equal(t, uint(1231233), submission.Runs[2].MemoryUsed)
		assert.Equal(t, uint(1234), submission.Runs[2].TimeUsed)
		assert.Equal(t, "2333", submission.Runs[2].OutputStrippedHash)
		assert.Equal(t, "ACCEPTED", submission.Runs[2].Status)
		assert.Equal(t, true, submission.Runs[2].Judged)
		assert.Equal(t, uint(100), submission.Score)
		assert.Equal(t, true, submission.Judged)
		assert.Equal(t, "ACCEPTED", submission.Status)

		httpResp = makeResp(req)
		assert.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
		jsonEQ(t, response.Response{
			Message: "ALREADY_SUBMITTED",
		}, httpResp)
	})

	t.Run("SuccessNotDefaultScore", func(t *testing.T) {
		t.Parallel()
		compareScript := compareScript
		user := createUserForTest(t, "update_run", 3)
		problem := createProblemForTest(t, "update_run", 3, nil, user)
		submission := createSubmissionForTest(t, "update_run", 3, &problem, &user, newFileContent(
			"", "code.test_language", b64Encode("balh"),
		), 3)
		var language models.Language
		assert.NoError(t, base.DB.Model(&problem).Association("CompareScript").Append(&compareScript))
		assert.NoError(t, base.DB.Model(&submission).Preload("RunScript").Preload("BuildScript").Association("Language").Find(&language))
		submission.Runs[0].Status = "JUDGING"
		submission.Runs[0].JudgerName = "test_judger"
		submission.Runs[1].Status = "JUDGING"
		submission.Runs[1].JudgerName = "test_judger"
		submission.Runs[2].Status = "JUDGING"
		submission.Runs[2].JudgerName = "test_judger"
		problem.TestCases[0].Score = 1
		problem.TestCases[1].Score = 2
		problem.TestCases[2].Score = 3

		assert.NoError(t, base.DB.Save(&submission.Runs[0]).Error)
		assert.NoError(t, base.DB.Save(&submission.Runs[1]).Error)
		assert.NoError(t, base.DB.Save(&submission.Runs[2]).Error)
		assert.NoError(t, base.DB.Save(&problem.TestCases[0]).Error)
		assert.NoError(t, base.DB.Save(&problem.TestCases[1]).Error)
		assert.NoError(t, base.DB.Save(&problem.TestCases[2]).Error)
		runIDs := []uint{
			submission.Runs[0].ID,
			submission.Runs[1].ID,
			submission.Runs[2].ID,
		}
		_ = runIDs

		output := newFileContent("output_file", "c", b64Encode("output"))
		comparerOutput := newFileContent("comparer_output_file", "c", b64Encode("comparer_output"))
		compilerOutput := newFileContent("compiler_output_file", "c", b64Encode("compiler_output"))

		req := makeReq(t, "PUT", base.Echo.Reverse("judger.updateRun", submission.Runs[0].ID), []reqContent{
			output,
			comparerOutput,
			compilerOutput,
			&fieldContent{
				key:   "status",
				value: "ACCEPTED",
			},
			&fieldContent{
				key:   "memory_used",
				value: "123123",
			},
			&fieldContent{
				key:   "time_used",
				value: "1234",
			},
			&fieldContent{
				key:   "output_stripped_hash",
				value: "2333",
			},
		}, judgerAuthorize)
		httpResp := makeResp(req)
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		jsonEQ(t, response.Response{
			Message: "SUCCESS",
		}, httpResp)
		assert.NoError(t, base.DB.Preload("Runs").First(&submission, submission.ID).Error)
		assert.Equal(t, uint(123123), submission.Runs[0].MemoryUsed)
		assert.Equal(t, uint(1234), submission.Runs[0].TimeUsed)
		assert.Equal(t, "2333", submission.Runs[0].OutputStrippedHash)
		assert.Equal(t, "ACCEPTED", submission.Runs[0].Status)
		assert.Equal(t, true, submission.Runs[0].Judged)
		assert.Equal(t, uint(1), submission.Score)
		assert.Equal(t, false, submission.Judged)
		assert.Equal(t, "PENDING", submission.Status)

		req = makeReq(t, "PUT", base.Echo.Reverse("judger.updateRun", submission.Runs[1].ID), []reqContent{
			output,
			comparerOutput,
			compilerOutput,
			&fieldContent{
				key:   "status",
				value: "ACCEPTED",
			},
			&fieldContent{
				key:   "memory_used",
				value: "1231235",
			},
			&fieldContent{
				key:   "time_used",
				value: "1234",
			},
			&fieldContent{
				key:   "output_stripped_hash",
				value: "2333",
			},
		}, judgerAuthorize)
		httpResp = makeResp(req)
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		jsonEQ(t, response.Response{
			Message: "SUCCESS",
		}, httpResp)
		assert.NoError(t, base.DB.Preload("Runs").First(&submission, submission.ID).Error)
		assert.Equal(t, uint(1231235), submission.Runs[1].MemoryUsed)
		assert.Equal(t, uint(1234), submission.Runs[1].TimeUsed)
		assert.Equal(t, "2333", submission.Runs[1].OutputStrippedHash)
		assert.Equal(t, "ACCEPTED", submission.Runs[1].Status)
		assert.Equal(t, true, submission.Runs[1].Judged)
		assert.Equal(t, uint(3), submission.Score)
		assert.Equal(t, false, submission.Judged)
		assert.Equal(t, "PENDING", submission.Status)

		req = makeReq(t, "PUT", base.Echo.Reverse("judger.updateRun", submission.Runs[2].ID), []reqContent{
			output,
			comparerOutput,
			compilerOutput,
			&fieldContent{
				key:   "status",
				value: "ACCEPTED",
			},
			&fieldContent{
				key:   "memory_used",
				value: "1231233",
			},
			&fieldContent{
				key:   "time_used",
				value: "1234",
			},
			&fieldContent{
				key:   "output_stripped_hash",
				value: "2333",
			},
		}, judgerAuthorize)
		httpResp = makeResp(req)
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		jsonEQ(t, response.Response{
			Message: "SUCCESS",
		}, httpResp)
		assert.NoError(t, base.DB.Preload("Runs").First(&submission, submission.ID).Error)
		assert.Equal(t, uint(1231233), submission.Runs[2].MemoryUsed)
		assert.Equal(t, uint(1234), submission.Runs[2].TimeUsed)
		assert.Equal(t, "2333", submission.Runs[2].OutputStrippedHash)
		assert.Equal(t, "ACCEPTED", submission.Runs[2].Status)
		assert.Equal(t, true, submission.Runs[2].Judged)
		assert.Equal(t, uint(6), submission.Score)
		assert.Equal(t, true, submission.Judged)
		assert.Equal(t, "ACCEPTED", submission.Status)

		httpResp = makeResp(req)
		assert.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
		jsonEQ(t, response.Response{
			Message: "ALREADY_SUBMITTED",
		}, httpResp)
	})

	t.Run("Fail", func(t *testing.T) {
		t.Parallel()
		compareScript := compareScript
		user := createUserForTest(t, "update_run", 4)
		problem := createProblemForTest(t, "update_run", 4, nil, user)
		submission := createSubmissionForTest(t, "update_run", 4, &problem, &user, newFileContent(
			"", "code.test_language", b64Encode("balh"),
		), 3)
		var language models.Language
		assert.NoError(t, base.DB.Model(&problem).Association("CompareScript").Append(&compareScript))
		assert.NoError(t, base.DB.Model(&submission).Preload("RunScript").Preload("BuildScript").Association("Language").Find(&language))
		submission.Runs[0].Status = "JUDGING"
		submission.Runs[1].Status = "JUDGING"
		submission.Runs[2].Status = "JUDGING"
		submission.Runs[2].JudgerName = "test_judger"
		problem.TestCases[0].Score = 1
		problem.TestCases[1].Score = 2
		problem.TestCases[2].Score = 3

		assert.NoError(t, base.DB.Save(&submission.Runs[0]).Error)
		assert.NoError(t, base.DB.Save(&submission.Runs[1]).Error)
		assert.NoError(t, base.DB.Save(&submission.Runs[2]).Error)
		assert.NoError(t, base.DB.Save(&problem.TestCases[0]).Error)
		assert.NoError(t, base.DB.Save(&problem.TestCases[1]).Error)
		assert.NoError(t, base.DB.Save(&problem.TestCases[2]).Error)
		runIDs := []uint{
			submission.Runs[0].ID,
			submission.Runs[1].ID,
			submission.Runs[2].ID,
		}
		_ = runIDs

		output := newFileContent("output_file", "c", b64Encode("output"))
		comparerOutput := newFileContent("comparer_output_file", "c", b64Encode("comparer_output"))
		compilerOutput := newFileContent("compiler_output_file", "c", b64Encode("compiler_output"))

		req := makeReq(t, "PUT", base.Echo.Reverse("judger.updateRun", 2147483647), []reqContent{
			&fieldContent{
				key:   "status",
				value: "ACCEPTED",
			},
			&fieldContent{
				key:   "memory_used",
				value: "1231233",
			},
			&fieldContent{
				key:   "time_used",
				value: "1234",
			},
			&fieldContent{
				key:   "output_stripped_hash",
				value: "2333",
			},
		}, judgerAuthorize)
		httpResp := makeResp(req)
		assert.Equal(t, http.StatusNotFound, httpResp.StatusCode)
		jsonEQ(t, response.ErrorResp("NOT_FOUND", nil), httpResp)

		req = makeReq(t, "PUT", base.Echo.Reverse("judger.updateRun", submission.Runs[0].ID), []reqContent{
			output,
			comparerOutput,
			compilerOutput,
			&fieldContent{
				key:   "status",
				value: "ACCEPTED",
			},
			&fieldContent{
				key:   "memory_used",
				value: "1231233",
			},
			&fieldContent{
				key:   "time_used",
				value: "1234",
			},
			&fieldContent{
				key:   "output_stripped_hash",
				value: "2333",
			},
		}, judgerAuthorize)
		httpResp = makeResp(req)
		assert.Equal(t, http.StatusForbidden, httpResp.StatusCode)
		jsonEQ(t, response.ErrorResp("WRONG_RUN_ID", nil), httpResp)

		req = makeReq(t, "PUT", base.Echo.Reverse("judger.updateRun", submission.Runs[2].ID), []reqContent{
			comparerOutput,
			compilerOutput,
			&fieldContent{
				key:   "status",
				value: "ACCEPTED",
			},
			&fieldContent{
				key:   "memory_used",
				value: "1231233",
			},
			&fieldContent{
				key:   "time_used",
				value: "1234",
			},
			&fieldContent{
				key:   "output_stripped_hash",
				value: "2333",
			},
		}, judgerAuthorize)
		httpResp = makeResp(req)
		assert.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
		jsonEQ(t, response.ErrorResp("MISSING_OUTPUT", nil), httpResp)

		req = makeReq(t, "PUT", base.Echo.Reverse("judger.updateRun", submission.Runs[2].ID), []reqContent{
			output,
			compilerOutput,
			&fieldContent{
				key:   "status",
				value: "ACCEPTED",
			},
			&fieldContent{
				key:   "memory_used",
				value: "1231233",
			},
			&fieldContent{
				key:   "time_used",
				value: "1234",
			},
			&fieldContent{
				key:   "output_stripped_hash",
				value: "2333",
			},
		}, judgerAuthorize)
		httpResp = makeResp(req)
		assert.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
		jsonEQ(t, response.ErrorResp("MISSING_COMPARER_OUTPUT", nil), httpResp)

		req = makeReq(t, "PUT", base.Echo.Reverse("judger.updateRun", submission.Runs[2].ID), []reqContent{
			output,
			comparerOutput,
			&fieldContent{
				key:   "status",
				value: "ACCEPTED",
			},
			&fieldContent{
				key:   "memory_used",
				value: "1231233",
			},
			&fieldContent{
				key:   "time_used",
				value: "1234",
			},
			&fieldContent{
				key:   "output_stripped_hash",
				value: "2333",
			},
		}, judgerAuthorize)
		httpResp = makeResp(req)
		assert.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
		jsonEQ(t, response.ErrorResp("MISSING_COMPILER_OUTPUT", nil), httpResp)
	})

}
