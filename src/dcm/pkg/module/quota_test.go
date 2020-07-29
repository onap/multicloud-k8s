package module

import (
	"testing"

	"github.com/pkg/errors"
)

func TestCreateQuota(t *testing.T) {

	mData := QMetaDataList{
		QuotaName: "test_quota",
	}

	q := Quota{
		MetaData: mData,
	}
	data1 := [][]byte{}

	key := QuotaKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		QuotaName:        "test_quota",
	}
	myMocks := new(mockValues)
	// just to get an error value
	err1 := errors.New("math: square root of negative number")

	myMocks.On("CheckProject", "test_project").Return(nil)
	myMocks.On("CheckLogicalCloud", "test_project", "test_asdf").Return(nil)
	myMocks.On("DBInsert", "test_dcm", key, nil, "test_meta", q).Return(nil)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, err1)

	qClient := QuotaClient{"test_dcm", "test_meta", myMocks}
	_, err := qClient.CreateQuota("test_project", "test_asdf", q)
	if err != nil {
		t.Errorf("Some error occured!")
	}
}

func TestGetQuota(t *testing.T) {
	key := QuotaKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		QuotaName:        "test_quota",
	}

	data1 := [][]byte{
		[]byte("abc"),
	}

	data2 := []byte("abc")

	myMocks := new(mockValues)

	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)
	myMocks.On("DBUnmarshal", data2).Return(nil)
	qClient := QuotaClient{"test_dcm", "test_meta", myMocks}
	_, err := qClient.GetQuota("test_project", "test_asdf", "test_quota")
	if err != nil {
		t.Errorf("Some error occured!")
	}
}

func TestDeleteQuota(t *testing.T) {

	key := QuotaKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		QuotaName:        "test_quota",
	}

	myMocks := new(mockValues)

	myMocks.On("DBRemove", "test_dcm", key).Return(nil)

	qClient := QuotaClient{"test_dcm", "test_meta", myMocks}
	err := qClient.DeleteQuota("test_project", "test_asdf", "test_quota")
	if err != nil {
		t.Errorf("Some error occured!")
	}

}

func TestUpdateQuota(t *testing.T) {
	key := QuotaKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		QuotaName:        "test_quota",
	}
	mData := QMetaDataList{
		QuotaName: "test_quota",
	}
	q := Quota{
		MetaData: mData,
	}
	data1 := [][]byte{
		[]byte("abc"),
	}
	data2 := []byte("abc")

	myMocks := new(mockValues)

	myMocks.On("DBInsert", "test_dcm", key, nil, "test_meta", q).Return(nil)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)
	myMocks.On("DBUnmarshal", data2).Return(nil)
	qClient := QuotaClient{"test_dcm", "test_meta", myMocks}
	_, err := qClient.UpdateQuota("test_project", "test_asdf", "test_quota", q)
	if err != nil {
		t.Errorf("Some error occured!")
	}
}
