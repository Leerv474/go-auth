package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"jwt-auth/config"
	"jwt-auth/webhook"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type JWTManager struct {
	secretKey            string
	accessTokenDuration  time.Duration
	refreshTokenLength   int
	refreshTokenDuration time.Duration
	db                   *sql.DB
}

type Claims struct {
	UserID    int64 `json:"user_id"`
	KeyPairID string
	jwt.RegisteredClaims
}

func NewJWTManager(jwtConfig *config.JWTConfig, DB *sql.DB) *JWTManager {
	return &JWTManager{
		secretKey:            jwtConfig.Secret,
		accessTokenDuration:  jwtConfig.AccessDuration,
		refreshTokenLength:   jwtConfig.RefreshLength,
		refreshTokenDuration: jwtConfig.RefreshDuration,
		db:                   DB,
	}
}

func (j *JWTManager) GenerateAccessToken(userID int64, keyPairID string) (string, error) {
	claims := &Claims{
		UserID:    userID,
		KeyPairID: keyPairID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	return token.SignedString([]byte(j.secretKey))
}

func (j *JWTManager) GenerateRefreshToken(userID int64, keyPairID, userAgent, agentIp string) (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("Failed to generate random bytes: %w", err)
	}

	token := base64.StdEncoding.EncodeToString(bytes)
	hashedToken, err := bcrypt.GenerateFromPassword(bytes, bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("Failed to hash random bytes: %w", err)
	}

	tokenId := uuid.New().String()

	tx, err := j.db.Begin()

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	_, err = tx.Exec(
		"insert into refresh_tokens (id, token, userId, keyPairId, userAgent, agentIp, issued_at, expires_at) values($1, $2, $3, $4, $5, $6, $7, $8)",
		tokenId, hashedToken, userID, keyPairID, userAgent, agentIp, time.Now(), time.Now().Add(j.refreshTokenDuration),
	)
	if err != nil {
		return "", fmt.Errorf("Failed to save refresh token: %w", err)
	}

	return tokenId + ":" + token, nil
}

func (j *JWTManager) ValidateAccessToken(accessToken string) (int64, string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return 0, "", err
	}

	if !token.Valid {
		return 0, "", fmt.Errorf("invalid token")
	}

	return claims.UserID, claims.KeyPairID, nil
}

func (j *JWTManager) ValidateRefreshToken(ctx context.Context, userId int64, keyPairID, userAgent, agentIp, refreshTokenString string) error {
	parts := strings.Split(refreshTokenString, ":")
	if len(parts) != 2 {
		return fmt.Errorf("Invalid token format")
	}
	tokenId := parts[0]
	refreshTokenBase64 := parts[1]
	refreshToken, err := base64.StdEncoding.DecodeString(refreshTokenBase64)
	if err != nil {
		return fmt.Errorf("Failed base64 decode")
	}

	var extractedUserId int64
	var hashedToken []byte
	var expires_at time.Time
	var extractedKeyPairId, extractedUserAgent, extractedAgentIp string

	err = j.db.QueryRowContext(
		ctx,
		"select token, userId, keyPairId, userAgent, agentIp, expires_at from refresh_tokens where id = $1",
		tokenId,
	).Scan(&hashedToken, &extractedUserId, &keyPairID, &userAgent, &extractedAgentIp, &expires_at)

	if extractedUserId != userId {
		return fmt.Errorf("User mismatch")
	}
	if extractedKeyPairId != keyPairID {
		return fmt.Errorf("Key pair mismatch")
	}
	if extractedUserAgent != userAgent {
		return fmt.Errorf("User-Agent mismatch")
	}
	if extractedAgentIp != agentIp {
		go webhook.NotifyWebhook(userId, agentIp)
	}
	err = bcrypt.CompareHashAndPassword(hashedToken, refreshToken)
	if err != nil || time.Now().After(expires_at) {
		_, err = j.db.ExecContext(
			ctx,
			"delete from refresh_tokens where id = $1",
			tokenId,
		)
		return fmt.Errorf("Invalid token")
	}
	return nil
}

func (j *JWTManager) InvalidateRefreshToken(ctx context.Context, refreshTokenString string) error {
	parts := strings.Split(refreshTokenString, ":")
	if len(parts) != 2 {
		return fmt.Errorf("Invalid token format")
	}
	tokenId := parts[0]
	_, err := j.db.ExecContext(
		ctx,
		"delete from refresh_tokens where id = $1",
		tokenId,
	)
	if err != nil {
		return fmt.Errorf("Unknown refresh token")
	}
	return nil
}

