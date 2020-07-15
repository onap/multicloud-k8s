/*
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
// Based on Code: https://github.com/johandry/klient
package client

import (
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery"
	diskcached "k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubectl/pkg/util/openapi"
	openapivalidation "k8s.io/kubectl/pkg/util/openapi/validation"
	"k8s.io/kubectl/pkg/validation"
)

// factory implements the kubectl Factory interface which also requieres to
// implement the genericclioptions.RESTClientGetter interface.
// The Factory inerface provides abstractions that allow the Kubectl command to
// be extended across multiple types of resources and different API sets.
type factory struct {
	KubeConfig            string
	Context               string
	initOpenAPIGetterOnce sync.Once
	openAPIGetter         openapi.Getter
}

// If multiple clients are created, this sync.once make sure the CRDs are added
// only once into the API extensions v1 and v1beta schemes
var addToSchemeOnce sync.Once

var _ genericclioptions.RESTClientGetter = &factory{}

// newFactory creates a new client factory which encapsulate a REST client getter
func newFactory(context, kubeconfig string) *factory {
	factory := &factory{
		KubeConfig: kubeconfig,
		Context:    context,
	}

	// From: helm/pkg/kube/client.go > func New()
	// Add CRDs to the scheme. They are missing by default.
	addToSchemeOnce.Do(func() {
		if err := apiextv1.AddToScheme(scheme.Scheme); err != nil {
			panic(err)
		}
		if err := apiextv1beta1.AddToScheme(scheme.Scheme); err != nil {
			panic(err)
		}
	})

	return factory
}

// BuildRESTConfig builds a kubernetes REST client factory using the following
// rules from ToRawKubeConfigLoader()
// func BuildRESTConfig(context, kubeconfig string) (*rest.Config, error) {
// 	return newFactory(context, kubeconfig).ToRESTConfig()
// }

// ToRESTConfig creates a kubernetes REST client factory.
// It's required to implement the interface genericclioptions.RESTClientGetter
func (f *factory) ToRESTConfig() (*rest.Config, error) {
	// From: k8s.io/kubectl/pkg/cmd/util/kubectl_match_version.go > func setKubernetesDefaults()
	config, err := f.ToRawKubeConfigLoader().ClientConfig()
	if err != nil {
		return nil, err
	}

	if config.GroupVersion == nil {
		config.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	}
	if config.APIPath == "" {
		config.APIPath = "/api"
	}
	if config.NegotiatedSerializer == nil {
		// This codec config ensures the resources are not converted. Therefore, resources
		// will not be round-tripped through internal versions. Defaulting does not happen
		// on the client.
		config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	}

	rest.SetKubernetesDefaults(config)
	return config, nil
}

// ToRawKubeConfigLoader creates a client factory using the following rules:
// 1. builds from the given kubeconfig path, if not empty
// 2. use the in cluster factory if running in-cluster
// 3. gets the factory from KUBECONFIG env var
// 4. Uses $HOME/.kube/factory
// It's required to implement the interface genericclioptions.RESTClientGetter
func (f *factory) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	if len(f.KubeConfig) != 0 {
		loadingRules.ExplicitPath = f.KubeConfig
	}
	configOverrides := &clientcmd.ConfigOverrides{
		ClusterDefaults: clientcmd.ClusterDefaults,
	}
	if len(f.Context) != 0 {
		configOverrides.CurrentContext = f.Context
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
}

// overlyCautiousIllegalFileCharacters matches characters that *might* not be supported.  Windows is really restrictive, so this is really restrictive
var overlyCautiousIllegalFileCharacters = regexp.MustCompile(`[^(\w/\.)]`)

// ToDiscoveryClient returns a CachedDiscoveryInterface using a computed RESTConfig
// It's required to implement the interface genericclioptions.RESTClientGetter
func (f *factory) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	// From: k8s.io/cli-runtime/pkg/genericclioptions/config_flags.go > func (*configFlags) ToDiscoveryClient()
	factory, err := f.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	factory.Burst = 100
	defaultHTTPCacheDir := filepath.Join(homedir.HomeDir(), ".kube", "http-cache")

	// takes the parentDir and the host and comes up with a "usually non-colliding" name for the discoveryCacheDir
	parentDir := filepath.Join(homedir.HomeDir(), ".kube", "cache", "discovery")
	// strip the optional scheme from host if its there:
	schemelessHost := strings.Replace(strings.Replace(factory.Host, "https://", "", 1), "http://", "", 1)
	// now do a simple collapse of non-AZ09 characters.  Collisions are possible but unlikely.  Even if we do collide the problem is short lived
	safeHost := overlyCautiousIllegalFileCharacters.ReplaceAllString(schemelessHost, "_")
	discoveryCacheDir := filepath.Join(parentDir, safeHost)

	return diskcached.NewCachedDiscoveryClientForConfig(factory, discoveryCacheDir, defaultHTTPCacheDir, time.Duration(10*time.Minute))
}

// ToRESTMapper returns a mapper
// It's required to implement the interface genericclioptions.RESTClientGetter
func (f *factory) ToRESTMapper() (meta.RESTMapper, error) {
	// From: k8s.io/cli-runtime/pkg/genericclioptions/config_flags.go > func (*configFlags) ToRESTMapper()
	discoveryClient, err := f.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient)
	return expander, nil
}

// KubernetesClientSet creates a kubernetes clientset from the configuration
// It's required to implement the Factory interface
func (f *factory) KubernetesClientSet() (*kubernetes.Clientset, error) {
	// From: k8s.io/kubectl/pkg/cmd/util/factory_client_access.go > func (*factoryImpl) KubernetesClientSet()
	factory, err := f.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(factory)
}

// DynamicClient creates a dynamic client from the configuration
// It's required to implement the Factory interface
func (f *factory) DynamicClient() (dynamic.Interface, error) {
	// From: k8s.io/kubectl/pkg/cmd/util/factory_client_access.go > func (*factoryImpl) DynamicClient()
	factory, err := f.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(factory)
}

// NewBuilder returns a new resource builder for structured api objects.
// It's required to implement the Factory interface
func (f *factory) NewBuilder() *resource.Builder {
	// From: k8s.io/kubectl/pkg/cmd/util/factory_client_access.go > func (*factoryImpl) NewBuilder()
	return resource.NewBuilder(f)
}

// RESTClient creates a REST client from the configuration
// It's required to implement the Factory interface
func (f *factory) RESTClient() (*rest.RESTClient, error) {
	// From: k8s.io/kubectl/pkg/cmd/util/factory_client_access.go > func (*factoryImpl) RESTClient()
	factory, err := f.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	return rest.RESTClientFor(factory)
}

func (f *factory) configForMapping(mapping *meta.RESTMapping) (*rest.Config, error) {
	factory, err := f.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	gvk := mapping.GroupVersionKind
	factory.APIPath = "/apis"
	if gvk.Group == corev1.GroupName {
		factory.APIPath = "/api"
	}
	gv := gvk.GroupVersion()
	factory.GroupVersion = &gv

	return factory, nil
}

// ClientForMapping creates a resource REST client from the given mappings
// It's required to implement the Factory interface
func (f *factory) ClientForMapping(mapping *meta.RESTMapping) (resource.RESTClient, error) {
	// From: k8s.io/kubectl/pkg/cmd/util/factory_client_access.go > func (*factoryImpl) ClientForMapping()
	factory, err := f.configForMapping(mapping)
	if err != nil {
		return nil, err
	}

	return rest.RESTClientFor(factory)
}

// UnstructuredClientForMapping creates a unstructured resource REST client from the given mappings
// It's required to implement the Factory interface
func (f *factory) UnstructuredClientForMapping(mapping *meta.RESTMapping) (resource.RESTClient, error) {
	// From: k8s.io/kubectl/pkg/cmd/util/factory_client_access.go > func (*factoryImpl) UnstructuredClientForMapping()
	factory, err := f.configForMapping(mapping)
	if err != nil {
		return nil, err
	}
	factory.ContentConfig = resource.UnstructuredPlusDefaultContentConfig()

	return rest.RESTClientFor(factory)
}

// Validator returns a schema that can validate objects stored on disk.
// It's required to implement the Factory interface
func (f *factory) Validator(validate bool) (validation.Schema, error) {
	// From: k8s.io/kubectl/pkg/cmd/util/factory_client_access.go > func (*factoryImpl) Validator(bool)
	if !validate {
		return validation.NullSchema{}, nil
	}

	resources, err := f.OpenAPISchema()
	if err != nil {
		return nil, err
	}

	return validation.ConjunctiveSchema{
		openapivalidation.NewSchemaValidation(resources),
		validation.NoDoubleKeySchema{},
	}, nil
}

// OpenAPISchema returns metadata and structural information about Kubernetes object definitions.
// It's required to implement the Factory interface
func (f *factory) OpenAPISchema() (openapi.Resources, error) {
	// From: k8s.io/kubectl/pkg/cmd/util/factory_client-access.go > func (*factoryImpl) OpenAPISchema()
	discovery, err := f.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	f.initOpenAPIGetterOnce.Do(func() {
		// Create the caching OpenAPIGetter
		f.openAPIGetter = openapi.NewOpenAPIGetter(discovery)
	})

	// Delegate to the OpenAPIGetter
	return f.openAPIGetter.Get()
}
