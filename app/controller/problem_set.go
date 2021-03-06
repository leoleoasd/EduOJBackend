package controller

import (
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
	problemSet := models.ProblemSet{}
	if err := base.DB.Preload("Problems").Preload("Grades").
		First(&problemSet, "id = ? and class_id = ?", c.Param("id"), c.Param("class_id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		}
		panic(errors.Wrap(err, "could not get class for getting problem set"))
	}
	user := c.Get("user").(models.User)
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
	var users []models.User
	if err := base.DB.Model(&class).Association("Students").Find(&users, user.ID); err != nil {
		panic(errors.Wrap(err, "could not check student in class for getting problem set"))
	}
	if len(users) == 0 {
		return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
	}
	if time.Now().Before(problemSet.StartTime) || time.Now().After(problemSet.EndTime) {
		problemSet.Problems = nil
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
	if err := base.DB.Preload("Problems").Preload("Grades").
		First(&problemSet, "id = ? and class_id = ?", c.Param("id"), c.Param("class_id")).Error; err != nil {
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
	if err := base.DB.Preload("Problems").Preload("Grades").
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
	if err := base.DB.Preload("Problems").Preload("Grades").
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
	if err := base.DB.First(&problemSet, "id = ? and class_id = ?", c.Param("id"), c.Param("class_id")).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		panic(errors.Wrap(err, "could not get problem set for deleting problem set"))
	}
	utils.PanicIfDBError(base.DB.Delete(&problemSet), "could not delete problem set for deleting problem set")
	return c.JSON(http.StatusOK, response.Response{
		Message: "SUCCESS",
		Error:   nil,
		Data:    nil,
	})
}
