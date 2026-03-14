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

package v1alpha1

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	keycloakv1alpha1 "github.com/OSC/keycloak-cr-operator/api/v1alpha1"
	"github.com/OSC/keycloak-cr-operator/internal/models"
)

var (
	keycloakclientlog = logf.Log.WithName("keycloakclient-resource")
	clientSecretType  = "client-secret"
)

// SetupKeycloakClientWebhookWithManager registers the webhook for KeycloakClient in the manager.
func SetupKeycloakClientWebhookWithManager(mgr ctrl.Manager, keycloakConfig *models.KeycloakConfig) error {
	return ctrl.NewWebhookManagedBy(mgr, &keycloakv1alpha1.KeycloakClient{}).
		WithValidator(&KeycloakClientCustomValidator{keycloakConfig: keycloakConfig}).
		WithDefaulter(&KeycloakClientCustomDefaulter{keycloakConfig: keycloakConfig}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-keycloak-osc-edu-v1alpha1-keycloakclient,mutating=true,failurePolicy=fail,sideEffects=None,groups=keycloak.osc.edu,resources=keycloakclients,verbs=create;update;delete,versions=v1alpha1,name=mkeycloakclient-v1alpha1.kb.io,admissionReviewVersions=v1

// KeycloakClientCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind KeycloakClient when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type KeycloakClientCustomDefaulter struct {
	keycloakConfig *models.KeycloakConfig
}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind KeycloakClient.
func (d *KeycloakClientCustomDefaulter) Default(_ context.Context, obj *keycloakv1alpha1.KeycloakClient) error {
	keycloakclientlog.Info("Defaulting for KeycloakClient", "name", obj.GetName(), "namespace", obj.GetNamespace())

	// Set default ClientID if not set
	if obj.Spec.ClientID == nil || *obj.Spec.ClientID == "" {
		clientIDPrefix := d.keycloakConfig.ClientIDPrefix
		if clientIDPrefix == "" {
			// If no prefix is set, just use namespace-name
			clientID := fmt.Sprintf("%s-%s", obj.GetNamespace(), obj.GetName())
			obj.Spec.ClientID = &clientID
		} else {
			// If prefix is set, use prefix-namespace-name
			clientID := fmt.Sprintf("%s-%s-%s", clientIDPrefix, obj.GetNamespace(), obj.GetName())
			obj.Spec.ClientID = &clientID
		}
	}

	// Set default Realm if not set
	if obj.Spec.Realm == nil || *obj.Spec.Realm == "" {
		defaultRealm := d.keycloakConfig.DefaultRealm
		obj.Spec.Realm = &defaultRealm
	}

	// Set default ID if not set
	if obj.Spec.ID == nil || *obj.Spec.ID == "" {
		if obj.Spec.ClientID != nil && *obj.Spec.ClientID != "" {
			clientID := *obj.Spec.ClientID
			obj.Spec.ID = &clientID
		}
	}

	return nil
}

// +kubebuilder:webhook:path=/validate-keycloak-osc-edu-v1alpha1-keycloakclient,mutating=false,failurePolicy=fail,sideEffects=None,groups=keycloak.osc.edu,resources=keycloakclients,verbs=create;update;delete,versions=v1alpha1,name=vkeycloakclient-v1alpha1.kb.io,admissionReviewVersions=v1

// KeycloakClientCustomValidator struct is responsible for validating the KeycloakClient resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type KeycloakClientCustomValidator struct {
	keycloakConfig *models.KeycloakConfig
}

// validateKeycloakClient validates a KeycloakClient resource based on the specified rules.
func (v *KeycloakClientCustomValidator) validateKeycloakClient(obj *keycloakv1alpha1.KeycloakClient) error {
	var allErrs field.ErrorList

	// ClientID must be set
	if obj.Spec.ClientID == nil || *obj.Spec.ClientID == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("spec", "clientID"), "clientID must be set"))
	}

	// Realm must be set
	if obj.Spec.Realm == nil || *obj.Spec.Realm == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("spec", "realm"), "realm must be set"))
	}

	// Realm matched AllowedRealms
	if obj.Spec.Realm != nil && *obj.Spec.Realm != "" && len(v.keycloakConfig.AllowedRealms) > 0 {
		if !slices.Contains(v.keycloakConfig.AllowedRealms, *obj.Spec.Realm) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "realm"), *obj.Spec.Realm, fmt.Sprintf("realm must be one of: %s", strings.Join(v.keycloakConfig.AllowedRealms, ","))))
		}
	}

	// If ClientIDPrefix is set, the ClientID must begin with the prefix from ClientIDPrefix
	if v.keycloakConfig.ClientIDPrefix != "" {
		if obj.Spec.ClientID != nil && *obj.Spec.ClientID != "" {
			if !strings.HasPrefix(*obj.Spec.ClientID, v.keycloakConfig.ClientIDPrefix) {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "clientID"), *obj.Spec.ClientID, fmt.Sprintf("clientID must begin with the prefix %s", v.keycloakConfig.ClientIDPrefix)))
			}
		}
	}

	// Validate ClientSecretRef is present if ClientAuthenticatorType=client-secret and Public=false
	if obj.Spec.ClientAuthenticatorType != nil && *obj.Spec.ClientAuthenticatorType == clientSecretType {
		if obj.Spec.PublicClient == nil || !*obj.Spec.PublicClient {
			if obj.Spec.ClientSecretRef == nil {
				allErrs = append(allErrs, field.Required(field.NewPath("spec", "clientSecretRef"), fmt.Sprintf("clientSecretRef must be set when clientAuthenticatorType is %s and public is false", clientSecretType)))
			}
		}
	}

	if len(allErrs) > 0 {
		return errors.NewInvalid(keycloakv1alpha1.GroupVersion.WithKind("KeycloakClient").GroupKind(), obj.Name, allErrs)
	}

	return nil
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type KeycloakClient.
func (v *KeycloakClientCustomValidator) ValidateCreate(_ context.Context, obj *keycloakv1alpha1.KeycloakClient) (admission.Warnings, error) {
	keycloakclientlog.Info("Validation for KeycloakClient upon creation", "name", obj.GetName(), "namespace", obj.GetNamespace())

	// Validate the KeycloakClient resource
	if err := v.validateKeycloakClient(obj); err != nil {
		return nil, err
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type KeycloakClient.
func (v *KeycloakClientCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj *keycloakv1alpha1.KeycloakClient) (admission.Warnings, error) {
	keycloakclientlog.Info("Validation for KeycloakClient upon update", "name", newObj.GetName(), "namespace", newObj.GetNamespace())

	// Validate the KeycloakClient resource
	if err := v.validateKeycloakClient(newObj); err != nil {
		return nil, err
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type KeycloakClient.
func (v *KeycloakClientCustomValidator) ValidateDelete(_ context.Context, obj *keycloakv1alpha1.KeycloakClient) (admission.Warnings, error) {
	keycloakclientlog.Info("Validation for KeycloakClient upon deletion", "name", obj.GetName(), "namespace", obj.GetNamespace())

	// For deletion, we don't perform any validation as the resource is being deleted
	// but we can add validation logic here if needed in the future

	return nil, nil
}
