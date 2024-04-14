package middleware

import (
	"context"
	"net/http"
	"strings"
)

const (
	AdminRole = 1
	UserRole  = 2
)

func AuthAdminPermission(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")

		if token != "user_token" && token != "admin_token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if token == "user_token" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func AuthUserPermission(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")

		if token != "user_token" && token != "admin_token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		var role int8
		if token == "user_token" {
			role = UserRole
		} else if token == "admin_token" {
			role = AdminRole
		}

		ctx := context.WithValue(r.Context(), "role", role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
