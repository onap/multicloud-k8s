<!-- Copyright 2018 Intel Corporation.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. -->

# MultiCloud-k8s-plugin

MultiCloud Kubernetes plugin for ONAP multicloud.

# Installation

Requirements:
* Go 1.16

Steps:

* Clone repo:
    * `git clone https://git.onap.org/multicloud/k8s`

* Build everything:
    *  `cd k8s/src/k8splugin && make all`

* Run Tests:
    *  `cd k8s/src/k8splugin && make test`

* Generate Coverage Report:
    * `cd k8s/src/k8splugin && make cover`

* Run the plugin:
    * `cd k8s/deployments && ./start.sh`

# Optional API Authentication

The standalone Kubernetes plugin supports application-level HTTP Basic authentication:

| Environment variable | Default | Meaning |
| --- | --- | --- |
| `MULTICLOUD_AUTH_ENABLED` | `false` | Set to `true` to require API credentials. |
| `MULTICLOUD_AUTH_USERS_FILE` | unset | Path to a mounted JSON credentials file; required when authentication is enabled. |

The file is a JSON array of username/password objects, for example (replace the
placeholder password before use):

```json
[
    {"username": "api-client", "password": "<strong-secret>"}
]
```

All listed users are accepted with the same API access; this is authentication,
not role-based authorization. An invalid enable flag, unreadable file, malformed
JSON, empty list, blank username or password, duplicate username, or username
containing `:` prevents startup. Passwords may contain `:`. Nonblank credential
values are used exactly as written, without trimming. Configuration errors do not
include credential values or file contents. When authentication is disabled, the
users file is ignored and existing anonymous access is preserved.

Enforcement covers the whole plugin HTTP router on both HTTP and HTTPS, including
direct and localhost requests, unknown routes, and all API families. Missing,
malformed, or incorrect credentials receive `401` with a `WWW-Authenticate: Basic`
challenge. Only exact `GET /v1/healthcheck` remains anonymous for Robot health
checks; other methods, trailing slashes, and encoded lookalikes require credentials.
Existing TCP probes are unchanged. This setting does not protect other MultiCloud
services or configure callers automatically.

Basic credentials are encoded, not encrypted. Require TLS or mesh mTLS for their
transport; enabling authentication does not enable TLS or change the existing TLS
configuration. Restrict access to the credentials file and do not store real
credentials in source control or logs.

Credentials are loaded once at startup. Restart the plugin after changing the
environment or rotating the file, including externally managed mounted secrets.
For overlap during rotation, temporarily list both old and new users, restart the
plugin, switch callers to the new credentials, then remove the old user and restart
again. Update callers before enabling enforcement.

# Troubleshooting

* Logs can be set as "warn" or "info" in the config.json file of the orchestrator directory. By default the log level is set as "warn".
