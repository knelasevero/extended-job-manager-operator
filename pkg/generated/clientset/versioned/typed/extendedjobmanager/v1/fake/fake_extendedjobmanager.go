/*
Copyright The Kubernetes Authors.

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
	json "encoding/json"
	"fmt"

	v1 "github.com/knelasevero/extended-job-manager-operator/pkg/apis/extendedjobmanager/v1"
	extendedjobmanagerv1 "github.com/knelasevero/extended-job-manager-operator/pkg/generated/applyconfiguration/extendedjobmanager/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeExtendedJobManagers implements ExtendedJobManagerInterface
type FakeExtendedJobManagers struct {
	Fake *FakeExtendedjobmanagerV1
	ns   string
}

var extendedjobmanagersResource = v1.SchemeGroupVersion.WithResource("extendedjobmanagers")

var extendedjobmanagersKind = v1.SchemeGroupVersion.WithKind("ExtendedJobManager")

// Get takes name of the extendedJobManager, and returns the corresponding extendedJobManager object, and an error if there is any.
func (c *FakeExtendedJobManagers) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.ExtendedJobManager, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(extendedjobmanagersResource, c.ns, name), &v1.ExtendedJobManager{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ExtendedJobManager), err
}

// List takes label and field selectors, and returns the list of ExtendedJobManagers that match those selectors.
func (c *FakeExtendedJobManagers) List(ctx context.Context, opts metav1.ListOptions) (result *v1.ExtendedJobManagerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(extendedjobmanagersResource, extendedjobmanagersKind, c.ns, opts), &v1.ExtendedJobManagerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.ExtendedJobManagerList{ListMeta: obj.(*v1.ExtendedJobManagerList).ListMeta}
	for _, item := range obj.(*v1.ExtendedJobManagerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested extendedJobManagers.
func (c *FakeExtendedJobManagers) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(extendedjobmanagersResource, c.ns, opts))

}

// Create takes the representation of a extendedJobManager and creates it.  Returns the server's representation of the extendedJobManager, and an error, if there is any.
func (c *FakeExtendedJobManagers) Create(ctx context.Context, extendedJobManager *v1.ExtendedJobManager, opts metav1.CreateOptions) (result *v1.ExtendedJobManager, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(extendedjobmanagersResource, c.ns, extendedJobManager), &v1.ExtendedJobManager{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ExtendedJobManager), err
}

// Update takes the representation of a extendedJobManager and updates it. Returns the server's representation of the extendedJobManager, and an error, if there is any.
func (c *FakeExtendedJobManagers) Update(ctx context.Context, extendedJobManager *v1.ExtendedJobManager, opts metav1.UpdateOptions) (result *v1.ExtendedJobManager, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(extendedjobmanagersResource, c.ns, extendedJobManager), &v1.ExtendedJobManager{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ExtendedJobManager), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeExtendedJobManagers) UpdateStatus(ctx context.Context, extendedJobManager *v1.ExtendedJobManager, opts metav1.UpdateOptions) (*v1.ExtendedJobManager, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(extendedjobmanagersResource, "status", c.ns, extendedJobManager), &v1.ExtendedJobManager{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ExtendedJobManager), err
}

// Delete takes name of the extendedJobManager and deletes it. Returns an error if one occurs.
func (c *FakeExtendedJobManagers) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(extendedjobmanagersResource, c.ns, name, opts), &v1.ExtendedJobManager{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeExtendedJobManagers) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(extendedjobmanagersResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1.ExtendedJobManagerList{})
	return err
}

// Patch applies the patch and returns the patched extendedJobManager.
func (c *FakeExtendedJobManagers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.ExtendedJobManager, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(extendedjobmanagersResource, c.ns, name, pt, data, subresources...), &v1.ExtendedJobManager{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ExtendedJobManager), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied extendedJobManager.
func (c *FakeExtendedJobManagers) Apply(ctx context.Context, extendedJobManager *extendedjobmanagerv1.ExtendedJobManagerApplyConfiguration, opts metav1.ApplyOptions) (result *v1.ExtendedJobManager, err error) {
	if extendedJobManager == nil {
		return nil, fmt.Errorf("extendedJobManager provided to Apply must not be nil")
	}
	data, err := json.Marshal(extendedJobManager)
	if err != nil {
		return nil, err
	}
	name := extendedJobManager.Name
	if name == nil {
		return nil, fmt.Errorf("extendedJobManager.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(extendedjobmanagersResource, c.ns, *name, types.ApplyPatchType, data), &v1.ExtendedJobManager{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ExtendedJobManager), err
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *FakeExtendedJobManagers) ApplyStatus(ctx context.Context, extendedJobManager *extendedjobmanagerv1.ExtendedJobManagerApplyConfiguration, opts metav1.ApplyOptions) (result *v1.ExtendedJobManager, err error) {
	if extendedJobManager == nil {
		return nil, fmt.Errorf("extendedJobManager provided to Apply must not be nil")
	}
	data, err := json.Marshal(extendedJobManager)
	if err != nil {
		return nil, err
	}
	name := extendedJobManager.Name
	if name == nil {
		return nil, fmt.Errorf("extendedJobManager.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(extendedjobmanagersResource, c.ns, *name, types.ApplyPatchType, data, "status"), &v1.ExtendedJobManager{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ExtendedJobManager), err
}
