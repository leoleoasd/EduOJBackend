package response

import (
	"github.com/EduOJ/backend/app/response/resource"
)

type ProblemSetCreateSubmissionResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.SubmissionDetail `json:"submission"`
	} `json:"data"`
}

type ProblemSetGetSubmissionResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.SubmissionDetail `json:"submission"`
	} `json:"data"`
}

type ProblemSetGetSubmissionsResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		Submissions []resource.Submission `json:"submissions"`
		Total       int                   `json:"total"`
		Count       int                   `json:"count"`
		Offset      int                   `json:"offset"`
		Prev        *string               `json:"prev"`
		Next        *string               `json:"next"`
	} `json:"data"`
}
