package models

import (
	"encoding/json"
	"fmt"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/database"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
	"hash/fnv"
	"testing"
	"time"
)

func hashStringToTime(s string) time.Time {
	h := fnv.New32()
	if _, err := h.Write([]byte(s)); err != nil {
		panic(err)
	}
	return time.Unix(int64(h.Sum32()), 0).UTC()
}

func createJSONForTest(t *testing.T, in interface{}) datatypes.JSON {
	j := datatypes.JSON{}
	b, err := json.Marshal(in)
	assert.NoError(t, err)
	assert.NoError(t, j.UnmarshalJSON(b))
	return j
}

func createProblemForTest(t *testing.T, name string, id uint) *Problem {
	problem := Problem{
		Name:        fmt.Sprintf("%s_%d_name", name, id),
		Description: fmt.Sprintf("%s_%d_description", name, id),
		TestCases: []TestCase{
			{
				Score:          10,
				Sample:         true,
				InputFileName:  fmt.Sprintf("%s_%d_1.in", name, id),
				OutputFileName: fmt.Sprintf("%s_%d_1.out", name, id),
			},
			{
				Score:          20,
				Sample:         false,
				InputFileName:  fmt.Sprintf("%s_%d_2.in", name, id),
				OutputFileName: fmt.Sprintf("%s_%d_2.out", name, id),
			},
		},
		LanguageAllowed: database.StringArray([]string{""}),
	}
	assert.NoError(t, base.DB.Create(&problem).Error)
	return &problem
}

func TestAddProblemsAndDeleteProblemsByID(t *testing.T) {
	t.Parallel()
	//problem1 := Problem{
	//	Name:        "test_add_and_delete_problems_1_name",
	//	Description: "test_add_and_delete_problems_1_description",
	//	TestCases: []TestCase{
	//		{
	//			Score:          10,
	//			Sample:         true,
	//			InputFileName:  "test_add_and_delete_problems_1_test_case_1.in",
	//			OutputFileName: "test_add_and_delete_problems_1_test_case_1.out",
	//		},
	//		{
	//			Score:          20,
	//			Sample:         false,
	//			InputFileName:  "test_add_and_delete_problems_1_test_case_2.in",
	//			OutputFileName: "test_add_and_delete_problems_1_test_case_2.out",
	//		},
	//	},
	//	LanguageAllowed: database.StringArray([]string{""}),
	//}
	//problem2 := Problem{
	//	Name:        "test_add_and_delete_problems_2_name",
	//	Description: "test_add_and_delete_problems_2_description",
	//	TestCases: []TestCase{
	//		{
	//			Score:          10,
	//			Sample:         true,
	//			InputFileName:  "test_add_and_delete_problems_2_test_case_1.in",
	//			OutputFileName: "test_add_and_delete_problems_2_test_case_1.out",
	//		},
	//		{
	//			Score:          20,
	//			Sample:         false,
	//			InputFileName:  "test_add_and_delete_problems_2_test_case_2.in",
	//			OutputFileName: "test_add_and_delete_problems_2_test_case_2.out",
	//		},
	//	},
	//	LanguageAllowed: database.StringArray([]string{""}),
	//}
	//problem3 := Problem{
	//	Name:        "test_add_and_delete_problems_3_name",
	//	Description: "test_add_and_delete_problems_3_description",
	//	TestCases: []TestCase{
	//		{
	//			Score:          10,
	//			Sample:         true,
	//			InputFileName:  "test_add_and_delete_problems_3_test_case_1.in",
	//			OutputFileName: "test_add_and_delete_problems_3_test_case_1.out",
	//		},
	//		{
	//			Score:          20,
	//			Sample:         false,
	//			InputFileName:  "test_add_and_delete_problems_3_test_case_2.in",
	//			OutputFileName: "test_add_and_delete_problems_3_test_case_2.out",
	//		},
	//	},
	//	LanguageAllowed: database.StringArray([]string{""}),
	//}
	//problem4 := Problem{
	//	Name:        "test_add_and_delete_problems_4_name",
	//	Description: "test_add_and_delete_problems_4_description",
	//	TestCases: []TestCase{
	//		{
	//			Score:          10,
	//			Sample:         true,
	//			InputFileName:  "test_add_and_delete_problems_4_test_case_1.in",
	//			OutputFileName: "test_add_and_delete_problems_4_test_case_1.out",
	//		},
	//		{
	//			Score:          20,
	//			Sample:         false,
	//			InputFileName:  "test_add_and_delete_problems_4_test_case_2.in",
	//			OutputFileName: "test_add_and_delete_problems_4_test_case_2.out",
	//		},
	//	},
	//	LanguageAllowed: database.StringArray([]string{""}),
	//}
	//assert.NoError(t, base.DB.Create(&problem1).Error)
	//assert.NoError(t, base.DB.Create(&problem2).Error)
	//assert.NoError(t, base.DB.Create(&problem3).Error)
	//assert.NoError(t, base.DB.Create(&problem4).Error)

	t.Run("AddSuccess", func(t *testing.T) {
		problem1 := createProblemForTest(t, "add_success", 1)
		problem2 := createProblemForTest(t, "add_success", 2)
		problem3 := createProblemForTest(t, "add_success", 3)
		problem4 := createProblemForTest(t, "add_success", 4)
		t.Parallel()
		problemSet := ProblemSet{
			Name:        "test_add_students_success_problem_set_name",
			Description: "test_add_students_success_problem_set_description",
			Problems: []*Problem{
				problem1,
				problem2,
			},
			StartTime: hashStringToTime("test_add_students_success_problem_set_start_time"),
			EndTime:   hashStringToTime("test_add_students_success_problem_set_end_time"),
		}
		assert.NoError(t, base.DB.Create(&problemSet).Error)
		assert.NoError(t, problemSet.AddProblems([]uint{
			problem3.ID,
			problem4.ID,
		}))
		assert.Equal(t, ProblemSet{
			ID:          problemSet.ID,
			Name:        "test_add_students_success_problem_set_name",
			Description: "test_add_students_success_problem_set_description",
			Problems: []*Problem{
				problem1,
				problem2,
				problem3,
				problem4,
			},
			StartTime: hashStringToTime("test_add_students_success_problem_set_start_time"),
			EndTime:   hashStringToTime("test_add_students_success_problem_set_end_time"),
			CreatedAt: problemSet.CreatedAt,
			UpdatedAt: problemSet.UpdatedAt,
		}, problemSet)
	})
	t.Run("AddExistingInSet", func(t *testing.T) {
		problem1 := createProblemForTest(t, "add_existing_in_set", 1)
		problem2 := createProblemForTest(t, "add_existing_in_set", 2)
		problem3 := createProblemForTest(t, "add_existing_in_set", 3)
		t.Parallel()
		problemSet := ProblemSet{
			Name:        "test_add_students_existing_in_set_problem_set_name",
			Description: "test_add_students_existing_in_set_problem_set_description",
			Problems: []*Problem{
				problem1,
				problem2,
			},
			StartTime: hashStringToTime("test_add_students_existing_in_set_problem_set_start_time"),
			EndTime:   hashStringToTime("test_add_students_existing_in_set_problem_set_end_time"),
		}
		assert.NoError(t, base.DB.Create(&problemSet).Error)
		assert.NoError(t, problemSet.AddProblems([]uint{
			problem2.ID,
			problem3.ID,
		}))
		assert.Equal(t, ProblemSet{
			ID:          problemSet.ID,
			Name:        "test_add_students_existing_in_set_problem_set_name",
			Description: "test_add_students_existing_in_set_problem_set_description",
			Problems: []*Problem{
				problem1,
				problem2,
				problem3,
			},
			StartTime: hashStringToTime("test_add_students_existing_in_set_problem_set_start_time"),
			EndTime:   hashStringToTime("test_add_students_existing_in_set_problem_set_end_time"),
			CreatedAt: problemSet.CreatedAt,
			UpdatedAt: problemSet.UpdatedAt,
		}, problemSet)
	})
	t.Run("AddNonExisting", func(t *testing.T) {
		problem1 := createProblemForTest(t, "add_non_exist", 1)
		problem2 := createProblemForTest(t, "add_non_exist", 2)
		problem3 := createProblemForTest(t, "add_non_exist", 3)
		t.Parallel()
		problemSet := ProblemSet{
			Name:        "test_add_students_non_existing_problem_set_name",
			Description: "test_add_students_non_existing_problem_set_description",
			Problems: []*Problem{
				problem1,
				problem2,
			},
			StartTime: hashStringToTime("test_add_students_non_existing_problem_set_start_time"),
			EndTime:   hashStringToTime("test_add_students_non_existing_problem_set_end_time"),
		}
		assert.NoError(t, base.DB.Create(&problemSet).Error)
		assert.NoError(t, problemSet.AddProblems([]uint{
			0,
			problem3.ID,
		}))
		assert.Equal(t, ProblemSet{
			ID:          problemSet.ID,
			Name:        "test_add_students_non_existing_problem_set_name",
			Description: "test_add_students_non_existing_problem_set_description",
			Problems: []*Problem{
				problem1,
				problem2,
				problem3,
			},
			StartTime: hashStringToTime("test_add_students_non_existing_problem_set_start_time"),
			EndTime:   hashStringToTime("test_add_students_non_existing_problem_set_end_time"),
			CreatedAt: problemSet.CreatedAt,
			UpdatedAt: problemSet.UpdatedAt,
		}, problemSet)
	})
	t.Run("DeleteSuccess", func(t *testing.T) {
		problem1 := createProblemForTest(t, "delete_success", 1)
		problem2 := createProblemForTest(t, "delete_success", 2)
		problem3 := createProblemForTest(t, "delete_success", 3)
		t.Parallel()
		problemSet := ProblemSet{
			Name:        "test_delete_students_success_problem_set_name",
			Description: "test_delete_students_success_problem_set_description",
			Problems: []*Problem{
				problem1,
				problem2,
				problem3,
			},
			StartTime: hashStringToTime("test_delete_students_success_problem_set_start_time"),
			EndTime:   hashStringToTime("test_delete_students_success_problem_set_end_time"),
		}
		assert.NoError(t, base.DB.Create(&problemSet).Error)
		assert.NoError(t, problemSet.DeleteProblems([]uint{
			problem2.ID,
			problem3.ID,
		}))
		assert.Equal(t, ProblemSet{
			ID:          problemSet.ID,
			Name:        "test_delete_students_success_problem_set_name",
			Description: "test_delete_students_success_problem_set_description",
			Problems: []*Problem{
				problem1,
			},
			StartTime: hashStringToTime("test_delete_students_success_problem_set_start_time"),
			EndTime:   hashStringToTime("test_delete_students_success_problem_set_end_time"),
			CreatedAt: problemSet.CreatedAt,
			UpdatedAt: problemSet.UpdatedAt,
		}, problemSet)
	})
	t.Run("DeleteNotBelongTo", func(t *testing.T) {
		problem1 := createProblemForTest(t, "delete_not_belong_to", 1)
		problem2 := createProblemForTest(t, "delete_not_belong_to", 2)
		problem3 := createProblemForTest(t, "delete_not_belong_to", 3)
		t.Parallel()
		problemSet := ProblemSet{
			Name:        "test_delete_students_not_belong_to_problem_set_name",
			Description: "test_delete_students_not_belong_to_problem_set_description",
			Problems: []*Problem{
				problem1,
				problem2,
			},
			StartTime: hashStringToTime("test_delete_students_not_belong_to_problem_set_start_time"),
			EndTime:   hashStringToTime("test_delete_students_not_belong_to_problem_set_end_time"),
		}
		assert.NoError(t, base.DB.Create(&problemSet).Error)
		assert.NoError(t, problemSet.DeleteProblems([]uint{
			problem2.ID,
			problem3.ID,
		}))
		assert.Equal(t, ProblemSet{
			ID:          problemSet.ID,
			Name:        "test_delete_students_not_belong_to_problem_set_name",
			Description: "test_delete_students_not_belong_to_problem_set_description",
			Problems: []*Problem{
				problem1,
			},
			StartTime: hashStringToTime("test_delete_students_not_belong_to_problem_set_start_time"),
			EndTime:   hashStringToTime("test_delete_students_not_belong_to_problem_set_end_time"),
			CreatedAt: problemSet.CreatedAt,
			UpdatedAt: problemSet.UpdatedAt,
		}, problemSet)
	})

}

func TestUpdateGrade(t *testing.T) {
	t.Parallel()

	user1 := User{
		Username: "test_update_grade_1_username",
		Nickname: "test_update_grade_1_nickname",
		Email:    "test_update_grade_1@email.com",
		Password: "test_update_grade_1_password",
	}
	user2 := User{
		Username: "test_update_grade_2_username",
		Nickname: "test_update_grade_2_nickname",
		Email:    "test_update_grade_2@email.com",
		Password: "test_update_grade_2_password",
	}
	assert.NoError(t, base.DB.Create(&user1).Error)
	assert.NoError(t, base.DB.Create(&user2).Error)
	problem1 := Problem{
		Name:        "test_update_grade_1_name",
		Description: "test_update_grade_1_description",
		MemoryLimit: 1024,
		TimeLimit:   1000,
	}
	problem2 := Problem{
		Name:        "test_update_grade_2_name",
		Description: "test_update_grade_2_description",
		MemoryLimit: 2048,
		TimeLimit:   2000,
	}
	assert.NoError(t, base.DB.Create(&problem1).Error)
	assert.NoError(t, base.DB.Create(&problem2).Error)
	problemSet := ProblemSet{
		Name:        "test_update_grade_name",
		Description: "test_update_grade_description",
		Problems: []*Problem{
			&problem1,
			&problem2,
		},
		Grades: []*Grade{
			{
				UserID: user1.ID,
				Detail: createJSONForTest(t, map[uint]uint{
					problem1.ID: 11,
					problem2.ID: 12,
				}),
				Total: 23,
			},
		},
		StartTime: hashStringToTime("test_update_grade_start_time"),
		EndTime:   hashStringToTime("test_update_grade_end_time"),
	}
	assert.NoError(t, base.DB.Create(&problemSet).Error)
	assert.NoError(t, UpdateGrade(&Submission{
		ProblemSetId: problemSet.ID,
		UserID:       user1.ID,
		ProblemID:    problem2.ID,
		Score:        120,
	}))
	assert.NoError(t, base.DB.Preload("Grades").First(&problemSet, problemSet.ID).Error)
	assert.Equal(t, []*Grade{
		{
			ID:           problemSet.Grades[0].ID,
			UserID:       user1.ID,
			ProblemSetID: problemSet.ID,
			Detail: createJSONForTest(t, map[uint]uint{
				problem1.ID: 11,
				problem2.ID: 120,
			}),
			Total:     131,
			CreatedAt: problemSet.Grades[0].CreatedAt,
			UpdatedAt: problemSet.Grades[0].UpdatedAt,
		},
	}, problemSet.Grades)
	assert.NoError(t, UpdateGrade(&Submission{
		ProblemSetId: problemSet.ID,
		UserID:       user2.ID,
		ProblemID:    problem1.ID,
		Score:        21,
	}))
	assert.NoError(t, base.DB.Preload("Grades").First(&problemSet, problemSet.ID).Error)
	assert.Equal(t, []*Grade{
		{
			ID:           problemSet.Grades[0].ID,
			UserID:       user1.ID,
			ProblemSetID: problemSet.ID,
			Detail: createJSONForTest(t, map[uint]uint{
				problem1.ID: 11,
				problem2.ID: 120,
			}),
			Total:     131,
			CreatedAt: problemSet.Grades[0].CreatedAt,
			UpdatedAt: problemSet.Grades[0].UpdatedAt,
		},
		{
			ID:           problemSet.Grades[1].ID,
			UserID:       user2.ID,
			ProblemSetID: problemSet.ID,
			Detail: createJSONForTest(t, map[uint]uint{
				problem1.ID: 21,
			}),
			Total:     21,
			CreatedAt: problemSet.Grades[1].CreatedAt,
			UpdatedAt: problemSet.Grades[1].UpdatedAt,
		},
	}, problemSet.Grades)
	assert.NoError(t, UpdateGrade(&Submission{
		ProblemSetId: problemSet.ID,
		UserID:       user2.ID,
		ProblemID:    problem2.ID,
		Score:        22,
	}))
	assert.NoError(t, base.DB.Preload("Grades").First(&problemSet, problemSet.ID).Error)
	assert.Equal(t, []*Grade{
		{
			ID:           problemSet.Grades[0].ID,
			UserID:       user1.ID,
			ProblemSetID: problemSet.ID,
			Detail: createJSONForTest(t, map[uint]uint{
				problem1.ID: 11,
				problem2.ID: 120,
			}),
			Total:     131,
			CreatedAt: problemSet.Grades[0].CreatedAt,
			UpdatedAt: problemSet.Grades[0].UpdatedAt,
		},
		{
			ID:           problemSet.Grades[1].ID,
			UserID:       user2.ID,
			ProblemSetID: problemSet.ID,
			Detail: createJSONForTest(t, map[uint]uint{
				problem1.ID: 21,
				problem2.ID: 22,
			}),
			Total:     43,
			CreatedAt: problemSet.Grades[1].CreatedAt,
			UpdatedAt: problemSet.Grades[1].UpdatedAt,
		},
	}, problemSet.Grades)
	assert.NoError(t, UpdateGrade(&Submission{
		ProblemSetId: problemSet.ID,
		UserID:       user2.ID,
		ProblemID:    problem2.ID,
		Score:        220,
	}))
	assert.NoError(t, base.DB.Preload("Grades").First(&problemSet, problemSet.ID).Error)
	assert.Equal(t, []*Grade{
		{
			ID:           problemSet.Grades[0].ID,
			UserID:       user1.ID,
			ProblemSetID: problemSet.ID,
			Detail: createJSONForTest(t, map[uint]uint{
				problem1.ID: 11,
				problem2.ID: 120,
			}),
			Total:     131,
			CreatedAt: problemSet.Grades[0].CreatedAt,
			UpdatedAt: problemSet.Grades[0].UpdatedAt,
		},
		{
			ID:           problemSet.Grades[1].ID,
			UserID:       user2.ID,
			ProblemSetID: problemSet.ID,
			Detail: createJSONForTest(t, map[uint]uint{
				problem1.ID: 21,
				problem2.ID: 220,
			}),
			Total:     241,
			CreatedAt: problemSet.Grades[1].CreatedAt,
			UpdatedAt: problemSet.Grades[1].UpdatedAt,
		},
	}, problemSet.Grades)
}
