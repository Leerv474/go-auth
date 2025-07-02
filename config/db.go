package config

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func LoadDBConfig() *DBConfig {
	config := &DBConfig{
		Host:     GetEnv("DB_HOST"),
		Port:     GetEnv("DB_PORT"),
		User:     GetEnv("DB_USER"),
		Password: GetEnv("DB_PASSWORD"),
		DBName:   GetEnv("DB_NAME"),
		SSLMode:  GetEnv("DB_SSLMODE"),
	}
	return config
}

func (config *DBConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)
}

func (config *DBConfig) DB_URL_string() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.DBName, config.SSLMode,
	)
}

func RunMigrationsUp(migrationPath, dbURL string) {
	m, err := migrate.New(
		"file://"+migrationPath,
		dbURL,
	)
	fmt.Println(dbURL)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Database migrated successfully")
}
