package resource_test

import (
	"github.com/go-playground/assert/v2"
	"github.com/leoleoasd/EduOJBackend/app/response/resource"
	"github.com/leoleoasd/EduOJBackend/database/models"
	"testing"
)

func TestGetUserGetUserForAdminAndGetUserSlice(t *testing.T) {
	role1 := createRoleForTest("get_user", 1, 1)
	role2 := createRoleForTest("get_user", 2, 2)
	user1 := createUserForTest("get_user", 1,
		roleWithTargetID{role: role1, id: 1},
		roleWithTargetID{role: role2, id: 2},
	)
	user2 := createUserForTest("get_user", 2)
	t.Run("testGetUser", func(t *testing.T) {
		actualU := resource.GetUser(&user1)
		expectedU := resource.User{
			ID:       1,
			Username: "test_get_user_user_1",
			Nickname: "test_get_user_user_1_nick",
			Email:    "test_get_user_user_1@e.e",
		}
		assert.Equal(t, expectedU, actualU)
	})
	t.Run("testGetUserNilUser", func(t *testing.T) {
		emptyUser := resource.User{}
		assert.Equal(t, emptyUser, resource.GetUser(nil))
	})
	t.Run("testGetUserForAdmin", func(t *testing.T) {
		actualU := resource.GetUserForAdmin(&user1)
		target1 := "test_get_user_role_1_target"
		target2 := "test_get_user_role_2_target"
		expectedU := resource.UserForAdmin{
			ID:       1,
			Username: "test_get_user_user_1",
			Nickname: "test_get_user_user_1_nick",
			Email:    "test_get_user_user_1@e.e",
			Roles: []resource.Role{
				{
					ID:     0,
					Name:   "test_get_user_role_1",
					Target: &target1,
					Permissions: []resource.Permission{
						{ID: 0, Name: "test_get_user_permission_0"},
					},
					TargetID: 1,
				},
				{
					ID:     0,
					Name:   "test_get_user_role_2",
					Target: &target2,
					Permissions: []resource.Permission{
						{ID: 0, Name: "test_get_user_permission_0"},
						{ID: 1, Name: "test_get_user_permission_1"},
					},
					TargetID: 2,
				},
			},
		}
		assert.Equal(t, &expectedU, actualU)
	})
	t.Run("testGetUserForAdminNilUser", func(t *testing.T) {
		emptyUser := resource.UserForAdmin{}
		assert.Equal(t, emptyUser, resource.GetUserForAdmin(nil))
	})
	t.Run("testGetUserSlice", func(t *testing.T) {
		actualUS := resource.GetUserSlice([]models.User{
			user1, user2,
		})
		expectedUS := []resource.User{
			{
				ID:       1,
				Username: "test_get_user_user_1",
				Nickname: "test_get_user_user_1_nick",
				Email:    "test_get_user_user_1@e.e",
			},
			{
				ID:       2,
				Username: "test_get_user_user_2",
				Nickname: "test_get_user_user_2_nick",
				Email:    "test_get_user_user_2@e.e",
			},
		}
		assert.Equal(t, expectedUS, actualUS)
	})
}
