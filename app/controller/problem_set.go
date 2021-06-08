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
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
	"time"
)

func CreateProblemSet(c echo.Context) error {
	req := request.CreateProblemSetRequest{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
	}
	class := models.Class{}
	if err := base.DB.First(&class, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("CLASS_NOT_FOUND", nil))
		}
		panic(errors.Wrap(err, "could not get class while creating problem set"))
	}
	problemSet := models.ProblemSet{
		ClassID:     class.ID,
		Name:        req.Name,
		Description: req.Description,
		Problems:    nil,
		Grades:      nil,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	}
	utils.PanicIfDBError(base.DB.Create(&problemSet), "could not create problem set for creating problem set")
	return c.JSON(http.StatusCreated, response.CreateProblemSetResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.ProblemSetDetail `json:"problem_set"`
		}{
			resource.GetProblemSetDetail(&problemSet),
		},
	})
}

func CloneProblemSet(c echo.Context) error {
	req := request.CloneProblemSetRequest{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
	}
	class := models.Class{}
	if err := base.DB.First(&class, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("CLASS_NOT_FOUND", nil))
		}
		panic(errors.Wrap(err, "could not get class while cloning problem set"))
	}
	sourceProblemSet := models.ProblemSet{}
	if err := base.DB.Preload("Problems").
		First(&sourceProblemSet, "id = ? and class_id = ?", req.SourceProblemSetID, req.SourceClassID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("SOURCE_NOT_FOUND", nil))
		} else {
			panic(errors.Wrap(err, "could not find source class when cloning problem set"))
		}
	}
	problemSet := models.ProblemSet{
		ClassID:     class.ID,
		Name:        sourceProblemSet.Name,
		Description: sourceProblemSet.Description,
		Problems:    sourceProblemSet.Problems,
		Grades:      nil,
		StartTime:   sourceProblemSet.StartTime,
		EndTime:     sourceProblemSet.EndTime,
	}
	utils.PanicIfDBError(base.DB.Create(&problemSet), "could not add problem set for class when cloning problem set")
	return c.JSON(http.StatusCreated, response.CloneProblemSetResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.ProblemSetDetail `json:"problem_set"`
		}{
			resource.GetProblemSetDetail(&problemSet),
		},
	})
}

func GetProblemSet(c echo.Context) error {
	class := models.Class{}
	if err := base.DB.First(&class, c.Param("class_id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("CLASS_NOT_FOUND", nil))
		}
		panic(errors.Wrap(err, "could not get class while creating problem set"))
	}

	user := c.Get("user").(models.User)
	problemSet := models.ProblemSet{}
	if err := base.DB.Preload("Problems").
		First(&problemSet, "id = ? and class_id = ?", c.Param("problem_set_id"), c.Param("class_id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		}
		panic(errors.Wrap(err, "could not get class for getting problem set"))
	}

	if user.Can("manage_problem_sets", class) || user.Can("manage_problem_sets") {
		return c.JSON(http.StatusOK, response.GetProblemSetResponseForAdmin{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.ProblemSetDetail `json:"problem_set"`
			}{
				resource.GetProblemSetDetail(&problemSet),
			},
		})
	}
	if time.Now().Before(problemSet.StartTime) {
		// TODO: add config to determine if students could read problems and submissions when problem sets end
		return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
	}
	var users []models.User
	if err := base.DB.Model(&class).Association("Students").Find(&users, user.ID); err != nil {
		panic(errors.Wrap(err, "could not check student in class for getting problem set"))
	}
	if len(users) == 0 {
		return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
	}
	return c.JSON(http.StatusOK, response.GetProblemSetResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.ProblemSet `json:"problem_set"`
		}{
			resource.GetProblemSet(&problemSet),
		},
	})
}

func UpdateProblemSet(c echo.Context) error {
	req := request.UpdateProblemSetRequest{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
	}
	problemSet := models.ProblemSet{}
	if err := base.DB.Preload("Problems").
		First(&problemSet, "id = ? and class_id = ?", c.Param("problem_set_id"), c.Param("class_id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		}
		panic(errors.Wrap(err, "could not get problem set for updating problem set"))
	}
	problemSet.Name = req.Name
	problemSet.Description = req.Description
	problemSet.StartTime = req.StartTime
	problemSet.EndTime = req.EndTime
	utils.PanicIfDBError(base.DB.Save(&problemSet), "could not update problem set for updating problem set")
	return c.JSON(http.StatusOK, response.UpdateProblemSetResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.ProblemSetDetail `json:"problem_set"`
		}{
			resource.GetProblemSetDetail(&problemSet),
		},
	})
}

func AddProblemsToSet(c echo.Context) error {
	req := request.AddProblemsToSetRequest{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
	}

	problemSet := models.ProblemSet{}
	if err := base.DB.Preload("Problems").
		First(&problemSet, "id = ? and class_id = ?", c.Param("id"), c.Param("class_id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		}
		panic(errors.Wrap(err, "could not get problem set for adding problems to problem set"))
	}
	if err := problemSet.AddProblems(req.ProblemIDs); err != nil {
		panic(errors.Wrap(err, "could not add problems to problem set"))
	}
	return c.JSON(http.StatusOK, response.AddProblemsToSetResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.ProblemSetDetail `json:"problem_set"`
		}{
			resource.GetProblemSetDetail(&problemSet),
		},
	})
}

func DeleteProblemsFromSet(c echo.Context) error {
	req := request.DeleteProblemsFromSetRequest{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
	}

	problemSet := models.ProblemSet{}
	if err := base.DB.Preload("Problems").
		First(&problemSet, "id = ? and class_id = ?", c.Param("id"), c.Param("class_id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		}
		panic(errors.Wrap(err, "could not get problem set for deleting problems from problem set"))
	}
	if err := problemSet.DeleteProblems(req.ProblemIDs); err != nil {
		panic(errors.Wrap(err, "could not delete problems from problem set"))
	}
	return c.JSON(http.StatusOK, response.DeleteProblemsFromSetResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.ProblemSetDetail `json:"problem_set"`
		}{
			resource.GetProblemSetDetail(&problemSet),
		},
	})
}

func DeleteProblemSet(c echo.Context) error {
	problemSet := models.ProblemSet{}
	if err := base.DB.First(&problemSet, "id = ? and class_id = ?", c.Param("problem_set_id"), c.Param("class_id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		}
		panic(errors.Wrap(err, "could not get problem set for deleting problem set"))
	}
	if problemSet.ID != 0 {
		utils.PanicIfDBError(base.DB.Delete(&problemSet), "could not delete problem set for deleting problem set")
	}
	return c.JSON(http.StatusOK, response.Response{
		Message: "SUCCESS",
		Error:   nil,
		Data:    nil,
	})
}

func GetProblemSetProblem(c echo.Context) error {

	problemSet := models.ProblemSet{}
	var class *models.Class
	isAdmin := false
	problemSetInContext := c.Get("problem_set")
	if problemSetInContext != nil {
		err := c.Get("find_problem_set_error")
		if err != nil {
			if errors.Is(err.(error), gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, response.ErrorResp("PROBLEM_SET_NOT_FOUND", nil))
			}
			panic(errors.Wrap(err.(error), "could not find problem set for getting problem set problem"))
		}
		problemSet = *problemSetInContext.(*models.ProblemSet)
		class = &models.Class{}
		utils.PanicIfDBError(base.DB.First(class, problemSet.ClassID), "could not find class while getting problem set problem")
	} else {
		if err := base.DB.Preload("Class").
			First(&problemSet, "id = ? and class_id = ?", c.Param("problem_set_id"), c.Param("class_id")).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, response.ErrorResp("PROBLEM_SET_NOT_FOUND", nil))
			}
			panic(errors.Wrap(err, "could not get problem set for getting problem set problem"))
		}
		class = problemSet.Class
		isAdmin = true
	}

	var problems []models.Problem
	if err := base.DB.Model(&problemSet).Association("Problems").Find(&problems, c.Param("id")); err != nil {
		panic(errors.Wrap(err, "could not check problem in problem set for getting problem set problem"))
	}
	if len(problems) == 0 {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	}

	// skip permission check in function FindProblem, check permission in this function instead
	problem, err := utils.FindProblem(c.Param("id"), nil)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	} else if err != nil {
		panic(errors.Wrap(err, "could not find problem for getting problem set problem"))
	}

	user := c.Get("user").(models.User)
	if isAdmin {
		return c.JSON(http.StatusOK, response.GetProblemSetProblemResponseForAdmin{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.ProblemForAdmin `json:"problem"`
			}{
				resource.GetProblemForAdmin(problem),
			},
		})
	}
	var users []models.User
	if err := base.DB.Model(class).Association("Students").Find(&users, user.ID); err != nil {
		panic(errors.Wrap(err, "could not check student in class for getting problem set problem"))
	}
	if len(users) == 0 {
		return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
	}
	return c.JSON(http.StatusOK, response.GetProblemSetProblemResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.Problem `json:"problem"`
		}{
			resource.GetProblem(problem),
		},
	})
}

func GetProblemSetProblemInputFile(c echo.Context) error {
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
		return c.JSON(http.StatusNotFound, response.ErrorResp("TEST_CASE_NOT_FOUND", nil))
	}
	if err != nil {
		panic(errors.Wrap(err, "could not find test case"))
	}

	problemSet := models.ProblemSet{}
	var class *models.Class
	isAdmin := false
	problemSetInContext := c.Get("problem_set")
	if problemSetInContext != nil {
		err := c.Get("find_problem_set_error")
		if err != nil {
			if errors.Is(err.(error), gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, response.ErrorResp("PROBLEM_SET_NOT_FOUND", nil))
			}
			panic(errors.Wrap(err.(error), "could not find problem set for getting problem set problem"))
		}
		problemSet = *problemSetInContext.(*models.ProblemSet)
		class = &models.Class{}
		utils.PanicIfDBError(base.DB.First(class, problemSet.ClassID), "could not find class while getting problem set problem")
	} else {
		if err := base.DB.Preload("Class").
			First(&problemSet, "id = ? and class_id = ?", c.Param("problem_set_id"), c.Param("class_id")).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, response.ErrorResp("PROBLEM_SET_NOT_FOUND", nil))
			}
			panic(errors.Wrap(err, "could not get problem set for getting problem set problem"))
		}
		class = problemSet.Class
		isAdmin = true
	}

	var problems []models.Problem
	if err := base.DB.Model(&problemSet).Association("Problems").Find(&problems, c.Param("id")); err != nil {
		panic(errors.Wrap(err, "could not check problem in problem set for getting problem set problem"))
	}
	if len(problems) == 0 {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	}

	user := c.Get("user").(models.User)
	var users []models.User
	if err := base.DB.Model(class).Association("Students").Find(&users, user.ID); err != nil {
		panic(errors.Wrap(err, "could not check student in class for getting problem set problem"))
	}
	if len(users) == 0 && !isAdmin {
		return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
	}
	presignedUrl, err := utils.GetPresignedURL("problems", fmt.Sprintf("%d/input/%d.in", problem.ID, testCase.ID), testCase.InputFileName)
	if err != nil {
		panic(errors.Wrap(err, "could not get presigned url"))
	}
	return c.Redirect(http.StatusFound, presignedUrl)
}

func GetProblemSetProblemOutputFile(c echo.Context) error {
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
		return c.JSON(http.StatusNotFound, response.ErrorResp("TEST_CASE_NOT_FOUND", nil))
	}
	if err != nil {
		panic(errors.Wrap(err, "could not find test case"))
	}

	problemSet := models.ProblemSet{}
	var class *models.Class
	isAdmin := false
	problemSetInContext := c.Get("problem_set")
	if problemSetInContext != nil {
		err := c.Get("find_problem_set_error")
		if err != nil {
			if errors.Is(err.(error), gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, response.ErrorResp("PROBLEM_SET_NOT_FOUND", nil))
			}
			panic(errors.Wrap(err.(error), "could not find problem set for getting problem set problem"))
		}
		problemSet = *problemSetInContext.(*models.ProblemSet)
		class = &models.Class{}
		utils.PanicIfDBError(base.DB.First(class, problemSet.ClassID), "could not find class while getting problem set problem")
	} else {
		if err := base.DB.Preload("Class").
			First(&problemSet, "id = ? and class_id = ?", c.Param("problem_set_id"), c.Param("class_id")).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, response.ErrorResp("PROBLEM_SET_NOT_FOUND", nil))
			}
			panic(errors.Wrap(err, "could not get problem set for getting problem set problem"))
		}
		class = problemSet.Class
		isAdmin = true
	}

	var problems []models.Problem
	if err := base.DB.Model(&problemSet).Association("Problems").Find(&problems, c.Param("id")); err != nil {
		panic(errors.Wrap(err, "could not check problem in problem set for getting problem set problem"))
	}
	if len(problems) == 0 {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	}

	user := c.Get("user").(models.User)
	var users []models.User
	if err := base.DB.Model(class).Association("Students").Find(&users, user.ID); err != nil {
		panic(errors.Wrap(err, "could not check student in class for getting problem set problem"))
	}
	if len(users) == 0 && !isAdmin {
		return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
	}
	presignedUrl, err := utils.GetPresignedURL("problems", fmt.Sprintf("%d/output/%d.out", problem.ID, testCase.ID), testCase.OutputFileName)
	if err != nil {
		panic(errors.Wrap(err, "could not get presigned url"))
	}
	return c.Redirect(http.StatusFound, presignedUrl)
}

func RefreshGrades(c echo.Context) error {
	problemSet := models.ProblemSet{}
	if err := base.DB.Preload("Problems").Preload("Class.Students").Preload("Grades").
		First(&problemSet, "id = ? and class_id = ?", c.Param("id"), c.Param("class_id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		}
		panic(errors.Wrap(err, "could not get problem set for refreshing grades"))
	}
	if err := utils.RefreshGrades(&problemSet); err != nil {
		panic(errors.Wrap(err, "could not refresh grades"))
	}
	return c.JSON(http.StatusOK, response.RefreshGradesResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.ProblemSetWithGrades `json:"grades"`
		}{
			resource.GetProblemSetWithGrades(&problemSet),
		},
	})
}
