package module

import (
	"fmt"
	"testing"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
)

type mockValues struct {
	mock.Mock
}

func (m *mockValues) DBInsert(name string, key db.Key, query interface{}, meta string, c interface{}) error {
	fmt.Println("Mocked Insert operation in Mongo")
	args := m.Called(name, key, nil, meta, c)

	return args.Error(0)
}

func (m *mockValues) DBFind(name string, key db.Key, meta string) ([][]byte, error) {
	fmt.Println("Mocked Mongo DB Find Operation")
	args := m.Called(name, key, meta)

	return args.Get(0).([][]byte), args.Error(1)
}

func (m *mockValues) DBUnmarshal(value []byte, out interface{}) error {
	fmt.Println("Mocked Mongo DB Unmarshal Operation")
	args := m.Called(value)

	return args.Error(0)
}

func (m *mockValues) DBRemove(name string, key db.Key) error {
	fmt.Println("Mocked Mongo DB Remove operation")
	args := m.Called(name, key)

	return args.Error(0)
}

func (m *mockValues) CheckProject(project string) error {
	fmt.Println("Mocked Check Project exists")
	args := m.Called(project)

	return args.Error(0)
}

func (m *mockValues) CheckLogicalCloud(project, logicalCloud string) error {
	fmt.Println("Mocked Check Logical Cloud exists")
	args := m.Called(project, logicalCloud)

	return args.Error(0)
}

func TestCreateLogicalCloud(t *testing.T) {

	mData := MetaDataList{
		LogicalCloudName: "test_asdf",
	}

	lc := LogicalCloud{
		MetaData: mData,
	}
	data1 := [][]byte{}

	key := LogicalCloudKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
	}
	myMocks := new(mockValues)
	// just to get an error value
	err1 := errors.New("math: square root of negative number")

	myMocks.On("CheckProject", "test_project").Return(nil)
	myMocks.On("DBInsert", "test_dcm", key, nil, "test_meta", lc).Return(nil)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, err1)

	lcClient := LogicalCloudClient{"test_dcm", "test_meta", "test_context", myMocks}
	_, err := lcClient.Create("test_project", lc)
	if err != nil {
		t.Errorf("Some error occured!")
	}
}

func TestGetLogicalCloud(t *testing.T) {
	key := LogicalCloudKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
	}

	data1 := [][]byte{
		[]byte("abc"),
	}

	data2 := []byte("abc")

	myMocks := new(mockValues)

	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)
	myMocks.On("DBUnmarshal", data2).Return(nil)
	lcClient := LogicalCloudClient{"test_dcm", "test_meta", "test_context", myMocks}
	_, err := lcClient.Get("test_project", "test_asdf")
	if err != nil {
		t.Errorf("Some error occured!")
	}
}

func TestDeleteLogicalCloud(t *testing.T) {

	key := LogicalCloudKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
	}

	myMocks := new(mockValues)

	myMocks.On("DBRemove", "test_dcm", key).Return(nil)

	lcClient := LogicalCloudClient{"test_dcm", "test_meta", "test_context", myMocks}
	err := lcClient.Delete("test_project", "test_asdf")
	if err != nil {
		t.Errorf("Some error occured!")
	}

}

func TestUpdateLogicalCloud(t *testing.T) {
	key := LogicalCloudKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
	}
	mData := MetaDataList{
		LogicalCloudName: "test_asdf",
	}
	lc := LogicalCloud{
		MetaData: mData,
	}
	data1 := [][]byte{
		[]byte("abc"),
	}
	data2 := []byte("abc")

	myMocks := new(mockValues)

	myMocks.On("DBInsert", "test_dcm", key, nil, "test_meta", lc).Return(nil)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)
	myMocks.On("DBUnmarshal", data2).Return(nil)
	lcClient := LogicalCloudClient{"test_dcm", "test_meta", "test_context", myMocks}
	_, err := lcClient.Update("test_project", "test_asdf", lc)
	if err != nil {
		t.Errorf("Some error occured!")
	}
}
