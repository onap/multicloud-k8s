/*
 * Copyright 2018 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"

	pkgerrors "github.com/pkg/errors"
)

// healthCheckHandler executes a db read to return health of k8splugin
// and its backing database
func TestHealthCheckHandler(t *testing.T) {

	t.Run("OK HealthCheck", func(t *testing.T) {
		db.DBconn = &db.MockDB{
			Err: nil,
		}
		request := httptest.NewRequest("GET", "/v1/healthcheck", nil)
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil))

		//Check returned code
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected %d; Got: %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("FAILED HealthCheck", func(t *testing.T) {
		db.DBconn = &db.MockDB{
			Err: pkgerrors.New("Runtime Error in DB"),
		}
		request := httptest.NewRequest("GET", "/v1/healthcheck", nil)
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil))

		//Check returned code
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("Expected %d; Got: %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})
}
