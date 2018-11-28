/*
Copyright 2018 Intel Corporation.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"k8splugin/rb"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"

	"k8splugin/db"
	"k8splugin/krd"
)

// CheckEnvVariables checks for required Environment variables
func CheckEnvVariables() error {
	envList := []string{"CSAR_DIR", "KUBE_CONFIG_DIR", "PLUGINS_DIR",
		"DATABASE_TYPE", "DATABASE_IP", "OVN_CENTRAL_ADDRESS"}
	for _, env := range envList {
		if _, ok := os.LookupEnv(env); !ok {
			return pkgerrors.New("environment variable " + env + " not set")
		}
	}

	return nil
}

// CheckDatabaseConnection checks if the database is up and running and
// plugin can talk to it
func CheckDatabaseConnection() error {
	err := db.CreateDBClient(os.Getenv("DATABASE_TYPE"))
	if err != nil {
		return pkgerrors.Cause(err)
	}

	err = db.DBconn.HealthCheck()
	if err != nil {
		return pkgerrors.Cause(err)
	}
	return nil
}

// LoadPlugins loads all the compiled .so plugins
func LoadPlugins() error {
	pluginsDir := os.Getenv("PLUGINS_DIR")
	err := filepath.Walk(pluginsDir,
		func(path string, info os.FileInfo, err error) error {
			if strings.Contains(path, ".so") {
				p, err := plugin.Open(path)
				if err != nil {
					return pkgerrors.Cause(err)
				}
				krd.LoadedPlugins[info.Name()[:len(info.Name())-3]] = p
			}
			return err
		})
	if err != nil {
		return err
	}

	return nil
}

// CheckInitialSettings is used to check initial settings required to start api
func CheckInitialSettings() error {
	err := CheckEnvVariables()
	if err != nil {
		return pkgerrors.Cause(err)
	}

	err = CheckDatabaseConnection()
	if err != nil {
		return pkgerrors.Cause(err)
	}

	err = LoadPlugins()
	if err != nil {
		return pkgerrors.Cause(err)
	}

	return nil
}

// NewRouter creates a router instance that serves the VNFInstance web methods
func NewRouter(kubeconfig string) *mux.Router {
	router := mux.NewRouter()

	vnfInstanceHandler := router.PathPrefix("/v1/vnf_instances").Subrouter()
	vnfInstanceHandler.HandleFunc("/", CreateHandler).Methods("POST").Name("VNFCreation")
	vnfInstanceHandler.HandleFunc("/{cloudRegionID}/{namespace}", ListHandler).Methods("GET")
	vnfInstanceHandler.HandleFunc("/{cloudRegionID}/{namespace}/{externalVNFID}", DeleteHandler).Methods("DELETE")
	vnfInstanceHandler.HandleFunc("/{cloudRegionID}/{namespace}/{externalVNFID}", GetHandler).Methods("GET")

	//rbd is resource bundle definition
	resRouter := router.PathPrefix("/v1/rb").Subrouter()
	rbdef := rbDefinitionHandler{client: rb.NewDefinitionClient()}
	resRouter.HandleFunc("/definition", rbdef.createHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbdID}/content", rbdef.uploadHandler).Methods("POST")
	resRouter.HandleFunc("/definition", rbdef.listHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbdID}", rbdef.getHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbdID}", rbdef.deleteHandler).Methods("DELETE")

	//rbp is resource bundle profile
	rbprofile := rbProfileHandler{client: rb.NewProfileClient()}
	resRouter.HandleFunc("/profile", rbprofile.createHandler).Methods("POST")
	resRouter.HandleFunc("/profile/{rbpID}/content", rbprofile.uploadHandler).Methods("POST")
	resRouter.HandleFunc("/profile/help", rbprofile.helpHandler).Methods("GET")
	resRouter.HandleFunc("/profile", rbprofile.listHandler).Methods("GET")
	resRouter.HandleFunc("/profile/{rbpID}", rbprofile.getHandler).Methods("GET")
	resRouter.HandleFunc("/profile/{rbpID}", rbprofile.deleteHandler).Methods("DELETE")

	// (TODO): Fix update method
	// vnfInstanceHandler.HandleFunc("/{vnfInstanceId}", UpdateHandler).Methods("PUT")

	return router
}
