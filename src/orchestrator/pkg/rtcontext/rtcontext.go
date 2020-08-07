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

package rtcontext

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/contextdb"
	pkgerrors "github.com/pkg/errors"
)

const maxrand = 0x7fffffffffffffff
const prefix string = "/context/"

type RunTimeContext struct {
	cid  interface{}
	meta interface{}
}

type Rtcontext interface {
	RtcInit() (interface{}, error)
	RtcLoad(interface{}) (interface{}, error)
	RtcCreate() (interface{}, error)
	RtcAddMeta(meta interface{}) error
	RtcGet() (interface{}, error)
	RtcAddLevel(handle interface{}, level string, value string) (interface{}, error)
	RtcAddResource(handle interface{}, resname string, value interface{}) (interface{}, error)
	RtcAddInstruction(handle interface{}, level string, insttype string, value interface{}) (interface{}, error)
	RtcDeletePair(handle interface{}) error
	RtcDeletePrefix(handle interface{}) error
	RtcGetHandles(handle interface{}) ([]interface{}, error)
	RtcGetValue(handle interface{}, value interface{}) error
	RtcUpdateValue(handle interface{}, value interface{}) error
	RtcGetMeta() (interface{}, error)
	RtcAddOneLevel(pl interface{}, level string, value interface{}) (interface{}, error)
}

//Intialize context by assiging a new id
func (rtc *RunTimeContext) RtcInit() (interface{}, error) {
	if rtc.cid != nil {
		return nil, pkgerrors.Errorf("Error, context already initialized")
	}
	ra := rand.New(rand.NewSource(time.Now().UnixNano()))
	rn := ra.Int63n(maxrand)
	id := fmt.Sprintf("%v", rn)
	cid := (prefix + id + "/")
	rtc.cid = interface{}(cid)
	return interface{}(id), nil

}

//Load context using the given id
func (rtc *RunTimeContext) RtcLoad(id interface{}) (interface{}, error) {
	str := fmt.Sprintf("%v", id)
	if str == "" {
		return nil, pkgerrors.Errorf("Not a valid context id")
	}
	cid := (prefix + str + "/")
	rtc.cid = interface{}(cid)
	handle, err := rtc.RtcGet()
	if err != nil {
		return nil, pkgerrors.Errorf("Error finding the context id: %s", err.Error())
	}
	return handle, nil
}

func (rtc *RunTimeContext) RtcCreate() (interface{}, error) {
	cid := fmt.Sprintf("%v", rtc.cid)
	if cid == "" {
		return nil, pkgerrors.Errorf("Error, context not intialized")
	}
	if !strings.HasPrefix(cid, prefix) {
		return nil, pkgerrors.Errorf("Not a valid run time context prefix")
	}
	id := strings.SplitN(cid, "/", 4)[2]
	err := contextdb.Db.Put(cid, id)
	if err != nil {
		return nil, pkgerrors.Errorf("Error creating run time context: %s", err.Error())
	}

	return rtc.cid, nil
}

//RtcAddMeta is used for saving meta data of appContext into ETCD.
func (rtc *RunTimeContext) RtcAddMeta(meta interface{}) error {
	cid := fmt.Sprintf("%v", rtc.cid)
	if cid == "" {
		return pkgerrors.Errorf("Error, context not intialized")
	}
	if !strings.HasPrefix(cid, prefix) {
		return pkgerrors.Errorf("Not a valid run time context prefix")
	}

	rtc.meta = meta
	k := cid + "meta" + "/"
	err := contextdb.Db.Put(k, rtc.meta)
	if err != nil {
		return pkgerrors.Errorf("Error saving metadata in run time context: %s", err.Error())
	}

	return nil
}

//Get the root handle
func (rtc *RunTimeContext) RtcGet() (interface{}, error) {
	str := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, prefix) {
		return nil, pkgerrors.Errorf("Not a valid run time context")
	}

	var value string
	err := contextdb.Db.Get(str, &value)
	if err != nil {
		return nil, pkgerrors.Errorf("Error getting run time context metadata: %s", err.Error())
	}
	if !strings.Contains(str, value) {
		return nil, pkgerrors.Errorf("Error matching run time context metadata")
	}

	return rtc.cid, nil
}

// RtcGetMeta method fetches the meta data of the rtc object and returns it.
func (rtc *RunTimeContext) RtcGetMeta() (interface{}, error) {
	str := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, prefix) {
		return nil, pkgerrors.Errorf("Not a valid run time context")
	}

	var value interface{}
	k := str + "meta" + "/"
	err := contextdb.Db.Get(k, &value)
	if err != nil {
		return nil, pkgerrors.Errorf("Error getting run time context metadata: %s", err.Error())
	}
	return value, nil

}

//Add a new level at a given handle and return the new handle
func (rtc *RunTimeContext) RtcAddLevel(handle interface{}, level string, value string) (interface{}, error) {
	str := fmt.Sprintf("%v", handle)
	sid := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, sid) {
		return nil, pkgerrors.Errorf("Not a valid run time context handle")
	}

	if level == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context level")
	}
	if value == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context level value")
	}

	key := str + level + "/" + value + "/"
	err := contextdb.Db.Put(key, value)
	if err != nil {
		return nil, pkgerrors.Errorf("Error adding run time context level: %s", err.Error())
	}

	return (interface{})(key), nil
}

// RtcAddOneLevel adds one more level to the existing context prefix.RtcAddOneLevel. It takes in PreviousContentLevel as inteface, new level to be appended as string and the value to be saved of any type. It returns the updated interface and nil if no error.
//
func (rtc *RunTimeContext) RtcAddOneLevel(pl interface{}, level string, value interface{}) (interface{}, error) {
	str := fmt.Sprintf("%v", pl)
	sid := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, sid) {
		return nil, pkgerrors.Errorf("Not a valid run time context handle")
	}

	if level == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context level")
	}
	if value == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context level value")
	}

	key := str + level + "/"
	err := contextdb.Db.Put(key, value)
	if err != nil {
		return nil, pkgerrors.Errorf("Error adding run time context level: %s", err.Error())
	}
	return (interface{})(key), nil
}

// Add a resource under the given level and return new handle
func (rtc *RunTimeContext) RtcAddResource(handle interface{}, resname string, value interface{}) (interface{}, error) {

	str := fmt.Sprintf("%v", handle)
	sid := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, sid) {
		return nil, pkgerrors.Errorf("Not a valid run time context handle")
	}
	if resname == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context resource name")
	}
	if value == nil {
		return nil, pkgerrors.Errorf("Not a valid run time context resource value")
	}

	k := str + "resource" + "/" + resname + "/"
	err := contextdb.Db.Put(k, value)
	if err != nil {
		return nil, pkgerrors.Errorf("Error adding run time context resource: %s", err.Error())
	}
	return (interface{})(k), nil
}

// Add instruction at a given level and type, return the new handle
func (rtc *RunTimeContext) RtcAddInstruction(handle interface{}, level string, insttype string, value interface{}) (interface{}, error) {
	str := fmt.Sprintf("%v", handle)
	sid := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, sid) {
		return nil, pkgerrors.Errorf("Not a valid run time context handle")
	}

	if level == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context level")
	}
	if insttype == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context instruction type")
	}
	if value == nil {
		return nil, pkgerrors.Errorf("Not a valid run time context instruction value")
	}
	k := str + level + "/" + "instruction" + "/" + insttype + "/"
	err := contextdb.Db.Put(k, fmt.Sprintf("%v", value))
	if err != nil {
		return nil, pkgerrors.Errorf("Error adding run time context instruction: %s", err.Error())
	}

	return (interface{})(k), nil
}

//Delete the key value pair using given handle
func (rtc *RunTimeContext) RtcDeletePair(handle interface{}) error {
	str := fmt.Sprintf("%v", handle)
	sid := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, sid) {
		return pkgerrors.Errorf("Not a valid run time context handle")
	}
	err := contextdb.Db.Delete(str)
	if err != nil {
		return pkgerrors.Errorf("Error deleting run time context pair: %s", err.Error())
	}

	return nil
}

// Delete all handles underneath the given handle
func (rtc *RunTimeContext) RtcDeletePrefix(handle interface{}) error {
	str := fmt.Sprintf("%v", handle)
	sid := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, sid) {
		return pkgerrors.Errorf("Not a valid run time context handle")
	}

	err := contextdb.Db.DeleteAll(str)
	if err != nil {
		return pkgerrors.Errorf("Error deleting run time context with prefix: %s", err.Error())
	}

	return nil
}

// Return the list of handles under the given handle
func (rtc *RunTimeContext) RtcGetHandles(handle interface{}) ([]interface{}, error) {
	str := fmt.Sprintf("%v", handle)
	sid := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, sid) {
		return nil, pkgerrors.Errorf("Not a valid run time context handle")
	}

	s, err := contextdb.Db.GetAllKeys(str)
	if err != nil {
		return nil, pkgerrors.Errorf("Error getting run time context handles: %s", err.Error())
	}
	r := make([]interface{}, len(s))
	for i, v := range s {
		r[i] = v
	}
	return r, nil
}

// Get the value for a given handle
func (rtc *RunTimeContext) RtcGetValue(handle interface{}, value interface{}) error {
	str := fmt.Sprintf("%v", handle)
	sid := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, sid) {
		return pkgerrors.Errorf("Not a valid run time context handle")
	}

	err := contextdb.Db.Get(str, value)
	if err != nil {
		return pkgerrors.Errorf("Error getting run time context value: %s", err.Error())
	}

	return nil
}

// Update the value of a given handle
func (rtc *RunTimeContext) RtcUpdateValue(handle interface{}, value interface{}) error {
	str := fmt.Sprintf("%v", handle)
	sid := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, sid) {
		return pkgerrors.Errorf("Not a valid run time context handle")
	}
	err := contextdb.Db.Put(str, value)
	if err != nil {
		return pkgerrors.Errorf("Error updating run time context value: %s", err.Error())
	}
	return nil

}
