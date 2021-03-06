package models

import (
	"github.com/EduOJ/backend/base"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestClass struct {
	ID uint
}

func (c TestClass) TypeName() string {
	return "test_class"
}

func (c TestClass) GetID() uint {
	return c.ID
}

func TestUser_GrantRole(t *testing.T) {
	t.Parallel()

	u := User{
		Username: "test_user_grant_role",
		Nickname: "test_user_grant_role",
		Email:    "test_user_grant_role",
		Password: "test_user_grant_role",
	}
	base.DB.Create(&u)
	r := Role{
		Name: "ttt",
	}
	base.DB.Create(&r)
	u.GrantRole(r.Name)
	assert.Equal(t, r, u.Roles[0].Role)
	{
		dummy := "test_class"
		r = Role{
			Name:   "ttt123",
			Target: &dummy,
		}
	}
	base.DB.Create(&r)
	u.GrantRole(r.Name, TestClass{ID: 2})
	assert.Equal(t, r, u.Roles[1].Role)
	assert.Equal(t, uint(2), u.Roles[1].TargetID)
}

func TestCan(t *testing.T) {
	t.Parallel()

	base.DB.AutoMigrate(&TestClass{})
	classA := TestClass{}
	classB := TestClass{}
	base.DB.Create(&classA)
	base.DB.Create(&classB)
	dummy := "test_class"
	teacher := Role{
		Name:   "testCanTeacher",
		Target: &dummy,
	}
	assistant := Role{
		Name:   "testCanAssistant",
		Target: &dummy,
	}
	admin := Role{
		Name:   "testCanAdmin",
		Target: &dummy,
	}
	globalRole := Role{
		Name: "testCanGlobalRole",
	}
	globalAdmin := Role{
		Name: "testCanGlobalAdmin",
	}
	base.DB.Create(&teacher)
	base.DB.Create(&assistant)
	base.DB.Create(&admin)
	base.DB.Create(&globalRole)
	base.DB.Create(&globalAdmin)
	teacher.AddPermission("permission_teacher")
	teacher.AddPermission("permission_both")
	assistant.AddPermission("permission_both")
	admin.AddPermission("all")
	globalRole.AddPermission("global_permission")
	globalAdmin.AddPermission("all")

	testUser0 := User{
		Username: "test_user_0",
		Nickname: "tu0",
		Email:    "tu0@e.com",
		Password: "",
	}
	testUser1 := User{
		Username: "test_user_1",
		Nickname: "tu1",
		Email:    "tu1@e.com",
		Password: "",
	}
	base.DB.Create(&testUser0)
	base.DB.Create(&testUser1)
	testUser0.GrantRole(teacher.Name, classA)
	testUser0.GrantRole(assistant.Name, classB)
	testUser1.GrantRole(admin.Name, classB)
	testUser0.GrantRole(globalRole.Name)
	testUser1.GrantRole(globalAdmin.Name)
	t.Run("scoped", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			thisAssert := assert.New(t)
			thisAssert.True(testUser0.Can("permission_teacher", classA))
			thisAssert.False(testUser0.Can("permission_teacher", classB))
			thisAssert.True(testUser0.Can("permission_both", classA))
			thisAssert.True(testUser0.Can("permission_both", classB))
			thisAssert.False(testUser0.Can("permission_both"))
			thisAssert.True(testUser1.Can("permission_teacher", classB))
			thisAssert.True(testUser1.Can("permission_both", classB))
		})
		t.Run("admin", func(t *testing.T) {
			thisAssert := assert.New(t)
			thisAssert.False(testUser1.Can("all", classA))
			thisAssert.True(testUser1.Can("all", classB))
			thisAssert.True(testUser1.Can("permission_teacher", classB))
			thisAssert.True(testUser1.Can("permission_both", classB))
			thisAssert.True(testUser1.Can("permission_non_existing", classB))
		})
	})
	t.Run("global", func(t *testing.T) {
		thisAssert := assert.New(t)
		thisAssert.True(testUser0.Can("global_permission"))
		thisAssert.False(testUser0.Can("non_existing_permission"))
		thisAssert.True(testUser1.Can("global_permission"))
		thisAssert.True(testUser1.Can("non_existing_permission"))
	})
	assert.Panics(t, func() {
		testUser0.Can("xxx", classA, classB)
	})
}
