package app

import (
	"sync"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"
)

var k8sNativeScheme *runtime.Scheme
var k8sNativeSchemeOnce sync.Once

// AsVersioned converts the given info into a runtime.Object with the correct
// group and version set
func AsVersioned(info *resource.Info) runtime.Object {
	return convertWithMapper(info.Object, info.Mapping)
}

// convertWithMapper converts the given object with the optional provided
// RESTMapping. If no mapping is provided, the default schema versioner is used
func convertWithMapper(obj runtime.Object, mapping *meta.RESTMapping) runtime.Object {
	s := kubernetesNativeScheme()
	var gv = runtime.GroupVersioner(schema.GroupVersions(s.PrioritizedVersionsAllGroups()))
	if mapping != nil {
		gv = mapping.GroupVersionKind.GroupVersion()
	}
	if obj, err := runtime.ObjectConvertor(s).ConvertToVersion(obj, gv); err == nil {
		return obj
	}
	return obj
}

// kubernetesNativeScheme returns a clean *runtime.Scheme with _only_ Kubernetes
// native resources added to it. This is required to break free of custom resources
// that may have been added to scheme.Scheme due to Helm being used as a package in
// combination with e.g. a versioned kube client. If we would not do this, the client
// may attempt to perform e.g. a 3-way-merge strategy patch for custom resources.
func kubernetesNativeScheme() *runtime.Scheme {
	k8sNativeSchemeOnce.Do(func() {
		k8sNativeScheme = runtime.NewScheme()
		scheme.AddToScheme(k8sNativeScheme)
		// API extensions are not in the above scheme set,
		// and must thus be added separately.
		apiextensionsv1beta1.AddToScheme(k8sNativeScheme)
		apiextensionsv1.AddToScheme(k8sNativeScheme)
	})
	return k8sNativeScheme
}

