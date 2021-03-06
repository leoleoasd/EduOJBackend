package controller

import (
	"fmt"
	"github.com/EduOJ/backend/app/request"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/app/response/resource"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/utils"
	"github.com/EduOJ/backend/database/models"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

func GetProblem(c echo.Context) error {
	user := c.Get("user").(models.User)
	problem, err := utils.FindProblem(c.Param("id"), &user)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	} else if err != nil {
		panic(err)
	}
	if user.Can("read_problem_secrets", problem) || user.Can("read_problem_secrets") {
		return c.JSON(http.StatusOK, response.GetProblemResponseForAdmin{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.ProblemForAdmin `json:"problem"`
			}{
				resource.GetProblemForAdmin(problem),
			},
		})
	}
	return c.JSON(http.StatusOK, response.GetProblemResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.Problem `json:"problem"`
		}{
			resource.GetProblem(problem),
		},
	})
}

func GetProblems(c echo.Context) error {
	req := request.GetProblemsRequest{}
	if err, ok := utils.BindAndValidate(&req, c); !ok {
		return err
	}

	if req.Tried && req.Passed {
		return c.JSON(http.StatusBadRequest, response.ErrorResp("INVALID_STATUS", nil))
	}

	query := base.DB.Model(&models.Problem{}).Order("id ASC").Omit("Description") // Force order by id asc.

	user := c.Get("user").(models.User)
	isAdmin := user.Can("manage_problem")
	if !isAdmin {
		query = query.Where("public = ?", true)
	}

	if req.Search != "" {
		id, _ := strconv.ParseUint(req.Search, 10, 64)
		query = query.Where("id = ? or name like ?", id, "%"+req.Search+"%")
	}

	where := (*gorm.DB)(nil)

	if req.Passed {
		where = base.DB.Where("id in (?)", base.DB.Table("submissions").
			Select("problem_id").
			Where("status = 'ACCEPTED' and user_id = ?", req.UserID).
			Group("problem_id"))
	}

	if req.Tried {
		where = base.DB.Where("id not in (?)",
			base.DB.Table("submissions").
				Select("problem_id").
				Where("status = 'ACCEPTED' and user_id = ?", req.UserID).
				Group("problem_id"),
		).Where("id in (?)",
			base.DB.Table("submissions").
				Select("problem_id").
				Where("status <> 'ACCEPTED' and user_id = ?", req.UserID).
				Group("problem_id"),
		)
	}

	if where != nil {
		query = query.Where(where)
	}

	var problems []*models.Problem
	total, prevUrl, nextUrl, err := utils.Paginator(query, req.Limit, req.Offset, c.Request().URL, &problems)
	if err != nil {
		if herr, ok := err.(utils.HttpError); ok {
			return herr.Response(c)
		}
		panic(err)
	}
	if isAdmin {
		return c.JSON(http.StatusOK, response.GetProblemsResponseForAdmin{
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
				Problems: resource.GetProblemForAdminSlice(problems),
				Total:    total,
				Count:    len(problems),
				Offset:   req.Offset,
				Prev:     prevUrl,
				Next:     nextUrl,
			},
		})
	}
	return c.JSON(http.StatusOK, response.GetProblemsResponse{
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
			Problems: resource.GetProblemSlice(problems),
			Total:    total,
			Count:    len(problems),
			Offset:   req.Offset,
			Prev:     prevUrl,
			Next:     nextUrl,
		},
	})
}

func GetProblemAttachmentFile(c echo.Context) error { // TODO: use MustGetObject
	user := c.Get("user").(models.User)
	problem, err := utils.FindProblem(c.Param("id"), &user)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	} else if err != nil {
		panic(err)
	}
	if problem.AttachmentFileName == "" {
		return c.JSON(http.StatusNotFound, response.ErrorResp("ATTACHMENT_NOT_FOUND", nil))
	}
	presignedUrl, err := utils.GetPresignedURL("problems", fmt.Sprintf("%d/attachment", problem.ID), problem.AttachmentFileName)
	if err != nil {
		panic(errors.Wrap(err, "could not get presigned url"))
	}
	return c.Redirect(http.StatusFound, presignedUrl)
}

func CreateProblem(c echo.Context) error {
	file, err := c.FormFile("attachment_file")
	if err != nil && err != http.ErrMissingFile && err.Error() != "request Content-Type isn't multipart/form-data" {
		panic(errors.Wrap(err, "could not read file"))
	}

	req := request.CreateProblemRequest{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
	}
	var public, privacy bool
	if req.Public == nil {
		public = false
	} else {
		public = *req.Public
	}
	if req.Privacy == nil {
		privacy = true
	} else {
		privacy = *req.Privacy
	}

	problem := models.Problem{
		Name:              req.Name,
		Description:       req.Description,
		Public:            public,
		Privacy:           privacy,
		MemoryLimit:       req.MemoryLimit,
		TimeLimit:         req.TimeLimit,
		LanguageAllowed:   strings.Split(req.LanguageAllowed, ","),
		BuildArg:          req.BuildArg,
		CompareScriptName: req.CompareScriptName,
	}
	if file != nil {
		problem.AttachmentFileName = file.Filename
	}
	utils.PanicIfDBError(base.DB.Create(&problem), "could not create problem")

	// Move this before "Must Put Object" to prevent creating a problem without "problem_creator" if put object fails.
	user := c.Get("user").(models.User)
	user.GrantRole("problem_creator", problem)

	if file != nil {
		utils.MustPutObject(file, c.Request().Context(), "problems", fmt.Sprintf("%d/attachment", problem.ID))
	}

	return c.JSON(http.StatusCreated, response.CreateProblemResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.ProblemForAdmin `json:"problem"`
		}{
			resource.GetProblemForAdmin(&problem),
		},
	})
}

func UpdateProblem(c echo.Context) error {
	req := request.UpdateProblemRequest{}
	if err, ok := utils.BindAndValidate(&req, c); !ok {
		return err
	}
	problem, err := utils.FindProblem(c.Param("id"), nil)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		}
		panic(err)
	}
	problem.Name = req.Name
	problem.Description = req.Description

	file, err := c.FormFile("attachment_file")
	if err != nil && err != http.ErrMissingFile && err.Error() != "request Content-Type isn't multipart/form-data" {
		panic(errors.Wrap(err, "could not read file"))
	}
	if file != nil { // TODO: use MustPutObject
		problem.AttachmentFileName = file.Filename
		src, err := file.Open()
		if err != nil {
			panic(err)
		}
		defer src.Close()
		_, err = base.Storage.PutObject(c.Request().Context(), "problems", fmt.Sprintf("%d/attachment", problem.ID), src, file.Size, minio.PutObjectOptions{})
		if err != nil {
			panic(errors.Wrap(err, "could write attachment file to s3 storage."))
		}
	}

	if req.Public != nil {
		problem.Public = *req.Public
	} else {
		problem.Public = false
	}
	if req.Privacy != nil {
		problem.Privacy = *req.Privacy
	} else {
		problem.Privacy = true
	}
	problem.MemoryLimit = req.MemoryLimit
	problem.TimeLimit = req.TimeLimit
	problem.LanguageAllowed = strings.Split(req.LanguageAllowed, ",")
	problem.BuildArg = req.BuildArg
	problem.CompareScriptName = req.CompareScriptName
	utils.PanicIfDBError(base.DB.Save(&problem), "could not update problem")
	return c.JSON(http.StatusOK, response.UpdateProblemResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.ProblemForAdmin `json:"problem"`
		}{
			resource.GetProblemForAdmin(problem),
		},
	})
}

func DeleteProblem(c echo.Context) error {
	problem, err := utils.FindProblem(c.Param("id"), nil)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	} else if err != nil {
		panic(err)
	}

	testCaseIds := make([]uint, len(problem.TestCases))

	for i, testCase := range problem.TestCases {
		testCaseIds[i] = testCase.ID
	}
	utils.PanicIfDBError(base.DB.Delete(&models.TestCase{}, "id IN (?)", testCaseIds), "could not delete test cases")

	var roles []models.Role
	if err := base.DB.Where("target = ?", "problem").Find(&roles).Error; err != gorm.ErrRecordNotFound && err != nil {
		panic(errors.Wrap(err, "could not find roles"))
	}

	roleIds := make([]uint, len(roles))
	for i, role := range roles {
		roleIds[i] = role.ID
	}
	utils.PanicIfDBError(base.DB.Delete(&models.UserHasRole{}, "role_id IN (?) and target_id = ?", roleIds, problem.ID), "could not delete user has roles")
	utils.PanicIfDBError(base.DB.Delete(&problem), "could not delete problem")
	return c.JSON(http.StatusOK, response.Response{
		Message: "SUCCESS",
		Error:   nil,
		Data:    nil,
	})
}

func CreateTestCase(c echo.Context) error {

	problem, err := utils.FindProblem(c.Param("id"), nil)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		}
		panic(err)
	}

	inputFile, err := c.FormFile("input_file")
	if err != nil && err != http.ErrMissingFile && err.Error() != "request Content-Type isn't multipart/form-data" {
		panic(errors.Wrap(err, "could not read input file"))
	}
	outputFile, err := c.FormFile("output_file")
	if err != nil && err != http.ErrMissingFile && err.Error() != "request Content-Type isn't multipart/form-data" {
		panic(errors.Wrap(err, "could not read output file"))
	}

	if inputFile == nil || outputFile == nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResp("INVALID_FILE", nil))
	}

	req := request.CreateTestCaseRequest{}
	if err, ok := utils.BindAndValidate(&req, c); !ok {
		return err
	}

	testCase := models.TestCase{
		Score:          req.Score,
		Sample:         *req.Sample,
		InputFileName:  inputFile.Filename,
		OutputFileName: outputFile.Filename,
	}

	if err := base.DB.Model(&problem).Association("TestCases").Append(&testCase); err != nil {
		panic(errors.Wrap(err, "could not create test case"))
	}

	utils.MustPutObject(inputFile, c.Request().Context(), "problems", fmt.Sprintf("%d/input/%d.in", problem.ID, testCase.ID))
	utils.MustPutObject(outputFile, c.Request().Context(), "problems", fmt.Sprintf("%d/output/%d.out", problem.ID, testCase.ID))

	return c.JSON(http.StatusCreated, response.CreateTestCaseResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.TestCaseForAdmin `json:"test_case"`
		}{
			resource.GetTestCaseForAdmin(&testCase),
		},
	})
}

func GetTestCaseInputFile(c echo.Context) error {
	testCase := c.Get("test_case").(*models.TestCase)
	problem := c.Get("problem").(*models.Problem)
	var err error
	// ferr finding error
	if ferr := c.Get("find_test_case_err"); ferr != nil {
		err = ferr.(error)
	}
	if problem == nil {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	}
	if testCase == nil {
		return c.JSON(http.StatusNotFound, response.ErrorResp("TEST_CASE_NOT_FOUND", err))
	}
	if err != nil {
		panic(errors.Wrap(err, "could not find test case"))
	}

	presignedUrl, err := utils.GetPresignedURL("problems", fmt.Sprintf("%d/input/%d.in", problem.ID, testCase.ID), testCase.InputFileName)
	if err != nil {
		panic(errors.Wrap(err, "could not get presigned url"))
	}
	return c.Redirect(http.StatusFound, presignedUrl)
}

func GetTestCaseOutputFile(c echo.Context) error {
	testCase := c.Get("test_case").(*models.TestCase)
	problem := c.Get("problem").(*models.Problem)
	var err error
	// ferr finding error
	if ferr := c.Get("find_test_case_err"); ferr != nil {
		err = ferr.(error)
	}
	if problem == nil {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	}
	if testCase == nil {
		return c.JSON(http.StatusNotFound, response.ErrorResp("TEST_CASE_NOT_FOUND", err))
	}
	if err != nil {
		panic(errors.Wrap(err, "could not find test case"))
	}

	presignedUrl, err := utils.GetPresignedURL("problems", fmt.Sprintf("%d/output/%d.out", problem.ID, testCase.ID), testCase.OutputFileName)
	if err != nil {
		panic(errors.Wrap(err, "could not get presigned url"))
	}
	return c.Redirect(http.StatusFound, presignedUrl)
}

func UpdateTestCase(c echo.Context) error {
	testCase, problem, err := utils.FindTestCase(c.Param("id"), c.Param("test_case_id"), nil)
	if problem == nil {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	} else if testCase == nil {
		return c.JSON(http.StatusNotFound, response.ErrorResp("TEST_CASE_NOT_FOUND", err))
	}

	req := request.UpdateTestCaseRequest{}
	if err, ok := utils.BindAndValidate(&req, c); !ok {
		return err
	}

	inputFile, err := c.FormFile("input_file")
	if err != nil && err != http.ErrMissingFile && err.Error() != "request Content-Type isn't multipart/form-data" {
		panic(errors.Wrap(err, "could not read input file"))
	}
	outputFile, err := c.FormFile("output_file")
	if err != nil && err != http.ErrMissingFile && err.Error() != "request Content-Type isn't multipart/form-data" {
		panic(errors.Wrap(err, "could not read output file"))
	}

	if inputFile != nil {
		utils.MustPutObject(inputFile, c.Request().Context(), "problems", fmt.Sprintf("%d/input/%d.in", problem.ID, testCase.ID))
		testCase.InputFileName = inputFile.Filename
	}
	if outputFile != nil {
		utils.MustPutObject(outputFile, c.Request().Context(), "problems", fmt.Sprintf("%d/output/%d.out", problem.ID, testCase.ID))
		testCase.OutputFileName = outputFile.Filename
	}

	testCase.Score = req.Score
	testCase.Sample = *req.Sample
	utils.PanicIfDBError(base.DB.Save(&testCase), "could not update testCase")

	return c.JSON(http.StatusOK, response.UpdateTestCaseResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.TestCaseForAdmin `json:"test_case"`
		}{
			resource.GetTestCaseForAdmin(testCase),
		},
	})
}

func DeleteTestCase(c echo.Context) error {
	testCase, problem, err := utils.FindTestCase(c.Param("id"), c.Param("test_case_id"), nil)
	if problem == nil {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	} else if testCase == nil {
		return c.JSON(http.StatusNotFound, response.ErrorResp("TEST_CASE_NOT_FOUND", err))
	}

	utils.PanicIfDBError(base.DB.Delete(&testCase), "could not remove test case")

	return c.JSON(http.StatusOK, response.Response{
		Message: "SUCCESS",
		Error:   nil,
		Data:    nil,
	})
}

func DeleteTestCases(c echo.Context) error {
	problem, err := utils.FindProblem(c.Param("id"), nil)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	} else if err != nil {
		panic(err)
	}

	testCaseIds := make([]uint, len(problem.TestCases))

	for i, testCase := range problem.TestCases {
		testCaseIds[i] = testCase.ID
	}
	utils.PanicIfDBError(base.DB.Delete(&models.TestCase{}, "id IN (?)", testCaseIds), "could not delete test cases")

	return c.JSON(http.StatusOK, response.Response{
		Message: "SUCCESS",
		Error:   nil,
		Data:    nil,
	})
}

func GetRandomProblem(c echo.Context) error {
	var count int64
	utils.PanicIfDBError(base.DB.Find(&models.Problem{}, "public = true").Count(&count),
		"could not get count of public problems for getting random problem")
	if count == 0 {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	}
	problem := models.Problem{}
	utils.PanicIfDBError(base.DB.Limit(1).Offset(rand.Intn(int(count))).Find(&problem),
		"could not get problem for getting random problem")
	return c.JSON(http.StatusOK, response.GetRandomProblemResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.Problem `json:"problem"`
		}{
			resource.GetProblem(&problem),
		},
	})
}

func GetUserProblemInfo(c echo.Context) error {
	userID := c.Param("id")
	var passedCount, triedCount int64

	utils.PanicIfDBError(base.DB.Model(&models.Problem{}).
		Where("id in (?)", base.DB.Table("submissions").
			Select("problem_id").
			Where("status = 'ACCEPTED' and user_id = ?", userID).
			Group("problem_id")).
		Count(&passedCount), "could not get count of passed problems for getting user problem info")

	utils.PanicIfDBError(base.DB.Model(&models.Problem{}).
		Where("id not in (?)", base.DB.Table("submissions").
			Select("problem_id").
			Where("status = 'ACCEPTED' and user_id = ?", userID).
			Group("problem_id"),
		).Where("id in (?)", base.DB.Table("submissions").
		Select("problem_id").
		Where("status <> 'ACCEPTED' and user_id = ?", userID).
		Group("problem_id")).
		Count(&triedCount), "could not get count of tried problems for getting user problem info")

	return c.JSON(http.StatusOK, response.GetUserProblemInfoResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			TriedCount  int `json:"tried_count"`
			PassedCount int `json:"passed_count"`
			Rank        int `json:"rank"`
		}{
			TriedCount:  int(triedCount),
			PassedCount: int(passedCount),
			Rank:        0, // TODO: develop this
		},
	})
}
