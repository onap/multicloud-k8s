/*
Copyright © 2021 Samsung Electronics
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package healthcheck

import (
	"time"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	log "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"

	"helm.sh/helm/v3/pkg/release"
	helmtime "helm.sh/helm/v3/pkg/time"

	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	pkgerrors "github.com/pkg/errors"
)

type HookStatus struct {
	StartedAt   helmtime.Time           `json:"started_at"`
	CompletedAt helmtime.Time           `json:"completed_at"`
	Status      release.HookPhase       `json:"status"`
	Name        string                  `json:"name"`
	KR          helm.KubernetesResource `json:"-"`
}

// Helper type to combine Hook definition with Status
type hookPair struct {
	Definition *helm.Hook
	Status     *HookStatus
}

// Wraper type to implement helper method for extraction
type hookPairs []hookPair

// Helper function to retrieve slice of Statuses from slice of hookPairs
func (hps hookPairs) statuses() (hsps []*HookStatus) {
	for _, hp := range hps {
		hsps = append(hsps, hp.Status)
	}
	return
}

// TODO Optimize by using k8s.io/client-go/tools/cache.NewListWatchFromClient just like
// in helm.sh/helm/v3/pkg/kube/client.go -> watchUntilReady()
func getHookState(hookStatus HookStatus, k8sClient app.KubernetesClient, namespace string) (release.HookPhase, error) {
	// Initial check of Hook Resource type
	switch hookStatus.KR.GVK.Kind {
	case "Job", "Pod":
	default:
		//We don't know how to check state of such resource
		return release.HookPhaseUnknown, nil
	}

	for {
		res, err := k8sClient.GetResourceStatus(hookStatus.KR, namespace)
		if err != nil {
			log.Error("Unable to check Resource Status", log.Fields{
				"Resource":  hookStatus.KR,
				"Namespace": namespace,
				"Error":     err,
			})
			return release.HookPhaseUnknown,
				pkgerrors.Wrap(err, "Unable to check Resource Status")
		}

		var parsedRes runtime.Object
		switch hookStatus.KR.GVK.Kind {
		case "Job":
			parsedRes = new(batch.Job)
		case "Pod":
			parsedRes = new(v1.Pod)
		}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(res.Status.Object, parsedRes)
		if err != nil {
			log.Error("Couldn't convert Response into runtime object", log.Fields{
				"Response": res.Status.Object,
				"Error":    err,
			})
			return release.HookPhaseUnknown,
				pkgerrors.Wrap(err, "Couldn't conver Response into runtime object")
		}

		var tempState release.HookPhase
		switch hookStatus.KR.GVK.Kind {
		case "Job":
			tempState = parseJobState(parsedRes)
		case "Pod":
			tempState = parsePodState(parsedRes)
		}
		if tempState != release.HookPhaseRunning {
			return tempState, nil
		}
		//TODO should be changed to "Watching" of resource as pointed earlier
		time.Sleep(5 * time.Second)
	}
}

// Based on kube/client.go -> waitForJob()
func parseJobState(obj runtime.Object) (state release.HookPhase) {
	job, ok := obj.(*batch.Job)
	if !ok {
		//Something went wrong, and we don't want to parse such resource again
		return release.HookPhaseUnknown
	}
	for _, c := range job.Status.Conditions {
		if c.Type == batch.JobComplete && c.Status == "True" {
			return release.HookPhaseSucceeded
		} else if c.Type == batch.JobFailed && c.Status == "True" {
			return release.HookPhaseFailed
		}
	}
	return release.HookPhaseRunning
}

// Based on kube/client.go -> waitForPodSuccess()
func parsePodState(obj runtime.Object) (state release.HookPhase) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return release.HookPhaseUnknown
	}

	switch pod.Status.Phase {
	case v1.PodSucceeded:
		return release.HookPhaseSucceeded
	case v1.PodFailed:
		return release.HookPhaseFailed
	default:
		return release.HookPhaseRunning
	}
}
