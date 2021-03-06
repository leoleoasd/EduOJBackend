package utils

import (
	"fmt"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/database/models"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func checkInviteCode(t *testing.T, code string) {
	assert.Regexp(t, regexp.MustCompile("^[a-zA-Z2-9]{5}$"), code)
	var count int64
	assert.NoError(t, base.DB.Model(models.Class{}).Where("invite_code = ?", code).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func createClassForTest(t *testing.T, name string, id int, managers, students []*models.User) models.Class {
	inviteCode := GenerateInviteCode()
	class := models.Class{
		Name:        fmt.Sprintf("test_%s_%d_name", name, id),
		CourseName:  fmt.Sprintf("test_%s_%d_course_name", name, id),
		Description: fmt.Sprintf("test_%s_%d_description", name, id),
		InviteCode:  inviteCode,
		Managers:    managers,
		Students:    students,
	}
	assert.NoError(t, base.DB.Create(&class).Error)
	return class
}

func TestGenerateInviteCode(t *testing.T) {
	t.Parallel()
	class := createClassForTest(t, "test_generate_invite_code_success", 0, nil, nil)
	checkInviteCode(t, class.InviteCode)
}
