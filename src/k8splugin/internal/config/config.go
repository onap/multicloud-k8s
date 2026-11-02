/*
 * Copyright 2019 Intel Corporation, Inc
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

package config

import (
	"encoding/json"
	"errors"
	log "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"
	"os"
	"reflect"
	"strings"
)

type AuthUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthConfiguration struct {
	Enabled bool
	Users   []AuthUser
}

func LoadAuthConfiguration() (AuthConfiguration, error) {
	var configuration AuthConfiguration
	if value, present := os.LookupEnv("MULTICLOUD_AUTH_ENABLED"); present {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "true":
			configuration.Enabled = true
		case "false":
			configuration.Enabled = false
		default:
			return AuthConfiguration{}, errors.New("MULTICLOUD_AUTH_ENABLED must be true or false")
		}
	}
	if !configuration.Enabled {
		return configuration, nil
	}
	path := os.Getenv("MULTICLOUD_AUTH_USERS_FILE")
	if strings.TrimSpace(path) == "" {
		return AuthConfiguration{}, errors.New("MULTICLOUD_AUTH_USERS_FILE is required when authentication is enabled")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return AuthConfiguration{}, errors.New("cannot read MULTICLOUD_AUTH_USERS_FILE")
	}
	if err := json.Unmarshal(contents, &configuration.Users); err != nil {
		return AuthConfiguration{}, errors.New("MULTICLOUD_AUTH_USERS_FILE must contain a JSON array of username/password objects")
	}
	if len(configuration.Users) == 0 {
		return AuthConfiguration{}, errors.New("MULTICLOUD_AUTH_USERS_FILE must contain at least one user")
	}
	names := make(map[string]bool, len(configuration.Users))
	for _, user := range configuration.Users {
		if strings.TrimSpace(user.Username) == "" || strings.TrimSpace(user.Password) == "" {
			return AuthConfiguration{}, errors.New("authentication usernames and passwords must not be blank")
		}
		if strings.Contains(user.Username, ":") {
			return AuthConfiguration{}, errors.New("authentication usernames must not contain a colon")
		}
		if names[user.Username] {
			return AuthConfiguration{}, errors.New("authentication usernames must be unique")
		}
		names[user.Username] = true
	}
	return configuration, nil
}

// Configuration loads up all the values that are used to configure
// backend implementations
type Configuration struct {
	CAFile              string `json:"ca-file"`
	ServerCert          string `json:"server-cert"`
	ServerKey           string `json:"server-key"`
	Password            string `json:"password"`
	DatabaseAddress     string `json:"database-address"`
	DatabaseType        string `json:"database-type"`
	PluginDir           string `json:"plugin-dir"`
	EtcdIP              string `json:"etcd-ip"`
	EtcdCert            string `json:"etcd-cert"`
	EtcdKey             string `json:"etcd-key"`
	EtcdCAFile          string `json:"etcd-ca-file"`
	ServicePort         string `json:"service-port"`
	KubernetesLabelName string `json:"kubernetes-label-name"`
}

// Config is the structure that stores the configuration
var gConfig *Configuration

// readConfigFile reads the specified smsConfig file to setup some env variables
func readConfigFile(file string) (*Configuration, error) {
	f, err := os.Open(file)
	if err != nil {
		log.Warn("Error loading config file. Using defaults.", log.Fields{"error": err})
		return defaultConfiguration(), err
	}
	defer f.Close()

	// Setup some defaults here
	// If the json file has values in it, the defaults will be overwritten
	conf := defaultConfiguration()

	// Read the configuration from json file
	decoder := json.NewDecoder(f)
	err = decoder.Decode(conf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}

func defaultConfiguration() *Configuration {
	cwd, err := os.Getwd()
	if err != nil {
		log.Warn("Error getting cwd. Using .", log.Fields{"error": err})
		cwd = "."
	}

	return &Configuration{
		CAFile:              "ca.cert",
		ServerCert:          "server.cert",
		ServerKey:           "server.key",
		Password:            "",
		DatabaseAddress:     "127.0.0.1",
		DatabaseType:        "mongo",
		PluginDir:           cwd,
		EtcdIP:              "127.0.0.1",
		EtcdCert:            "",
		EtcdKey:             "",
		EtcdCAFile:          "",
		ServicePort:         "9015",
		KubernetesLabelName: "k8splugin.io/rb-instance-id",
	}
}

// GetConfiguration returns the configuration for the app.
// It will try to load it if it is not already loaded.
func GetConfiguration() *Configuration {
	if gConfig == nil {
		conf, err := readConfigFile("k8sconfig.json")
		if err != nil {
			log.Error("Error loading config file", log.Fields{"error": err})
		}
		gConfig = conf
	}

	return gConfig
}

// SetConfigValue sets a value in the configuration
// This is mostly used to customize the application and
// should be used carefully.
func SetConfigValue(key string, value string) *Configuration {
	c := GetConfiguration()
	if value == "" || key == "" {
		return c
	}

	v := reflect.ValueOf(c).Elem()
	if v.Kind() == reflect.Struct {
		f := v.FieldByName(key)
		if f.IsValid() {
			if f.CanSet() {
				if f.Kind() == reflect.String {
					f.SetString(value)
				}
			}
		}
	}
	return c
}
