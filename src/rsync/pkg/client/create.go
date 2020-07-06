package client

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/resource"
)

// Create creates a resource with the given content
func (c *Client) Create(content []byte) error {
	r := c.ResultForContent(content, nil)
	return c.CreateResource(r)
}

// CreateFile creates a resource with the given content
func (c *Client) CreateFile(filenames ...string) error {
	r := c.ResultForFilenameParam(filenames, nil)
	return c.CreateResource(r)
}

// CreateResource creates the given resource. Create the resources with `ResultForFilenameParam` or `ResultForContent`
func (c *Client) CreateResource(r *resource.Result) error {
	if err := r.Err(); err != nil {
		return err
	}
	return r.Visit(create)
}

func create(info *resource.Info, err error) error {
	if err != nil {
		return failedTo("create", info, err)
	}

	// TODO: If will be allow to do create then apply, here must be added the annotation as in Apply/Patch

	options := metav1.CreateOptions{}
	obj, err := resource.NewHelper(info.Client, info.Mapping).Create(info.Namespace, true, info.Object, &options)
	if err != nil {
		return failedTo("create", info, err)
	}
	info.Refresh(obj, true)

	return nil
}
