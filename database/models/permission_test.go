package models

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestUserHasRoleMarshalJSON(t *testing.T) {
	dummy := "test_class"
	perm1ForRole1 := Permission{
		ID:     1001,
		RoleID: 2001,
		Name:   "perm1",
	}
	perm1ForRole2 := Permission{
		ID:     1002,
		RoleID: 2002,
		Name:   "perm1",
	}
	perm2ForRole1 := Permission{
		ID:     1003,
		RoleID: 2001,
		Name:   "perm2",
	}
	role1 := Role{
		ID:     2001,
		Name:   "role1",
		Target: &dummy,
		Permissions: []Permission{
			perm1ForRole1,
			perm2ForRole1,
		},
	}
	role2 := Role{
		ID:     2002,
		Name:   "role2",
		Target: &dummy,
		Permissions: []Permission{
			perm1ForRole2,
		},
	}
	userHasRole1ForClassA := UserHasRole{
		ID:       3001,
		UserID:   4001,
		RoleID:   2001,
		Role:     role1,
		TargetID: 5001, // classA
	}
	userHasRole2ForClassB := UserHasRole{
		ID:       3002,
		UserID:   4001,
		RoleID:   2002,
		Role:     role2,
		TargetID: 5002, // classB
	}
	userHasRole1Global := UserHasRole{
		ID:     3003,
		UserID: 4001,
		RoleID: 2001,
		Role:   role1,
	}
	location, _ := time.LoadLocation("Asia/Shanghai")
	baseTime := time.Date(2020, 8, 18, 13, 24, 24, 31972138, location)
	deleteTime := baseTime.Add(time.Minute * 5)
	user := User{
		ID:       4001,
		Username: "test_marshal_json_username",
		Nickname: "test_marshal_json_nickname",
		Email:    "test_marshal_json@mail.com",
		Password: "test_marshal_json_password",

		Roles: []UserHasRole{
			userHasRole1ForClassA,
			userHasRole2ForClassB,
			userHasRole1Global,
		},
		RoleLoaded: true,

		CreatedAt: baseTime.Add(time.Second),
		UpdatedAt: baseTime.Add(time.Minute),
		DeletedAt: gorm.DeletedAt{
			Valid: true,
			Time:  deleteTime,
		},
	}
	j, err := json.Marshal(user)
	assert.NoError(t, err)
	expected := `{"id":4001,"username":"test_marshal_json_username","nickname":"test_marshal_json_nickname","email":"test_marshal_json@mail.com","roles":[{"id":2001,"name":"role1","target":"test_class","Permissions":[{"id":1001,"role_id":2001,"name":"perm1"},{"id":1003,"role_id":2001,"name":"perm2"}],"target_id":5001},{"id":2002,"name":"role2","target":"test_class","Permissions":[{"id":1002,"role_id":2002,"name":"perm1"}],"target_id":5002},{"id":2001,"name":"role1","target":"test_class","Permissions":[{"id":1001,"role_id":2001,"name":"perm1"},{"id":1003,"role_id":2001,"name":"perm2"}],"target_id":0}],"class_managing":null,"class_taking":null,"grades":null,"created_at":"2020-08-18T13:24:25.031972138+08:00","deleted_at":"2020-08-18T13:29:24.031972138+08:00","Credentials":null}`
	assert.Equal(t, expected, string(j))
}
