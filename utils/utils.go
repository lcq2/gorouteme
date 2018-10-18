package utils

import (
	"gorouteme/session"
	"net/http"
)

func LoginRequired(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := session.Manager().SessionGet(r)
		if s == nil || !s.UserLoggedIn() {
			http.Error(w, "User not logged in", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	})
}
