package response

import (
	"github.com/EduOJ/backend/database/models"
	"time"
)

type GetTaskResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		RunID             uint            `json:"run_id"`
		Language          models.Language `json:"language"`
		TestCaseID        uint            `json:"test_case_id"`
		InputFile         string          `json:"input_file"`  // pre-signed url
		OutputFile        string          `json:"output_file"` // same as above
		CodeFile          string          `json:"code_file"`
		TestCaseUpdatedAt time.Time       `json:"test_case_updated_at"`
		MemoryLimit       uint64          `json:"memory_limit"` // Byte
		TimeLimit         uint            `json:"time_limit"`   // ms
		BuildArg          string          `json:"build_arg"`    // E.g.  O2=false
		CompareScript     models.Script   `json:"compare_script"`
	} `json:"data"`
}
