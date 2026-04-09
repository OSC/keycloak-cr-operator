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
	appsv1ac "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;update;patch

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
		log.V(1).Info("No deployment found for checksum update")
		return nil
	}
	for _, deployment := range deploymentList.Items {
		// Ensure template annotations exist
		if deployment.Spec.Template.Annotations == nil {
			deployment.Spec.Template.Annotations = make(map[string]string)
		}

		// Check if checksum already matches
		if currentChecksum, ok := deployment.Spec.Template.Annotations[annotationKey]; ok {
			if currentChecksum == checksum {
				log.V(1).Info("Checksum matches, skip", "resource", deployment.Name, "annotation", annotationKey, "checksum", checksum)
				continue
			}
		}

		deployment.Spec.Template.Annotations[annotationKey] = checksum
		deploymentCopy := appsv1ac.Deployment(deployment.Name, deployment.Namespace).
			WithSpec(appsv1ac.DeploymentSpec().
				WithTemplate(corev1ac.PodTemplateSpec().
					WithAnnotations(deployment.Spec.Template.Annotations)))

		log.Info("Apply Deployment checksum with SSA", "resource", deployment.Name, "annotation", annotationKey, "checksum", checksum)

		// Use Server Side Apply with SSA field manager
		if err := r.Apply(ctx, deploymentCopy, client.FieldOwner(operatorName), client.ForceOwnership); err != nil {
			log.Error(err, "Unable to apply Deployment checksum with SSA", "resource", deployment.Name, "namespace", keycloakClient.Namespace)
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
		log.V(1).Info("No StatefulSet found for checksum update")
		return nil
	}
	for _, statefulset := range statefulsetList.Items {
		// Ensure template annotations exist
		if statefulset.Spec.Template.Annotations == nil {
			statefulset.Spec.Template.Annotations = make(map[string]string)
		}

		// Check if checksum already matches
		if currentChecksum, ok := statefulset.Spec.Template.Annotations[annotationKey]; ok {
			if currentChecksum == checksum {
				log.V(1).Info("Checksum matches, skip", "resource", statefulset.Name, "annotation", annotationKey, "checksum", checksum)
				continue
			}
		}

		statefulset.Spec.Template.Annotations[annotationKey] = checksum
		statefulsetCopy := appsv1ac.StatefulSet(statefulset.Name, statefulset.Namespace).
			WithSpec(appsv1ac.StatefulSetSpec().
				WithTemplate(corev1ac.PodTemplateSpec().
					WithAnnotations(statefulset.Spec.Template.Annotations)))

		log.Info("Apply StatefulSet checksum with SSA", "resource", statefulset.Name, "annotation", annotationKey, "checksum", checksum)

		// Use Server Side Apply with SSA field manager
		if err := r.Apply(ctx, statefulsetCopy, client.FieldOwner(operatorName), client.ForceOwnership); err != nil {
			log.Error(err, "Unable to apply StatefulSet checksum with SSA", "resource", statefulset.Name, "namespace", keycloakClient.Namespace)
			errs = append(errs, err)
		}
	}
	return errs
}
