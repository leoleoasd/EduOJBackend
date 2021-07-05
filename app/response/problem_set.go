package response

import "github.com/EduOJ/backend/app/response/resource"

type CreateProblemSetResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.ProblemSetDetail `json:"problem_set"`
	} `json:"data"`
}

type CloneProblemSetResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.ProblemSetDetail `json:"problem_set"`
	} `json:"data"`
}

type GetProblemSetResponseForAdmin struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.ProblemSetDetail `json:"problem_set"`
	} `json:"data"`
}

type GetProblemSetResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.ProblemSet `json:"problem_set"`
	} `json:"data"`
}

type GetProblemSetResponseSummary struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.ProblemSetSummary `json:"problem_set"`
	} `json:"data"`
}

type UpdateProblemSetResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.ProblemSetDetail `json:"problem_set"`
	} `json:"data"`
}

type AddProblemsToSetResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.ProblemSetDetail `json:"problem_set"`
	} `json:"data"`
}

type DeleteProblemsFromSetResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.ProblemSetDetail `json:"problem_set"`
	} `json:"data"`
}

type GetProblemSetProblemResponseForAdmin struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.ProblemForAdmin `json:"problem"`
	} `json:"data"`
}

type GetProblemSetProblemResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.Problem `json:"problem"`
	} `json:"data"`
}

type GetGradesResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.ProblemSetWithGrades `json:"grades"`
	} `json:"data"`
}

type RefreshGradesResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.ProblemSetWithGrades `json:"grades"`
	} `json:"data"`
}
