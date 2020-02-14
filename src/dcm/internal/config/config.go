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
    "log"
    "os"
)

// Configurations loads up all the values that are used to configure
// backend implementations

type Configuration struct {
        CAFile              string `json:"ca-file"`
        ServerCert          string `json:"server-cert"`
        ServerKey           string `json:"server-key"`
        Password            string `json:"password"`
        DatabaseIP          string `json:"database-ip"`
        DatabaseType        string `json:"database-type"`
        ServicePort         string `json:"service-port"`
}

// Config is the structure that stores the configuration
var gConfig *Configuration

// readConfigFile reads the specified the specified smsConfig file to set up
// some env variables
func readConfigFile(file string) (*Configuration, error) {
        f, err := os.Open(file)
        if err != nil {
            return defaultConfiguration(), err
        }
        defer f.Close()

        // Set up some defaults here
        // If the json file has values in it, the defaults will be overwritten
        conf := defaultConfiguration()

        // Read the configuration from json file
        decoder := json.NewDecoder(f)
        decoder.DisallowUnknownFields()
        err = decoder.Decode(conf)
        if err != nil {
                return conf, err
        }

        return conf, nil
}

func defaultConfiguration() *Configuration {
        // Is there need for the working directory as part of the config???
        return &Configuration{
                CAFile:             "ca.cert",
                ServerCert:         "server.cert",
                ServerKey:          "server.key",
                Password:           "",
                DatabaseIP:         "127.0.0.1",
                DatabaseType:       "mongo",
                ServicePort:        "9015",
        }
}


// GetConfiguration returns the configuration for the app.
// It will try to load it if it is not already loaded.
func GetConfiguration() *Configuration {
        if gConfig == nil {
                conf, err := readConfigFile("DCMconfig.json")
                if err != nil {
                        log.Println("Error loading config file. Using defaults")
                }

                gConfig = conf
        }

        return gConfig
}
