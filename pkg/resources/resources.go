/*
	SPDX-License-Identifier: MIT
*/

package resources

import (
	"fmt"
	"time"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	"github.com/imdario/mergo"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const (
	waitTimeoutSeconds       = 600
	waitCheckIntervalSeconds = 5
)

// ToUnstructured returns an unstructured representation of a Resource.
func (resource *Resource) ToUnstructured() (*unstructured.Unstructured, error) {
	innerObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&resource.Object)
	if err != nil {
		return nil, err
	}

	return &unstructured.Unstructured{Object: innerObject}, nil
}

// ToCommonResource converts a resources.Resource into a common API resource.
func (resource *Resource) ToCommonResource() *ResourceCommon {
	return &ResourceCommon{
		Group:     resource.Group,
		Version:   resource.Version,
		Kind:      resource.Kind,
		Name:      resource.Name,
		Namespace: resource.Namespace,
	}
}

func (resource *Resource) setResourceChecker(name, namespace string) {
	switch resource.Kind {
	case NamespaceKind:
		resource.resourceChecker = NewNamespaceResource(name, namespace)
	case CustomResourceDefinitionKind:
		resource.resourceChecker = NewCRDResource(name, namespace)
	case SecretKind:
		resource.resourceChecker = NewSecretResource(name, namespace)
	case ConfigMapKind:
		resource.resourceChecker = NewConfigMapResource(name, namespace)
	case DeploymentKind:
		resource.resourceChecker = NewDeploymentResource(name, namespace)
	case DaemonSetKind:
		resource.resourceChecker = NewDaemonSetResource(name, namespace)
	case StatefulSetKind:
		resource.resourceChecker = NewStatefulSetResource(name, namespace)
	case JobKind:
		resource.resourceChecker = NewJobResource(name, namespace)
	case ServiceKind:
		resource.resourceChecker = NewServiceResource(name, namespace)
	default:
		resource.resourceChecker = NewUnknownResource(name, namespace)
	}
}

// Wait waits for a resource to enter a ready state.
func (resource *Resource) Wait() error {
	timeout := time.After(waitTimeoutSeconds * time.Second)
	interval := time.Tick(waitCheckIntervalSeconds * time.Second)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timed out waiting for resource")
		case <-interval:
			ready, err := resource.IsReady()
			if err != nil {
				return fmt.Errorf("error waiting for resource to be ready, %w", err)
			}

			if ready {
				return nil
			}
		}
	}
}

// IsReady returns whether a specific known resource is ready.  Always returns true for unknown resources
// so that dependency checks will not fail and reconciliation of resources can happen with errors rather
// than stopping entirely.
func (resource *Resource) IsReady() (bool, error) {
	if _, isUnknown := resource.resourceChecker.(*UnknownResource); !isUnknown {
		// get the object from the kubernetes cluster
		if err := GetObject(resource); err != nil {
			return false, err
		}
	}

	return resource.resourceChecker.IsReady(resource)
}

// AreReady returns whether resources are ready.  All resources must be ready in order
// to satisfy the requirement that resources are ready.
func AreReady(resources ...*Resource) (bool, error) {
	for _, rsrc := range resources {
		ready, err := rsrc.IsReady()
		if !ready || err != nil {
			return false, err
		}
	}

	return true, nil
}

// AreEqual determines if two resources are equal.
func AreEqual(desired, actual *Resource) (bool, error) {
	mergedResource, err := actual.ToUnstructured()
	if err != nil {
		return false, err
	}

	actualResource, err := actual.ToUnstructured()
	if err != nil {
		return false, err
	}

	desiredResource, err := desired.ToUnstructured()
	if err != nil {
		return false, err
	}

	// ensure that resource versions and observed generation do not interfere
	// with calculating equality
	desiredResource.SetResourceVersion(actualResource.GetResourceVersion())
	desiredResource.SetGeneration(actualResource.GetGeneration())

	// ensure that a current cluster-scoped resource is not evaluated against
	// a manifest which may include a namespace
	if actualResource.GetNamespace() == "" {
		desiredResource.SetNamespace(actualResource.GetNamespace())
	}

	// merge the overrides from the desired resource into the actual resource
	mergo.Merge(
		&mergedResource.Object,
		desiredResource.Object,
		mergo.WithOverride,
		mergo.WithSliceDeepCopy,
	)

	// calculate the actual differences
	diffOptions := []patch.CalculateOption{
		reconciler.IgnoreManagedFields(),
		patch.IgnoreStatusFields(),
		patch.IgnoreVolumeClaimTemplateTypeMetaAndStatus(),
		patch.IgnorePDBSelector(),
	}

	diffResults, err := patch.DefaultPatchMaker.Calculate(
		actualResource,
		mergedResource,
		diffOptions...,
	)
	if err != nil {
		return false, err
	}

	return diffResults.IsEmpty(), nil
}

// EqualNamespaceName will compare the namespace and name of two resource objects for equality.
func (resource *Resource) EqualNamespaceName(compared *Resource) bool {
	return (resource.Name == compared.Name) && (resource.Namespace == compared.Namespace)
}

// EqualGVK will compare the GVK of two resource objects for equality.
func (resource *Resource) EqualGVK(compared *Resource) bool {
	return resource.Group == compared.Group &&
		resource.Version == compared.Version &&
		resource.Kind == compared.Kind
}

// GetObject returns an object based on an input object, and a destination
// object.
func GetObject(source *Resource) error {
	namespacedName := types.NamespacedName{
		Name:      source.Name,
		Namespace: source.Namespace,
	}

	if err := source.Client.Get(
		source.Context,
		namespacedName,
		source.resourceChecker.GetParent(),
	); err != nil {
		// if we have anything but an IsNotFound error, return the error
		if !errors.IsNotFound(err) {
			return fmt.Errorf(
				"unable to fetch resource of kind: [%s] in namespaced name: [%v]; %w",
				source.Kind,
				namespacedName,
				err,
			)
		}
	}

	return nil
}
