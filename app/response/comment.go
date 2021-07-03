package response

import (
	"github.com/EduOJ/backend/database/models"
)

type CreateCommentResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		Id uint
	} `json:"data"`
}

type GetCommentResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		ComsRoot     []models.Comment
		ComsNoneRoot []models.Comment
	} `json:"data"`
}

type AddReactionResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		Cont string
	} `json:"data"`
}
