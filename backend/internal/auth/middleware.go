package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string
		var errResp ErrorResponse
		errResp.Timestamp = time.Now()

		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			cookie, err := r.Cookie("auth")
			if err != nil {
				errResp.Message = "Unauthorized: missing auth token"
				errResp.Status = http.StatusUnauthorized
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(errResp.Status)
				json.NewEncoder(w).Encode(errResp)
				return
			}
			tokenString = cookie.Value
		}

		claims, err := ParseToken(tokenString)
		if err != nil {
			errResp.Message = fmt.Sprintf("Invalid token: %v", err)
			errResp.Status = http.StatusUnauthorized
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(errResp.Status)
			json.NewEncoder(w).Encode(errResp)
			return
		}

		ctx := WithUserID(r.Context(), claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

type contextKey string

const userIDKey contextKey = "userID"

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func GetUserIDFromContext(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		return "", errors.New("user ID not found in context")
	}
	return userID, nil
}
