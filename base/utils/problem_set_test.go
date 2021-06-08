package utils

import (
	"encoding/json"
	"fmt"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/database/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
	"testing"
	"time"
)

func createJSONForTest(t *testing.T, in interface{}) datatypes.JSON {
	j := datatypes.JSON{}
	b, err := json.Marshal(in)
	assert.NoError(t, err)
	assert.NoError(t, j.UnmarshalJSON(b))
	return j
}

func TestUpdateGrade(t *testing.T) {
	t.Parallel()

	user1 := models.User{
		Username: "test_update_grade_1_username",
		Nickname: "test_update_grade_1_nickname",
		Email:    "test_update_grade_1@email.com",
		Password: "test_update_grade_1_password",
	}
	user2 := models.User{
		Username: "test_update_grade_2_username",
		Nickname: "test_update_grade_2_nickname",
		Email:    "test_update_grade_2@email.com",
		Password: "test_update_grade_2_password",
	}
	assert.NoError(t, base.DB.Create(&user1).Error)
	assert.NoError(t, base.DB.Create(&user2).Error)
	problem1 := models.Problem{
		Name:        "test_update_grade_1_name",
		Description: "test_update_grade_1_description",
		MemoryLimit: 1024,
		TimeLimit:   1000,
	}
	problem2 := models.Problem{
		Name:        "test_update_grade_2_name",
		Description: "test_update_grade_2_description",
		MemoryLimit: 2048,
		TimeLimit:   2000,
	}
	assert.NoError(t, base.DB.Create(&problem1).Error)
	assert.NoError(t, base.DB.Create(&problem2).Error)

	t.Run("SubmissionNotInProblemSet", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, UpdateGrade(&models.Submission{
			ProblemSetID: 0,
			UserID:       user1.ID,
			ProblemID:    problem2.ID,
			Score:        100,
		}))
	})
	t.Run("EndedProblemSet", func(t *testing.T) {
		t.Parallel()
		problemSet := models.ProblemSet{
			Name:        "test_update_grade_ended_name",
			Description: "test_update_grade_ended_description",
			Problems: []*models.Problem{
				&problem1,
				&problem2,
			},
			Grades: []*models.Grade{
				{
					UserID: user1.ID,
					Detail: createJSONForTest(t, map[uint]uint{
						problem1.ID: 33,
						problem2.ID: 44,
					}),
					Total: 77,
				},
			},
			StartTime: time.Now().Add(-2 * time.Hour),
			EndTime:   time.Now().Add(-1 * time.Hour),
		}
		assert.NoError(t, base.DB.Create(&problemSet).Error)
		assert.NoError(t, UpdateGrade(&models.Submission{
			ProblemSetID: problemSet.ID,
			UserID:       user1.ID,
			ProblemID:    problem2.ID,
			Score:        100,
		}))
		assert.NoError(t, base.DB.Preload("Grades").First(&problemSet, problemSet.ID).Error)
		assert.Equal(t, []*models.Grade{
			{
				ID:           problemSet.Grades[0].ID,
				UserID:       user1.ID,
				ProblemSetID: problemSet.ID,
				Detail: createJSONForTest(t, map[uint]uint{
					problem1.ID: 33,
					problem2.ID: 44,
				}),
				Total:     77,
				CreatedAt: problemSet.Grades[0].CreatedAt,
				UpdatedAt: problemSet.Grades[0].UpdatedAt,
			},
		}, problemSet.Grades)
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		problemSet := models.ProblemSet{
			Name:        "test_update_grade_name",
			Description: "test_update_grade_description",
			Problems: []*models.Problem{
				&problem1,
				&problem2,
			},
			Grades: []*models.Grade{
				{
					UserID: user1.ID,
					Detail: createJSONForTest(t, map[uint]uint{
						problem1.ID: 11,
						problem2.ID: 12,
					}),
					Total: 23,
				},
			},
			StartTime: time.Now().Add(-1 * time.Hour),
			EndTime:   time.Now().Add(time.Hour),
		}
		assert.NoError(t, base.DB.Create(&problemSet).Error)
		assert.NoError(t, UpdateGrade(&models.Submission{
			ProblemSetID: problemSet.ID,
			UserID:       user1.ID,
			ProblemID:    problem2.ID,
			Score:        120,
		}))
		assert.NoError(t, base.DB.Preload("Grades").First(&problemSet, problemSet.ID).Error)
		assert.Equal(t, []*models.Grade{
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
		assert.NoError(t, UpdateGrade(&models.Submission{
			ProblemSetID: problemSet.ID,
			UserID:       user2.ID,
			ProblemID:    problem1.ID,
			Score:        21,
		}))
		assert.NoError(t, base.DB.Preload("Grades").First(&problemSet, problemSet.ID).Error)
		assert.Equal(t, []*models.Grade{
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
		assert.NoError(t, UpdateGrade(&models.Submission{
			ProblemSetID: problemSet.ID,
			UserID:       user2.ID,
			ProblemID:    problem2.ID,
			Score:        22,
		}))
		assert.NoError(t, base.DB.Preload("Grades").First(&problemSet, problemSet.ID).Error)
		assert.Equal(t, []*models.Grade{
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
		assert.NoError(t, UpdateGrade(&models.Submission{
			ProblemSetID: problemSet.ID,
			UserID:       user2.ID,
			ProblemID:    problem2.ID,
			Score:        220,
		}))
		assert.NoError(t, base.DB.Preload("Grades").First(&problemSet, problemSet.ID).Error)
		assert.Equal(t, []*models.Grade{
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
	})
}

func checkGrade(t *testing.T, expectedGrade *models.Grade) {
	databaseGrade := models.Grade{}
	err := base.DB.
		Where("user_id = ?", expectedGrade.UserID).
		Where("problem_set_id = ?", expectedGrade.ProblemSetID).
		First(&databaseGrade).Error
	assert.NoError(t, err)
	expectedDetail := make(map[uint]uint)
	assert.NoError(t, json.Unmarshal(expectedGrade.Detail, &expectedDetail))
	databaseDetail := make(map[uint]uint)
	assert.NoError(t, json.Unmarshal(databaseGrade.Detail, &databaseDetail))
	assert.Equal(t, expectedDetail, databaseDetail)
	assert.Equal(t, expectedGrade.Total, databaseGrade.Total)
}

func createSubmissionForTest(t *testing.T, problemSet *models.ProblemSet, userID, problemID uint, score uint, status string, timeOffset time.Duration) *models.Submission {
	submission := models.Submission{
		UserID:       userID,
		ProblemID:    problemID,
		ProblemSetID: problemSet.ID,
		LanguageName: "",
		Language:     nil,
		FileName:     "",
		Priority:     0,
		Judged:       true,
		Score:        score,
		Status:       status,
		Runs:         nil,
		CreatedAt:    time.Now().Add(timeOffset),
	}
	assert.NoError(t, base.DB.Create(&submission).Error)
	return &submission
}

func TestRefreshGrades(t *testing.T) {
	t.Parallel()

	problem1 := models.Problem{
		Name:        "test_refresh_grades_1_name",
		Description: "test_refresh_grades_1_description",
		MemoryLimit: 1024,
		TimeLimit:   1000,
	}
	problem2 := models.Problem{
		Name:        "test_refresh_grades_2_name",
		Description: "test_refresh_grades_2_description",
		MemoryLimit: 2048,
		TimeLimit:   2000,
	}
	assert.NoError(t, base.DB.Create(&problem1).Error)
	assert.NoError(t, base.DB.Create(&problem2).Error)
	init := func(id int) (u1, u2 *models.User, ps *models.ProblemSet) {
		user1 := models.User{
			Username: fmt.Sprintf("test_refresh_grades_%d_1_username", id),
			Nickname: fmt.Sprintf("test_refresh_grades_%d_1_nickname", id),
			Email:    fmt.Sprintf("test_refresh_grades_%d_1@mail.com", id),
			Password: fmt.Sprintf("test_refresh_grades_%d_1_password", id),
		}
		user2 := models.User{
			Username: fmt.Sprintf("test_refresh_grades_%d_2_username", id),
			Nickname: fmt.Sprintf("test_refresh_grades_%d_2_nickname", id),
			Email:    fmt.Sprintf("test_refresh_grades_%d_2@mail.com", id),
			Password: fmt.Sprintf("test_refresh_grades_%d_2_password", id),
		}
		assert.NoError(t, base.DB.Create(&user1).Error)
		assert.NoError(t, base.DB.Create(&user2).Error)
		class := models.Class{
			Name:        fmt.Sprintf("test_refresh_grades_%d_name", id),
			CourseName:  fmt.Sprintf("test_refresh_grades_%d_course_name", id),
			Description: fmt.Sprintf("test_refresh_grades_%d_description", id),
			InviteCode:  GenerateInviteCode(),
			Managers:    nil,
			Students: []*models.User{
				&user1,
				&user2,
			},
			ProblemSets: nil,
		}
		assert.NoError(t, base.DB.Create(&class).Error)
		problemSet := models.ProblemSet{
			ClassID:     class.ID,
			Name:        fmt.Sprintf("test_refresh_grades_%d_name", id),
			Description: fmt.Sprintf("test_refresh_grades_%d_description", id),
			Problems: []*models.Problem{
				&problem1,
				&problem2,
			},
			Grades:    nil,
			StartTime: time.Now().Add(time.Hour),
			EndTime:   time.Now().Add(2 * time.Hour),
		}
		assert.NoError(t, base.DB.Create(&problemSet).Error)
		problemSet.Class = &class
		return &user1, &user2, &problemSet
	}

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		u1, u2, ps := init(0)
		assert.NoError(t, RefreshGrades(ps))
		j, err := json.Marshal(map[uint]uint{
			problem1.ID: 0,
			problem2.ID: 0,
		})
		assert.NoError(t, err)
		checkGrade(t, &models.Grade{
			UserID:       u1.ID,
			ProblemSetID: ps.ID,
			Detail:       j,
			Total:        0,
		})
		checkGrade(t, &models.Grade{
			UserID:       u2.ID,
			ProblemSetID: ps.ID,
			Detail:       j,
			Total:        0,
		})
	})
	t.Run("MaxScore", func(t *testing.T) {
		t.Parallel()
		u1, u2, ps := init(1)
		// user1 problem1
		createSubmissionForTest(t, ps, u1.ID, problem1.ID, 30, "RUNTIME_ERROR", time.Hour+time.Minute*1)
		createSubmissionForTest(t, ps, u1.ID, problem1.ID, 100, "ACCEPTED", time.Hour+time.Minute*2)
		createSubmissionForTest(t, ps, u1.ID, problem1.ID, 10, "WRONG_ANSWER", time.Hour+time.Minute*3)
		// user1 problem2
		createSubmissionForTest(t, ps, u1.ID, problem2.ID, 100, "ACCEPTED", time.Hour+time.Minute*5)
		// user2 problem1
		// user2 problem2
		createSubmissionForTest(t, ps, u2.ID, problem2.ID, 20, "RUNTIME_ERROR", time.Hour+time.Minute*1)
		createSubmissionForTest(t, ps, u2.ID, problem2.ID, 60, "WRONG_ANSWER", time.Hour+time.Minute*2)
		createSubmissionForTest(t, ps, u2.ID, problem2.ID, 45, "WRONG_ANSWER", time.Hour+time.Minute*3)
		assert.NoError(t, RefreshGrades(ps))
		j1, err := json.Marshal(map[uint]uint{
			problem1.ID: 100,
			problem2.ID: 100,
		})
		j2, err := json.Marshal(map[uint]uint{
			problem1.ID: 0,
			problem2.ID: 60,
		})
		assert.NoError(t, err)
		checkGrade(t, &models.Grade{
			UserID:       u1.ID,
			ProblemSetID: ps.ID,
			Detail:       j1,
			Total:        200,
		})
		checkGrade(t, &models.Grade{
			UserID:       u2.ID,
			ProblemSetID: ps.ID,
			Detail:       j2,
			Total:        60,
		})
	})
	t.Run("TimeLimit", func(t *testing.T) {
		t.Parallel()
		u1, u2, ps := init(2)
		// user1 problem1
		createSubmissionForTest(t, ps, u1.ID, problem1.ID, 30, "RUNTIME_ERROR", time.Hour+time.Minute*1)
		createSubmissionForTest(t, ps, u1.ID, problem1.ID, 100, "ACCEPTED", time.Hour+time.Minute*2)
		createSubmissionForTest(t, ps, u1.ID, problem1.ID, 10, "WRONG_ANSWER", time.Hour*2+time.Minute*3)
		// user1 problem2
		createSubmissionForTest(t, ps, u1.ID, problem2.ID, 100, "ACCEPTED", time.Hour*2+time.Minute*5)
		// user2 problem1
		// user2 problem2
		createSubmissionForTest(t, ps, u2.ID, problem2.ID, 80, "RUNTIME_ERROR", time.Hour+time.Minute*1)
		createSubmissionForTest(t, ps, u2.ID, problem2.ID, 60, "WRONG_ANSWER", time.Hour+time.Minute*2)
		createSubmissionForTest(t, ps, u2.ID, problem2.ID, 90, "WRONG_ANSWER", time.Hour*2+time.Minute*3)
		assert.NoError(t, RefreshGrades(ps))
		j1, err := json.Marshal(map[uint]uint{
			problem1.ID: 100,
			problem2.ID: 0,
		})
		j2, err := json.Marshal(map[uint]uint{
			problem1.ID: 0,
			problem2.ID: 80,
		})
		assert.NoError(t, err)
		checkGrade(t, &models.Grade{
			UserID:       u1.ID,
			ProblemSetID: ps.ID,
			Detail:       j1,
			Total:        100,
		})
		checkGrade(t, &models.Grade{
			UserID:       u2.ID,
			ProblemSetID: ps.ID,
			Detail:       j2,
			Total:        80,
		})
	})
}
