/*
Copyright 2026 Deutsche Telekom AG
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

package app

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/connection"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

// TestMain shrinks the scheduleResources startup delay for the whole package so
// config-touching tests do not each pay the production 5s grace period. The
// delay must stay non-zero: applyConfig sends to the scheduler goroutine with a
// non-blocking select, so the goroutine has to reach its receive first or the
// message is dropped and applyConfig blocks forever.
func TestMain(m *testing.M) {
	scheduleResourcesStartupDelay = 50 * time.Millisecond
	scheduleNotificationsStartupDelay = 50 * time.Millisecond
	os.Exit(m.Run())
}

// fakeHookGVKs are the kinds exercised by the hook and instance tests. A static
// RESTMapper carrying them lets GetResourceStatus / DeleteKind resolve a mapping
// without a DeferredDiscoveryRESTMapper dialing the (unreachable) apiserver.
var fakeHookGVKs = []schema.GroupVersionKind{
	{Group: "batch", Version: "v1", Kind: "Job"},
	{Group: "apps", Version: "v1", Kind: "Deployment"},
	{Group: "", Version: "v1", Kind: "Service"},
	{Group: "", Version: "v1", Kind: "ServiceAccount"},
}

// useFakeClients points the buildClients seam at fakes that perform no network
// I/O: a fake clientset, a fake dynamic client seeded with objects, and a static
// RESTMapper covering fakeHookGVKs. It returns a restore func to be deferred.
func useFakeClients(objects ...runtime.Object) func() {
	mapper := meta.NewDefaultRESTMapper(nil)
	for _, gvk := range fakeHookGVKs {
		mapper.Add(gvk, meta.RESTScopeNamespace)
	}
	scheme := runtime.NewScheme()
	dyn := dynamicfake.NewSimpleDynamicClient(scheme, objects...)
	cs := fake.NewSimpleClientset()

	return setBuildClients(func(*rest.Config) (kubernetes.Interface, dynamic.Interface, meta.RESTMapper, error) {
		return cs, dyn, mapper, nil
	})
}

// mockConnectionDB returns a MockDB carrying the mock kubeconfig so that
// KubernetesClient.Init can build a rest.Config without touching a cluster.
func mockConnectionDB(t *testing.T) *db.MockDB {
	t.Helper()
	fd, err := ioutil.ReadFile("../../mock_files/mock_configs/mock_kube_config")
	if err != nil {
		t.Fatal("Unable to read mock_kube_config")
	}
	return &db.MockDB{
		Items: map[string]map[string][]byte{
			connection.ConnectionKey{CloudRegion: "mock_connection"}.String(): {
				"metadata": []byte(
					"{\"cloud-region\":\"mock_connection\"," +
						"\"cloud-owner\":\"mock_owner\"," +
						"\"kubeconfig\": \"" + base64.StdEncoding.EncodeToString(fd) + "\"}"),
			},
		},
	}
}

// TestInitUsesInjectedClientFactory proves that Init builds its clients through
// the overridable buildClients seam, so tests can supply fakes that perform no
// network I/O instead of the real, cluster-dialing clients.
func TestInitUsesInjectedClientFactory(t *testing.T) {
	fakeCS := fake.NewSimpleClientset()
	fakeDyn := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	fakeMapper := meta.NewDefaultRESTMapper(nil)

	restore := setBuildClients(func(*rest.Config) (kubernetes.Interface, dynamic.Interface, meta.RESTMapper, error) {
		return fakeCS, fakeDyn, fakeMapper, nil
	})
	defer restore()

	db.DBconn = mockConnectionDB(t)

	k := KubernetesClient{}
	if err := k.Init(context.TODO(), "mock_connection", "iid"); err != nil {
		t.Fatalf("Init returned an error: %s", err)
	}

	if k.GetStandardClient() != fakeCS {
		t.Fatal("Init did not use the injected clientset")
	}
	if k.GetDynamicClient() != fakeDyn {
		t.Fatal("Init did not use the injected dynamic client")
	}
	if k.GetMapper() != fakeMapper {
		t.Fatal("Init did not use the injected RESTMapper")
	}
}
