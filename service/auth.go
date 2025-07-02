package service

import (
	"context"
	"database/sql"
	"jwt-auth/auth"
	"jwt-auth/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	RegisterUser(ctx context.Context, user model.User) (int64, error)
	LoginUser(ctx context.Context, user model.User, userAgent, agentIp string) (int64, string, string, error)
	RefreshTokens(ctx context.Context, accessToken, userAgent, agentIp, refreshToken string) (string, string, error)
	LogoutUser(ctx context.Context, refreshTokenString string) error
	UserAccessPoint(ctx context.Context, accessToken string) (int64, error)
}

type authService struct {
	db         *sql.DB
	jwtManager *auth.JWTManager
}

func NewAuthService(db *sql.DB, jwtManager *auth.JWTManager) AuthService {
	return &authService{
		db:         db,
		jwtManager: jwtManager,
	}
}

func (a *authService) RegisterUser(ctx context.Context, user model.User) (int64, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	var userId int64

	tx, err := a.db.Begin()
	if err != nil {
		return 0, err
	}

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
	err = tx.QueryRowContext(ctx,
		"INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id",
		user.Username, string(hashedPassword),
	).Scan(&userId)
	if err != nil {
		return 0, err
	}

	return userId, nil
}

func (a *authService) LoginUser(ctx context.Context, user model.User, userAgent, agentIp string) (int64, string, string, error) {
	var userId int64
	var storedPassword string
	err := a.db.QueryRowContext(ctx,
		"SELECT id, password FROM users WHERE username = $1",
		user.Username,
	).Scan(&userId, &storedPassword)
	if err != nil {
		return 0, "", "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password))
	if err != nil {
		return 0, "", "", err
	}

	pairID := uuid.New().String()
	accessToken, err := a.jwtManager.GenerateAccessToken(userId, pairID)
	if err != nil {
		return 0, "", "", err
	}

	refreshToken, err := a.jwtManager.GenerateRefreshToken(userId, pairID, userAgent, agentIp)
	if err != nil {
		return 0, "", "", err
	}

	return userId, accessToken, refreshToken, nil
}

func (a *authService) RefreshTokens(ctx context.Context, accessToken, userAgent, agentIp, refreshToken string) (string, string, error) {
	userId, keyPairID, err := a.jwtManager.ValidateAccessToken(accessToken)
	if err != nil {
		return "", "", err
	}
	err = a.jwtManager.ValidateRefreshToken(ctx, userId, refreshToken, userAgent, keyPairID, agentIp)
	if err != nil {
		return "", "", err
	}

	pairID := uuid.New().String()

	newAccessToken, err := a.jwtManager.GenerateAccessToken(userId, pairID)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := a.jwtManager.GenerateRefreshToken(userId, pairID, userAgent, agentIp)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (a *authService) LogoutUser(ctx context.Context, refreshTokenString string) error {
	err := a.jwtManager.InvalidateRefreshToken(ctx, refreshTokenString)
	if err != nil {
		return err
	}
	return nil
}

func (a *authService) UserAccessPoint(ctx context.Context, accessToken string) (int64, error) {
	extractedUserId, _, err := a.jwtManager.ValidateAccessToken(accessToken)
	if err != nil {
		return 0, err
	}
	var userId int64
	err = a.db.QueryRowContext(
		ctx,
		"select id from users where id = $1",
		extractedUserId,
	).Scan(&userId)
	if err != nil {
		return 0, err
	}
	return userId, nil
}
