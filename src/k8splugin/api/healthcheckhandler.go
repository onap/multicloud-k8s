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

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
	logr "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"
)

// healthCheckHandler executes a db read to return health of k8splugin
// and its backing database
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	err := db.DBconn.HealthCheck()
	if err != nil {
		logr.WithFields("http.StatusInternalServerError", "Error", "StatusInternalServerError")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
