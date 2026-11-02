package api

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gorilla/mux"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/auth"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/config"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
)

func TestBasicAuthProtectsAllRoutes(t *testing.T) {
	router := NewRouter(nil, nil, nil, nil, nil, nil, nil, &mockInstanceStatusSubClient{}, nil)
	handler := auth.NewBasicAuth(config.AuthConfiguration{Enabled: true, Users: []config.AuthUser{
		{Username: "first", Password: "one"},
	}})(router)
	variables := regexp.MustCompile(`\{[^}]+\}`)
	checked := 0
	err := router.Walk(func(route *mux.Route, parent *mux.Router, ancestors []*mux.Route) error {
		methods, err := route.GetMethods()
		if err != nil {
			return nil
		}
		template, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		path := variables.ReplaceAllString(template, "sample")
		for _, method := range methods {
			if method == http.MethodGet && path == "/v1/healthcheck" {
				continue
			}
			checked++
			t.Run(method+" "+template, func(t *testing.T) {
				response := httptest.NewRecorder()
				handler.ServeHTTP(response, httptest.NewRequest(method, path, nil))
				if response.Code != http.StatusUnauthorized || response.Header().Get("WWW-Authenticate") == "" {
					t.Fatalf("unprotected route: status %d", response.Code)
				}
			})
		}
		return nil
	})
	if err != nil || checked == 0 {
		t.Fatalf("route walk failed: checked %d routes, error %v", checked, err)
	}
	t.Logf("checked %d protected route/method registrations", checked)
	for _, path := range []string{"/unknown", "/v1//healthcheck", "/v1/../v1/healthcheck", "/v1/healthcheck/", "/v1/%68ealthcheck"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", path, nil))
		if response.Code != http.StatusUnauthorized {
			t.Errorf("path %q bypassed auth: %d", path, response.Code)
		}
	}
	for _, method := range []string{"HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(method, "/v1/healthcheck", nil))
		if response.Code != http.StatusUnauthorized {
			t.Errorf("health method %s bypassed auth: %d", method, response.Code)
		}
	}
	previousDB, previousEtcd := db.DBconn, db.Etcd
	t.Cleanup(func() {
		db.DBconn, db.Etcd = previousDB, previousEtcd
	})
	db.DBconn = &db.MockDB{}
	db.Etcd = &db.MockEtcdClient{}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/v1/healthcheck", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("anonymous Robot health failed: %d", response.Code)
	}
}
