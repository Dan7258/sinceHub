package models

import "time"

type VerificationEmail struct {
	Code      string
	ExpiresAt time.Time
}

type ChangeEmail struct {
	Code      string
	ExpiresAt time.Time
}

type ChangePassword struct {
	Code      string
	ExpiresAt time.Time
}

type VerifyProfile struct {
	Profile Profiles
	Code    string `json:"code"`
}

var VerificationEmailCodes = make(map[string]VerificationEmail)

var ChangeEmailCodes = make(map[uint64]ChangeEmail)

var ChangePasswordCodes = make(map[uint64]ChangePassword)
