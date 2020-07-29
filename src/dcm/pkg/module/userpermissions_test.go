package module

import (
	"testing"

	"github.com/pkg/errors"
)

func TestCreateUserPerm(t *testing.T) {

	up := UserPermission{
		UserPermissionName: "test_user_perm",
	}
	data1 := [][]byte{}

	// data2 := []byte("abc")

	key := UserPermissionKey{
		Project:            "test_project",
		LogicalCloudName:   "test_asdf",
		UserPermissionName: "test_user_perm",
	}
	myMocks := new(mockValues)
	// just to get an error value
	err1 := errors.New("math: square root of negative number")

	myMocks.On("CheckProject", "test_project").Return(nil)
	myMocks.On("CheckLogicalCloud", "test_project", "test_asdf").Return(nil)
	myMocks.On("DBInsert", "test_dcm", key, nil, "test_meta", up).Return(nil)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, err1)
	// myMocks.On("DBUnmarshal", data2).Return(nil)

	upClient := UserPermissionClient{"test_dcm", "test_meta", myMocks}
	_, err := upClient.CreateUserPerm("test_project", "test_asdf", up)
	if err != nil {
		t.Errorf("Some error occured!")
	}
}

func TestGetUserPerm(t *testing.T) {
	key := UserPermissionKey{
		Project:            "test_project",
		LogicalCloudName:   "test_asdf",
		UserPermissionName: "test_user_perm",
	}

	data1 := [][]byte{
		[]byte("abc"),
	}

	data2 := []byte("abc")

	myMocks := new(mockValues)

	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)
	myMocks.On("DBUnmarshal", data2).Return(nil)
	upClient := UserPermissionClient{"test_dcm", "test_meta", myMocks}
	_, err := upClient.GetUserPerm("test_project", "test_asdf", "test_user_perm")
	if err != nil {
		t.Errorf("Some error occured!")
	}
}

func TestDeleteUserPerm(t *testing.T) {

	key := UserPermissionKey{
		Project:            "test_project",
		LogicalCloudName:   "test_asdf",
		UserPermissionName: "test_user_perm",
	}

	myMocks := new(mockValues)

	myMocks.On("DBRemove", "test_dcm", key).Return(nil)

	upClient := UserPermissionClient{"test_dcm", "test_meta", myMocks}
	err := upClient.DeleteUserPerm("test_project", "test_asdf", "test_user_perm")
	if err != nil {
		t.Errorf("Some error occured!")
	}

}

func TestUpdateUserPerm(t *testing.T) {
	key := UserPermissionKey{
		Project:            "test_project",
		LogicalCloudName:   "test_asdf",
		UserPermissionName: "test_user_perm",
	}
	up := UserPermission{
		UserPermissionName: "test_user_perm",
	}
	data1 := [][]byte{
		[]byte("abc"),
	}
	data2 := []byte("abc")

	myMocks := new(mockValues)

	myMocks.On("DBInsert", "test_dcm", key, nil, "test_meta", up).Return(nil)
	myMocks.On("DBFind", "test_dcm", key, "test_meta").Return(data1, nil)
	myMocks.On("DBUnmarshal", data2).Return(nil)
	upClient := UserPermissionClient{"test_dcm", "test_meta", myMocks}
	_, err := upClient.UpdateUserPerm("test_project", "test_asdf", "test_user_perm", up)
	if err != nil {
		t.Errorf("Some error occured!")
	}
}
