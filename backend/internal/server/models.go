package server

import (
	"smart_budget/internal/auth"
	"smart_budget/internal/importer"
	"smart_budget/internal/push"
)

type SuccessResponse struct {
	Success bool `json:"success" example:"true"`
	Data    any  `json:"data"`
}

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"invalid_request"`
}

type UserResponse struct {
	Success bool      `json:"success" example:"true"`
	Data    auth.User `json:"data"`
}

type ImportResponse struct {
	Success bool            `json:"success" example:"true"`
	Data    importer.Import `json:"data"`
}

type ImportsResponse struct {
	Success bool              `json:"success" example:"true"`
	Data    []importer.Import `json:"data"`
}

type OperationsResponse struct {
	Success bool                 `json:"success" example:"true"`
	Data    []importer.Operation `json:"data"`
}

type PublicKeySuccessResponse struct {
	Success bool              `json:"success" example:"true"`
	Data    PublicKeyResponse `json:"data"`
}

type QueuedSuccessResponse struct {
	Success bool           `json:"success" example:"true"`
	Data    QueuedResponse `json:"data"`
}

type NotificationsResponse struct {
	Success bool         `json:"success" example:"true"`
	Data    []push.Event `json:"data"`
}

type PublicKeyResponse struct {
	PublicKey string `json:"publicKey"`
}

type QueuedResponse struct {
	ID     string `json:"id"`
	Status string `json:"status" enums:"queued"`
}

// Тело подписки передаётся в notifications без изменений.
type SubscriptionInput struct {
	Endpoint string `json:"endpoint" format:"uri" maxLength:"4096"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

type UnsubscribeInput struct {
	Endpoint string `json:"endpoint" format:"uri"`
}
