package utils

import (
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/database/models"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"sync"
	"time"
)

var initAuth sync.Once
var SessionTimeout time.Duration
var RememberMeTimeout time.Duration
var SessionCount int

func initAuthConfig() {
	viper.SetDefault("auth.session_timeout", 1200)
	viper.SetDefault("auth.remember_me_timeout", 604800)
	viper.SetDefault("auth.session_count", 10)
	SessionTimeout = time.Second * viper.GetDuration("auth.session_timeout")
	RememberMeTimeout = time.Second * viper.GetDuration("auth.remember_me_timeout")
	SessionCount = viper.GetInt("auth.session_timeout")
}

func IsTokenExpired(token models.Token) bool {
	initAuth.Do(initAuthConfig)
	if token.RememberMe {
		return token.UpdatedAt.Add(RememberMeTimeout).Before(time.Now())
	} else {
		return token.UpdatedAt.Add(SessionTimeout).Before(time.Now())
	}
}

// TODO: Use this function in timed tasks
func CleanUpExpiredTokens() error {
	initAuth.Do(initAuthConfig)
	var users []models.User
	err := base.DB.Model(models.User{}).Find(&users).Error
	if err != nil {
		return errors.Wrap(err, "could not find users")
	}
	for _, user := range users {
		var tokens []models.Token
		var tokenIds []uint
		storedTokenCount := 0
		err = base.DB.Where("user_id = ?", user.ID).Order("updated_at desc").Find(&tokens).Error
		// TODO: select updated_at > xxx limit 5;  delete not in
		if err != nil {
			return errors.Wrap(err, "could not find tokens")
		}
		for _, token := range tokens {
			if storedTokenCount < SessionCount && !IsTokenExpired(token) {
				storedTokenCount++
				continue
			}
			tokenIds = append(tokenIds, token.ID)
		}
		err = base.DB.Delete(&models.Token{}, "id in (?)", tokenIds).Error
		if err != nil {
			return errors.Wrap(err, "could not delete tokens")
		}
	}
	return nil
}
