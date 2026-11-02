package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/config"
)

func TestBasicAuth(t *testing.T) {
	configuration := config.AuthConfiguration{Enabled: true, Users: []config.AuthUser{
		{Username: "first", Password: "one"},
		{Username: "second", Password: "two:with:colons"},
	}}
	for _, test := range []struct {
		name, method, path, header, username, password string
		disabled                                       bool
		want                                           int
	}{
		{name: "missing", method: "GET", path: "/v1/instance", want: 401},
		{name: "first user", method: "POST", path: "/v1/instance", username: "first", password: "one", want: 204},
		{name: "second user", method: "GET", path: "/v1/instance", username: "second", password: "two:with:colons", want: 204},
		{name: "wrong password", method: "GET", path: "/v1/instance", username: "first", password: "wrong", want: 401},
		{name: "wrong username", method: "GET", path: "/v1/instance", username: "unknown", password: "one", want: 401},
		{name: "crossed users", method: "GET", path: "/v1/instance", username: "first", password: "two:with:colons", want: 401},
		{name: "malformed base64", method: "GET", path: "/v1/instance", header: "Basic !!!", want: 401},
		{name: "missing colon", method: "GET", path: "/v1/instance", header: "Basic Zmlyc3Q=", want: 401},
		{name: "other scheme", method: "GET", path: "/v1/instance", header: "Bearer token", want: 401},
		{name: "health", method: "GET", path: "/v1/healthcheck", want: 204},
		{name: "health query", method: "GET", path: "/v1/healthcheck?probe=true", want: 204},
		{name: "health head", method: "HEAD", path: "/v1/healthcheck", want: 401},
		{name: "health post", method: "POST", path: "/v1/healthcheck", want: 401},
		{name: "health trailing slash", method: "GET", path: "/v1/healthcheck/", want: 401},
		{name: "health suffix", method: "GET", path: "/v1/healthcheck-extra", want: 401},
		{name: "health child", method: "GET", path: "/v1/healthcheck/child", want: 401},
		{name: "health case", method: "GET", path: "/v1/Healthcheck", want: 401},
		{name: "health escaped", method: "GET", path: "/v1/%68ealthcheck", want: 401},
		{name: "unknown route", method: "GET", path: "/unknown", want: 401},
		{name: "disabled", method: "POST", path: "/v1/instance", disabled: true, want: 204},
		{name: "disabled malformed header", method: "GET", path: "/v1/instance", header: "Basic !!!", disabled: true, want: 204},
	} {
		t.Run(test.name, func(t *testing.T) {
			settings := configuration
			settings.Enabled = !test.disabled
			called := false
			handler := NewBasicAuth(settings)(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				called = true
				writer.WriteHeader(http.StatusNoContent)
			}))
			request := httptest.NewRequest(test.method, test.path, nil)
			request.RemoteAddr = "127.0.0.1:12345"
			if test.username != "" {
				request.SetBasicAuth(test.username, test.password)
			} else if test.header != "" {
				request.Header.Set("Authorization", test.header)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.want || called != (test.want == 204) {
				t.Fatalf("status = %d, handler called = %v, want %d", response.Code, called, test.want)
			}
			if test.want == 401 && response.Header().Get("WWW-Authenticate") != `Basic realm="multicloud"` {
				t.Fatal("missing Basic authentication challenge")
			}
			if test.want != 401 && response.Header().Get("WWW-Authenticate") != "" {
				t.Fatal("unexpected authentication challenge")
			}
		})
	}
}

func TestBasicAuthHTTPAndHTTPS(t *testing.T) {
	handler := NewBasicAuth(config.AuthConfiguration{Enabled: true, Users: []config.AuthUser{
		{Username: "first", Password: "one"},
	}})(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))
	for _, tlsEnabled := range []bool{false, true} {
		name := "HTTP"
		if tlsEnabled {
			name = "HTTPS"
		}
		t.Run(name, func(t *testing.T) {
			var server *httptest.Server
			if tlsEnabled {
				server = httptest.NewTLSServer(handler)
			} else {
				server = httptest.NewServer(handler)
			}
			defer server.Close()
			for _, authenticated := range []bool{false, true} {
				request, err := http.NewRequest(http.MethodGet, server.URL+"/v1/instance", nil)
				if err != nil {
					t.Fatal(err)
				}
				want := http.StatusUnauthorized
				if authenticated {
					request.SetBasicAuth("first", "one")
					want = http.StatusNoContent
				}
				response, err := server.Client().Do(request)
				if err != nil {
					t.Fatal(err)
				}
				response.Body.Close()
				if response.StatusCode != want {
					t.Fatalf("authenticated = %v, status = %d, want %d", authenticated, response.StatusCode, want)
				}
			}
		})
	}
}
