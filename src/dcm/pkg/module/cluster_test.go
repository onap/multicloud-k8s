package module

import (
	"testing"

	"github.com/pkg/errors"
)

func TestCreateCluster(t *testing.T) {

	mData := ClusterMeta{
		ClusterReference: "test_cluster",
	}

	cl := Cluster{
		MetaData: mData,
	}
	data1 := [][]byte{}

	key := ClusterKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		ClusterReference: "test_cluster",
	}
	myMocks := new(mockValues)
	// just to get an error value
	err1 := errors.New("math: square root of negative number")

	myMocks.On("CheckProject", "test_project").Return(nil)
	myMocks.On("CheckLogicalCloud", "test_project", "test_asdf").Return(nil)
	myMocks.On("DBInsert", "test_dcm", key, nil, "test_meta", cl).Return(nil)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, err1)

	clClient := ClusterClient{"test_dcm", "test_meta", myMocks}
	_, err := clClient.CreateCluster("test_project", "test_asdf", cl)
	if err != nil {
		t.Errorf("Some error occured!")
	}
}

func TestGetCluster(t *testing.T) {
	key := ClusterKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		ClusterReference: "test_cluster",
	}

	data1 := [][]byte{
		[]byte("abc"),
	}

	data2 := []byte("abc")

	myMocks := new(mockValues)

	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)
	myMocks.On("DBUnmarshal", data2).Return(nil)
	clClient := ClusterClient{"test_dcm", "test_meta", myMocks}
	_, err := clClient.GetCluster("test_project", "test_asdf", "test_cluster")
	if err != nil {
		t.Errorf("Some error occured!")
	}
}

func TestDeleteCluster(t *testing.T) {

	key := ClusterKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		ClusterReference: "test_cluster",
	}

	myMocks := new(mockValues)

	myMocks.On("DBRemove", "test_dcm", key).Return(nil)

	clClient := ClusterClient{"test_dcm", "test_meta", myMocks}
	err := clClient.DeleteCluster("test_project", "test_asdf", "test_cluster")
	if err != nil {
		t.Errorf("Some error occured!")
	}

}

func TestUpdateCluster(t *testing.T) {
	key := ClusterKey{
		Project:          "test_project",
		LogicalCloudName: "test_asdf",
		ClusterReference: "test_cluster",
	}
	mData := ClusterMeta{
		ClusterReference: "test_cluster",
	}
	cl := Cluster{
		MetaData: mData,
	}
	data1 := [][]byte{
		[]byte("abc"),
	}
	data2 := []byte("abc")

	myMocks := new(mockValues)

	myMocks.On("DBInsert", "test_dcm", key, nil, "test_meta", cl).Return(nil)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)
	myMocks.On("DBUnmarshal", data2).Return(nil)
	clClient := ClusterClient{"test_dcm", "test_meta", myMocks}
	_, err := clClient.UpdateCluster("test_project", "test_asdf", "test_cluster", cl)
	if err != nil {
		t.Errorf("Some error occured!")
	}
}
