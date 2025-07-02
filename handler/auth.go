package handler

import (
	"encoding/json"
	"jwt-auth/model"
	"jwt-auth/service"
	"net"
	"net/http"
	"strings"
	"time"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type Handler struct {
	AuthService service.AuthService
}

type HTTPError struct {
	Message string `json:"message" example:"error message"`
}

// RegisterHandler godoc
// @Summary      Register new user
// @Description  Registers a new user with username and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        registerRequest body RegisterRequest true "User registration info"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  HTTPError
// @Failure      500  {object}  HTTPError
// @Router       /register [post]
func (h *Handler) RegisterHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Method now allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	err := json.NewDecoder(request.Body).Decode(&req)
	if err != nil {
		http.Error(writer, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user := model.User{
		Username: req.Username,
		Password: req.Password,
	}

	_, err = h.AuthService.RegisterUser(request.Context(), user)
	if err != nil {
		http.Error(writer, "Internal server error "+err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	json.NewEncoder(writer).Encode(map[string]interface{}{})
}

// LoginHandler godoc
// @Summary      User login
// @Description  Login with username and password, returns access and refresh tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        loginRequest body LoginRequest true "User login info"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  HTTPError
// @Failure      401  {object}  HTTPError
// @Router       /login [post]
func (h *Handler) LoginHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	err := json.NewDecoder(request.Body).Decode(&req)
	if err != nil {
		http.Error(writer, "Invalid JSON", http.StatusBadRequest)
	}

	user := model.User{
		Username: req.Username,
		Password: req.Password,
	}

	ip, _, _ := net.SplitHostPort(request.RemoteAddr)

	userId, accessToken, refreshToken, err := h.AuthService.LoginUser(request.Context(), user, request.Header.Get("User-Agent"), ip)
	if err != nil {
		http.Error(writer, "Invalid username or password: "+err.Error(), http.StatusUnauthorized)
		return
	}

	http.SetCookie(writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Authorization", "Bearer "+accessToken)
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(map[string]interface{}{
		"id": userId,
	})
}

// RefreshTokensHandler godoc
// @Summary      Refresh tokens
// @Description  Refresh access and refresh tokens using a refresh token cookie
// @Tags         auth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  HTTPError
// @Failure      400  {object}  HTTPError
// @Router       /refresh [get]
func (h *Handler) RefreshTokensHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := request.Cookie("refresh_token")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(writer, "Unauthorized - no cookie", http.StatusUnauthorized)
			return
		}
		http.Error(writer, "Bad request", http.StatusBadRequest)
		return
	}
	ip, _, _ := net.SplitHostPort(request.RemoteAddr)

	refreshTokenString := cookie.Value
	accessTokenHeader := request.Header.Get("Authorization")
	accessToken := strings.TrimPrefix(accessTokenHeader, "Bearer ")
	newAccessToken, newRefreshToken, err := h.AuthService.RefreshTokens(request.Context(), accessToken, request.Header.Get("User-Agent"), refreshTokenString, ip)
	if err != nil {
		http.Error(writer, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}
	http.SetCookie(writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Path:     "/",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Authorization", "Bearer "+newAccessToken)
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(map[string]interface{}{})
}

// LogoutHandler godoc
// @Summary      Logout user
// @Description  Logs out the user by invalidating the refresh token cookie
// @Tags         auth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  HTTPError
// @Failure      400  {object}  HTTPError
// @Failure      500  {object}  HTTPError
// @Router       /logout [get]
func (h *Handler) LogoutHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := request.Cookie("refresh_token")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(writer, "Unauthorized - no cookie", http.StatusUnauthorized)
			return
		}
		http.Error(writer, "Bad request", http.StatusBadRequest)
		return
	}

	refreshTokenString := cookie.Value
	err = h.AuthService.LogoutUser(request.Context(), refreshTokenString)
	if err != nil {
		http.Error(writer, "Internal error", http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(map[string]interface{}{})
}

// UserAccessPointHandler godoc
// @Summary      Get user ID from access token
// @Description  Retrieves the user ID from the access token in Authorization header
// @Tags         auth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  HTTPError
// @Failure      405  {object}  HTTPError
// @Router       /user [get]
func (h *Handler) UserAccessPointHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	accessTokenHeader := request.Header.Get("Authorization")
	accessToken := strings.TrimPrefix(accessTokenHeader, "Bearer ")
	userId, err := h.AuthService.UserAccessPoint(request.Context(), accessToken)
	if err != nil {
		http.Error(writer, "Access denied", http.StatusUnauthorized)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(map[string]interface{}{
		"userId": userId,
	})
}
