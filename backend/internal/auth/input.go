package auth

import (
	"errors"
	"net/mail"
	"strings"
	"time"
	_ "time/tzdata"
	"unicode/utf8"
)

var (
	ErrInvalidEmail       = errors.New("invalid_email")
	ErrInvalidPassword    = errors.New("invalid_password")
	ErrInvalidTimezone    = errors.New("invalid_timezone")
	ErrInvalidCredentials = errors.New("invalid_credentials")
	ErrAccountExists      = errors.New("account_exists")
	ErrTooManyAttempts    = errors.New("too_many_attempts")
)

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterInput struct {
	LoginInput
	Timezone string `json:"timezone"`
}

func (input *LoginInput) validate() error {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	address, err := mail.ParseAddress(input.Email)
	if err != nil || address.Address != input.Email || len(input.Email) > 254 {
		return ErrInvalidEmail
	}

	if !utf8.ValidString(input.Password) || utf8.RuneCountInString(input.Password) < 8 || len(input.Password) > 72 {
		return ErrInvalidPassword
	}

	return nil
}

func (input *RegisterInput) validate() error {
	if err := input.LoginInput.validate(); err != nil {
		return err
	}

	if input.Timezone == "" {
		input.Timezone = "UTC"
	}

	if input.Timezone == "Local" {
		return ErrInvalidTimezone
	}

	if _, err := time.LoadLocation(input.Timezone); err != nil {
		return ErrInvalidTimezone
	}

	return nil
}
