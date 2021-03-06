package response

import (
	"github.com/EduOJ/backend/app/response/resource"
)

type AdminCreateUserResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.UserForAdmin `json:"user"`
	} `json:"data"`
}

type AdminUpdateUserResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.UserForAdmin `json:"user"`
	} `json:"data"`
}

type AdminGetUserResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.UserForAdmin `json:"user"`
	} `json:"data"`
}

type AdminGetUsersResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		Users  []resource.User `json:"users"`
		Total  int             `json:"total"`
		Count  int             `json:"count"`
		Offset int             `json:"offset"`
		Prev   *string         `json:"prev"`
		Next   *string         `json:"next"`
	} `json:"data"`
}
