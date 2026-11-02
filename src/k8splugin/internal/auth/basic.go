package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/config"
)

func NewBasicAuth(configuration config.AuthConfiguration) func(http.Handler) http.Handler {
	type credential struct {
		username [sha256.Size]byte
		password [sha256.Size]byte
	}
	users := make([]credential, len(configuration.Users))
	for index, user := range configuration.Users {
		users[index] = credential{
			username: sha256.Sum256([]byte(user.Username)),
			password: sha256.Sum256([]byte(user.Password)),
		}
	}
	return func(next http.Handler) http.Handler {
		if !configuration.Enabled {
			return next
		}
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.Method == http.MethodGet && request.URL.EscapedPath() == "/v1/healthcheck" {
				next.ServeHTTP(writer, request)
				return
			}
			username, password, ok := request.BasicAuth()
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			matched := 0
			for _, user := range users {
				matched |= subtle.ConstantTimeCompare(usernameHash[:], user.username[:]) &
					subtle.ConstantTimeCompare(passwordHash[:], user.password[:])
			}
			if !ok || matched != 1 {
				writer.Header().Set("WWW-Authenticate", `Basic realm="multicloud"`)
				http.Error(writer, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(writer, request)
		})
	}
}
