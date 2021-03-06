package controller

import (
	"context"
	"fmt"
	"github.com/EduOJ/backend/app/request"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/event"
	"github.com/EduOJ/backend/base/utils"
	"github.com/EduOJ/backend/database/models"
	runEvent "github.com/EduOJ/backend/event/run"
	submissionEvent "github.com/EduOJ/backend/event/submission"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"net/http"
	"sync"
	"time"
)

var taskLock sync.Mutex
var runLock sync.Mutex

// remember to lock taskLock when using this function
func getRun() *models.Run {
	run := models.Run{}
	err := base.DB.Order("priority desc").
		Order("id asc").
		Preload("Problem.CompareScript").
		Preload("TestCase").
		Preload("Submission.Language.RunScript").
		Preload("Submission.Language.BuildScript").
		First(&run, "status = 'PENDING'").Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		} else {
			panic(errors.Wrap(err, "could not query run"))
		}
	}
	return &run
}

func generateResponse(run *models.Run) response.GetTaskResponse {
	inputUrl, err := utils.GetPresignedURL("problems", fmt.Sprintf("%d/input/%d.in", run.Problem.ID, run.TestCase.ID), run.TestCase.InputFileName)
	if err != nil {
		panic(errors.Wrap(err, "could not get problem input file"))
	}
	outputUrl, err := utils.GetPresignedURL("problems", fmt.Sprintf("%d/output/%d.out", run.Problem.ID, run.TestCase.ID), run.TestCase.OutputFileName)
	if err != nil {
		panic(errors.Wrap(err, "could not get problem output file"))
	}
	codeUrl, err := utils.GetPresignedURL("submissions", fmt.Sprintf("%d/code", run.Submission.ID), run.Submission.FileName)
	if err != nil {
		panic(errors.Wrap(err, "could not get problem code file"))
	}
	return response.GetTaskResponse{
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
			RunID:             run.ID,
			Language:          *run.Submission.Language,
			TestCaseID:        run.TestCaseID,
			InputFile:         inputUrl,
			OutputFile:        outputUrl,
			CodeFile:          codeUrl,
			TestCaseUpdatedAt: run.TestCase.UpdatedAt,
			MemoryLimit:       run.Problem.MemoryLimit,
			TimeLimit:         run.Problem.TimeLimit,
			BuildArg:          run.Problem.BuildArg,
			CompareScript:     run.Problem.CompareScript,
		},
	}
}

func GetTask(c echo.Context) error {
	var poll bool
	var err error
	var run *models.Run
	if c.QueryParam("poll") == "1" {
		poll = true
	}

	taskLock.Lock()
	run = getRun()
	if run == nil {
		taskLock.Unlock()
		if poll {
			goto poll
		}
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	}
	run.Status = "JUDGING"
	run.JudgerName = c.Request().Header.Get("Judger-Name")
	err = base.DB.Save(&run).Error
	taskLock.Unlock()
	if err != nil {
		panic(errors.Wrap(err, "could not update run"))
	}
	return c.JSON(http.StatusOK, generateResponse(run))

poll:
	timeoutChan := time.After(viper.GetDuration("polling_timeout"))
	timeout := false
	sub := base.Redis.Subscribe(c.Request().Context(), "runs")
	defer sub.Close()
	for {
		select {
		case <-sub.Channel():
			taskLock.Lock()
			run = getRun()
			if run == nil {
				taskLock.Unlock()
				break
			}
			run.Status = "JUDGING"
			run.JudgerName = c.Request().Header.Get("Judger-Name")
			err := base.DB.Save(&run).Error
			taskLock.Unlock()
			if err != nil {
				panic(errors.Wrap(err, "could not update run"))
			}
			break
		case <-c.Request().Context().Done():
			// context cancelled
			return nil
		case <-timeoutChan:
			timeout = true
		}
		if timeout {
			break
		}
		if run != nil {
			break
		}
	}
	if run == nil {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	}
	return c.JSON(http.StatusOK, generateResponse(run))
}

func UpdateRun(c echo.Context) error {
	runLock.Lock()
	unlocked := false
	defer func() {
		if !unlocked {
			runLock.Unlock()
		}
	}()
	run := models.Run{}
	err := base.DB.Preload("TestCase").Preload("Submission").First(&run, c.Param("id")).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		}
		panic(errors.Wrap(err, "could not query run"))
	}
	if run.JudgerName != c.Request().Header.Get("Judger-Name") {
		return c.JSON(http.StatusForbidden, response.ErrorResp("WRONG_RUN_ID", nil))
	}
	if run.Judged {
		return c.JSON(http.StatusBadRequest, response.ErrorResp("ALREADY_SUBMITTED", nil))
	}
	req := request.UpdateRunRequest{}
	if err, ok := utils.BindAndValidate(&req, c); !ok {
		return err
	}
	compiler, err := c.FormFile("compiler_output_file")
	if err != nil {
		if err != http.ErrMissingFile && err.Error() != "request Content-Type isn't multipart/form-data" {
			panic(errors.Wrap(err, "could not read input file"))
		}
		return c.JSON(http.StatusBadRequest, response.ErrorResp("MISSING_COMPILER_OUTPUT", nil))
	}
	comparer, err := c.FormFile("comparer_output_file")
	if err != nil {
		if err != http.ErrMissingFile && err.Error() != "request Content-Type isn't multipart/form-data" {
			panic(errors.Wrap(err, "could not read input file"))
		}
		return c.JSON(http.StatusBadRequest, response.ErrorResp("MISSING_COMPARER_OUTPUT", nil))
	}
	output, err := c.FormFile("output_file")
	if err != nil {
		if err != http.ErrMissingFile && err.Error() != "request Content-Type isn't multipart/form-data" {
			panic(errors.Wrap(err, "could not read input file"))
		}
		return c.JSON(http.StatusBadRequest, response.ErrorResp("MISSING_OUTPUT", nil))
	}
	testcaseCount := base.DB.Model(&run.Submission).Association("Runs").Count()
	run.MemoryUsed = *req.MemoryUsed
	run.TimeUsed = *req.TimeUsed
	run.Status = req.Status
	run.OutputStrippedHash = *req.OutputStrippedHash
	run.JudgerMessage = req.Message
	run.Judged = true
	if req.Status == "ACCEPTED" {
		if run.TestCase.Score != 0 {
			run.Submission.Score += run.TestCase.Score
		} else {
			run.Submission.Score += uint(100 / testcaseCount)
			if run.Submission.Score == uint(100-(100%testcaseCount)) {
				run.Submission.Score = 100
			}
		}
	}
	utils.PanicIfDBError(base.DB.Session(&gorm.Session{FullSaveAssociations: true}).Save(&run), "could not save run")
	var isLast = false
	if base.DB.Model(&run.Submission).Where("status IN ? AND id <> ?", []string{"PENDING", "JUDGING"}, run.ID).Association("Runs").Count() == 0 {
		// this is the last judged run
		isLast = true
		run.Submission.Judged = true
		var runs []models.Run
		if err := base.DB.Model(&run.Submission).Association("Runs").Find(&runs); err != nil {
			panic(errors.Wrap(err, "could not query runs"))
		}
		for _, r := range runs {
			if r.Status != "ACCEPTED" {
				run.Submission.Status = r.Status
				break
			}
		}
		if run.Submission.Status == "PENDING" {
			run.Submission.Status = "ACCEPTED"
		}
	}
	run.Submission.UpdatedAt = time.Now()
	utils.PanicIfDBError(base.DB.Session(&gorm.Session{FullSaveAssociations: true}).Save(&run), "could not save run")
	unlocked = true
	runLock.Unlock()

	utils.MustPutObject(output, context.Background(), "submissions", fmt.Sprintf("%d/run/%d/output", run.Submission.ID, run.ID))
	utils.MustPutObject(comparer, context.Background(), "submissions", fmt.Sprintf("%d/run/%d/comparer_output", run.Submission.ID, run.ID))
	utils.MustPutObject(compiler, context.Background(), "submissions", fmt.Sprintf("%d/run/%d/compiler_output", run.Submission.ID, run.ID))
	if _, err := event.FireEvent("run", runEvent.EventArgs(&run)); err != nil {
		panic(errors.Wrap(err, "could not fire run events"))
	}
	if isLast {
		eventResults, err := event.FireEvent("submission", submissionEvent.EventArgs(run.Submission))
		if err != nil {
			panic(errors.Wrap(err, "could not fire submission events"))
		}
		for _, ret := range eventResults {
			if ret[0] != nil {
				panic(err)
			}
		}
	}

	return c.JSON(http.StatusOK, response.Response{
		Message: "SUCCESS",
	})
}
