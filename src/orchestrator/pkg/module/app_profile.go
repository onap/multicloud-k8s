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
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

// AppProfile contains the parameters needed for AppProfiles
// It implements the interface for managing the AppProfiles
type AppProfile struct {
	Metadata AppProfileMetadata `json:"metadata"`
	Spec     AppProfileSpec     `json:"spec"`
}

type AppProfileContent struct {
	Profile string `json:"profile"`
}

// AppProfileMetadata contains the metadata for AppProfiles
type AppProfileMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

// AppProfileSpec contains the Spec for AppProfiles
type AppProfileSpec struct {
	AppName string `json:"app-name"`
}

// AppProfileKey is the key structure that is used in the database
type AppProfileKey struct {
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeapp"`
	CompositeAppVersion string `json:"compositeappversion"`
	CompositeProfile    string `json:"compositeprofile"`
	Profile             string `json:"profile"`
}

type AppProfileQueryKey struct {
	AppName string `json:"app-name"`
}

type AppProfileFindByAppKey struct {
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeapp"`
	CompositeAppVersion string `json:"compositeappversion"`
	CompositeProfile    string `json:"compositeprofile"`
	AppName             string `json:"app-name"`
}

// AppProfileManager exposes the AppProfile functionality
type AppProfileManager interface {
	CreateAppProfile(provider, compositeApp, compositeAppVersion, compositeProfile string, ap AppProfile, ac AppProfileContent) (AppProfile, error)
	GetAppProfile(project, compositeApp, compositeAppVersion, compositeProfile, profile string) (AppProfile, error)
	GetAppProfiles(project, compositeApp, compositeAppVersion, compositeProfile string) ([]AppProfile, error)
	GetAppProfileByApp(project, compositeApp, compositeAppVersion, compositeProfile, appName string) (AppProfile, error)
	GetAppProfileContent(project, compositeApp, compositeAppVersion, compositeProfile, profile string) (AppProfileContent, error)
	GetAppProfileContentByApp(project, compositeApp, compositeAppVersion, compositeProfile, appName string) (AppProfileContent, error)
	DeleteAppProfile(project, compositeApp, compositeAppVersion, compositeProfile, profile string) error
}

// AppProfileClient implements the Manager
// It will also be used to maintain some localized state
type AppProfileClient struct {
	storeName  string
	tagMeta    string
	tagContent string
}

// NewAppProfileClient returns an instance of the AppProfileClient
// which implements the Manager
func NewAppProfileClient() *AppProfileClient {
	return &AppProfileClient{
		storeName:  "orchestrator",
		tagMeta:    "profilemetadata",
		tagContent: "profilecontent",
	}
}

// CreateAppProfile creates an entry for AppProfile in the database.
func (c *AppProfileClient) CreateAppProfile(project, compositeApp, compositeAppVersion, compositeProfile string, ap AppProfile, ac AppProfileContent) (AppProfile, error) {
	key := AppProfileKey{
		Project:             project,
		CompositeApp:        compositeApp,
		CompositeAppVersion: compositeAppVersion,
		CompositeProfile:    compositeProfile,
		Profile:             ap.Metadata.Name,
	}
	qkey := AppProfileQueryKey{
		AppName: ap.Spec.AppName,
	}

	res, err := c.GetAppProfile(project, compositeApp, compositeAppVersion, compositeProfile, ap.Metadata.Name)
	if res != (AppProfile{}) {
		return AppProfile{}, pkgerrors.New("AppProfile already exists")
	}

	res, err = c.GetAppProfileByApp(project, compositeApp, compositeAppVersion, compositeProfile, ap.Spec.AppName)
	if res != (AppProfile{}) {
		return AppProfile{}, pkgerrors.New("App already has an AppProfile")
	}

	//Check if composite profile exists (success assumes existance of all higher level 'parent' objects)
	_, err = NewCompositeProfileClient().GetCompositeProfile(compositeProfile, project, compositeApp, compositeAppVersion)
	if err != nil {
		return AppProfile{}, pkgerrors.New("Unable to find the compositeProfile")
	}

	// TODO: (after app api is ready) check that the app Spec.AppName exists as part of the composite app

	err = db.DBconn.Insert(c.storeName, key, qkey, c.tagMeta, ap)
	if err != nil {
		return AppProfile{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}
	err = db.DBconn.Insert(c.storeName, key, qkey, c.tagContent, ac)
	if err != nil {
		return AppProfile{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return ap, nil
}

// GetAppProfile - return specified App Profile
func (c *AppProfileClient) GetAppProfile(project, compositeApp, compositeAppVersion, compositeProfile, profile string) (AppProfile, error) {
	key := AppProfileKey{
		Project:             project,
		CompositeApp:        compositeApp,
		CompositeAppVersion: compositeAppVersion,
		CompositeProfile:    compositeProfile,
		Profile:             profile,
	}

	value, err := db.DBconn.Find(c.storeName, key, c.tagMeta)
	if err != nil {
		return AppProfile{}, pkgerrors.Wrap(err, "Get App Profile error")
	}

	if value != nil {
		ap := AppProfile{}
		err = db.DBconn.Unmarshal(value[0], &ap)
		if err != nil {
			return AppProfile{}, pkgerrors.Wrap(err, "Unmarshalling AppProfile")
		}
		return ap, nil
	}

	return AppProfile{}, pkgerrors.New("Error getting AppProfile")

}

// GetAppProfile - return all App Profiles for given composite profile
func (c *AppProfileClient) GetAppProfiles(project, compositeApp, compositeAppVersion, compositeProfile string) ([]AppProfile, error) {

	key := AppProfileKey{
		Project:             project,
		CompositeApp:        compositeApp,
		CompositeAppVersion: compositeAppVersion,
		CompositeProfile:    compositeProfile,
		Profile:             "",
	}

	var resp []AppProfile
	values, err := db.DBconn.Find(c.storeName, key, c.tagMeta)
	if err != nil {
		return []AppProfile{}, pkgerrors.Wrap(err, "Get AppProfiles")
	}

	for _, value := range values {
		ap := AppProfile{}
		err = db.DBconn.Unmarshal(value, &ap)
		if err != nil {
			return []AppProfile{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		resp = append(resp, ap)
	}

	return resp, nil
}

// GetAppProfileByApp - return all App Profiles for given composite profile
func (c *AppProfileClient) GetAppProfileByApp(project, compositeApp, compositeAppVersion, compositeProfile, appName string) (AppProfile, error) {

	key := AppProfileFindByAppKey{
		Project:             project,
		CompositeApp:        compositeApp,
		CompositeAppVersion: compositeAppVersion,
		CompositeProfile:    compositeProfile,
		AppName:             appName,
	}

	value, err := db.DBconn.Find(c.storeName, key, c.tagMeta)
	if err != nil {
		return AppProfile{}, pkgerrors.Wrap(err, "Get AppProfile by App")
	}

	if value != nil {
		ap := AppProfile{}
		err = db.DBconn.Unmarshal(value[0], &ap)
		if err != nil {
			return AppProfile{}, pkgerrors.Wrap(err, "Unmarshalling AppProfile")
		}
		return ap, nil
	}

	return AppProfile{}, pkgerrors.New("Error getting AppProfile by App")
}

func (c *AppProfileClient) GetAppProfileContent(project, compositeApp, compositeAppVersion, compositeProfile, profile string) (AppProfileContent, error) {
	key := AppProfileKey{
		Project:             project,
		CompositeApp:        compositeApp,
		CompositeAppVersion: compositeAppVersion,
		CompositeProfile:    compositeProfile,
		Profile:             profile,
	}

	value, err := db.DBconn.Find(c.storeName, key, c.tagContent)
	if err != nil {
		return AppProfileContent{}, pkgerrors.Wrap(err, "Get Cluster Content")
	}

	//value is a byte array
	if value != nil {
		ac := AppProfileContent{}
		err = db.DBconn.Unmarshal(value[0], &ac)
		if err != nil {
			return AppProfileContent{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return ac, nil
	}

	return AppProfileContent{}, pkgerrors.New("Error getting App Profile Content")
}

func (c *AppProfileClient) GetAppProfileContentByApp(project, compositeApp, compositeAppVersion, compositeProfile, appName string) (AppProfileContent, error) {
	key := AppProfileFindByAppKey{
		Project:             project,
		CompositeApp:        compositeApp,
		CompositeAppVersion: compositeAppVersion,
		CompositeProfile:    compositeProfile,
		AppName:             appName,
	}

	value, err := db.DBconn.Find(c.storeName, key, c.tagContent)
	if err != nil {
		return AppProfileContent{}, pkgerrors.Wrap(err, "Get Cluster Content")
	}

	//value is a byte array
	if value != nil {
		ac := AppProfileContent{}
		err = db.DBconn.Unmarshal(value[0], &ac)
		if err != nil {
			return AppProfileContent{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return ac, nil
	}

	return AppProfileContent{}, pkgerrors.New("Error getting App Profile Content")
}

// Delete AppProfile from the database
func (c *AppProfileClient) DeleteAppProfile(project, compositeApp, compositeAppVersion, compositeProfile, profile string) error {
	key := AppProfileKey{
		Project:             project,
		CompositeApp:        compositeApp,
		CompositeAppVersion: compositeAppVersion,
		CompositeProfile:    compositeProfile,
		Profile:             profile,
	}

	err := db.DBconn.Remove(c.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete AppProfile entry;")
	}
	return nil
}
