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
	"bytes"
	"fmt"
	"io"

	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/validation"
)

// DefaultValidation default action to validate. If `true` all resources by
// default will be validated.
const DefaultValidation = false

// Client is a kubernetes client, like `kubectl`
type Client struct {
	Clientset        *kubernetes.Clientset
	factory          *factory
	validator        validation.Schema
	namespace        string
	enforceNamespace bool
	forceConflicts   bool
	ServerSideApply  bool
}

// Result is an alias for the Kubernetes CLI runtime resource.Result
type Result = resource.Result

// BuilderOptions parameters to create a Resource Builder
type BuilderOptions struct {
	Unstructured  bool
	Validate      bool
	Namespace     string
	LabelSelector string
	FieldSelector string
	All           bool
	AllNamespaces bool
}

// NewBuilderOptions creates a BuilderOptions with the default values for
// the parameters to create a Resource Builder
func NewBuilderOptions() *BuilderOptions {
	return &BuilderOptions{
		Unstructured: true,
		Validate:     true,
	}
}

// NewE creates a kubernetes client, returns an error if fail
func NewE(context, kubeconfig string, ns string) (*Client, error) {
	var namespace string
	var enforceNamespace bool
	var err error
	factory := newFactory(context, kubeconfig)

	// If `true` it will always validate the given objects/resources
	// Unless something different is specified in the NewBuilderOptions
	validator, _ := factory.Validator(DefaultValidation)

	if ns == "" {
		namespace, enforceNamespace, err = factory.ToRawKubeConfigLoader().Namespace()
		if err != nil {
			namespace = v1.NamespaceDefault
			enforceNamespace = true
		}
	} else {
		namespace = ns
		enforceNamespace = false
	}
	clientset, err := factory.KubernetesClientSet()
	if err != nil {
		return nil, err
	}
	if clientset == nil {
		return nil, fmt.Errorf("cannot create a clientset from given context and kubeconfig")
	}

	return &Client{
		factory:          factory,
		Clientset:        clientset,
		validator:        validator,
		namespace:        namespace,
		enforceNamespace: enforceNamespace,
	}, nil
}

// New creates a kubernetes client
func New(context, kubeconfig string, namespace string) *Client {
	client, _ := NewE(context, kubeconfig, namespace)
	return client
}

// Builder creates a resource builder
func (c *Client) builder(opt *BuilderOptions) *resource.Builder {
	validator := c.validator
	namespace := c.namespace

	if opt == nil {
		opt = NewBuilderOptions()
	} else {
		if opt.Validate != DefaultValidation {
			validator, _ = c.factory.Validator(opt.Validate)
		}
		if opt.Namespace != "" {
			namespace = opt.Namespace
		}
	}

	b := c.factory.NewBuilder()
	if opt.Unstructured {
		b = b.Unstructured()
	}

	return b.
		Schema(validator).
		ContinueOnError().
		NamespaceParam(namespace).DefaultNamespace()
}

// ResultForFilenameParam returns the builder results for the given list of files or URLs
func (c *Client) ResultForFilenameParam(filenames []string, opt *BuilderOptions) *Result {
	filenameOptions := &resource.FilenameOptions{
		Recursive: false,
		Filenames: filenames,
	}

	return c.builder(opt).
		FilenameParam(c.enforceNamespace, filenameOptions).
		Flatten().
		Do()
}

// ResultForReader returns the builder results for the given reader
func (c *Client) ResultForReader(r io.Reader, opt *BuilderOptions) *Result {
	return c.builder(opt).
		Stream(r, "").
		Flatten().
		Do()
}

// func (c *Client) ResultForName(opt *BuilderOptions, names ...string) *Result {
// 	return c.builder(opt).
// 		LabelSelectorParam(opt.LabelSelector).
// 		FieldSelectorParam(opt.FieldSelector).
// 		SelectAllParam(opt.All).
// 		AllNamespaces(opt.AllNamespaces).
// 		ResourceTypeOrNameArgs(false, names...).RequireObject(false).
// 		Flatten().
// 		Do()
// }

// ResultForContent returns the builder results for the given content
func (c *Client) ResultForContent(content []byte, opt *BuilderOptions) *Result {
	b := bytes.NewBuffer(content)
	return c.ResultForReader(b, opt)
}

func failedTo(action string, info *resource.Info, err error) error {
	var resKind string
	if info.Mapping != nil {
		resKind = info.Mapping.GroupVersionKind.Kind + " "
	}

	return fmt.Errorf("cannot %s object Kind: %q,	Name: %q, Namespace: %q. %s", action, resKind, info.Name, info.Namespace, err)
}

// IsReachable tests connectivity to the cluster
func (c *Client) IsReachable() error {
	client, _ := c.factory.KubernetesClientSet()
	_, err := client.ServerVersion()
	if err != nil {
		return fmt.Errorf("Kubernetes cluster unreachable")
	}
	return nil
}
