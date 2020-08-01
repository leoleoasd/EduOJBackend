package request

type GetUserRequest struct {
}

type GetUsersRequest struct { //min=1 can be removed
	Username string `json:"username" form:"username" query:"username" validate:"min=1,max=30,printascii"`
	Nickname string `json:"username" form:"username" query:"username" validate:"min=1,max=30"`

	Limit  int `json:"limit" form:"limit" query:"limit" validate:"max=100,min=1"`
	Offset int `json:"offset" form:"offset" query:"offset" validate:"min=0"`

	OrderBy string `json:"order_by" form:"order_by" query:"order_by"`
}