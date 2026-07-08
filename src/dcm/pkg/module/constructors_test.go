/*
* Copyright 2020 Intel Corporation, Inc
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

package module

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogicalCloudClient(t *testing.T) {
	c := NewLogicalCloudClient()
	assert.NotNil(t, c)
	assert.Equal(t, "orchestrator", c.storeName)
	assert.Equal(t, "logicalcloud", c.tagMeta)
	assert.Equal(t, "lccontext", c.tagContext)
	assert.NotNil(t, c.util)
}

func TestNewClusterClient(t *testing.T) {
	c := NewClusterClient()
	assert.NotNil(t, c)
	assert.Equal(t, "orchestrator", c.storeName)
	assert.Equal(t, "cluster", c.tagMeta)
	assert.NotNil(t, c.util)
}

func TestNewUserPermissionClient(t *testing.T) {
	c := NewUserPermissionClient()
	assert.NotNil(t, c)
	assert.Equal(t, "orchestrator", c.storeName)
	assert.Equal(t, "userpermission", c.tagMeta)
	assert.NotNil(t, c.util)
}

func TestNewQuotaClient(t *testing.T) {
	c := NewQuotaClient()
	assert.NotNil(t, c)
	assert.Equal(t, "orchestrator", c.storeName)
	assert.Equal(t, "quota", c.tagMeta)
	assert.NotNil(t, c.util)
}

func TestNewKeyValueClient(t *testing.T) {
	c := NewKeyValueClient()
	assert.NotNil(t, c)
	assert.Equal(t, "orchestrator", c.storeName)
	assert.Equal(t, "keyvalue", c.tagMeta)
	assert.NotNil(t, c.util)
}

func TestNewClient(t *testing.T) {
	c := NewClient()
	assert.NotNil(t, c)
	assert.NotNil(t, c.LogicalCloud)
	assert.NotNil(t, c.Cluster)
	assert.NotNil(t, c.Quota)
	assert.NotNil(t, c.UserPermission)
	assert.NotNil(t, c.KeyValue)

	// spot-check that the sub-clients were wired with the expected tags
	assert.Equal(t, "logicalcloud", c.LogicalCloud.tagMeta)
	assert.Equal(t, "cluster", c.Cluster.tagMeta)
	assert.Equal(t, "quota", c.Quota.tagMeta)
	assert.Equal(t, "userpermission", c.UserPermission.tagMeta)
	assert.Equal(t, "keyvalue", c.KeyValue.tagMeta)
}
