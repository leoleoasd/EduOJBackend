package utils

import (
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/database/models"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (hashed string) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	hashed = string(hash)
	return
}

func VerifyPassword(password, hashed string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) == nil
}

func GetToken(tokenString string) (token models.Token, err error) {
	t := models.Token{}
	err = base.DB.Preload("User").Where("token = ?", tokenString).First(&t).Error
	if err != nil {
		return
	}
	return t, nil
}
