/*
Copyright 2026 Ohio Supercomputer Center.

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

package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	keycloakv1alpha1 "github.com/OSC/keycloak-cr-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	configmapChecksumAnnotation = "keycloak.osc.edu/configmap-checksum"
	secretChecksumAnnotation    = "keycloak.osc.edu/secret-checksum"
)

// computeChecksum computes a checksum of a map with sorted keys
// The input is a Kubernetes object (either a Secret or ConfigMap)
// For Secrets and ConfigMaps, it accesses the Data field to compute the checksum
// Returns the computed checksum as a hex-encoded string
func computeChecksum(obj client.Object) (string, error) {
	// We need to handle different types of Kubernetes objects
	// For now, we'll handle Secrets and ConfigMaps which have Data field

	// Convert to map[string]string for processing
	dataMap := make(map[string]string)

	switch v := obj.(type) {
	case *corev1.Secret:
		for key, value := range v.StringData {
			dataMap[key] = value
		}
	case *corev1.ConfigMap:
		for key, value := range v.Data {
			dataMap[key] = value
		}
	default:
		// If we don't recognize the object type, return error
		return "", fmt.Errorf("unsupported object type for checksum: %T", obj)
	}

	// Sort keys for consistent ordering
	keys := make([]string, 0, len(dataMap))
	for k := range dataMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Calculate SHA256 checksum using sha256.New() and write each sorted item
	hasher := sha256.New()
	for _, key := range keys {
		// Format: key:value|
		item := key + ":" + dataMap[key] + "|"
		hasher.Write([]byte(item))
	}

	// Return hex-encoded checksum
	checksum := hasher.Sum(nil)
	return hex.EncodeToString(checksum), nil
}

func (r *KeycloakClientReconciler) updateChecksum(ctx context.Context, obj client.Object, keycloakClient *keycloakv1alpha1.KeycloakClient) error {
	// Check if ChecksumRef is configured
	if keycloakClient.Spec.ChecksumRef == nil {
		return nil
	}

	// Get the resource type from ChecksumRef
	resourceKind := keycloakClient.Spec.ChecksumRef.Kind
	if resourceKind == nil {
		return nil
	}

	// Get the resource name from ChecksumRef
	resourceName := keycloakClient.Spec.ChecksumRef.Name
	if resourceName == nil {
		return nil
	}

	// Determine the annotation key based on object type (not resource type)
	var annotationKey string
	switch obj.(type) {
	case *corev1.ConfigMap:
		annotationKey = configmapChecksumAnnotation
	case *corev1.Secret:
		annotationKey = secretChecksumAnnotation
	default:
		// For unsupported object types, return error
		return fmt.Errorf("unsupported object type for checksum: %T", obj)
	}

	// Get the target object using a single variable approach
	var targetObj client.Object
	switch *resourceKind {
	case "Deployment":
		deployment := &appsv1.Deployment{}
		targetObj = deployment
	case "StatefulSet":
		statefulSet := &appsv1.StatefulSet{}
		targetObj = statefulSet
	default:
		return nil
	}

	err := r.Get(ctx, types.NamespacedName{Name: *resourceName, Namespace: keycloakClient.Namespace}, targetObj)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object doesn't exist, do nothing
			return nil
		}
		return err
	}

	// Compute the checksum for the object
	checksum, err := computeChecksum(obj)
	if err != nil {
		return err
	}

	// Prepare patch for the pod template annotations
	patch := client.MergeFrom(targetObj.DeepCopyObject().(client.Object))

	// Update the annotations in the pod template
	switch *resourceKind {
	case "Deployment":
		deployment := targetObj.(*appsv1.Deployment)
		if deployment.Spec.Template.Annotations == nil {
			deployment.Spec.Template.Annotations = make(map[string]string)
		}
		deployment.Spec.Template.Annotations[annotationKey] = checksum
	case "StatefulSet":
		statefulSet := targetObj.(*appsv1.StatefulSet)
		if statefulSet.Spec.Template.Annotations == nil {
			statefulSet.Spec.Template.Annotations = make(map[string]string)
		}
		statefulSet.Spec.Template.Annotations[annotationKey] = checksum
	}

	// Apply the patch
	return r.Patch(ctx, targetObj, patch)
}
