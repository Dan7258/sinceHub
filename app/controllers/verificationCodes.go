package controllers

import (
	"crypto/rand"
	"fmt"
	"github.com/revel/revel"
	"math/big"
	"scinceHub/app/models"
	"time"
)

type codeType int

const (
	emailCode codeType = iota
	passwordCode
)

func (p Profiles) GenerateRandomNumber() (*big.Int, error) {
	max := big.NewInt(100000)
	randomNumber, err := rand.Int(rand.Reader, max)
	if err != nil {
		fmt.Println("Ошибка при генерации случайного числа:", err)
		return nil, err
	}
	return randomNumber, nil
}

func SetVerifyCode(key, code string, timeLife time.Duration, codeType codeType) {
	switch codeType {
	case emailCode:
		SetVerificationEmailCode(key, code, timeLife)
	case passwordCode:

	}
}

func SetVerificationEmailCode(key, code string, timeLife time.Duration) {
	key = key + "_" + "verifyEmailCode"
	err := models.SetDataInRedis(key, []byte(code), timeLife)
	if err != nil {
		revel.AppLog.Error(err.Error())
		models.SetVerificationEmailCode(key, code, timeLife)
	} else {
		revel.AppLog.Debug("установили код верификации " + code + "по ключу: " + key)
	}

}

func GetVerificationEmailCode(key string) (string, bool) {
	value, err := models.GetDataFromRedis(key + "_" + "verifyEmailCode")
	revel.AppLog.Info(key + "_" + "verifyEmailCode")
	if err != nil || value == nil {
		revel.AppLog.Error(err.Error())
		return models.GetVerificationEmailCode(key)
	}
	revel.AppLog.Info("получили код верификации из Redis" + string(value))
	return string(value), true
}
