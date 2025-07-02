package config

import (
	"strconv"
	"time"
)

type JWTConfig struct {
	Secret          string
	AccessDuration  time.Duration
	RefreshDuration time.Duration
	RefreshLength   int
}

func LoadJWTConfig() (*JWTConfig, error) {
	accessDuration, err := strconv.Atoi(GetEnv("JWT_ACCESS_EXPIRATION"))
	if err != nil {
		return nil, err
	}
	refreshDuration, err := strconv.Atoi(GetEnv("JWT_REFRESH_EXPIRATION"))
	if err != nil {
		return nil, err
	}
	refreshLength, err := strconv.Atoi(GetEnv("JWT_REFRESH_LENGTH"))
	if err != nil {
		return nil, err
	}

	config := &JWTConfig{
		Secret:          GetEnv("JWT_SECRET"),
		AccessDuration:  time.Duration(accessDuration),
		RefreshDuration: time.Duration(refreshDuration),
		RefreshLength:   refreshLength,
	}
	return config, nil
}
