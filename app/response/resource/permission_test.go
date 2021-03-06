package resource_test

import (
	"fmt"
	"github.com/EduOJ/backend/app/response/resource"
	"github.com/EduOJ/backend/base/utils"
	"github.com/EduOJ/backend/database/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
	"time"
)

type roleWithTargetID struct {
	role models.Role
	id   uint
}

func createPermissionForTest(name string, id uint, roleId uint) models.Permission {
	return models.Permission{
		ID:     id,
		RoleID: roleId,
		Name:   fmt.Sprintf("test_%s_permission_%d", name, id),
	}
}

func createRoleForTest(name string, id uint, permissionCount uint) models.Role {
	target := fmt.Sprintf("test_%s_role_%d_target", name, id)
	permissions := make([]models.Permission, permissionCount)
	for i := range permissions {
		permissions[i] = createPermissionForTest(name, uint(i), id)
	}
	return models.Role{
		ID:          id,
		Name:        fmt.Sprintf("test_%s_role_%d", name, id),
		Target:      &target,
		Permissions: permissions,
	}
}

func createUserForTest(name string, id uint, roles ...roleWithTargetID) (user models.User) {
	user = models.User{
		ID:         id,
		Username:   fmt.Sprintf("test_%s_user_%d", name, id),
		Nickname:   fmt.Sprintf("test_%s_user_%d_nick", name, id),
		Email:      fmt.Sprintf("test_%s_user_%d@e.e", name, id),
		Password:   utils.HashPassword(fmt.Sprintf("test_%s_user_%d_pwd", name, id)),
		Roles:      make([]models.UserHasRole, len(roles)),
		RoleLoaded: true,
		CreatedAt:  time.Date(int(id), 1, 1, 1, 1, 1, 1, time.FixedZone("test_zone", 0)),
		UpdatedAt:  time.Date(int(id), 2, 2, 2, 2, 2, 2, time.FixedZone("test_zone", 0)),
		DeletedAt:  gorm.DeletedAt{},
	}
	for i, role := range roles {
		user.Roles[i] = models.UserHasRole{
			ID:       uint(i),
			UserID:   user.ID,
			RoleID:   role.role.ID,
			Role:     role.role,
			TargetID: role.id,
		}
	}
	return
}

func TestGetPermission(t *testing.T) {
	permission := createPermissionForTest("get_permission", 0, 0)
	actualPermission := resource.GetPermission(&permission)
	expectedPermission := resource.Permission{
		ID:   0,
		Name: "test_get_permission_permission_0",
	}
	assert.Equal(t, expectedPermission, *actualPermission)
}

func TestGetRoleAndGetRoleSlice(t *testing.T) {
	role1 := createRoleForTest("get_role", 1, 1)
	role2 := createRoleForTest("get_role", 2, 2)
	user := createUserForTest("get_role", 0,
		roleWithTargetID{role: role1, id: 1},
		roleWithTargetID{role: role2, id: 2},
	)
	t.Run("testGetRole", func(t *testing.T) {
		actualRole := resource.GetRole(&user.Roles[0])
		target := "test_get_role_role_1_target"
		expectedRole := resource.Role{
			ID:     0,
			Name:   "test_get_role_role_1",
			Target: &target,
			Permissions: []resource.Permission{
				{ID: 0, Name: "test_get_role_permission_0"},
			},
			TargetID: 1,
		}
		assert.Equal(t, expectedRole, *actualRole)
	})
	t.Run("testGetRoleSlice", func(t *testing.T) {
		actualRoleSlice := resource.GetRoleSlice(user.Roles)
		target1 := "test_get_role_role_1_target"
		target2 := "test_get_role_role_2_target"
		expectedRoleSlice := []resource.Role{
			{
				ID:     0,
				Name:   "test_get_role_role_1",
				Target: &target1,
				Permissions: []resource.Permission{
					{ID: 0, Name: "test_get_role_permission_0"},
				},
				TargetID: 1,
			},
			{
				ID:     0,
				Name:   "test_get_role_role_2",
				Target: &target2,
				Permissions: []resource.Permission{
					{ID: 0, Name: "test_get_role_permission_0"},
					{ID: 1, Name: "test_get_role_permission_1"},
				},
				TargetID: 2,
			},
		}
		assert.Equal(t, expectedRoleSlice, actualRoleSlice)
	})
}
