package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/silvernemesis/jwtproject/internal/security"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type errorResponse struct {
	Message string `json:"message"`
}

// HandleLogin implements a POST request to retrieve a token based on a loginRequest
func HandleLogin(w http.ResponseWriter, req *http.Request) {
	var requestData loginRequest
	if error := json.NewDecoder(req.Body).Decode(&requestData); error == nil {
		if user, error := security.VerifyUser(requestData.Username, requestData.Password); error == nil {
			if tokenString, error := user.CreateToken(); error == nil {
				json.NewEncoder(w).Encode(loginResponse{Token: tokenString})
			}
		} else {
			message := "invalid username or password"
			w.WriteHeader(http.StatusUnauthorized)
			message = fmt.Sprintf(`{"error": {"code": "%v", "message": "%v"}}`, http.StatusUnauthorized, message)
			w.Write([]byte(message))
		}
	}
}

// ValidateMiddleware implements a wrapper for requests that are protected by a JWT
func ValidateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		message := ""
		authorizationHeader := req.Header.Get("authorization")
		if authorizationHeader != "" {
			bearerToken := strings.Split(authorizationHeader, " ")
			if len(bearerToken) == 2 && bearerToken[0] == "Bearer" {
				claims, err := security.VerifyToken(bearerToken[1])
				if err == nil {
					ctx := context.WithValue(req.Context(), security.Claims, claims)
					next(w, req.WithContext(ctx))
					return
				}
				message = err.Error()
			} else {
				message = "authorization header invalid"
			}
		} else {
			message = "authorization header missing"
		}
		w.WriteHeader(http.StatusUnauthorized)
		message = fmt.Sprintf(`{"error": {"code": "%v", "message": "%v"}}`, http.StatusUnauthorized, message)
		w.Write([]byte(message))
	})
}
