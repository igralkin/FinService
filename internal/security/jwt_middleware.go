package security

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
)

// ContextKey is a type-safe string key used for storing values in context.
type ContextKey string

// ContextKeyUserID is the context key for storing the user ID.
const ContextKeyUserID ContextKey = "userID"

// AuthMiddleware validates JWT tokens from the Authorization header.
// If valid, injects the userID into the request context.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Warn("Missing Authorization header")
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		// Extract token from header
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			log.Warn("Malformed Authorization header")
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		claims := &jwt.RegisteredClaims{}
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			log.Error("JWT_SECRET is not set in environment")
			http.Error(w, "Server misconfiguration", http.StatusInternalServerError)
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.WithField("alg", token.Header["alg"]).Warn("Unexpected signing method")
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			log.WithError(err).Warn("Invalid or expired JWT token")
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		userID := claims.Subject
		if userID == "" {
			log.Warn("Token does not contain subject (userID)")
			http.Error(w, "Invalid token: subject missing", http.StatusUnauthorized)
			return
		}

		log.WithField("userID", userID).Info("JWT token verified successfully")
		ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext extracts the user ID from request context.
// Returns an error if the user ID is not found or is not a string.
func GetUserIDFromContext(r *http.Request) (string, error) {
	value := r.Context().Value(ContextKeyUserID)
	userID, ok := value.(string)
	if !ok || userID == "" {
		log.Warn("userID not found in context or invalid type")
		return "", errors.New("userID not found in context")
	}

	log.WithField("userID", userID).Debug("Extracted userID from context")
	return userID, nil
}
