package models

import (
	"github.com/EduOJ/backend/base"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
)

func TestAddStudentsAndDeleteStudentsByID(t *testing.T) {
	t.Parallel()

	user1 := User{
		Username: "test_add_and_delete_students_1_username",
		Nickname: "test_add_and_delete_students_1_nickname",
		Email:    "test_add_and_delete_students_1@email.com",
		Password: "test_add_and_delete_students_1_password",
	}
	user2 := User{
		Username: "test_add_and_delete_students_2_username",
		Nickname: "test_add_and_delete_students_2_nickname",
		Email:    "test_add_and_delete_students_2@email.com",
		Password: "test_add_and_delete_students_2_password",
	}
	user3 := User{
		Username: "test_add_and_delete_students_3_username",
		Nickname: "test_add_and_delete_students_3_nickname",
		Email:    "test_add_and_delete_students_3@email.com",
		Password: "test_add_and_delete_students_3_password",
	}
	user4 := User{
		Username: "test_add_and_delete_students_4_username",
		Nickname: "test_add_and_delete_students_4_nickname",
		Email:    "test_add_and_delete_students_4@email.com",
		Password: "test_add_and_delete_students_4_password",
	}
	assert.NoError(t, base.DB.Create(&user1).Error)
	assert.NoError(t, base.DB.Create(&user2).Error)
	assert.NoError(t, base.DB.Create(&user3).Error)
	assert.NoError(t, base.DB.Create(&user4).Error)

	t.Run("AddSuccess", func(t *testing.T) {
		t.Parallel()
		class := Class{
			Name:        "test_add_students_success_class_name",
			CourseName:  "test_add_students_success_class_course_name",
			Description: "test_add_students_success_class_description",
			InviteCode:  "test_add_students_success_class_invite_code",
			Managers:    []*User{},
			Students: []*User{
				&user1,
				&user2,
			},
		}
		assert.NoError(t, base.DB.Create(&class).Error)
		assert.NoError(t, class.AddStudents([]uint{
			user3.ID,
			user4.ID,
		}))
		assert.Equal(t, Class{
			ID:          class.ID,
			Name:        "test_add_students_success_class_name",
			CourseName:  "test_add_students_success_class_course_name",
			Description: "test_add_students_success_class_description",
			InviteCode:  "test_add_students_success_class_invite_code",
			Managers:    []*User{},
			Students: []*User{
				&user1,
				&user2,
				&user3,
				&user4,
			},
			CreatedAt: class.CreatedAt,
			UpdatedAt: class.UpdatedAt,
			DeletedAt: gorm.DeletedAt{},
		}, class)
	})
	t.Run("AddExistingInClass", func(t *testing.T) {
		t.Parallel()
		class := Class{
			Name:        "test_add_students_existing_in_class_name",
			CourseName:  "test_add_students_existing_in_class_course_name",
			Description: "test_add_students_existing_in_class_description",
			InviteCode:  "test_add_students_existing_in_class_invite_code",
			Managers:    []*User{},
			Students: []*User{
				&user1,
				&user2,
			},
		}
		assert.NoError(t, base.DB.Create(&class).Error)
		assert.NoError(t, class.AddStudents([]uint{
			user2.ID,
			user3.ID,
		}))
		assert.Equal(t, Class{
			ID:          class.ID,
			Name:        "test_add_students_existing_in_class_name",
			CourseName:  "test_add_students_existing_in_class_course_name",
			Description: "test_add_students_existing_in_class_description",
			InviteCode:  "test_add_students_existing_in_class_invite_code",
			Managers:    []*User{},
			Students: []*User{
				&user1,
				&user2,
				&user3,
			},
			CreatedAt: class.CreatedAt,
			UpdatedAt: class.UpdatedAt,
			DeletedAt: gorm.DeletedAt{},
		}, class)
	})
	t.Run("AddNonExist", func(t *testing.T) {
		t.Parallel()
		class := Class{
			Name:        "test_add_students_non_exist_class_name",
			CourseName:  "test_add_students_non_exist_class_course_name",
			Description: "test_add_students_non_exist_class_description",
			InviteCode:  "test_add_students_non_exist_class_invite_code",
			Managers:    []*User{},
			Students: []*User{
				&user1,
				&user2,
			},
		}
		assert.NoError(t, base.DB.Create(&class).Error)
		assert.NoError(t, class.AddStudents([]uint{
			user3.ID,
			0,
		}))
		assert.Equal(t, Class{
			ID:          class.ID,
			Name:        "test_add_students_non_exist_class_name",
			CourseName:  "test_add_students_non_exist_class_course_name",
			Description: "test_add_students_non_exist_class_description",
			InviteCode:  "test_add_students_non_exist_class_invite_code",
			Managers:    []*User{},
			Students: []*User{
				&user1,
				&user2,
				&user3,
			},
			CreatedAt: class.CreatedAt,
			UpdatedAt: class.UpdatedAt,
			DeletedAt: gorm.DeletedAt{},
		}, class)
	})
	t.Run("DeleteSuccess", func(t *testing.T) {
		t.Parallel()
		class := Class{
			Name:        "test_delete_students_success_class_name",
			CourseName:  "test_delete_students_success_class_course_name",
			Description: "test_delete_students_success_class_description",
			InviteCode:  "test_delete_students_success_class_invite_code",
			Managers:    []*User{},
			Students: []*User{
				&user1,
				&user2,
				&user3,
			},
		}
		assert.NoError(t, base.DB.Create(&class).Error)
		assert.NoError(t, class.DeleteStudents([]uint{
			user1.ID,
			user3.ID,
		}))
		assert.Equal(t, Class{
			ID:          class.ID,
			Name:        "test_delete_students_success_class_name",
			CourseName:  "test_delete_students_success_class_course_name",
			Description: "test_delete_students_success_class_description",
			InviteCode:  "test_delete_students_success_class_invite_code",
			Managers:    []*User{},
			Students: []*User{
				&user2,
			},
			CreatedAt: class.CreatedAt,
			UpdatedAt: class.UpdatedAt,
			DeletedAt: gorm.DeletedAt{},
		}, class)
	})
	t.Run("DeleteNotBelongTo", func(t *testing.T) {
		t.Parallel()
		class := Class{
			Name:        "test_delete_students_not_belong_to_class_name",
			CourseName:  "test_delete_students_not_belong_to_class_course_name",
			Description: "test_delete_students_not_belong_to_class_description",
			InviteCode:  "test_delete_students_not_belong_to_class_invite_code",
			Managers:    []*User{},
			Students: []*User{
				&user1,
				&user2,
				&user3,
			},
		}
		assert.NoError(t, base.DB.Create(&class).Error)
		assert.NoError(t, class.DeleteStudents([]uint{
			user1.ID,
			user4.ID,
		}))
		assert.Equal(t, Class{
			ID:          class.ID,
			Name:        "test_delete_students_not_belong_to_class_name",
			CourseName:  "test_delete_students_not_belong_to_class_course_name",
			Description: "test_delete_students_not_belong_to_class_description",
			InviteCode:  "test_delete_students_not_belong_to_class_invite_code",
			Managers:    []*User{},
			Students: []*User{
				&user2,
				&user3,
			},
			CreatedAt: class.CreatedAt,
			UpdatedAt: class.UpdatedAt,
			DeletedAt: gorm.DeletedAt{},
		}, class)
	})
}
