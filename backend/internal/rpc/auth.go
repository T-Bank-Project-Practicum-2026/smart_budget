package rpc

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"time"

	"connectrpc.com/connect"
)

func Authenticate(token string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if token == "" || subtle.ConstantTimeCompare([]byte(req.Header().Get("Authorization")), []byte("Bearer "+token)) != 1 {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid service credential"))
			}

			return next(ctx, req)
		}
	}
}

func Client(client *http.Client) *http.Client {
	if client != nil {
		return client
	}

	return &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
}

func ValidID(id string) bool { return id != "" && len(id) <= 100 }
