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
	"errors"
	"fmt"
	"maps"
	"sort"

	keycloakv1alpha1 "github.com/OSC/keycloak-cr-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	configmapChecksumAnnotation = "keycloak.osc.edu/configmap-checksum"
	secretChecksumAnnotation    = "keycloak.osc.edu/secret-checksum"
)

func computeChecksum(obj client.Object) (string, error) {
	dataMap := make(map[string]string)

	switch v := obj.(type) {
	case *corev1.Secret:
		for key, value := range v.Data {
			dataMap[key] = string(value)
		}
	case *corev1.ConfigMap:
		maps.Copy(dataMap, v.Data)
	default:
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

	checksum := hasher.Sum(nil)
	return hex.EncodeToString(checksum), nil
}

func (r *KeycloakClientReconciler) updateChecksum(ctx context.Context, obj client.Object, keycloakClient *keycloakv1alpha1.KeycloakClient) error {
	// log := logf.FromContext(ctx)

	var annotationKey string
	switch obj.(type) {
	case *corev1.ConfigMap:
		annotationKey = configmapChecksumAnnotation
	case *corev1.Secret:
		annotationKey = secretChecksumAnnotation
	default:
		return fmt.Errorf("unsupported object type for checksum: %T", obj)
	}

	checksum, err := computeChecksum(obj)
	if err != nil {
		return err
	}

	errs := r.updateChecksumDeployment(ctx, checksum, annotationKey, keycloakClient)
	if errs != nil {
		return fmt.Errorf("failures to update deployment checksum: %s", errors.Join(errs...))
	}

	errs = r.updateChecksumStatefulSet(ctx, checksum, annotationKey, keycloakClient)
	if errs != nil {
		return fmt.Errorf("failures to update statefulset checksum: %s", errors.Join(errs...))
	}

	return nil
}

func (r *KeycloakClientReconciler) updateChecksumDeployment(ctx context.Context, checksum string, annotationKey string, keycloakClient *keycloakv1alpha1.KeycloakClient) []error {
	log := logf.FromContext(ctx)

	var errs []error
	deploymentList := &appsv1.DeploymentList{}
	err := r.List(ctx, deploymentList, client.InNamespace(keycloakClient.Namespace), client.MatchingLabels{keycloakClientMatchLabel: keycloakClient.Name})
	if err != nil {
		return []error{err}
	}
	if len(deploymentList.Items) == 0 {
		log.V(1).Info("No deployment found for checksum update", "name", keycloakClient.Name, "namespace", keycloakClient.Namespace)
		return nil
	}
	for _, deployment := range deploymentList.Items {
		patch := client.MergeFrom(deployment.DeepCopy())
		if deployment.Spec.Template.Annotations == nil {
			deployment.Spec.Template.Annotations = make(map[string]string)
		}
		if currentChecksum, ok := deployment.Spec.Template.Annotations[annotationKey]; ok {
			if currentChecksum == checksum {
				log.V(1).Info("Checksum matches, skip", "name", keycloakClient.Name, "namespace", keycloakClient.Namespace,
					"resource", deployment.Name, "annotation", annotationKey, "checksum", checksum)
				continue
			}
		}
		deployment.Spec.Template.Annotations[annotationKey] = checksum
		log.V(1).Info("Patch Deployment with checksum", "name", keycloakClient.Name, "namespace", keycloakClient.Namespace,
			"resource", deployment.Name, "annotation", annotationKey, "checksum", checksum)
		if err := r.Patch(ctx, &deployment, patch); err != nil {
			log.Error(err, "Unable to patch Deployment", "resource", deployment.Name, "namespace", keycloakClient.Namespace)
			errs = append(errs, err)
		}
	}
	return errs
}

func (r *KeycloakClientReconciler) updateChecksumStatefulSet(ctx context.Context, checksum string, annotationKey string, keycloakClient *keycloakv1alpha1.KeycloakClient) []error {
	log := logf.FromContext(ctx)

	var errs []error
	statefulsetList := &appsv1.StatefulSetList{}
	err := r.List(ctx, statefulsetList, client.InNamespace(keycloakClient.Namespace), client.MatchingLabels{keycloakClientMatchLabel: keycloakClient.Name})
	if err != nil {
		return []error{err}
	}
	if len(statefulsetList.Items) == 0 {
		log.V(1).Info("No deployment found for checksum update", "name", keycloakClient.Name, "namespace", keycloakClient.Namespace)
		return nil
	}
	for _, statefulset := range statefulsetList.Items {
		patch := client.MergeFrom(statefulset.DeepCopy())
		if statefulset.Spec.Template.Annotations == nil {
			statefulset.Spec.Template.Annotations = make(map[string]string)
		}
		if currentChecksum, ok := statefulset.Spec.Template.Annotations[annotationKey]; ok {
			if currentChecksum == checksum {
				log.V(1).Info("Checksum matches, skip", "name", keycloakClient.Name, "namespace", keycloakClient.Namespace,
					"resource", statefulset.Name, "annotation", annotationKey, "checksum", checksum)
				continue
			}
		}
		statefulset.Spec.Template.Annotations[annotationKey] = checksum
		log.V(1).Info("Patch StatefulSet with checksum", "name", keycloakClient.Name, "namespace", keycloakClient.Namespace,
			"resource", statefulset.Name, "annotation", annotationKey, "checksum", checksum)
		if err := r.Patch(ctx, &statefulset, patch); err != nil {
			log.Error(err, "Unable to patch StatefulSet", "resource", statefulset.Name, "namespace", keycloakClient.Namespace)
			errs = append(errs, err)
		}
	}
	return errs
}
