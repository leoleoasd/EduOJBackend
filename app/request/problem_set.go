package request

import "time"

type CreateProblemSetRequest struct {
	Name        string `json:"name" form:"name" query:"name" validate:"required,max=255"`
	Description string `json:"description" form:"description" query:"description" validate:"required"`

	StartTime time.Time `json:"start_time" form:"start_time" query:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" form:"end_time" query:"end_time" validate:"required,gtefield=StartTime"`
}

type CloneProblemSetRequest struct {
	SourceClassID      uint `json:"source_class_id" form:"source_class_id" query:"source_class_id" validate:"required"`
	SourceProblemSetID uint `json:"source_problem_set_id" form:"source_problem_set_id" query:"source_problem_set_id" validate:"required"`
}

type GetProblemSetRequest struct {
}

type UpdateProblemSetRequest struct {
	Name        string `json:"name" form:"name" query:"name" validate:"required,max=255"`
	Description string `json:"description" form:"description" query:"description" validate:"required"`

	StartTime time.Time `json:"start_time" form:"start_time" query:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" form:"end_time" query:"end_time" validate:"required,gtefield=StartTime"`
}

type AddProblemsToSetRequest struct {
	ProblemIDs []uint `json:"problem_ids" form:"problem_ids" query:"problem_ids" validate:"required"`
}

type DeleteProblemsFromSetRequest struct {
	ProblemIDs []uint `json:"problem_ids" form:"problem_ids" query:"problem_ids" validate:"required"`
}

type DeleteProblemSetRequest struct {
}
