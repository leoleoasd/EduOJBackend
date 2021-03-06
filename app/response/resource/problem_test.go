package resource_test

import (
	"fmt"
	"github.com/EduOJ/backend/app/response/resource"
	"github.com/EduOJ/backend/database/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
	"time"
)

func createTestCaseForTest(name string, problemId uint, id uint) models.TestCase {
	return models.TestCase{
		ID:             id,
		ProblemID:      problemId,
		Score:          id,
		Sample:         true,
		InputFileName:  fmt.Sprintf("test_%s_problem_%d_test_case_%d_input", name, problemId, id),
		OutputFileName: fmt.Sprintf("test_%s_problem_%d_test_case_%d_output", name, problemId, id),
		CreatedAt:      time.Date(int(problemId), 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),
		UpdatedAt:      time.Date(int(problemId), 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
		DeletedAt:      gorm.DeletedAt{},
	}
}

func createProblemForTest(name string, id uint, testCaseCount uint) (problem models.Problem) {
	problem = models.Problem{
		ID:                 id,
		Name:               fmt.Sprintf("test_%s_problem_%d", name, id),
		Description:        fmt.Sprintf("test_%s_problem_%d_desc", name, id),
		AttachmentFileName: fmt.Sprintf("test_%s_problem_%d_attachment", name, id),
		Public:             true,
		Privacy:            false,
		MemoryLimit:        1024,
		TimeLimit:          1000,
		LanguageAllowed:    []string{fmt.Sprintf("test_%s_language_allowed_%d", name, id), "test_language"},
		BuildArg:           fmt.Sprintf("test_%s_build_arg_%d", name, id),
		CompareScriptName:  "cmp1",
		TestCases:          make([]models.TestCase, testCaseCount),
		CreatedAt:          time.Date(int(id), 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),
		UpdatedAt:          time.Date(int(id), 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
		DeletedAt:          gorm.DeletedAt{},
	}
	for i := range problem.TestCases {
		problem.TestCases[i] = createTestCaseForTest(name, id, uint(i))
	}
	return
}

func TestGetTestCaseAndGetTestCaseForAdmin(t *testing.T) {
	testCase := createTestCaseForTest("get_test_case", 0, 0)
	t.Run("testGetTestCase", func(t *testing.T) {
		actualTestCase := resource.GetTestCase(&testCase)
		expectedTestCase := resource.TestCase{
			ID:        0,
			ProblemID: 0,
			Score:     0,
			Sample:    true,
		}
		assert.Equal(t, expectedTestCase, *actualTestCase)
	})
	t.Run("testGetTestCaseForAdmin", func(t *testing.T) {
		actualTestCase := resource.GetTestCaseForAdmin(&testCase)
		expectedTestCase := resource.TestCaseForAdmin{
			ID:             0,
			ProblemID:      0,
			Score:          0,
			Sample:         true,
			InputFileName:  "test_get_test_case_problem_0_test_case_0_input",
			OutputFileName: "test_get_test_case_problem_0_test_case_0_output",
		}
		assert.Equal(t, expectedTestCase, *actualTestCase)
	})
}

func TestGetProblemAndGetProblemForAdmin(t *testing.T) {
	problem := createProblemForTest("get_problem", 0, 2)
	t.Run("testGetProblem", func(t *testing.T) {
		actualProblem := resource.GetProblem(&problem)
		expectedProblem := resource.Problem{
			ID:                 0,
			Name:               "test_get_problem_problem_0",
			Description:        "test_get_problem_problem_0_desc",
			AttachmentFileName: "test_get_problem_problem_0_attachment",
			MemoryLimit:        1024,
			TimeLimit:          1000,
			LanguageAllowed:    []string{"test_get_problem_language_allowed_0", "test_language"},
			CompareScriptName:  "cmp1",
			TestCases: []resource.TestCase{
				{
					ID:        0,
					ProblemID: 0,
					Score:     0,
					Sample:    true,
				}, {
					ID:        1,
					ProblemID: 0,
					Score:     1,
					Sample:    true,
				},
			},
		}
		assert.Equal(t, expectedProblem, *actualProblem)
	})
	t.Run("testGetProblemForAdmin", func(t *testing.T) {
		actualProblem := resource.GetProblemForAdmin(&problem)
		expectedProblem := resource.ProblemForAdmin{
			ID:                 0,
			Name:               "test_get_problem_problem_0",
			Description:        "test_get_problem_problem_0_desc",
			AttachmentFileName: "test_get_problem_problem_0_attachment",
			Public:             true,
			Privacy:            false,
			MemoryLimit:        1024,
			TimeLimit:          1000,
			LanguageAllowed:    []string{"test_get_problem_language_allowed_0", "test_language"},
			BuildArg:           "test_get_problem_build_arg_0",
			CompareScriptName:  "cmp1",
			TestCases: []resource.TestCaseForAdmin{
				{
					ID:             0,
					ProblemID:      0,
					Score:          0,
					Sample:         true,
					InputFileName:  "test_get_problem_problem_0_test_case_0_input",
					OutputFileName: "test_get_problem_problem_0_test_case_0_output",
				}, {
					ID:             1,
					ProblemID:      0,
					Score:          1,
					Sample:         true,
					InputFileName:  "test_get_problem_problem_0_test_case_1_input",
					OutputFileName: "test_get_problem_problem_0_test_case_1_output",
				},
			},
		}
		assert.Equal(t, expectedProblem, *actualProblem)
	})
}

func TestGetProblemSliceAndGetProblemForAdminSlice(t *testing.T) {
	problem1 := createProblemForTest("get_problem_slice", 1, 1)
	problem2 := createProblemForTest("get_problem_slice", 2, 2)
	t.Run("testGetProblemSlice", func(t *testing.T) {
		actualProblemSlice := resource.GetProblemSlice([]*models.Problem{&problem1, &problem2})
		expectedProblemSlice := []resource.Problem{
			{
				ID:                 1,
				Name:               "test_get_problem_slice_problem_1",
				Description:        "test_get_problem_slice_problem_1_desc",
				AttachmentFileName: "test_get_problem_slice_problem_1_attachment",
				MemoryLimit:        1024,
				TimeLimit:          1000,
				LanguageAllowed:    []string{"test_get_problem_slice_language_allowed_1", "test_language"},
				CompareScriptName:  "cmp1",
				TestCases: []resource.TestCase{
					{
						ID:        0,
						ProblemID: 1,
						Score:     0,
						Sample:    true,
					},
				},
			}, {
				ID:                 2,
				Name:               "test_get_problem_slice_problem_2",
				Description:        "test_get_problem_slice_problem_2_desc",
				AttachmentFileName: "test_get_problem_slice_problem_2_attachment",
				MemoryLimit:        1024,
				TimeLimit:          1000,
				LanguageAllowed:    []string{"test_get_problem_slice_language_allowed_2", "test_language"},
				CompareScriptName:  "cmp1",
				TestCases: []resource.TestCase{
					{
						ID:        0,
						ProblemID: 2,
						Score:     0,
						Sample:    true,
					}, {
						ID:        1,
						ProblemID: 2,
						Score:     1,
						Sample:    true,
					},
				},
			},
		}
		assert.Equal(t, expectedProblemSlice, actualProblemSlice)
	})
	t.Run("testGetProblemForAdminSlice", func(t *testing.T) {
		actualProblemSlice := resource.GetProblemForAdminSlice([]*models.Problem{&problem1, &problem2})
		expectedProblemSlice := []resource.ProblemForAdmin{
			{
				ID:                 1,
				Name:               "test_get_problem_slice_problem_1",
				Description:        "test_get_problem_slice_problem_1_desc",
				AttachmentFileName: "test_get_problem_slice_problem_1_attachment",
				Public:             true,
				Privacy:            false,
				MemoryLimit:        1024,
				TimeLimit:          1000,
				LanguageAllowed:    []string{"test_get_problem_slice_language_allowed_1", "test_language"},
				BuildArg:           "test_get_problem_slice_build_arg_1",
				CompareScriptName:  "cmp1",
				TestCases: []resource.TestCaseForAdmin{
					{
						ID:             0,
						ProblemID:      1,
						Score:          0,
						Sample:         true,
						InputFileName:  "test_get_problem_slice_problem_1_test_case_0_input",
						OutputFileName: "test_get_problem_slice_problem_1_test_case_0_output",
					},
				},
			}, {
				ID:                 2,
				Name:               "test_get_problem_slice_problem_2",
				Description:        "test_get_problem_slice_problem_2_desc",
				AttachmentFileName: "test_get_problem_slice_problem_2_attachment",
				Public:             true,
				Privacy:            false,
				MemoryLimit:        1024,
				TimeLimit:          1000,
				LanguageAllowed:    []string{"test_get_problem_slice_language_allowed_2", "test_language"},
				BuildArg:           "test_get_problem_slice_build_arg_2",
				CompareScriptName:  "cmp1",
				TestCases: []resource.TestCaseForAdmin{
					{
						ID:             0,
						ProblemID:      2,
						Score:          0,
						Sample:         true,
						InputFileName:  "test_get_problem_slice_problem_2_test_case_0_input",
						OutputFileName: "test_get_problem_slice_problem_2_test_case_0_output",
					}, {
						ID:             1,
						ProblemID:      2,
						Score:          1,
						Sample:         true,
						InputFileName:  "test_get_problem_slice_problem_2_test_case_1_input",
						OutputFileName: "test_get_problem_slice_problem_2_test_case_1_output",
					},
				},
			},
		}
		assert.Equal(t, expectedProblemSlice, actualProblemSlice)
	})
}
