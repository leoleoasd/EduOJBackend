package utils

import (
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/database/models"
	"sync"
)

var inviteCodeLock sync.Mutex

func GenerateInviteCode() (code string) {
	inviteCodeLock.Lock()
	defer inviteCodeLock.Unlock()
	var count int64 = 1
	for count > 0 {
		// 5: Fixed invite code length
		code = RandStr(5)
		PanicIfDBError(base.DB.Model(models.Class{}).Where("invite_code = ?", code).Count(&count),
			"could not check if invite code crashed for generating invite code")
	}
	return
}
