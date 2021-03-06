package middleware_test

import (
	"fmt"
	"github.com/EduOJ/backend/app/middleware"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/database/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type testClass struct {
	ID uint `gorm:"primaryKey" json:"id"`
}

func (c testClass) TypeName() string {
	return "test_class"
}

func (c testClass) GetID() uint {
	return c.ID
}

func setUser(user models.User) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user", user)
			return next(c)
		}
	}
}

func responseWithUser(user models.User) response.Response {
	return response.Response{
		Message: "SUCCESS",
		Error:   nil,
		Data:    user,
	}
}

func TestHasPermission(t *testing.T) {
	t.Parallel()
	e := echo.New()
	classA := testClass{ID: 1}
	classB := testClass{ID: 2}
	dummy := "test_class"
	adminRole := models.Role{
		Name:   "testHasRoleAdmin",
		Target: &dummy,
	}
	permRole := models.Role{
		Name:   "testRole",
		Target: &dummy,
	}
	globalAdminRole := models.Role{
		Name: "testHasRoleGlobalAdmin",
	}
	globalPermRole := models.Role{
		Name: "testHasRoleGlobalTestRole",
	}
	base.DB.Create(&adminRole)
	base.DB.Create(&permRole)
	base.DB.Create(&globalAdminRole)
	base.DB.Create(&globalPermRole)
	adminRole.AddPermission("all")
	permRole.AddPermission("testPerm")
	globalAdminRole.AddPermission("all")
	globalPermRole.AddPermission("testPerm")

	testHasPermUserWithoutPerms := models.User{
		Username: "testHasPermUserWithoutPerms",
		Nickname: "uwp",
		Email:    "uwop@e.com",
		Password: "",
	}
	testHasPermUserWithClassAPerm := models.User{
		Username: "testHasPermUserWithClassAPerm",
		Nickname: "uwcap",
		Email:    "uwcap@e.com",
		Password: "",
	}
	testHasPermUserWithAllClassAPerms := models.User{
		Username: "testHasPermUserWithAllClassAPerms",
		Nickname: "uwacap",
		Email:    "uwacap@e.com",
		Password: "",
	}
	testHasPermUserWithPerm := models.User{
		Username: "testHasPermUserWithPerm",
		Nickname: "uwp",
		Email:    "uwp@e.com",
		Password: "",
	}
	testHasPermAdministrator := models.User{
		Username: "testHasPermAdministrator",
		Nickname: "a",
		Email:    "a@e.com",
		Password: "",
	}
	testHasPermUserWithClassAAndGlobalPerm := models.User{
		Username: "testHasPermUserWithClassAAndGlobalPerm",
		Nickname: "uwaagp",
		Email:    "uwaagp@e.com",
		Password: "",
	}

	base.DB.Create(&testHasPermUserWithClassAPerm)
	base.DB.Create(&testHasPermUserWithAllClassAPerms)
	base.DB.Create(&testHasPermUserWithPerm)
	base.DB.Create(&testHasPermAdministrator)
	base.DB.Create(&testHasPermUserWithClassAAndGlobalPerm)
	testHasPermUserWithClassAPerm.GrantRole(permRole.Name, classA)
	testHasPermUserWithAllClassAPerms.GrantRole(adminRole.Name, classA)
	testHasPermUserWithPerm.GrantRole(globalPermRole.Name)
	testHasPermAdministrator.GrantRole(globalAdminRole.Name)
	testHasPermUserWithClassAAndGlobalPerm.GrantRole(permRole.Name, classA)
	testHasPermUserWithClassAAndGlobalPerm.GrantRole(globalPermRole.Name)

	users := []models.User{
		testHasPermUserWithoutPerms,
		testHasPermUserWithClassAPerm,
		testHasPermUserWithAllClassAPerms,
		testHasPermUserWithPerm,
		testHasPermAdministrator,
		testHasPermUserWithClassAAndGlobalPerm,
	}

	permTests := []struct {
		name       string
		path       string
		permName   string
		targetType *string
		logicMode  string
		targetID   uint
	}{
		{
			name:       "perm_global",
			path:       "test_perm_global",
			permName:   "testPerm",
			targetType: nil,
			logicMode:  "",
			targetID:   0,
		},
		{
			name:       "perm_a",
			path:       "test_perm",
			permName:   "testPerm",
			targetType: &dummy,
			logicMode:  "",
			targetID:   classA.ID,
		},
		{
			name:       "perm_b",
			path:       "test_perm",
			permName:   "testPerm",
			targetType: &dummy,
			logicMode:  "",
			targetID:   classB.ID,
		},
		{
			name:       "perm_a_and_global",
			path:       "test_perm_and",
			permName:   "testPerm",
			targetType: &dummy,
			logicMode:  "and",
			targetID:   classA.ID,
		},
		{
			name:       "perm_a_or_global",
			path:       "test_perm_or",
			permName:   "testPerm",
			targetType: &dummy,
			logicMode:  "or",
			targetID:   classA.ID,
		},
		{
			name:       "all_global",
			path:       "test_all_global",
			permName:   "nonExitingPerm",
			targetType: nil,
			logicMode:  "",
			targetID:   0,
		},
		{
			name:       "all_a",
			path:       "test_all",
			permName:   "non_exiting",
			targetType: &dummy,
			logicMode:  "",
			targetID:   classA.ID,
		},
		{
			name:       "all_b",
			path:       "test_all",
			permName:   "nonExitingPerm",
			targetType: &dummy,
			logicMode:  "",
			targetID:   classB.ID,
		},
	}

	expectedRet := map[string]map[string]bool{
		"testHasPermUserWithoutPerms": {},
		"testHasPermUserWithClassAPerm": {
			"perm_a":           true,
			"perm_a_or_global": true,
		},
		"testHasPermUserWithAllClassAPerms": {
			"perm_a":           true,
			"all_a":            true,
			"perm_a_or_global": true,
		},
		"testHasPermUserWithPerm": {
			"perm_global":      true,
			"perm_a_or_global": true,
		},
		"testHasPermAdministrator": {
			"perm_global":      true,
			"all_global":       true,
			"perm_a_or_global": true,
		},
		"testHasPermUserWithClassAAndGlobalPerm": {
			"perm_a":            true,
			"perm_global":       true,
			"perm_a_and_global": true,
			"perm_a_or_global":  true,
		},
	}

	userGroups := make([]*echo.Group, len(users))
	for i, user := range users {
		userGroups[i] = e.Group("/"+user.Username, setUser(user))
		for _, permTest := range permTests {
			if permTest.logicMode == "" {
				if permTest.targetType == nil {
					userGroups[i].POST("/"+permTest.path, testController, middleware.HasPermission(middleware.UnscopedPermission{P: permTest.permName}))
				} else {
					userGroups[i].POST("/"+permTest.path+"/:id", testController, middleware.HasPermission(middleware.ScopedPermission{P: permTest.permName, T: *permTest.targetType}))
				}
				// We only test logical operations for a global permission and a specified target permission
			} else if permTest.logicMode == "and" {
				userGroups[i].POST("/"+permTest.path+"/:id", testController, middleware.HasPermission(middleware.AndPermission{
					A: middleware.ScopedPermission{P: permTest.permName, T: *permTest.targetType},
					B: middleware.UnscopedPermission{P: permTest.permName},
				}))
			} else if permTest.logicMode == "or" {
				userGroups[i].POST("/"+permTest.path+"/:id", testController, middleware.HasPermission(middleware.OrPermission{
					A: middleware.ScopedPermission{P: permTest.permName, T: *permTest.targetType},
					B: middleware.UnscopedPermission{P: permTest.permName},
				}))
			}
		}
	}
	e.POST("/noUser/test_perm_global", testController, middleware.HasPermission(middleware.UnscopedPermission{P: "testPerm"}))

	for _, user := range users {
		t.Run(user.Username, func(t *testing.T) {
			t.Parallel()
			for _, permTest := range permTests {
				permTest := permTest
				t.Run(permTest.name, func(t *testing.T) {
					t.Parallel()
					httpResp := (*http.Response)(nil)
					resp := response.Response{}
					if permTest.targetType == nil {
						httpResp = makeResp(makeReq(t, "POST", "/"+user.Username+"/"+permTest.path, nil), e)
					} else {
						httpResp = makeResp(makeReq(t, "POST", fmt.Sprintf("/%s/%s/%d", user.Username, permTest.path, permTest.targetID), nil), e)
					}
					mustJsonDecode(httpResp, &resp)
					expectedResult := expectedRet[user.Username][permTest.name]
					if expectedResult {
						assert.Equal(t, http.StatusOK, httpResp.StatusCode)
						jsonEQ(t, responseWithUser(user), resp)
					} else {
						assert.Equal(t, http.StatusForbidden, httpResp.StatusCode)
						assert.Equal(t, response.ErrorResp("PERMISSION_DENIED", nil), resp)
					}
				})
			}
		})
	}

	t.Run("testIllegalRouteParam", func(t *testing.T) {
		t.Parallel()
		httpResp := makeResp(makeReq(t, "POST", "/testHasPermUserWithoutPerms/test_perm/aaa", nil), e)
		resp := response.Response{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, http.StatusForbidden, httpResp.StatusCode)
		assert.Equal(t, response.ErrorResp("PERMISSION_DENIED", nil), resp)
	})

	userGroups[1].POST("/scoped_multiple_id/:id1/:id2",
		testController,
		middleware.HasPermission(middleware.ScopedPermission{
			P:           "testPerm",
			IdFieldName: "id2",
			T:           "test_class",
		}))

	t.Run("scopedPermissionWithMultipleId", func(t *testing.T) {
		t.Parallel()
		httpResp := makeResp(makeReq(t, "POST", "/testHasPermUserWithClassAPerm/scoped_multiple_id/2/1", nil), e)
		resp := response.Response{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		jsonEQ(t, responseWithUser(testHasPermUserWithClassAPerm), resp)

		httpResp = makeResp(makeReq(t, "POST", "/testHasPermUserWithClassAPerm/scoped_multiple_id/1/2", nil), e)
		resp = response.Response{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, http.StatusForbidden, httpResp.StatusCode)
		assert.Equal(t, response.ErrorResp("PERMISSION_DENIED", nil), resp)
	})
}
