package request

type CreateCommentRequest struct {
	Content string `json:"content" form:"content" query:"content" `

	FatherID   uint   `json:"father_id" form:"father_id" query:"father_id"`
	TargetID   uint   `json:"target_id" form:"target_id" query:"target_id" `
	TargetType string `json:"target_type" form:"target_type" query:"target_type"`
}

type GetCommentRequest struct {
	TargetType string `json:"target_type" form:"target_type" query:"target_type"`
	TargetID   uint   `json:"target_id" form:"target_id" query:"target_id"`
	begin      uint   `json:"begin" form:"begin" query:"begin"` //param for paginator
	end        uint   `json:"end" form:"end" query:"end "`
}


type AddReacitonRequest struct {
	EmojiType string `json:"emoji_type" form:"emoji_type" query:"emoji_type"`
	ReactionID uint `json:"reaction_id" form:"reaction_id" query:"reaction_id"`
	IFAddAction bool `json:"if_add_action" form:"if_add_action" query:"if_add_action"`
}
