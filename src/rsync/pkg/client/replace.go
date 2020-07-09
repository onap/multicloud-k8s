package client

import (
	"k8s.io/cli-runtime/pkg/resource"
)

// Replace creates a resource with the given content
func (c *Client) Replace(content []byte) error {
	r := c.ResultForContent(content, nil)
	return c.ReplaceResource(r)
}

// ReplaceFiles create the resource(s) from the given filenames (file, directory or STDIN) or HTTP URLs
func (c *Client) ReplaceFiles(filenames ...string) error {
	r := c.ResultForFilenameParam(filenames, nil)
	return c.ReplaceResource(r)
}

// ReplaceResource applies the given resource. Create the resources with `ResultForFilenameParam` or `ResultForContent`
func (c *Client) ReplaceResource(r *resource.Result) error {
	if err := r.Err(); err != nil {
		return err
	}
	return r.Visit(replace)
}

func replace(info *resource.Info, err error) error {
	if err != nil {
		return failedTo("replace", info, err)
	}

	obj, err := resource.NewHelper(info.Client, info.Mapping).Replace(info.Namespace, info.Name, true, info.Object)
	if err != nil {
		return failedTo("replace", info, err)
	}
	info.Refresh(obj, true)

	return nil
}
