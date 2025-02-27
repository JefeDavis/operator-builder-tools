/*
	SPDX-License-Identifier: MIT
*/

package resources

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SecretKind    = "Secret"
	SecretVersion = "v1"
)

type SecretResource struct {
	Object v1.Secret
}

// NewSecretResource creates and returns a new SecretResource.
func NewSecretResource(object metav1.Object) (*SecretResource, error) {
	secret := &v1.Secret{}

	err := ToProper(secret, object)
	if err != nil {
		return nil, err
	}

	return &SecretResource{Object: *secret}, nil
}

// IsReady checks to see if a secret is ready.
func (secret *SecretResource) IsReady() (bool, error) {
	// if we have a name that is empty, we know we did not find the object
	if secret.Object.Name == "" {
		return false, nil
	}

	return true, nil
}
