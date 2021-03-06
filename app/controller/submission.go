package controller

import (
	"context"
	"fmt"
	"github.com/EduOJ/backend/app/request"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/app/response/resource"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/log"
	"github.com/EduOJ/backend/base/utils"
	"github.com/EduOJ/backend/database/models"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"net/http"
	"path/filepath"
	"time"
)

var inTest bool

func CreateSubmission(c echo.Context) error {
	req := request.CreateSubmissionRequest{}
	if err, ok := utils.BindAndValidate(&req, c); !ok {
		return err
	}

	problem := models.Problem{}
	if err := base.DB.First(&problem, c.Param("pid")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		} else {
			panic(errors.Wrap(err, "could not find problem"))
		}
	}
	problem.LoadTestCases()

	user := c.Get("user").(models.User)

	if !problem.Public && !user.Can("read_problem", problem) && !user.Can("read_problem") {
		return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
	}

	if !utils.Contain(req.Language, problem.LanguageAllowed) {
		return c.JSON(http.StatusBadRequest, response.ErrorResp("INVALID_LANGUAGE", nil))
	}

	language := models.Language{}

	if err := base.DB.First(&language, "name = ?", req.Language).Error; err != nil {
		panic(errors.Wrap(err, "could not find language"))
	}

	file, err := c.FormFile("code")
	if err != nil && err != http.ErrMissingFile && err.Error() != "request Content-Type isn't multipart/form-data" {
		panic(errors.Wrap(err, "could not read file"))
	}
	if file == nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResp("INVALID_FILE", nil))
	}

	ext := filepath.Ext(file.Filename)

	if ext != "" {
		ext = ext[1:]
	}

	priority := models.PriorityDefault

	submission := models.Submission{
		UserID:       user.ID,
		User:         &user,
		ProblemID:    problem.ID,
		Problem:      &problem,
		ProblemSetId: 0,
		LanguageName: language.Name,
		Language:     &language,
		FileName:     file.Filename,
		Priority:     priority,
		Judged:       false,
		Score:        0,
		Status:       "PENDING",
		Runs:         make([]models.Run, len(problem.TestCases)),
	}
	for i, testCase := range problem.TestCases {
		submission.Runs[i] = models.Run{
			UserID:             user.ID,
			ProblemID:          problem.ID,
			ProblemSetId:       0,
			TestCaseID:         testCase.ID,
			Sample:             testCase.Sample,
			Priority:           priority,
			Judged:             false,
			Status:             "PENDING",
			MemoryUsed:         0,
			TimeUsed:           0,
			OutputStrippedHash: "",
		}
	}
	utils.PanicIfDBError(base.DB.Create(&submission), "could not create submission and runs")

	utils.MustPutObject(file, c.Request().Context(), "submissions", fmt.Sprintf("%d/code", submission.ID))

	if !inTest {
		base.Redis.Publish(context.Background(), "runs", nil)
	}

	return c.JSON(http.StatusCreated, response.CreateSubmissionResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.SubmissionDetail `json:"submission"`
		}{
			resource.GetSubmissionDetail(&submission),
		},
	})
}

func GetSubmission(c echo.Context) error {
	var startedAt time.Time
	poll := false
	if c.QueryParam("poll") == "1" {
		poll = true
	}
	if err := echo.QueryParamsBinder(c).Time("before", &startedAt, time.RFC3339).BindError(); err != nil {
		// Ignore error.
		log.Error(err)
	}

	user := c.Get("user").(models.User)
	submission := models.Submission{}
	if err := base.DB.Preload("Problem").Preload("User").First(&submission, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if user.Can("read_submission") {
				return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
			} else {
				return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
			}
		} else {
			panic(errors.Wrap(err, "could not find problem"))
		}
	}
	if user.ID != submission.UserID && !user.Can("read_submission", submission.Problem) && !user.Can("read_submission") {
		return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
	}
	if !(submission.UpdatedAt.Before(startedAt.Add(time.Nanosecond)) && poll) {
		submission.LoadRuns()
		return c.JSON(http.StatusOK, response.GetSubmissionResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.SubmissionDetail `json:"submission"`
			}{
				resource.GetSubmissionDetail(&submission),
			},
		})
	}
	timeoutChan := time.After(viper.GetDuration("polling_timeout"))
	timeout := false
	sub := base.Redis.Subscribe(c.Request().Context(), fmt.Sprintf("submission_update:%d", submission.ID))
	defer sub.Close()
	for {
		select {
		case <-sub.Channel():
			if err := base.DB.Preload("Problem").Preload("User").First(&submission, c.Param("id")).Error; err != nil {
				panic(errors.Wrap(err, "could not find submission"))
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
		if submission.UpdatedAt.After(startedAt) {
			break
		}
	}
	submission.LoadRuns()
	return c.JSON(http.StatusOK, response.GetSubmissionResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.SubmissionDetail `json:"submission"`
		}{
			resource.GetSubmissionDetail(&submission),
		},
	})
}

func GetSubmissions(c echo.Context) error {
	req := request.GetSubmissionsRequest{}
	if err, ok := utils.BindAndValidate(&req, c); !ok {
		return err
	}

	query := base.DB.Model(&models.Submission{}).Preload("User").Preload("Problem").
		Where("problem_set_id = 0").Order("id DESC") // Force order by id desc.

	if req.ProblemId != 0 {
		query = query.Where("problem_id = ?", req.ProblemId)
	}
	if req.UserId != 0 {
		query = query.Where("user_id = ?", req.UserId)
	}

	var submissions []models.Submission
	total, prevUrl, nextUrl, err := utils.Paginator(query, req.Limit, req.Offset, c.Request().URL, &submissions)
	if err != nil {
		if herr, ok := err.(utils.HttpError); ok {
			return herr.Response(c)
		}
		panic(err)
	}
	return c.JSON(http.StatusOK, response.GetSubmissionsResponse{
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
			Submissions: resource.GetSubmissionSlice(submissions),
			Total:       total,
			Count:       len(submissions),
			Offset:      req.Offset,
			Prev:        prevUrl,
			Next:        nextUrl,
		},
	})
}

func GetSubmissionCode(c echo.Context) error {
	user := c.Get("user").(models.User)
	submission := models.Submission{}
	if err := base.DB.Preload("Problem").First(&submission, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if user.Can("read_submission") {
				return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
			} else {
				return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
			}
		} else {
			panic(errors.Wrap(err, "could not find problem"))
		}
	}
	if user.ID != submission.UserID && !user.Can("read_submission", submission.Problem) && !user.Can("read_submission") {
		return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
	}

	presignedUrl, err := utils.GetPresignedURL("submissions", fmt.Sprintf("%d/code", submission.ID), submission.FileName)
	if err != nil {
		panic(errors.Wrap(err, "could not get presigned url"))
	}
	return c.Redirect(http.StatusFound, presignedUrl)
}

func GetRunCompilerOutput(c echo.Context) error {
	user := c.Get("user").(models.User)
	submission := models.Submission{}
	if err := base.DB.Preload("Problem").First(&submission, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if user.Can("read_submission") {
				return c.JSON(http.StatusNotFound, response.ErrorResp("SUBMISSION_NOT_FOUND", nil))
			} else {
				return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
			}
		} else {
			panic(errors.Wrap(err, "could not find problem"))
		}
	}
	if user.ID != submission.UserID && !user.Can("read_submission", submission.Problem) && !user.Can("read_submission") {
		return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
	}

	run := models.Run{}
	if err := base.DB.Preload("TestCase").First(&run, c.Param("run_id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if user.Can("read_submission") {
				return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
			} else {
				return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
			}
		} else {
			panic(err)
		}
	}
	presignedUrl, err := utils.GetPresignedURL("submissions", fmt.Sprintf("%d/run/%d/compiler_output", submission.ID, run.ID), "compiler_output.txt")
	if err != nil {
		panic(errors.Wrap(err, "could not get presigned url"))
	}
	return c.Redirect(http.StatusFound, presignedUrl)
}

func GetRunOutput(c echo.Context) error {
	user := c.Get("user").(models.User)
	submission := models.Submission{}
	if err := base.DB.Preload("Problem").First(&submission, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if user.Can("create_problem") {
				return c.JSON(http.StatusNotFound, response.ErrorResp("SUBMISSION_NOT_FOUND", nil))
			} else {
				return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
			}
		} else {
			panic(errors.Wrap(err, "could not find problem"))
		}
	}

	canRead := false
	canReadSecret := false

	if user.Can("read_problem_secret", submission.Problem) || user.Can("read_problem_secret") {
		canReadSecret = true
		canRead = true
	} else if user.ID == submission.UserID {
		canRead = true
		canReadSecret = false
	}

	if !canRead {
		return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
	}

	run := models.Run{}
	if err := base.DB.Preload("TestCase").First(&run, "submission_id = ? and id = ?", submission.ID, c.Param("run_id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if canReadSecret {
				return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
			} else {
				return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
			}
		} else {
			panic(err)
		}
	}

	if !run.Sample && !canReadSecret {
		return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
	}

	presignedUrl, err := utils.GetPresignedURL("submissions", fmt.Sprintf("%d/run/%d/output", submission.ID, run.ID), run.TestCase.OutputFileName)

	if err != nil {
		panic(errors.Wrap(err, "could not get presigned url"))
	}
	return c.Redirect(http.StatusFound, presignedUrl)
}

func GetRunInput(c echo.Context) error {
	user := c.Get("user").(models.User)
	submission := models.Submission{}
	if err := base.DB.Preload("Problem").First(&submission, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if user.Can("create_problem") {
				return c.JSON(http.StatusNotFound, response.ErrorResp("SUBMISSION_NOT_FOUND", nil))
			} else {
				return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
			}
		} else {
			panic(errors.Wrap(err, "could not find problem"))
		}
	}

	canRead := false
	canReadSecret := false

	if user.Can("read_problem_secret", submission.Problem) || user.Can("read_problem_secret") {
		canReadSecret = true
		canRead = true
	} else if user.ID == submission.UserID {
		canRead = true
		canReadSecret = false
	}

	if !canRead {
		return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
	}

	run := models.Run{}
	if err := base.DB.Preload("TestCase").Preload("Problem").First(&run, "submission_id = ? and id = ?", submission.ID, c.Param("run_id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if canReadSecret {
				return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
			} else {
				return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
			}
		} else {
			panic(err)
		}
	}

	if !run.Sample && !canReadSecret {
		return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
	}

	presignedUrl, err := utils.GetPresignedURL("problems", fmt.Sprintf("%d/input/%d.in", run.Problem.ID, run.TestCase.ID), run.TestCase.InputFileName)

	if err != nil {
		panic(errors.Wrap(err, "could not get presigned url"))
	}
	return c.Redirect(http.StatusFound, presignedUrl)
}

func GetRunComparerOutput(c echo.Context) error {
	user := c.Get("user").(models.User)
	submission := models.Submission{}
	if err := base.DB.Preload("Problem").First(&submission, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if user.Can("create_problem") {
				return c.JSON(http.StatusNotFound, response.ErrorResp("SUBMISSION_NOT_FOUND", nil))
			} else {
				return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
			}
		} else {
			panic(errors.Wrap(err, "could not find problem"))
		}
	}

	canRead := false
	canReadSecret := false

	if user.Can("read_problem_secret", submission.Problem) || user.Can("read_problem_secret") {
		canReadSecret = true
		canRead = true
	} else if user.ID == submission.UserID {
		canRead = true
		canReadSecret = false
	}

	if !canRead {
		return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
	}

	run := models.Run{}
	if err := base.DB.Preload("TestCase").First(&run, "submission_id = ? and id = ?", submission.ID, c.Param("run_id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if canReadSecret {
				return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
			} else {
				return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
			}
		} else {
			panic(err)
		}
	}

	if !run.Sample && !canReadSecret {
		return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
	}

	presignedUrl, err := utils.GetPresignedURL("submissions", fmt.Sprintf("%d/run/%d/comparer_output", submission.ID, run.ID), "comparer_output.txt")

	if err != nil {
		panic(errors.Wrap(err, "could not get presigned url"))
	}
	return c.Redirect(http.StatusFound, presignedUrl)
}
