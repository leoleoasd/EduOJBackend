package controller

import (
	"encoding/json"
	"github.com/EduOJ/backend/app/request"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/utils"
	"github.com/EduOJ/backend/database/models"
	"github.com/labstack/echo/v4"
	"net/http"
)

// CreateComment creates a comment by father_id(0 for root node), target_id,target_type
func CreateComment(c echo.Context) error {
	req := request.CreateCommentRequest{}
	if err, ok := utils.BindAndValidate(&req, c); !ok {
		return err
	}

	user, ok := c.Get("user").(models.User)
	if !ok {
		panic("could not convert my user into type models.User")
	}

	TargetID := req.TargetID
	TargetType := req.TargetType

	content := req.Content

	fatherID := req.FatherID

	newReaction := models.Reaction{
		TargetType: "comment",
	}
	utils.PanicIfDBError(base.DB.Save(&newReaction), "could not save reaction")

	if fatherID != 0 {
		//father is comment
		var father models.Comment
		base.DB.Model(&models.Comment{}).First(&father, uint(fatherID))
		newComment := models.Comment{
			User:       user,
			Reaction:   newReaction,
			FatherID:   uint(fatherID),
			TargetType: father.TargetType,
			TargetID:   father.TargetID,
			Content:    content,
		}
		if father.FatherID != 0 {
			// father is comment node
			newComment.RecursiveFatherId = father.RecursiveFatherId
		} else {
			// father is root node
			newComment.RecursiveFatherId = father.ID
		}
		utils.PanicIfDBError(base.DB.Save(&newComment), "could not save comment")

	} else {
		//root node
		newComment := models.Comment{
			User:       user,
			Reaction:   newReaction,
			TargetType: TargetType,
			TargetID:   uint(TargetID),
			Content:    content,
		}
		utils.PanicIfDBError(base.DB.Save(&newComment), "could not save comment")
	}

	return c.JSON(http.StatusCreated, response.CreateCommentResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			Id uint
		}{
			uint(1),
		},
	})

}

// GetComment query a comment for a problem or a homework by id(uint) and type(string), returns rootNodes and noneRootNodes
func GetComment(c echo.Context) error {
	req := request.GetCommentRequest{}
	if err, ok := utils.BindAndValidate(&req, c); !ok {
		return err
	}
	if req.TargetType == "problem" {
		ids := req.TargetID
		var commentsNoneRoot []models.Comment
		var commentsRoot []models.Comment
		query := base.DB.Model(&models.Comment{}).Preload("User").Preload("Reaction")
		query.Order("ID").Find(&commentsRoot, " target_type = ? AND target_id = ? AND father_id = ?", "problem", uint(ids), 0)
		var RecursiveFatherIds []uint
		for _, v := range commentsRoot {
			RecursiveFatherIds = append(RecursiveFatherIds, v.ID)
		}

		base.DB.Model(&models.Comment{}).
			Preload("User").
			Preload("Reaction").
			Order("updated_at desc").
			Find(&commentsNoneRoot, "recursive_father_id in ?", RecursiveFatherIds)

		return c.JSON(http.StatusCreated, response.GetCommentResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				ComsRoot     []models.Comment
				ComsNoneRoot []models.Comment
			}{
				commentsRoot,
				commentsNoneRoot,
			},
		})
	} else {
		//todo: implement this.
		panic("we don't have this function now")
		return c.JSON(http.StatusBadRequest, response.ErrorResp("INVALID_STATUS", nil))
	}
}

// AddReaction makes a reaction, assert frontend have checked if the operation is illegaled
func AddReaction(c echo.Context) error {
	req := request.AddReacitonRequest{}
	if err, ok := utils.BindAndValidate(&req, c); !ok {
		return err
	}
	user, ok := c.Get("user").(models.User)
	if !ok {
		panic("could not convert my user into type models.User")
	}

	reactionId := req.ReactionID
	compType := req.IFAddAction
	typeId := req.EmojiType
	query := base.DB.Model(&models.Reaction{})
	var reaction models.Reaction
	query = query.First(&reaction, uint(reactionId))
	maps := make(map[string][]uint)
	if compType {
		if reaction.Details == "" {
			maps[typeId] = make([]uint, 2)
			maps[typeId][0] = 1
			maps[typeId][1] = uint(user.ID)
		} else {
			err := json.Unmarshal([]byte(reaction.Details), &maps)
			if err != nil {
				panic(err)
			}
			_, key := maps[typeId]
			if key {
				//operator is not zero
				for _, v := range maps[typeId] {
					if uint(v) == uint(user.ID) {
						return c.JSON(http.StatusBadRequest, response.ErrorResp("can't action twice!", nil))
					}
				}
				maps[typeId] = append(maps[typeId], uint(user.ID))
				maps[typeId][0] += 1
			} else {
				maps[typeId] = make([]uint, 2)
				maps[typeId][0] = 1
				maps[typeId][1] = uint(user.ID)
			}
		}
		jsonStr, err := json.Marshal(maps)
		if err != nil {
			panic("parse error")
			panic(err)
		}
		reaction.Details = string(jsonStr)
		base.DB.Save(&reaction)
	} else {
		//delete action
		if reaction.Details == "" {
			panic("this should not happen, logic in error!")
		} else {
			json.Unmarshal([]byte(reaction.Details), &maps)
			_, key := maps[typeId]
			if key {
				pos := -1
				for k, v := range maps[typeId] {
					//skip the first, standing for counts
					if k != 0 && v == uint(user.ID) {
						pos = k
						break
					}
				}
				if pos == -1 {
					panic("should not happen, you have not actived")
				} else {
					maps[typeId] = append(maps[typeId][:pos], maps[typeId][pos+1:]...)
					maps[typeId][0] -= 1
				}
				jsonStr, err := json.Marshal(maps)
				if err != nil {
					panic("parse error")
					panic(err)
				}
				reaction.Details = string(jsonStr)
				base.DB.Save(&reaction)
			} else {
				return c.JSON(http.StatusBadRequest, response.ErrorResp("you have not actived!", nil))
			}
		}
	}

	return c.JSON(http.StatusCreated, response.AddReactionResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			Cont string
		}{
			"you have successfully "+ req.EmojiType + "ed the comment",
		},
	})

}
