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

var ChangeEmailCodes = make(map[string]ChangeEmail)

var ChangePasswordCodes = make(map[string]ChangePassword)

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

func DeleteVerificationEmailCode(key string) {
	delete(VerificationEmailCodes, key)
}

func SetChangeEmailCode(key, code string, timeLife time.Duration) {
	ChangeEmailCodes[key] = ChangeEmail{
		Code:     code,
		TimeLife: time.Now().UTC().Add(timeLife),
	}
}

func GetChangeEmailCode(key string) (string, bool) {
	data, ok := ChangeEmailCodes[key]
	if !ok || time.Now().UTC().After(data.TimeLife) {
		return "", false
	}
	return data.Code, ok
}

func DeleteChangeEmailCode(key string) {
	delete(ChangeEmailCodes, key)
}

func SetChangePasswordCode(key, code string, timeLife time.Duration) {
	ChangePasswordCodes[key] = ChangePassword{
		Code:     code,
		TimeLife: time.Now().UTC().Add(timeLife),
	}
}

func GetChangePasswordCode(key string) (string, bool) {
	data, ok := ChangePasswordCodes[key]
	if !ok || time.Now().UTC().After(data.TimeLife) {
		return "", false
	}
	return data.Code, ok
}

func DeleteChangePasswordCode(key string) {
	delete(ChangePasswordCodes, key)
}
