/*
 * Copyright © 2022 Orange
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

package app

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
	log "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/rb"
	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

// QueryStatus is what is returned when status is queried for an instance
type StatusSubscription struct {
	Name              string                 `json:"name"`
	MinNotifyInterval int32                  `json:"min-notify-interval"`
	LastUpdateTime    time.Time              `json:"last-update-time"`
	CallbackUrl       string                 `json:"callback-url"`
	LastNotifyTime    time.Time              `json:"last-notify-time"`
	LastNotifyStatus  int32                  `json:"last-notify-status"`
	NotifyMetadata    map[string]interface{} `json:"metadata"`
}

type SubscriptionRequest struct {
	Name              string                 `json:"name"`
	MinNotifyInterval int32                  `json:"min-notify-interval"`
	NotifyMetadata    map[string]interface{} `json:"metadata"`
	CallbackUrl       string                 `json:"callback-url"`
}

// StatusSubscriptionKey is used as the primary key in the db
type StatusSubscriptionKey struct {
	InstanceId       string `json:"instanceid"`
	SubscriptionName string `json:"subscriptionname"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (dk StatusSubscriptionKey) String() string {
	out, err := json.Marshal(dk)
	if err != nil {
		return ""
	}

	return string(out)
}

// InstanceStatusSubClient implements InstanceStatusSubManager
type InstanceStatusSubClient struct {
	storeName string
	tagInst   string
}

func NewInstanceStatusSubClient() *InstanceStatusSubClient {
	return &InstanceStatusSubClient{
		storeName: "rbdef",
		tagInst:   "instanceStatusSub",
	}
}

type notifyResult struct {
	result int32
	time   time.Time
}

type resourceStatusDelta struct {
	Created  []ResourceStatus `json:"created"`
	Deleted  []ResourceStatus `json:"deleted"`
	Modified []ResourceStatus `json:"modified"`
}

type notifyRequestPayload struct {
	InstanceId   string                 `json:"instance-id"`
	Subscription string                 `json:"subscription-name"`
	Metadata     map[string]interface{} `json:"metadata"`
	Delta        resourceStatusDelta    `json:"resource-changes"`
}

func (rsd resourceStatusDelta) Delta() bool {
	return len(rsd.Created) > 0 || len(rsd.Deleted) > 0 || len(rsd.Modified) > 0
}

type notifyChannelData struct {
	instanceId   string
	subscription StatusSubscription
	action       string
	delta        resourceStatusDelta
	notifyResult chan notifyResult
}

type subscriptionWatch struct {
	watcherStop    chan struct{}
	lastUpdateTime time.Time
}

type subscriptionWatchManager struct {
	watchersStatus map[string]subscriptionWatch
}

type subscriptionNotifyManager struct {
	notifyLockMap  map[string]*sync.Mutex
	notifyChannel  map[string]chan notifyChannelData
	watchersStatus map[string]subscriptionWatchManager
	sync.Mutex
}

var subscriptionNotifyData = subscriptionNotifyManager{
	notifyLockMap:  map[string]*sync.Mutex{},
	notifyChannel:  map[string]chan notifyChannelData{},
	watchersStatus: map[string]subscriptionWatchManager{},
}

// InstanceStatusSubManager is an interface exposes the status subscription functionality
type InstanceStatusSubManager interface {
	Create(instanceId string, subDetails SubscriptionRequest) (StatusSubscription, error)
	Get(instanceId, subId string) (StatusSubscription, error)
	Update(instanceId, subId string, subDetails SubscriptionRequest) (StatusSubscription, error)
	List(instanceId string) ([]StatusSubscription, error)
	Delete(instanceId, subId string) error
	Cleanup(instanceId string) error
	RestoreWatchers()
}

// Create Status Subscription
func (iss *InstanceStatusSubClient) Create(instanceId string, subDetails SubscriptionRequest) (StatusSubscription, error) {

	_, err := iss.Get(instanceId, subDetails.Name)
	if err == nil {
		return StatusSubscription{}, pkgerrors.New("Subscription already exists")
	}

	lock, _, _ := getSubscriptionData(instanceId)

	key := StatusSubscriptionKey{
		InstanceId:       instanceId,
		SubscriptionName: subDetails.Name,
	}

	sub := StatusSubscription{
		Name:              subDetails.Name,
		MinNotifyInterval: subDetails.MinNotifyInterval,
		LastNotifyStatus:  0,
		CallbackUrl:       subDetails.CallbackUrl,
		LastUpdateTime:    time.Now(),
		LastNotifyTime:    time.Now(),
		NotifyMetadata:    subDetails.NotifyMetadata,
	}
	if sub.NotifyMetadata == nil {
		sub.NotifyMetadata = make(map[string]interface{})
	}

	err = iss.refreshWatchers(instanceId, subDetails.Name)
	if err != nil {
		return sub, pkgerrors.Wrap(err, "Creating Status Subscription DB Entry")
	}

	lock.Lock()
	defer lock.Unlock()

	err = db.DBconn.Create(iss.storeName, key, iss.tagInst, sub)
	if err != nil {
		return sub, pkgerrors.Wrap(err, "Creating Status Subscription DB Entry")
	}
	log.Info("Successfully created Status Subscription", log.Fields{
		"InstanceId":       instanceId,
		"SubscriptionName": subDetails.Name,
	})

	go runNotifyThread(instanceId, sub.Name)

	return sub, nil
}

// Get Status subscription
func (iss *InstanceStatusSubClient) Get(instanceId, subId string) (StatusSubscription, error) {
	lock, _, _ := getSubscriptionData(instanceId)
	// Acquire Mutex
	lock.Lock()
	defer lock.Unlock()
	key := StatusSubscriptionKey{
		InstanceId:       instanceId,
		SubscriptionName: subId,
	}
	DBResp, err := db.DBconn.Read(iss.storeName, key, iss.tagInst)
	if err != nil || DBResp == nil {
		return StatusSubscription{}, pkgerrors.Wrap(err, "Error retrieving Subscription data")
	}

	sub := StatusSubscription{}
	err = db.DBconn.Unmarshal(DBResp, &sub)
	if err != nil {
		return StatusSubscription{}, pkgerrors.Wrap(err, "Demarshalling Subscription Value")
	}
	return sub, nil
}

// Update status subscription
func (iss *InstanceStatusSubClient) Update(instanceId, subId string, subDetails SubscriptionRequest) (StatusSubscription, error) {
	sub, err := iss.Get(instanceId, subDetails.Name)
	if err != nil {
		return StatusSubscription{}, pkgerrors.Wrap(err, "Subscription does not exist")
	}

	lock, _, _ := getSubscriptionData(instanceId)

	key := StatusSubscriptionKey{
		InstanceId:       instanceId,
		SubscriptionName: subDetails.Name,
	}

	sub.MinNotifyInterval = subDetails.MinNotifyInterval
	sub.CallbackUrl = subDetails.CallbackUrl
	sub.NotifyMetadata = subDetails.NotifyMetadata
	if sub.NotifyMetadata == nil {
		sub.NotifyMetadata = make(map[string]interface{})
	}

	err = iss.refreshWatchers(instanceId, subDetails.Name)
	if err != nil {
		return sub, pkgerrors.Wrap(err, "Updating Status Subscription DB Entry")
	}

	lock.Lock()
	defer lock.Unlock()

	err = db.DBconn.Update(iss.storeName, key, iss.tagInst, sub)
	if err != nil {
		return sub, pkgerrors.Wrap(err, "Updating Status Subscription DB Entry")
	}
	log.Info("Successfully updated Status Subscription", log.Fields{
		"InstanceId":       instanceId,
		"SubscriptionName": subDetails.Name,
	})

	return sub, nil
}

// Get list of status subscriptions
func (iss *InstanceStatusSubClient) List(instanceId string) ([]StatusSubscription, error) {

	lock, _, _ := getSubscriptionData(instanceId)
	// Acquire Mutex
	lock.Lock()
	defer lock.Unlock()
	// Retrieve info about created status subscriptions
	dbResp, err := db.DBconn.ReadAll(iss.storeName, iss.tagInst)
	if err != nil {
		if !strings.Contains(err.Error(), "Did not find any objects with tag") {
			return []StatusSubscription{}, pkgerrors.Wrap(err, "Getting Status Subscription data")
		}
	}
	subList := make([]StatusSubscription, 0)
	for key, value := range dbResp {
		if key != "" {
			subKey := StatusSubscriptionKey{}
			err = json.Unmarshal([]byte(key), &subKey)
			if err != nil {
				log.Error("Error demarshaling Status Subscription Key DB data", log.Fields{
					"error": err.Error(),
					"key":   key})
				return []StatusSubscription{}, pkgerrors.Wrap(err, "Demarshalling subscription key")
			}
			if subKey.InstanceId != instanceId {
				continue
			}
		}
		//value is a byte array
		if value != nil {
			sub := StatusSubscription{}
			err = db.DBconn.Unmarshal(value, &sub)
			if err != nil {
				log.Error("Error demarshaling Status Subscription DB data", log.Fields{
					"error": err.Error(),
					"key":   key})
			}
			subList = append(subList, sub)
		}
	}

	return subList, nil
}

// Delete status subscription
func (iss *InstanceStatusSubClient) Delete(instanceId, subId string) error {
	_, err := iss.Get(instanceId, subId)
	if err != nil {
		return pkgerrors.Wrap(err, "Subscription does not exist")
	}
	lock, _, watchers := getSubscriptionData(instanceId)
	// Acquire Mutex
	lock.Lock()
	defer lock.Unlock()

	close(watchers.watchersStatus[subId].watcherStop)
	delete(watchers.watchersStatus, subId)

	key := StatusSubscriptionKey{
		InstanceId:       instanceId,
		SubscriptionName: subId,
	}
	err = db.DBconn.Delete(iss.storeName, key, iss.tagInst)
	if err != nil {
		return pkgerrors.Wrap(err, "Removing Status Subscription in DB")
	}
	return nil
}

// Cleanup status subscriptions for instance
func (iss *InstanceStatusSubClient) Cleanup(instanceId string) error {
	subList, err := iss.List(instanceId)
	if err != nil {
		return err
	}

	for _, sub := range subList {
		err = iss.Delete(instanceId, sub.Name)
		if err != nil {
			log.Error("Error deleting ", log.Fields{
				"error": err.Error(),
				"key":   sub.Name})
		}
	}
	removeSubscriptionData(instanceId)
	return err
}

// Restore status subscriptions notify threads
func (iss *InstanceStatusSubClient) RestoreWatchers() {
	go func() {
		time.Sleep(time.Second * 10)
		log.Info("Restoring status subscription notifications", log.Fields{})
		v := NewInstanceClient()
		instances, err := v.List("", "", "")
		if err != nil {
			log.Error("Error reading instance list", log.Fields{
				"error": err.Error(),
			})
		}
		for _, instance := range instances {
			subList, err := iss.List(instance.ID)
			if err != nil {
				log.Error("Error reading subscription list for instance", log.Fields{
					"error":    err.Error(),
					"instance": instance.ID,
				})
				continue
			}

			for _, sub := range subList {
				err = iss.refreshWatchers(instance.ID, sub.Name)
				if err != nil {
					log.Error("Error on refreshing watchers", log.Fields{
						"error":        err.Error(),
						"instance":     instance.ID,
						"subscription": sub.Name,
					})
					continue
				}
				go runNotifyThread(instance.ID, sub.Name)
			}
		}
	}()
}

func (iss *InstanceStatusSubClient) refreshWatchers(instanceId, subId string) error {
	log.Info("REFRESH WATCHERS", log.Fields{
		"instance":     instanceId,
		"subscription": subId,
	})
	v := NewInstanceClient()
	k8sClient := KubernetesClient{}
	instance, err := v.Get(instanceId)
	if err != nil {
		return pkgerrors.Wrap(err, "Cannot get instance for notify thread")
	}
	profile, err := rb.NewProfileClient().Get(instance.Request.RBName, instance.Request.RBVersion,
		instance.Request.ProfileName)
	if err != nil {
		return pkgerrors.Wrap(err, "Unable to find Profile instance status")
	}
	err = k8sClient.Init(instance.Request.CloudRegion, instanceId)
	if err != nil {
		return pkgerrors.Wrap(err, "Cannot set k8s client for instance")
	}

	lock, _, watchers := getSubscriptionData(instanceId)
	// Acquire Mutex
	lock.Lock()
	defer lock.Unlock()
	watcher, ok := watchers.watchersStatus[subId]
	if ok {
		close(watcher.watcherStop)
	} else {
		watchers.watchersStatus[subId] = subscriptionWatch{
			lastUpdateTime: time.Now(),
		}
	}

	watcher.watcherStop = make(chan struct{})

	for _, gvk := range gvkListForInstance(instance, profile) {
		informer, _ := k8sClient.GetInformer(gvk)
		handlers := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				lock.Lock()
				watcher.lastUpdateTime = time.Now()
				watchers.watchersStatus[subId] = watcher
				lock.Unlock()
			},
			UpdateFunc: func(oldObj, obj interface{}) {
				lock.Lock()
				watcher.lastUpdateTime = time.Now()
				watchers.watchersStatus[subId] = watcher
				lock.Unlock()
			},
			DeleteFunc: func(obj interface{}) {
				lock.Lock()
				watcher.lastUpdateTime = time.Now()
				watchers.watchersStatus[subId] = watcher
				lock.Unlock()
			},
		}
		informer.AddEventHandler(handlers)
		go func(informer cache.SharedInformer, stopper chan struct{}, fields log.Fields) {
			log.Info("[START] Watcher", fields)
			informer.Run(stopper)
			log.Info("[STOP] Watcher", fields)
		}(informer, watcher.watcherStop, log.Fields{
			"Kind":         gvk.Kind,
			"Instance":     instanceId,
			"Subscription": subId,
		})
	}
	return nil
}

// Get the Mutex for the Subscription
func getSubscriptionData(instanceId string) (*sync.Mutex, chan notifyChannelData, subscriptionWatchManager) {
	var key string = instanceId
	subscriptionNotifyData.Lock()
	defer subscriptionNotifyData.Unlock()
	_, ok := subscriptionNotifyData.notifyLockMap[key]
	if !ok {
		subscriptionNotifyData.notifyLockMap[key] = &sync.Mutex{}
	}
	_, ok = subscriptionNotifyData.notifyChannel[key]
	if !ok {
		subscriptionNotifyData.notifyChannel[key] = make(chan notifyChannelData)
		go scheduleNotifications(instanceId, subscriptionNotifyData.notifyChannel[key])
		time.Sleep(time.Second * 5)
	}
	_, ok = subscriptionNotifyData.watchersStatus[key]
	if !ok {
		subscriptionNotifyData.watchersStatus[key] = subscriptionWatchManager{
			watchersStatus: make(map[string]subscriptionWatch),
		}
	}
	return subscriptionNotifyData.notifyLockMap[key], subscriptionNotifyData.notifyChannel[key], subscriptionNotifyData.watchersStatus[key]
}

func removeSubscriptionData(instanceId string) {
	var key string = instanceId
	subscriptionNotifyData.Lock()
	defer subscriptionNotifyData.Unlock()
	_, ok := subscriptionNotifyData.notifyLockMap[key]
	if ok {
		delete(subscriptionNotifyData.notifyLockMap, key)
	}
	_, ok = subscriptionNotifyData.notifyChannel[key]
	if ok {
		crl := notifyChannelData{
			instanceId: instanceId,
			action:     "STOP",
		}
		subscriptionNotifyData.notifyChannel[key] <- crl
		delete(subscriptionNotifyData.notifyChannel, key)
	}
	_, ok = subscriptionNotifyData.watchersStatus[key]
	if !ok {
		delete(subscriptionNotifyData.watchersStatus, key)
	}
}

// notify request timeout
func notifyTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, time.Duration(time.Second*5))
}

// Per Subscription Go routine to send notification about status change
func scheduleNotifications(instanceId string, c chan notifyChannelData) {
	// Keep thread running
	log.Info("[START] status notify thread for ", log.Fields{
		"instance": instanceId,
	})
	for {
		data := <-c
		breakThread := false
		switch {
		case data.action == "NOTIFY":
			var result = notifyResult{}
			var err error = nil
			var notifyPayload = notifyRequestPayload{
				Delta:        data.delta,
				InstanceId:   data.instanceId,
				Subscription: data.subscription.Name,
				Metadata:     data.subscription.NotifyMetadata,
			}
			notifyBody, err := json.Marshal(notifyPayload)
			if err == nil {
				notifyBodyBuffer := bytes.NewBuffer(notifyBody)
				transport := http.Transport{
					Dial: notifyTimeout,
				}
				client := http.Client{
					Transport: &transport,
				}
				resp, errReq := client.Post(data.subscription.CallbackUrl, "application/json", notifyBodyBuffer)
				if errReq == nil {
					result.result = int32(resp.StatusCode)
					if resp.StatusCode >= 400 {
						respBody, _ := ioutil.ReadAll(resp.Body)
						log.Error("Status notification request failed", log.Fields{
							"instance": instanceId,
							"name":     data.subscription.Name,
							"url":      data.subscription.CallbackUrl,
							"code":     resp.StatusCode,
							"status":   resp.Status,
							"body":     string(respBody),
						})
						resp.Body.Close()
					}
				} else {
					err = errReq
				}
			}

			if err != nil {
				log.Error("Error for status notify thread", log.Fields{
					"instance": instanceId,
					"name":     data.subscription.Name,
					"err":      err.Error(),
				})
				result.result = 500
			}
			result.time = time.Now()

			data.notifyResult <- result

		case data.action == "STOP":
			breakThread = true
		}
		if breakThread {
			break
		}
	}
	log.Info("[STOP] status notify thread for ", log.Fields{
		"instance": instanceId,
	})
}

func gvkListForInstance(instance InstanceResponse, profile rb.Profile) []schema.GroupVersionKind {
	list := make([]schema.GroupVersionKind, 0)
	gvkMap := make(map[string]schema.GroupVersionKind)
	gvk := schema.FromAPIVersionAndKind("v1", "Pod")
	gvkMap[gvk.String()] = gvk
	for _, res := range instance.Resources {
		gvk = res.GVK
		_, ok := gvkMap[gvk.String()]
		if !ok {
			gvkMap[gvk.String()] = gvk
		}
	}
	for _, gvk := range profile.ExtraResourceTypes {
		_, ok := gvkMap[gvk.String()]
		if !ok {
			gvkMap[gvk.String()] = gvk
		}
	}
	for _, gvk := range gvkMap {
		list = append(list, gvk)
	}
	return list
}

func runNotifyThread(instanceId, subName string) {
	v := NewInstanceClient()
	iss := NewInstanceStatusSubClient()
	var status = InstanceStatus{
		ResourceCount: -1,
	}
	key := StatusSubscriptionKey{
		InstanceId:       instanceId,
		SubscriptionName: subName,
	}
	time.Sleep(time.Second * 5)
	log.Info("[START] status verification thread", log.Fields{
		"InstanceId":       instanceId,
		"SubscriptionName": subName,
	})

	lastChange := time.Now()
	var timeInSeconds time.Duration = 5
	for {
		time.Sleep(time.Second * timeInSeconds)

		lock, subData, watchers := getSubscriptionData(instanceId)
		var changeDetected = false
		lock.Lock()
		watcherStatus, ok := watchers.watchersStatus[subName]
		if ok {
			changeDetected = watcherStatus.lastUpdateTime.After(lastChange)
		}
		lock.Unlock()
		if !ok {
			break
		}
		if changeDetected || status.ResourceCount < 0 {
			currentSub, err := iss.Get(instanceId, subName)
			if err != nil {
				log.Error("Error getting current status", log.Fields{
					"error":    err.Error(),
					"instance": instanceId})
				break
			}
			if currentSub.MinNotifyInterval > 5 {
				timeInSeconds = time.Duration(currentSub.MinNotifyInterval)
			} else {
				timeInSeconds = 5
			}
			newStatus, err := v.Status(instanceId, false)
			if err != nil {
				log.Error("Error getting current status", log.Fields{
					"error":    err.Error(),
					"instance": instanceId})
				break
			} else {
				if status.ResourceCount >= 0 {
					var delta = statusDelta(status, newStatus)
					if delta.Delta() {
						log.Info("CHANGE DETECTED", log.Fields{
							"Instance":     instanceId,
							"Subscription": subName,
						})
						lastChange = watcherStatus.lastUpdateTime
						for _, res := range delta.Created {
							log.Info("CREATED", log.Fields{
								"Kind": res.GVK.Kind,
								"Name": res.Name,
							})
						}
						for _, res := range delta.Modified {
							log.Info("MODIFIED", log.Fields{
								"Kind": res.GVK.Kind,
								"Name": res.Name,
							})
						}
						for _, res := range delta.Deleted {
							log.Info("DELETED", log.Fields{
								"Kind": res.GVK.Kind,
								"Name": res.Name,
							})
						}
						// Acquire Mutex
						lock.Lock()
						currentSub.LastUpdateTime = time.Now()
						var notifyResultCh = make(chan notifyResult)
						var newData = notifyChannelData{
							instanceId:   instanceId,
							subscription: currentSub,
							action:       "NOTIFY",
							delta:        delta,
							notifyResult: notifyResultCh,
						}
						subData <- newData
						var notifyResult notifyResult = <-notifyResultCh
						log.Info("Notification sent", log.Fields{
							"InstanceId":       instanceId,
							"SubscriptionName": subName,
							"Result":           notifyResult.result,
						})
						currentSub.LastNotifyStatus = notifyResult.result
						currentSub.LastNotifyTime = notifyResult.time
						err = db.DBconn.Update(iss.storeName, key, iss.tagInst, currentSub)
						if err != nil {
							log.Error("Error updating subscription status", log.Fields{
								"error":    err.Error(),
								"instance": instanceId})
						}
						lock.Unlock()
					}
				}

				status = newStatus
			}
		}
	}
	log.Info("[STOP] status verification thread", log.Fields{
		"InstanceId":       instanceId,
		"SubscriptionName": subName,
	})
}

func statusDelta(first, second InstanceStatus) resourceStatusDelta {
	var delta resourceStatusDelta = resourceStatusDelta{
		Created:  make([]ResourceStatus, 0),
		Deleted:  make([]ResourceStatus, 0),
		Modified: make([]ResourceStatus, 0),
	}
	var firstResList map[string]ResourceStatus = make(map[string]ResourceStatus)
	for _, res := range first.ResourcesStatus {
		firstResList[res.Key()] = res
	}
	for _, res := range second.ResourcesStatus {
		var key string = res.Key()
		if prevRes, ok := firstResList[key]; ok {
			if prevRes.Value() != res.Value() {
				delta.Modified = append(delta.Modified, res)
			}
			delete(firstResList, res.Key())
		} else {
			delta.Created = append(delta.Created, res)
		}
	}
	for _, res := range firstResList {
		delta.Deleted = append(delta.Deleted, res)
	}
	return delta
}
