package router

import (
	"jwt-auth/handler"
	"jwt-auth/service"
	"net/http"

	"github.com/swaggo/http-swagger"
	_ "jwt-auth/docs"
)

func SetupRoutes(authService service.AuthService) http.Handler {
	handler := &handler.Handler{
		AuthService: authService,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/register", handler.RegisterHandler)
	mux.HandleFunc("/login", handler.LoginHandler)
	mux.HandleFunc("/refresh_tokens", handler.RefreshTokensHandler)
	mux.HandleFunc("/logout", handler.LogoutHandler)
	mux.HandleFunc("/user", handler.UserAccessPointHandler)

	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	return mux
}
