package request

type ProblemSetCreateSubmissionRequest struct {
	Language string `json:"language" form:"language" query:"language" validate:"required"`
	// code(required)
}

type ProblemSetGetSubmissionRequest struct {
}

type ProblemSetGetSubmissionsRequest struct {
	ProblemId uint `json:"problem_id" form:"problem_id" query:"problem_id"`
	UserId    uint `json:"user_id" form:"user_id" query:"user_id"`

	Limit  int `json:"limit" form:"limit" query:"limit" validate:"max=100,min=0"`
	Offset int `json:"offset" form:"offset" query:"offset" validate:"min=0"`
}
