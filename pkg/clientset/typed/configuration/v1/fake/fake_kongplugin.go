/*
Copyright 2021 Kong, Inc.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeKongPlugins implements KongPluginInterface
type FakeKongPlugins struct {
	Fake *FakeConfigurationV1
	ns   string
}

var kongpluginsResource = v1.SchemeGroupVersion.WithResource("kongplugins")

var kongpluginsKind = v1.SchemeGroupVersion.WithKind("KongPlugin")

// Get takes name of the kongPlugin, and returns the corresponding kongPlugin object, and an error if there is any.
func (c *FakeKongPlugins) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.KongPlugin, err error) {
	emptyResult := &v1.KongPlugin{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(kongpluginsResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.KongPlugin), err
}

// List takes label and field selectors, and returns the list of KongPlugins that match those selectors.
func (c *FakeKongPlugins) List(ctx context.Context, opts metav1.ListOptions) (result *v1.KongPluginList, err error) {
	emptyResult := &v1.KongPluginList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(kongpluginsResource, kongpluginsKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.KongPluginList{ListMeta: obj.(*v1.KongPluginList).ListMeta}
	for _, item := range obj.(*v1.KongPluginList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested kongPlugins.
func (c *FakeKongPlugins) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(kongpluginsResource, c.ns, opts))

}

// Create takes the representation of a kongPlugin and creates it.  Returns the server's representation of the kongPlugin, and an error, if there is any.
func (c *FakeKongPlugins) Create(ctx context.Context, kongPlugin *v1.KongPlugin, opts metav1.CreateOptions) (result *v1.KongPlugin, err error) {
	emptyResult := &v1.KongPlugin{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(kongpluginsResource, c.ns, kongPlugin, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.KongPlugin), err
}

// Update takes the representation of a kongPlugin and updates it. Returns the server's representation of the kongPlugin, and an error, if there is any.
func (c *FakeKongPlugins) Update(ctx context.Context, kongPlugin *v1.KongPlugin, opts metav1.UpdateOptions) (result *v1.KongPlugin, err error) {
	emptyResult := &v1.KongPlugin{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(kongpluginsResource, c.ns, kongPlugin, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.KongPlugin), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeKongPlugins) UpdateStatus(ctx context.Context, kongPlugin *v1.KongPlugin, opts metav1.UpdateOptions) (result *v1.KongPlugin, err error) {
	emptyResult := &v1.KongPlugin{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(kongpluginsResource, "status", c.ns, kongPlugin, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.KongPlugin), err
}

// Delete takes name of the kongPlugin and deletes it. Returns an error if one occurs.
func (c *FakeKongPlugins) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(kongpluginsResource, c.ns, name, opts), &v1.KongPlugin{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeKongPlugins) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(kongpluginsResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1.KongPluginList{})
	return err
}

// Patch applies the patch and returns the patched kongPlugin.
func (c *FakeKongPlugins) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.KongPlugin, err error) {
	emptyResult := &v1.KongPlugin{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(kongpluginsResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.KongPlugin), err
}
