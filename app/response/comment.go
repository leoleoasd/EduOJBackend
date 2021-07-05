package response

import (
	"github.com/EduOJ/backend/database/models"
)

type CreateCommentResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		Comment models.Comment
	} `json:"data"`
}

type GetCommentResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		ComsRoot     []models.Comment
		ComsNoneRoot []models.Comment
		Total    int                        `json:"total"`
		Count    int                        `json:"count"`
		Offset   int                        `json:"offset"`
		Prev     *string                    `json:"prev"`
		Next     *string                    `json:"next"`
	} `json:"data"`
}

type AddReactionResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		Content string
	} `json:"data"`
}
