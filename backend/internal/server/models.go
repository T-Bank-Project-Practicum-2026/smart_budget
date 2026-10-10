package server

type ErrorResponse struct {
	Error string `json:"error"`
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
