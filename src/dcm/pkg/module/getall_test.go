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

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// -----------------------------------------------------------------------------
// LogicalCloud.GetAll
// -----------------------------------------------------------------------------

func TestGetAllLogicalClouds(t *testing.T) {
	key := LogicalCloudKey{
		Project:          "test_project",
		LogicalCloudName: "",
	}

	data1 := [][]byte{
		[]byte("abc"),
		[]byte("def"),
	}

	myMocks := new(mockValues)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)
	myMocks.On("DBUnmarshal", []byte("abc")).Return(nil)
	myMocks.On("DBUnmarshal", []byte("def")).Return(nil)

	lcClient := LogicalCloudClient{"test_dcm", "test_meta", "test_context", myMocks}
	resp, err := lcClient.GetAll("test_project")
	assert.NoError(t, err)
	assert.Len(t, resp, 2)
}

func TestGetAllLogicalCloudsEmpty(t *testing.T) {
	key := LogicalCloudKey{
		Project:          "test_project",
		LogicalCloudName: "",
	}

	data1 := [][]byte{}

	myMocks := new(mockValues)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)

	lcClient := LogicalCloudClient{"test_dcm", "test_meta", "test_context", myMocks}
	resp, err := lcClient.GetAll("test_project")
	assert.NoError(t, err)
	assert.Len(t, resp, 0)
}

func TestGetAllLogicalCloudsDBError(t *testing.T) {
	key := LogicalCloudKey{
		Project:          "test_project",
		LogicalCloudName: "",
	}

	data1 := [][]byte{}
	err1 := errors.New("db find failed")

	myMocks := new(mockValues)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, err1)

	lcClient := LogicalCloudClient{"test_dcm", "test_meta", "test_context", myMocks}
	resp, err := lcClient.GetAll("test_project")
	assert.Error(t, err)
	assert.Len(t, resp, 0)
}

// -----------------------------------------------------------------------------
// Cluster.GetAllClusters
// -----------------------------------------------------------------------------

func TestGetAllClusters(t *testing.T) {
	key := ClusterKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		ClusterReference: "",
	}

	data1 := [][]byte{
		[]byte("abc"),
		[]byte("def"),
	}

	myMocks := new(mockValues)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)
	myMocks.On("DBUnmarshal", []byte("abc")).Return(nil)
	myMocks.On("DBUnmarshal", []byte("def")).Return(nil)

	clClient := ClusterClient{"test_dcm", "test_meta", myMocks}
	resp, err := clClient.GetAllClusters("test_project", "test_asdf")
	assert.NoError(t, err)
	assert.Len(t, resp, 2)
}

func TestGetAllClustersEmpty(t *testing.T) {
	key := ClusterKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		ClusterReference: "",
	}

	// GetAllClusters treats zero values as an error ("No Cluster References associated")
	data1 := [][]byte{}

	myMocks := new(mockValues)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)

	clClient := ClusterClient{"test_dcm", "test_meta", myMocks}
	resp, err := clClient.GetAllClusters("test_project", "test_asdf")
	assert.Error(t, err)
	assert.Len(t, resp, 0)
}

func TestGetAllClustersDBError(t *testing.T) {
	key := ClusterKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		ClusterReference: "",
	}

	data1 := [][]byte{}
	err1 := errors.New("db find failed")

	myMocks := new(mockValues)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, err1)

	clClient := ClusterClient{"test_dcm", "test_meta", myMocks}
	resp, err := clClient.GetAllClusters("test_project", "test_asdf")
	assert.Error(t, err)
	assert.Len(t, resp, 0)
}

// -----------------------------------------------------------------------------
// UserPermission.GetAllUserPerms
// -----------------------------------------------------------------------------

func TestGetAllUserPerms(t *testing.T) {
	key := UserPermissionKey{
		Project:            "test_project",
		LogicalCloudName:   "test_asdf",
		UserPermissionName: "",
	}

	data1 := [][]byte{
		[]byte("abc"),
		[]byte("def"),
	}

	myMocks := new(mockValues)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)
	myMocks.On("DBUnmarshal", []byte("abc")).Return(nil)
	myMocks.On("DBUnmarshal", []byte("def")).Return(nil)

	upClient := UserPermissionClient{"test_dcm", "test_meta", myMocks}
	resp, err := upClient.GetAllUserPerms("test_project", "test_asdf")
	assert.NoError(t, err)
	assert.Len(t, resp, 2)
}

func TestGetAllUserPermsEmpty(t *testing.T) {
	key := UserPermissionKey{
		Project:            "test_project",
		LogicalCloudName:   "test_asdf",
		UserPermissionName: "",
	}

	data1 := [][]byte{}

	myMocks := new(mockValues)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)

	upClient := UserPermissionClient{"test_dcm", "test_meta", myMocks}
	resp, err := upClient.GetAllUserPerms("test_project", "test_asdf")
	assert.NoError(t, err)
	assert.Len(t, resp, 0)
}

func TestGetAllUserPermsDBError(t *testing.T) {
	key := UserPermissionKey{
		Project:            "test_project",
		LogicalCloudName:   "test_asdf",
		UserPermissionName: "",
	}

	data1 := [][]byte{}
	err1 := errors.New("db find failed")

	myMocks := new(mockValues)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, err1)

	upClient := UserPermissionClient{"test_dcm", "test_meta", myMocks}
	resp, err := upClient.GetAllUserPerms("test_project", "test_asdf")
	assert.Error(t, err)
	assert.Len(t, resp, 0)
}

// -----------------------------------------------------------------------------
// Quota.GetAllQuotas
// -----------------------------------------------------------------------------

func TestGetAllQuotas(t *testing.T) {
	key := QuotaKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		QuotaName:        "",
	}

	data1 := [][]byte{
		[]byte("abc"),
		[]byte("def"),
	}

	myMocks := new(mockValues)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)
	myMocks.On("DBUnmarshal", []byte("abc")).Return(nil)
	myMocks.On("DBUnmarshal", []byte("def")).Return(nil)

	qClient := QuotaClient{"test_dcm", "test_meta", myMocks}
	resp, err := qClient.GetAllQuotas("test_project", "test_asdf")
	assert.NoError(t, err)
	assert.Len(t, resp, 2)
}

func TestGetAllQuotasEmpty(t *testing.T) {
	key := QuotaKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		QuotaName:        "",
	}

	data1 := [][]byte{}

	myMocks := new(mockValues)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)

	qClient := QuotaClient{"test_dcm", "test_meta", myMocks}
	resp, err := qClient.GetAllQuotas("test_project", "test_asdf")
	assert.NoError(t, err)
	assert.Len(t, resp, 0)
}

func TestGetAllQuotasDBError(t *testing.T) {
	key := QuotaKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		QuotaName:        "",
	}

	data1 := [][]byte{}
	err1 := errors.New("db find failed")

	myMocks := new(mockValues)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, err1)

	qClient := QuotaClient{"test_dcm", "test_meta", myMocks}
	resp, err := qClient.GetAllQuotas("test_project", "test_asdf")
	assert.Error(t, err)
	assert.Len(t, resp, 0)
}

// -----------------------------------------------------------------------------
// KeyValue.GetAllKVPairs
// -----------------------------------------------------------------------------

func TestGetAllKVPairs(t *testing.T) {
	key := KeyValueKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		KeyValueName:     "",
	}

	data1 := [][]byte{
		[]byte("abc"),
		[]byte("def"),
	}

	myMocks := new(mockValues)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)
	myMocks.On("DBUnmarshal", []byte("abc")).Return(nil)
	myMocks.On("DBUnmarshal", []byte("def")).Return(nil)

	kvClient := KeyValueClient{"test_dcm", "test_meta", myMocks}
	resp, err := kvClient.GetAllKVPairs("test_project", "test_asdf")
	assert.NoError(t, err)
	assert.Len(t, resp, 2)
}

func TestGetAllKVPairsEmpty(t *testing.T) {
	key := KeyValueKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		KeyValueName:     "",
	}

	data1 := [][]byte{}

	myMocks := new(mockValues)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)

	kvClient := KeyValueClient{"test_dcm", "test_meta", myMocks}
	resp, err := kvClient.GetAllKVPairs("test_project", "test_asdf")
	assert.NoError(t, err)
	assert.Len(t, resp, 0)
}

func TestGetAllKVPairsDBError(t *testing.T) {
	key := KeyValueKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		KeyValueName:     "",
	}

	data1 := [][]byte{}
	err1 := errors.New("db find failed")

	myMocks := new(mockValues)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, err1)

	kvClient := KeyValueClient{"test_dcm", "test_meta", myMocks}
	resp, err := kvClient.GetAllKVPairs("test_project", "test_asdf")
	assert.Error(t, err)
	assert.Len(t, resp, 0)
}
