package response

import (
	"github.com/EduOJ/backend/database/models"
)

type AdminGetLogsResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		Logs   []models.Log `json:"logs"`
		Total  int          `json:"total"`
		Count  int          `json:"count"`
		Offset int          `json:"offset"`
		Prev   *string      `json:"prev"`
		Next   *string      `json:"next"`
	} `json:"data"`
}
