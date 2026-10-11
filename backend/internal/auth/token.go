package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTTL  = 15 * time.Minute
	RefreshTTL = 30 * 24 * time.Hour
	issuer     = "smart-budget"
	audience   = "smart-budget-api"
)

var ErrUnauthorized = errors.New("authentication required")

type Claims struct {
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

type Tokens struct {
	key []byte
}

func NewTokens(secret string) (*Tokens, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT_SECRET must contain at least 32 bytes")
	}

	return &Tokens{key: []byte(secret)}, nil
}

func (t *Tokens) Issue(userID, sessionID string, now time.Time) (string, error) {
	claims := Claims{SessionID: sessionID, RegisteredClaims: jwt.RegisteredClaims{
		Issuer: issuer, Subject: userID, Audience: jwt.ClaimStrings{audience},
		IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(AccessTTL)),
	}}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.key)
}

func (t *Tokens) Parse(raw string, now time.Time) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(_ *jwt.Token) (any, error) { return t.key, nil },
		jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(issuer), jwt.WithAudience(audience),
		jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithTimeFunc(func() time.Time { return now }))
	if err != nil || claims.Subject == "" || claims.SessionID == "" || claims.IssuedAt == nil {
		return nil, ErrUnauthorized
	}

	return claims, nil
}

func Hash(token string) string {
	sum := sha256.Sum256([]byte(token))

	return hex.EncodeToString(sum[:])
}
