package controller_test

import (
	"fmt"
	"github.com/EduOJ/backend/app/request"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/app/response/resource"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/database/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"net/http"
	"testing"
	"time"
)

func TestCreateProblemSet(t *testing.T) {
	t.Parallel()

	class := createClassForTest(t, "create_problem_set_permission_denied", 0, nil, nil)

	failTests := []failTest{
		{
			name:       "WithoutParams",
			method:     "POST",
			path:       base.Echo.Reverse("problemSet.createProblemSet", class.ID),
			req:        request.CreateProblemSetRequest{},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusBadRequest,
			resp: response.ErrorResp("VALIDATION_ERROR", []interface{}{
				map[string]interface{}{
					"field":       "Name",
					"reason":      "required",
					"translation": "名称为必填字段",
				},
				map[string]interface{}{
					"field":       "Description",
					"reason":      "required",
					"translation": "描述为必填字段",
				},
				map[string]interface{}{
					"field":       "StartTime",
					"reason":      "required",
					"translation": "开始时间为必填字段",
				},
				map[string]interface{}{
					"field":       "EndTime",
					"reason":      "required",
					"translation": "结束时间为必填字段",
				},
			}),
		},
		{
			name:   "NonExistingClass",
			method: "POST",
			path:   base.Echo.Reverse("problemSet.createProblemSet", -1),
			req: request.CreateProblemSetRequest{
				Name:        "test_create_problem_set_non_existing_class_name",
				Description: "test_create_problem_set_non_existing_class_description",
				StartTime:   hashStringToTime("test_create_problem_set_non_existing_class_time"),
				EndTime:     hashStringToTime("test_create_problem_set_non_existing_class_time").Add(time.Hour),
			},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("CLASS_NOT_FOUND", nil),
		},
		{
			name:   "PermissionDenied",
			method: "POST",
			path:   base.Echo.Reverse("problemSet.createProblemSet", class.ID),
			req: request.CreateProblemSetRequest{
				Name:        "test_create_problem_set_permission_denied_name",
				Description: "test_create_problem_set_permission_denied_description",
				StartTime:   hashStringToTime("test_create_problem_set_permission_denied_time"),
				EndTime:     hashStringToTime("test_create_problem_set_permission_denied_time").Add(time.Hour),
			},
			reqOptions: []reqOption{applyNormalUser},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "")

	t.Run("Success", func(t *testing.T) {
		user := createUserForTest(t, "create_problem_set_success", 0)
		class := createClassForTest(t, "create_problem_set_success", 0, nil, nil)
		user.GrantRole("class_creator", class)

		httpResp := makeResp(makeReq(t, "POST", base.Echo.Reverse("problemSet.createProblemSet", class.ID), request.CreateProblemSetRequest{
			Name:        "test_create_problem_set_success_name",
			Description: "test_create_problem_set_success_description",
			StartTime:   hashStringToTime("test_create_problem_set_success_time"),
			EndTime:     hashStringToTime("test_create_problem_set_success_time").Add(time.Hour),
		}, applyUser(user)))
		assert.Equal(t, http.StatusCreated, httpResp.StatusCode)

		databaseProblemSet := models.ProblemSet{}
		assert.NoError(t, base.DB.Preload("Problems").Preload("Grades").First(&databaseProblemSet, "name = ?", "test_create_problem_set_success_name").Error)
		expectedProblemSet := models.ProblemSet{
			ID:          databaseProblemSet.ID,
			ClassID:     class.ID,
			Name:        "test_create_problem_set_success_name",
			Description: "test_create_problem_set_success_description",
			Problems:    []*models.Problem{},
			Grades:      []*models.Grade{},
			StartTime:   hashStringToTime("test_create_problem_set_success_time"),
			EndTime:     hashStringToTime("test_create_problem_set_success_time").Add(time.Hour),
			CreatedAt:   databaseProblemSet.CreatedAt,
			UpdatedAt:   databaseProblemSet.UpdatedAt,
			DeletedAt:   gorm.DeletedAt{},
		}
		assert.Equal(t, expectedProblemSet, databaseProblemSet)

		resp := response.CreateProblemSetResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.CreateProblemSetResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.ProblemSetDetail `json:"problem_set"`
			}{
				resource.GetProblemSetDetail(&expectedProblemSet),
			},
		}, resp)
	})
}

func createProblemSetForTest(t *testing.T, name string, id int, class *models.Class, problems []models.Problem) *models.ProblemSet {
	problemSet := models.ProblemSet{
		Name:        fmt.Sprintf("test_%s_%d_name", name, id),
		Description: fmt.Sprintf("test_%s_%d_description", name, id),
		Problems:    []*models.Problem{},
		Grades:      []*models.Grade{},
		StartTime:   hashStringToTime(fmt.Sprintf("test_%s_%d_time", name, id)),
		EndTime:     hashStringToTime(fmt.Sprintf("test_%s_%d_time", name, id)).Add(time.Hour),
	}
	assert.NoError(t, base.DB.Model(&class).Association("ProblemSets").Append(&problemSet))
	assert.NoError(t, base.DB.Model(&problemSet).Association("Problems").Append(problems))
	return &problemSet
}

func TestCloneProblemSet(t *testing.T) {
	t.Parallel()

	class1 := createClassForTest(t, "clone_problem_set_fail", 1, nil, nil)
	class2 := createClassForTest(t, "clone_problem_set_fail", 2, nil, nil)
	problemSet1 := createProblemSetForTest(t, "clone_problem_set_fail", 1, &class1, nil)
	problemSet2 := createProblemSetForTest(t, "clone_problem_set_fail", 2, &class2, nil)

	failTests := []failTest{
		{
			name:       "WithoutParams",
			method:     "POST",
			path:       base.Echo.Reverse("problemSet.cloneProblemSet", class1.ID),
			req:        request.CloneProblemSetRequest{},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusBadRequest,
			resp: response.ErrorResp("VALIDATION_ERROR", []interface{}{
				map[string]interface{}{
					"field":       "SourceClassID",
					"reason":      "required",
					"translation": "复制源班级ID为必填字段",
				},
				map[string]interface{}{
					"field":       "SourceProblemSetID",
					"reason":      "required",
					"translation": "复制源题目组ID为必填字段",
				},
			}),
		},
		{
			name:   "NonExistingClass",
			method: "POST",
			path:   base.Echo.Reverse("problemSet.cloneProblemSet", -1),
			req: request.CloneProblemSetRequest{
				SourceClassID:      class1.ID,
				SourceProblemSetID: problemSet1.ID,
			},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("CLASS_NOT_FOUND", nil),
		},
		{
			name:   "NonExistingSourceClass",
			method: "POST",
			path:   base.Echo.Reverse("problemSet.cloneProblemSet", class1.ID),
			req: request.CloneProblemSetRequest{
				SourceClassID:      9999999, // non-existing class id
				SourceProblemSetID: problemSet1.ID,
			},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("SOURCE_NOT_FOUND", nil),
		},
		{
			name:   "NonExistingSourceProblemSet",
			method: "POST",
			path:   base.Echo.Reverse("problemSet.cloneProblemSet", class1.ID),
			req: request.CloneProblemSetRequest{
				SourceClassID:      class1.ID,
				SourceProblemSetID: problemSet2.ID,
			},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("SOURCE_NOT_FOUND", nil),
		},
		{
			name:   "PermissionDenied",
			method: "POST",
			path:   base.Echo.Reverse("problemSet.cloneProblemSet", class1.ID),
			req: request.CloneProblemSetRequest{
				SourceClassID:      class1.ID,
				SourceProblemSetID: problemSet1.ID,
			},
			reqOptions: []reqOption{applyNormalUser},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}

	runFailTests(t, failTests, "")

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		user := createUserForTest(t, "clone_problem_set_success", 0)
		class := createClassForTest(t, "clone_problem_set_success", 0, nil, nil)
		user.GrantRole("class_creator", class)

		sourceClass := createClassForTest(t, "clone_problem_set_success_source", 0, nil, nil)
		problem1 := createProblemForTest(t, "clone_problem_set_success_source", 1, nil, user)
		problem2 := createProblemForTest(t, "clone_problem_set_success_source", 2, nil, user)
		sourceProblemSet := createProblemSetForTest(t, "clone_problem_set_success_source", 0, &sourceClass, []models.Problem{problem1, problem2})
		assert.NoError(t, models.UpdateGrade(&models.Submission{
			ProblemSetId: sourceProblemSet.ID,
			UserID:       user.ID,
			ProblemID:    problem1.ID,
			Score:        10,
		}))
		assert.NoError(t, models.UpdateGrade(&models.Submission{
			ProblemSetId: sourceProblemSet.ID,
			UserID:       user.ID,
			ProblemID:    problem2.ID,
			Score:        20,
		}))
		assert.NoError(t, base.DB.Preload("Grades").First(&sourceProblemSet, sourceProblemSet.ID).Error)

		httpResp := makeResp(makeReq(t, "POST", base.Echo.Reverse("problemSet.cloneProblemSet", class.ID), request.CloneProblemSetRequest{
			SourceClassID:      sourceClass.ID,
			SourceProblemSetID: sourceProblemSet.ID,
		}, applyUser(user)))
		assert.Equal(t, http.StatusCreated, httpResp.StatusCode)

		databaseProblemSet := models.ProblemSet{}
		assert.NoError(t, base.DB.Preload("Problems").Preload("Grades").
			First(&databaseProblemSet, "name = ? and class_id = ?", "test_clone_problem_set_success_source_0_name", class.ID).Error)
		expectedProblemSet := models.ProblemSet{
			ID:          databaseProblemSet.ID,
			ClassID:     class.ID,
			Name:        "test_clone_problem_set_success_source_0_name",
			Description: "test_clone_problem_set_success_source_0_description",
			Problems: []*models.Problem{
				&problem1,
				&problem2,
			},
			Grades:    []*models.Grade{},
			StartTime: hashStringToTime("test_clone_problem_set_success_source_0_time"),
			EndTime:   hashStringToTime("test_clone_problem_set_success_source_0_time").Add(time.Hour),
			CreatedAt: databaseProblemSet.CreatedAt,
			UpdatedAt: databaseProblemSet.UpdatedAt,
			DeletedAt: gorm.DeletedAt{},
		}
		assert.Equal(t, expectedProblemSet, databaseProblemSet)

		resp := response.CloneProblemSetResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.CloneProblemSetResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.ProblemSetDetail `json:"problem_set"`
			}{
				resource.GetProblemSetDetail(&expectedProblemSet),
			},
		}, resp)
	})
}

func TestGetProblemSet(t *testing.T) {
	t.Parallel()

	failClass := createClassForTest(t, "get_problem_set_fail", 0, nil, nil)
	failProblemSet := createProblemSetForTest(t, "get_problem_set_fail", 0, &failClass, nil)
	failTests := []failTest{
		{
			name:       "NonExistingClass",
			method:     "GET",
			path:       base.Echo.Reverse("problemSet.getProblemSet", -1, failProblemSet.ID),
			req:        request.GetProblemSetRequest{},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("CLASS_NOT_FOUND", nil),
		},
		{
			name:       "NonExistingProblemSet",
			method:     "GET",
			path:       base.Echo.Reverse("problemSet.getProblemSet", failClass.ID, -1),
			req:        request.GetProblemSetRequest{},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:       "PermissionDenied",
			method:     "GET",
			path:       base.Echo.Reverse("problemSet.getProblemSet", failClass.ID, failProblemSet.ID),
			req:        request.GetProblemSetRequest{},
			reqOptions: []reqOption{applyNormalUser},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}
	runFailTests(t, failTests, "")

	user := createUserForTest(t, "get_problem_set_success", 0)
	student := createUserForTest(t, "get_problem_set_success_student", 0)
	problem1 := createProblemForTest(t, "get_problem_set_success", 1, nil, user)
	problem2 := createProblemForTest(t, "get_problem_set_success", 2, nil, user)
	class := createClassForTest(t, "get_problem_set_success", 0, nil, []*models.User{&student})
	problemSetInProgress := createProblemSetForTest(t, "get_problem_set_success_in_progress", 0, &class, []models.Problem{problem1, problem2})
	assert.NoError(t, models.UpdateGrade(&models.Submission{
		ProblemSetId: problemSetInProgress.ID,
		UserID:       user.ID,
		ProblemID:    problem1.ID,
		Score:        10,
	}))
	assert.NoError(t, models.UpdateGrade(&models.Submission{
		ProblemSetId: problemSetInProgress.ID,
		UserID:       user.ID,
		ProblemID:    problem2.ID,
		Score:        20,
	}))
	assert.NoError(t, base.DB.Preload("Grades").First(&problemSetInProgress, problemSetInProgress.ID).Error)
	problemSetInProgress.StartTime = time.Now().Add(-1 * time.Hour)
	problemSetInProgress.EndTime = time.Now().Add(time.Hour)
	assert.NoError(t, base.DB.Save(&problemSetInProgress).Error)
	problemSetNotYetStarted := createProblemSetForTest(t, "get_problem_set_success_not_yet_started", 0, &class, []models.Problem{problem1, problem2})
	assert.NoError(t, models.UpdateGrade(&models.Submission{
		ProblemSetId: problemSetNotYetStarted.ID,
		UserID:       user.ID,
		ProblemID:    problem1.ID,
		Score:        50,
	}))
	assert.NoError(t, models.UpdateGrade(&models.Submission{
		ProblemSetId: problemSetNotYetStarted.ID,
		UserID:       user.ID,
		ProblemID:    problem2.ID,
		Score:        60,
	}))
	assert.NoError(t, base.DB.Preload("Grades").First(&problemSetInProgress, problemSetInProgress.ID).Error)
	problemSetNotYetStarted.StartTime = time.Now().Add(time.Hour)
	problemSetNotYetStarted.EndTime = time.Now().Add(2 * time.Hour)
	assert.NoError(t, base.DB.Save(&problemSetNotYetStarted).Error)
	user.GrantRole("class_creator", class)

	t.Run("SuccessAdmin", func(t *testing.T) {
		t.Parallel()

		httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("problemSet.getProblemSet", class.ID, problemSetInProgress.ID),
			request.GetProblemSetRequest{}, applyUser(user)))
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		resp := response.GetProblemSetResponseForAdmin{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.GetProblemSetResponseForAdmin{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.ProblemSetDetail `json:"problem_set"`
			}{
				resource.GetProblemSetDetail(problemSetInProgress),
			},
		}, resp)
	})
	t.Run("SuccessInProgressStudent", func(t *testing.T) {
		t.Parallel()

		httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("problemSet.getProblemSet", class.ID, problemSetInProgress.ID),
			request.GetProblemSetRequest{}, applyUser(student)))
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		resp := response.GetProblemSetResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.GetProblemSetResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.ProblemSet `json:"problem_set"`
			}{
				resource.GetProblemSet(problemSetInProgress),
			},
		}, resp)
	})
	t.Run("SuccessInProgressStudent", func(t *testing.T) {
		t.Parallel()

		httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("problemSet.getProblemSet", class.ID, problemSetNotYetStarted.ID),
			request.GetProblemSetRequest{}, applyUser(student)))
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		expectedProblemSetResource := resource.GetProblemSet(problemSetNotYetStarted)
		expectedProblemSetResource.Problems = []resource.Problem{}
		resp := response.GetProblemSetResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.GetProblemSetResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.ProblemSet `json:"problem_set"`
			}{
				expectedProblemSetResource,
			},
		}, resp)
	})

}

func TestUpdateProblemSet(t *testing.T) {
	t.Parallel()

	failClass := createClassForTest(t, "update_problem_set_fail", 0, nil, nil)
	failProblemSet := createProblemSetForTest(t, "update_problem_set_fail", 0, &failClass, nil)
	failTests := []failTest{
		{
			name:       "WithoutParams",
			method:     "PUT",
			path:       base.Echo.Reverse("problemSet.updateProblemSet", failClass.ID),
			req:        request.UpdateProblemSetRequest{},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusBadRequest,
			resp: response.ErrorResp("VALIDATION_ERROR", []interface{}{
				map[string]interface{}{
					"field":       "Name",
					"reason":      "required",
					"translation": "名称为必填字段",
				},
				map[string]interface{}{
					"field":       "Description",
					"reason":      "required",
					"translation": "描述为必填字段",
				},
				map[string]interface{}{
					"field":       "StartTime",
					"reason":      "required",
					"translation": "开始时间为必填字段",
				},
				map[string]interface{}{
					"field":       "EndTime",
					"reason":      "required",
					"translation": "结束时间为必填字段",
				},
			}),
		},
		{
			name:   "NonExistingClass",
			method: "PUT",
			path:   base.Echo.Reverse("problemSet.updateProblemSet", -1, failProblemSet.ID),
			req: request.UpdateProblemSetRequest{
				Name:        "test_update_problem_set_non_existing_class_name",
				Description: "test_update_problem_set_non_existing_class_description",
				StartTime:   hashStringToTime("test_update_problem_set_non_existing_class_time"),
				EndTime:     hashStringToTime("test_update_problem_set_non_existing_class_time").Add(time.Hour),
			},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "NonExistingProblemSet",
			method: "PUT",
			path:   base.Echo.Reverse("problemSet.updateProblemSet", failClass.ID, -1),
			req: request.UpdateProblemSetRequest{
				Name:        "test_update_problem_set_non_existing_problem_set_name",
				Description: "test_update_problem_set_non_existing_problem_set_description",
				StartTime:   hashStringToTime("test_update_problem_set_non_existing_problem_set_time"),
				EndTime:     hashStringToTime("test_update_problem_set_non_existing_problem_set_time").Add(time.Hour),
			},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "PermissionDenied",
			method: "PUT",
			path:   base.Echo.Reverse("problemSet.updateProblemSet", failClass.ID, failProblemSet.ID),
			req: request.UpdateProblemSetRequest{
				Name:        "test_update_problem_set_permission_denied_name",
				Description: "test_update_problem_set_permission_denied_description",
				StartTime:   hashStringToTime("test_update_problem_set_permission_denied_time"),
				EndTime:     hashStringToTime("test_update_problem_set_permission_denied_time").Add(time.Hour),
			},
			reqOptions: []reqOption{applyNormalUser},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}
	runFailTests(t, failTests, "")

	t.Run("Success", func(t *testing.T) {
		user := createUserForTest(t, "update_problem_set_success", 0)
		problem1 := createProblemForTest(t, "update_problem_set_success", 1, nil, user)
		problem2 := createProblemForTest(t, "update_problem_set_success", 2, nil, user)
		class := createClassForTest(t, "update_problem_set_success", 0, nil, nil)
		problemSet := createProblemSetForTest(t, "update_problem_set_success", 0, &class, []models.Problem{problem1, problem2})
		assert.NoError(t, models.UpdateGrade(&models.Submission{
			ProblemSetId: problemSet.ID,
			UserID:       user.ID,
			ProblemID:    problem1.ID,
			Score:        10,
		}))
		assert.NoError(t, models.UpdateGrade(&models.Submission{
			ProblemSetId: problemSet.ID,
			UserID:       user.ID,
			ProblemID:    problem2.ID,
			Score:        20,
		}))
		assert.NoError(t, base.DB.Preload("Grades").First(&problemSet, problemSet.ID).Error)
		assert.NoError(t, base.DB.Save(&problemSet).Error)
		user.GrantRole("class_creator", class)

		httpResp := makeResp(makeReq(t, "PUT", base.Echo.Reverse("problemSet.updateProblemSet", class.ID, problemSet.ID),
			request.UpdateProblemSetRequest{
				Name:        "test_update_problem_set_success_00_name",
				Description: "test_update_problem_set_success_00_description",
				StartTime:   hashStringToTime("test_update_problem_set_success_00_time"),
				EndTime:     hashStringToTime("test_update_problem_set_success_00_time").Add(time.Hour),
			}, applyUser(user)))
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)

		databaseProblemSet := models.ProblemSet{}
		assert.NoError(t, base.DB.Preload("Problems").Preload("Grades").First(&databaseProblemSet, problemSet.ID).Error)
		expectedProblemSet := models.ProblemSet{
			ID:          databaseProblemSet.ID,
			ClassID:     class.ID,
			Name:        "test_update_problem_set_success_00_name",
			Description: "test_update_problem_set_success_00_description",
			Problems:    problemSet.Problems,
			Grades:      problemSet.Grades,
			StartTime:   hashStringToTime("test_update_problem_set_success_00_time"),
			EndTime:     hashStringToTime("test_update_problem_set_success_00_time").Add(time.Hour),
			CreatedAt:   databaseProblemSet.CreatedAt,
			UpdatedAt:   databaseProblemSet.UpdatedAt,
			DeletedAt:   gorm.DeletedAt{},
		}
		assert.Equal(t, expectedProblemSet, databaseProblemSet)
		resp := response.UpdateProblemSetResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.UpdateProblemSetResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.ProblemSetDetail `json:"problem_set"`
			}{
				resource.GetProblemSetDetail(&expectedProblemSet),
			},
		}, resp)
	})
}

func TestAddProblemsToSetProblemSet(t *testing.T) {
	t.Parallel()

	failClass := createClassForTest(t, "add_problems_to_set_fail", 0, nil, nil)
	failProblemSet := createProblemSetForTest(t, "add_problems_to_set_fail", 0, &failClass, nil)
	failTests := []failTest{
		{
			name:       "WithoutParams",
			method:     "POST",
			path:       base.Echo.Reverse("problemSet.addProblemsToSet", failClass.ID),
			req:        request.AddProblemsToSetRequest{},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusBadRequest,
			resp: response.ErrorResp("VALIDATION_ERROR", []interface{}{
				map[string]interface{}{
					"field":       "ProblemIDs",
					"reason":      "required",
					"translation": "题目ID数组为必填字段",
				},
			}),
		},
		{
			name:   "NonExistingClass",
			method: "POST",
			path:   base.Echo.Reverse("problemSet.addProblemsToSet", -1, failProblemSet.ID),
			req: request.AddProblemsToSetRequest{
				ProblemIDs: []uint{0},
			},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "NonExistingProblemSet",
			method: "POST",
			path:   base.Echo.Reverse("problemSet.addProblemsToSet", failClass.ID, -1),
			req: request.AddProblemsToSetRequest{
				ProblemIDs: []uint{0},
			},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "PermissionDenied",
			method: "POST",
			path:   base.Echo.Reverse("problemSet.addProblemsToSet", failClass.ID, failProblemSet.ID),
			req: request.AddProblemsToSetRequest{
				ProblemIDs: []uint{0},
			},
			reqOptions: []reqOption{applyNormalUser},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}
	runFailTests(t, failTests, "")

	t.Run("Success", func(t *testing.T) {
		user := createUserForTest(t, "add_problems_to_set_success", 0)
		problem1 := createProblemForTest(t, "add_problems_to_set_success", 1, nil, user)
		problem2 := createProblemForTest(t, "add_problems_to_set_success", 2, nil, user)
		problem3 := createProblemForTest(t, "add_problems_to_set_success", 3, nil, user)
		class := createClassForTest(t, "add_problems_to_set_success", 0, nil, nil)
		problemSet := createProblemSetForTest(t, "add_problems_to_set_success", 0, &class, []models.Problem{problem1})
		assert.NoError(t, models.UpdateGrade(&models.Submission{
			ProblemSetId: problemSet.ID,
			UserID:       user.ID,
			ProblemID:    problem1.ID,
			Score:        10,
		}))
		assert.NoError(t, models.UpdateGrade(&models.Submission{
			ProblemSetId: problemSet.ID,
			UserID:       user.ID,
			ProblemID:    problem2.ID,
			Score:        20,
		}))
		assert.NoError(t, base.DB.Preload("Grades").First(&problemSet, problemSet.ID).Error)
		assert.NoError(t, base.DB.Save(&problemSet).Error)
		user.GrantRole("class_creator", class)

		httpResp := makeResp(makeReq(t, "POST", base.Echo.Reverse("problemSet.addProblemsToSet", class.ID, problemSet.ID),
			request.AddProblemsToSetRequest{
				ProblemIDs: []uint{
					problem1.ID,
					problem2.ID,
					problem3.ID,
					0,
				},
			}, applyUser(user)))
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)

		databaseProblemSet := models.ProblemSet{}
		assert.NoError(t, base.DB.Preload("Problems").Preload("Grades").First(&databaseProblemSet, problemSet.ID).Error)
		expectedProblemSet := models.ProblemSet{
			ID:          databaseProblemSet.ID,
			ClassID:     class.ID,
			Name:        "test_add_problems_to_set_success_0_name",
			Description: "test_add_problems_to_set_success_0_description",
			Problems: []*models.Problem{
				&problem1,
				&problem2,
				&problem3,
			},
			Grades:    problemSet.Grades,
			StartTime: hashStringToTime("test_add_problems_to_set_success_0_time"),
			EndTime:   hashStringToTime("test_add_problems_to_set_success_0_time").Add(time.Hour),
			CreatedAt: databaseProblemSet.CreatedAt,
			UpdatedAt: databaseProblemSet.UpdatedAt,
			DeletedAt: gorm.DeletedAt{},
		}
		assert.Equal(t, expectedProblemSet, databaseProblemSet)
		resp := response.AddProblemsToSetResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.AddProblemsToSetResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.ProblemSetDetail `json:"problem_set"`
			}{
				resource.GetProblemSetDetail(&expectedProblemSet),
			},
		}, resp)
	})
}

func TestDeleteProblemsFromSetProblemSet(t *testing.T) {
	t.Parallel()

	failClass := createClassForTest(t, "delete_problems_from_set_fail", 0, nil, nil)
	failProblemSet := createProblemSetForTest(t, "delete_problems_from_set_fail", 0, &failClass, nil)
	failTests := []failTest{
		{
			name:       "WithoutParams",
			method:     "DELETE",
			path:       base.Echo.Reverse("problemSet.deleteProblemsFromSet", failClass.ID),
			req:        request.DeleteProblemsFromSetRequest{},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusBadRequest,
			resp: response.ErrorResp("VALIDATION_ERROR", []interface{}{
				map[string]interface{}{
					"field":       "ProblemIDs",
					"reason":      "required",
					"translation": "题目ID数组为必填字段",
				},
			}),
		},
		{
			name:   "NonExistingClass",
			method: "DELETE",
			path:   base.Echo.Reverse("problemSet.deleteProblemsFromSet", -1, failProblemSet.ID),
			req: request.AddProblemsToSetRequest{
				ProblemIDs: []uint{0},
			},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "NonExistingProblemSet",
			method: "DELETE",
			path:   base.Echo.Reverse("problemSet.deleteProblemsFromSet", failClass.ID, -1),
			req: request.AddProblemsToSetRequest{
				ProblemIDs: []uint{0},
			},
			reqOptions: []reqOption{applyAdminUser},
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("NOT_FOUND", nil),
		},
		{
			name:   "PermissionDenied",
			method: "DELETE",
			path:   base.Echo.Reverse("problemSet.deleteProblemsFromSet", failClass.ID, failProblemSet.ID),
			req: request.AddProblemsToSetRequest{
				ProblemIDs: []uint{0},
			},
			reqOptions: []reqOption{applyNormalUser},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}
	runFailTests(t, failTests, "")

	t.Run("Success", func(t *testing.T) {
		user := createUserForTest(t, "delete_problems_from_set_success", 0)
		problem1 := createProblemForTest(t, "delete_problems_from_set_success", 1, nil, user)
		problem2 := createProblemForTest(t, "delete_problems_from_set_success", 2, nil, user)
		problem3 := createProblemForTest(t, "delete_problems_from_set_success", 3, nil, user)
		class := createClassForTest(t, "delete_problems_from_set_success", 0, nil, nil)
		problemSet := createProblemSetForTest(t, "delete_problems_from_set_success", 0, &class, []models.Problem{problem1, problem2})
		assert.NoError(t, models.UpdateGrade(&models.Submission{
			ProblemSetId: problemSet.ID,
			UserID:       user.ID,
			ProblemID:    problem1.ID,
			Score:        10,
		}))
		assert.NoError(t, models.UpdateGrade(&models.Submission{
			ProblemSetId: problemSet.ID,
			UserID:       user.ID,
			ProblemID:    problem2.ID,
			Score:        20,
		}))
		assert.NoError(t, base.DB.Preload("Grades").First(&problemSet, problemSet.ID).Error)
		assert.NoError(t, base.DB.Save(&problemSet).Error)
		user.GrantRole("class_creator", class)

		httpResp := makeResp(makeReq(t, "DELETE", base.Echo.Reverse("problemSet.deleteProblemsFromSet", class.ID, problemSet.ID),
			request.AddProblemsToSetRequest{
				ProblemIDs: []uint{
					problem2.ID,
					problem3.ID,
					0,
				},
			}, applyUser(user)))
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)

		databaseProblemSet := models.ProblemSet{}
		assert.NoError(t, base.DB.Preload("Problems").Preload("Grades").First(&databaseProblemSet, problemSet.ID).Error)
		expectedProblemSet := models.ProblemSet{
			ID:          databaseProblemSet.ID,
			ClassID:     class.ID,
			Name:        "test_delete_problems_from_set_success_0_name",
			Description: "test_delete_problems_from_set_success_0_description",
			Problems: []*models.Problem{
				&problem1,
			},
			Grades:    problemSet.Grades,
			StartTime: hashStringToTime("test_delete_problems_from_set_success_0_time"),
			EndTime:   hashStringToTime("test_delete_problems_from_set_success_0_time").Add(time.Hour),
			CreatedAt: databaseProblemSet.CreatedAt,
			UpdatedAt: databaseProblemSet.UpdatedAt,
			DeletedAt: gorm.DeletedAt{},
		}
		assert.Equal(t, expectedProblemSet, databaseProblemSet)
		resp := response.AddProblemsToSetResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, response.AddProblemsToSetResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.ProblemSetDetail `json:"problem_set"`
			}{
				resource.GetProblemSetDetail(&expectedProblemSet),
			},
		}, resp)
	})
}

func TestDeleteProblemSet(t *testing.T) {
	t.Parallel()

	failClass := createClassForTest(t, "delete_problem_set_fail", 0, nil, nil)
	failProblemSet := createProblemSetForTest(t, "delete_problem_set_fail", 0, &failClass, nil)
	failTests := []failTest{
		{
			name:   "PermissionDenied",
			method: "DELETE",
			path:   base.Echo.Reverse("problemSet.deleteProblemSet", failClass.ID, failProblemSet.ID),
			req: request.UpdateProblemSetRequest{
				Name:        "test_delete_problem_set_permission_denied_name",
				Description: "test_delete_problem_set_permission_denied_description",
				StartTime:   hashStringToTime("test_delete_problem_set_permission_denied_time"),
				EndTime:     hashStringToTime("test_delete_problem_set_permission_denied_time").Add(time.Hour),
			},
			reqOptions: []reqOption{applyNormalUser},
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("PERMISSION_DENIED", nil),
		},
	}
	runFailTests(t, failTests, "")

	t.Run("Success", func(t *testing.T) {
		user := createUserForTest(t, "delete_problem_set_success", 0)
		problem1 := createProblemForTest(t, "delete_problem_set_success", 1, nil, user)
		problem2 := createProblemForTest(t, "delete_problem_set_success", 2, nil, user)
		class := createClassForTest(t, "delete_problem_set_success", 0, nil, nil)
		user.GrantRole("class_creator", class)
		problemSet := createProblemSetForTest(t, "delete_problem_set_success", 0, &class, []models.Problem{problem1, problem2})
		assert.NoError(t, models.UpdateGrade(&models.Submission{
			ProblemSetId: problemSet.ID,
			UserID:       user.ID,
			ProblemID:    problem1.ID,
			Score:        10,
		}))
		assert.NoError(t, models.UpdateGrade(&models.Submission{
			ProblemSetId: problemSet.ID,
			UserID:       user.ID,
			ProblemID:    problem2.ID,
			Score:        20,
		}))
		assert.NoError(t, base.DB.Preload("Grades").First(&problemSet, problemSet.ID).Error)
		httpResp := makeResp(makeReq(t, "DELETE", base.Echo.Reverse("problemSet.deleteProblemSet", class.ID, problemSet.ID),
			request.DeleteProblemSetRequest{}, applyUser(user)))
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		databasePS := models.ProblemSet{}
		err := base.DB.First(&databasePS, problemSet.ID).Error
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		err = base.DB.First(&models.Grade{}, "problem_set_id = ?", problemSet.ID).Error
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}
