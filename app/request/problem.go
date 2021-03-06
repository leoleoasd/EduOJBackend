package request

type CreateProblemRequest struct {
	Name        string `json:"name" form:"name" query:"name" validate:"required,max=255"`
	Description string `json:"description" form:"description" query:"description" validate:"required"`
	// attachment_file(optional)
	Public  *bool `json:"public" form:"public" query:"public" validate:"required"`
	Privacy *bool `json:"privacy" form:"privacy" query:"privacy" validate:"required"`

	MemoryLimit       uint64 `json:"memory_limit" form:"memory_limit" query:"memory_limit" validate:"required"`                     // Byte
	TimeLimit         uint   `json:"time_limit" form:"time_limit" query:"time_limit" validate:"required"`                           // ms
	LanguageAllowed   string `json:"language_allowed" form:"language_allowed" query:"language_allowed" validate:"required,max=255"` // E.g.    cpp,c,java,python
	BuildArg          string `json:"build_arg" form:"build_arg" query:"build_arg" validate:"max=255"`                               // E.g.  O2=false
	CompareScriptName string `json:"compare_script_name" form:"compare_script_name" query:"compare_script_name" validate:"required"`
}

type UpdateProblemRequest struct {
	Name        string `json:"name" form:"name" query:"name" validate:"required,max=255"`
	Description string `json:"description" form:"description" query:"description" validate:"required"`
	// attachment_file(optional)
	Public  *bool `json:"public" form:"public" query:"public" validate:"required"`
	Privacy *bool `json:"privacy" form:"privacy" query:"privacy" validate:"required"`

	MemoryLimit       uint64 `json:"memory_limit" form:"memory_limit" query:"memory_limit" validate:"required"`                     // Byte
	TimeLimit         uint   `json:"time_limit" form:"time_limit" query:"time_limit" validate:"required"`                           // ms
	LanguageAllowed   string `json:"language_allowed" form:"language_allowed" query:"language_allowed" validate:"required,max=255"` // E.g.    cpp,c,java,python
	BuildArg          string `json:"build_arg" form:"build_arg" query:"build_arg" validate:"max=255"`                               // E.g.  O2=false
	CompareScriptName string `json:"compare_script_name" form:"compare_script_name" query:"compare_script_name" validate:"required"`
}

type DeleteProblemRequest struct {
}

type CreateTestCaseRequest struct {
	Score  uint  `json:"score" form:"score" query:"score"` // 0 for 平均分配
	Sample *bool `json:"sample" form:"sample" query:"sample" validate:"required"`
	// input_file(required)
	// output_file(required)
}

type GetTestCaseInputFileRequest struct {
}

type GetTestCaseOutputFileRequest struct {
}

type UpdateTestCaseRequest struct {
	Score  uint  `json:"score" form:"score" query:"score"` // 0 for 平均分配
	Sample *bool `json:"sample" form:"sample" query:"sample" validate:"required"`
	// input_file(optional)
	// output_file(optional)
}

type DeleteTestCaseRequest struct {
}

type DeleteTestCasesRequest struct {
}

type GetProblemRequest struct {
}

type GetProblemsRequest struct {
	Search string `json:"search" form:"search" query:"search"`
	UserID uint   `json:"user_id" form:"user_id" query:"user_id" validate:"min=0,required_with=Tried Passed"`

	Limit  int `json:"limit" form:"limit" query:"limit" validate:"max=100,min=0"`
	Offset int `json:"offset" form:"offset" query:"offset" validate:"min=0"`

	Tried  bool `json:"tried" form:"tried" query:"tried"`
	Passed bool `json:"passed" form:"passed" query:"passed"`
}

type GetRandomProblem struct {
}

type GetUserProblemInfo struct {
}
