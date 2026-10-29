/*
Copyright 2026 Deutsche Telekom.

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

package logutils

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	log "github.com/sirupsen/logrus"
)

// klogAdapter is the logr.Logger that klog (used by client-go) writes to.
// klog does its own verbosity filtering before calling it.
type klogAdapter struct {
	name   string
	fields log.Fields
}

func (k klogAdapter) entry(keysAndValues []interface{}) *log.Entry {
	fields := log.Fields{}
	for key, value := range k.fields {
		fields[key] = value
	}
	for i := 0; i+1 < len(keysAndValues); i += 2 {
		fields[fmt.Sprint(keysAndValues[i])] = keysAndValues[i+1]
	}
	if k.name != "" {
		fields["logger"] = k.name
	}
	return log.WithFields(fields)
}

func (k klogAdapter) Enabled() bool {
	return true
}

func (k klogAdapter) Info(msg string, keysAndValues ...interface{}) {
	k.entry(keysAndValues).Info(strings.TrimRight(msg, "\n"))
}

func (k klogAdapter) Error(err error, msg string, keysAndValues ...interface{}) {
	e := k.entry(keysAndValues)
	if err != nil {
		e = e.WithError(err)
	}
	e.Error(strings.TrimRight(msg, "\n"))
}

func (k klogAdapter) V(level int) logr.Logger {
	return k
}

func (k klogAdapter) WithValues(keysAndValues ...interface{}) logr.Logger {
	return klogAdapter{name: k.name, fields: k.entry(keysAndValues).Data}
}

func (k klogAdapter) WithName(name string) logr.Logger {
	if k.name != "" {
		name = k.name + "." + name
	}
	return klogAdapter{name: name, fields: k.fields}
}
