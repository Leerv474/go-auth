package main

import (
	"database/sql"
	"log"
	"net/http"

	"jwt-auth/auth"
	"jwt-auth/config"
	"jwt-auth/db"
	"jwt-auth/router"
	"jwt-auth/service"
)

func main() {
	var DB *sql.DB
	dbConfig := config.LoadDBConfig()
	db.Init(dbConfig, &DB)

	jwtConfig, err := config.LoadJWTConfig()
	if err != nil {
		log.Fatalf("Failed to read JWT env variables: %w", err)
		return
	}
	jwtManager := auth.NewJWTManager(jwtConfig, DB)
	authService := service.NewAuthService(DB, jwtManager)

	handler := router.SetupRoutes(authService)

	serverPort := config.GetServerPort()
	log.Printf("Server running on port %s", serverPort)
	err = http.ListenAndServe(":"+serverPort, handler)
	if err != nil {
		log.Fatalf("Server failed: %w", err)
	}
}
