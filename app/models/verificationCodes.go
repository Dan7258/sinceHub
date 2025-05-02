package models

import "time"

type VerificationEmail struct {
	Code     string
	TimeLife time.Time
}

type ChangeEmail struct {
	Code     string
	TimeLife time.Time
}

type ChangePassword struct {
	Code     string
	TimeLife time.Time
}

type VerifyProfile struct {
	Profile Profiles
	Code    string `json:"code"`
}

var VerificationEmailCodes = make(map[string]VerificationEmail)

var ChangeEmailCodes = make(map[uint64]ChangeEmail)

var ChangePasswordCodes = make(map[uint64]ChangePassword)

func SetVerificationEmailCode(key, code string, timeLife time.Duration) {
	VerificationEmailCodes[key] = VerificationEmail{
		Code:     code,
		TimeLife: time.Now().UTC().Add(timeLife),
	}
}

func GetVerificationEmailCode(key string) (string, bool) {
	data, ok := VerificationEmailCodes[key]
	if !ok || time.Now().UTC().After(data.TimeLife) {
		return "", false
	}
	return data.Code, ok
}
