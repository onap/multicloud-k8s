package module

import (
	"testing"

	"github.com/pkg/errors"
)

func TestCreateKVPair(t *testing.T) {

	mData := KVMetaDataList{
		KeyValueName: "test_kv_pair",
	}

	kv := KeyValue{
		MetaData: mData,
	}
	data1 := [][]byte{}

	key := KeyValueKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		KeyValueName:     "test_kv_pair",
	}
	myMocks := new(mockValues)
	// just to get an error value
	err1 := errors.New("math: square root of negative number")

	myMocks.On("CheckProject", "test_project").Return(nil)
	myMocks.On("CheckLogicalCloud", "test_project", "test_asdf").Return(nil)
	myMocks.On("DBInsert", "test_dcm", key, nil, "test_meta", kv).Return(nil)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, err1)

	kvClient := KeyValueClient{"test_dcm", "test_meta", myMocks}
	_, err := kvClient.CreateKVPair("test_project", "test_asdf", kv)
	if err != nil {
		t.Errorf("Some error occured!")
	}
}

func TestGetKVPair(t *testing.T) {
	key := KeyValueKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		KeyValueName:     "test_kv_pair",
	}

	data1 := [][]byte{
		[]byte("abc"),
	}

	data2 := []byte("abc")

	myMocks := new(mockValues)

	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)
	myMocks.On("DBUnmarshal", data2).Return(nil)
	kvClient := KeyValueClient{"test_dcm", "test_meta", myMocks}
	_, err := kvClient.GetKVPair("test_project", "test_asdf", "test_kv_pair")
	if err != nil {
		t.Errorf("Some error occured!")
	}
}

func TestDeleteKVPair(t *testing.T) {

	key := KeyValueKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		KeyValueName:     "test_kv_pair",
	}

	myMocks := new(mockValues)

	myMocks.On("DBRemove", "test_dcm", key).Return(nil)

	kvClient := KeyValueClient{"test_dcm", "test_meta", myMocks}
	err := kvClient.DeleteKVPair("test_project", "test_asdf", "test_kv_pair")
	if err != nil {
		t.Errorf("Some error occured!")
	}

}

func TestUpdateKVPair(t *testing.T) {
	key := KeyValueKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		KeyValueName:     "test_kv_pair",
	}
	mData := KVMetaDataList{
		KeyValueName: "test_kv_pair",
	}
	kv := KeyValue{
		MetaData: mData,
	}
	data1 := [][]byte{
		[]byte("abc"),
	}
	data2 := []byte("abc")

	myMocks := new(mockValues)

	myMocks.On("DBInsert", "test_dcm", key, nil, "test_meta", kv).Return(nil)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)
	myMocks.On("DBUnmarshal", data2).Return(nil)
	kvClient := KeyValueClient{"test_dcm", "test_meta", myMocks}
	_, err := kvClient.UpdateKVPair("test_project", "test_asdf", "test_kv_pair", kv)
	if err != nil {
		t.Errorf("Some error occured!")
	}
}
