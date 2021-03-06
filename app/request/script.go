package request

// JudgerGetScriptRequest
// No request params / bodies
type JudgerGetScriptRequest struct {
}

type CreateScriptRequest struct {
	Name string `json:"name" form:"name" query:"name" validate:"required,max=255"`
	// file(required)
}

type GetScriptRequest struct {
}

type GetScriptFileRequest struct {
}

type GetScriptsRequest struct {
	Limit  int `json:"limit" form:"limit" query:"limit" validate:"max=100,min=0"`
	Offset int `json:"offset" form:"offset" query:"offset" validate:"min=0"`
}

type UpdateScriptRequest struct {
	Name string `json:"name" form:"name" query:"name" validate:"required,max=255"`
	// file(optional)
}

type DeleteScriptRequest struct {
}
