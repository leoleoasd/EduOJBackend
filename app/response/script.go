package response

import "github.com/EduOJ/backend/app/response/resource"

// JudgerGetScriptResponse
// Will redirect to download url
type JudgerGetScriptResponse struct {
}

type CreateScriptResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.Script `json:"script"`
	} `json:"data"`
}

type GetScriptResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.Script `json:"script"`
	} `json:"data"`
}

type GetScriptFileResponse struct {
	// Redirect to presigned url of script file
}

type GetScriptsResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		Scripts []*resource.Script `json:"scripts"`
		Total   int                `json:"total"`
		Count   int                `json:"count"`
		Offset  int                `json:"offset"`
		Prev    *string            `json:"prev"`
		Next    *string            `json:"next"`
	} `json:"data"`
}

type UpdateScriptResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    struct {
		*resource.Script `json:"script"`
	} `json:"data"`
}

type DeleteScriptResponse struct {
	// No special response
}
