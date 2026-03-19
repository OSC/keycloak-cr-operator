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
	"bytes"
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
	defaultEnvVarKeys = true
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

type KeycloakClientData struct {
	Obj    keycloakv1alpha1.KeycloakClient
	Config models.KeycloakConfig
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

	// Apply ClientIDRequired template if set
	if d.keycloakConfig.ClientIDRequired != nil {
		data := KeycloakClientData{
			Obj:    *obj,
			Config: *d.keycloakConfig,
		}

		var buf bytes.Buffer
		err := d.keycloakConfig.ClientIDRequired.Execute(&buf, data)
		if err != nil {
			// Log error but don't fail - fall back to default behavior
			keycloakclientlog.Error(err, "Failed to execute ClientIDRequired template")
		} else {
			// Override the ClientID with the templated value
			clientID := buf.String()
			obj.Spec.ClientID = &clientID
		}
	}

	// Set default Realm if not set
	if obj.Spec.Realm == nil || *obj.Spec.Realm == "" {
		defaultRealm := d.keycloakConfig.DefaultRealm
		obj.Spec.Realm = &defaultRealm
	}

	if obj.Spec.ClientAuthenticatorType != nil && *obj.Spec.ClientAuthenticatorType == clientSecretType &&
		obj.Spec.PublicClient != nil && !*obj.Spec.PublicClient {
		if obj.Spec.ClientSecretRef == nil {
			obj.Spec.ClientSecretRef = &keycloakv1alpha1.KeycloakClientSecret{}
		}
		if obj.Spec.ClientSecretRef.Name == "" {
			obj.Spec.ClientSecretRef.Name = fmt.Sprintf("%s-secret", obj.Name)
		}
		if obj.Spec.ClientSecretRef.Key == "" {
			obj.Spec.ClientSecretRef.Key = "CLIENT_SECRET"
		}
		if obj.Spec.ClientSecretRef.Create == nil {
			create := true
			obj.Spec.ClientSecretRef.Create = &create
		}
		// Set default EnvVarKeys to true if not set
		if obj.Spec.ClientSecretRef.EnvVarKeys == nil {
			obj.Spec.ClientSecretRef.EnvVarKeys = &defaultEnvVarKeys
		}
	}

	// Handle ConfigMap structure
	defaultConfigMapName := fmt.Sprintf("%s-config", obj.Name)
	if obj.Spec.ConfigMap != nil {
		// Set default ConfigMap.Name if not set
		if obj.Spec.ConfigMap.Name == nil || *obj.Spec.ConfigMap.Name == "" {
			obj.Spec.ConfigMap.Name = &defaultConfigMapName
		}
		// Set default EnvVarKeys to true if not set
		if obj.Spec.ConfigMap.EnvVarKeys == nil {
			obj.Spec.ConfigMap.EnvVarKeys = &defaultEnvVarKeys
		}
	} else {
		// If ConfigMap is nil, create a new one with defaults
		obj.Spec.ConfigMap = &keycloakv1alpha1.KeycloakClientConfigMap{
			Name:       &defaultConfigMapName,
			EnvVarKeys: &defaultEnvVarKeys,
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
	if v.keycloakConfig.ClientIDPrefix != "" && v.keycloakConfig.ClientIDRequired == nil {
		if obj.Spec.ClientID != nil && *obj.Spec.ClientID != "" {
			if !strings.HasPrefix(*obj.Spec.ClientID, v.keycloakConfig.ClientIDPrefix) {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "clientID"), *obj.Spec.ClientID, fmt.Sprintf("clientID must begin with the prefix %s", v.keycloakConfig.ClientIDPrefix)))
			}
		}
	}

	// If ClientIDRequired is set, the ClientID must match the required template
	if v.keycloakConfig.ClientIDRequired != nil && obj.Spec.ClientID != nil && *obj.Spec.ClientID != "" {
		// Create a template data structure to evaluate the template
		data := KeycloakClientData{
			Obj:    *obj,
			Config: *v.keycloakConfig,
		}

		// Execute the template to get the expected client ID
		var buf bytes.Buffer
		err := v.keycloakConfig.ClientIDRequired.Execute(&buf, data)
		if err != nil {
			keycloakclientlog.Error(err, "Failed to execute ClientIDRequired template during validation")
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "clientID"), *obj.Spec.ClientID, fmt.Sprintf("failed to apply ClientIDRequired template: %s", err)))
		} else {
			expectedClientID := buf.String()
			if *obj.Spec.ClientID != expectedClientID {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "clientID"), *obj.Spec.ClientID, fmt.Sprintf("clientID must match the required template: %s", expectedClientID)))
			}
		}
	}

	// Validate ClientSecretRef is present if ClientAuthenticatorType=client-secret and Public=false
	clientSecretErrs := v.validateClientSecretRef(obj)
	if clientSecretErrs != nil {
		allErrs = append(allErrs, clientSecretErrs...)
	}

	// Validate ConfigMap structure
	if obj.Spec.ConfigMap != nil {
		if obj.Spec.ConfigMap.Name == nil || *obj.Spec.ConfigMap.Name == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "configMap", "name"), "configMap.name must be set"))
		}
		if obj.Spec.ConfigMap.EnvVarKeys == nil {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "configMap", "envVarKeys"), "configMap.envVarKeys must be set"))
		}
	} else {
		allErrs = append(allErrs, field.Required(field.NewPath("spec", "configMap"), "configMap data must be provided"))
	}

	if len(allErrs) > 0 {
		return errors.NewInvalid(keycloakv1alpha1.GroupVersion.WithKind("KeycloakClient").GroupKind(), obj.Name, allErrs)
	}

	return nil
}

func (v *KeycloakClientCustomValidator) validateClientSecretRef(obj *keycloakv1alpha1.KeycloakClient) field.ErrorList {
	var allErrs field.ErrorList
	if (obj.Spec.ClientAuthenticatorType != nil && *obj.Spec.ClientAuthenticatorType == clientSecretType) && (obj.Spec.PublicClient == nil || !*obj.Spec.PublicClient) {
		if obj.Spec.ClientSecretRef == nil {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "clientSecretRef"), fmt.Sprintf("clientSecretRef must be set when clientAuthenticatorType is %s and public is false", clientSecretType)))
		} else {
			if obj.Spec.ClientSecretRef.Name == "" {
				allErrs = append(allErrs, field.Required(field.NewPath("spec", "clientSecretRef", "name"), fmt.Sprintf("clientSecretRef name must be set when clientAuthenticatorType is %s and public is false", clientSecretType)))
			}
			if obj.Spec.ClientSecretRef.Key == "" {
				allErrs = append(allErrs, field.Required(field.NewPath("spec", "clientSecretRef", "key"), fmt.Sprintf("clientSecretRef key must be set when clientAuthenticatorType is %s and public is false", clientSecretType)))
			}
			if obj.Spec.ClientSecretRef.Create == nil {
				allErrs = append(allErrs, field.Required(field.NewPath("spec", "clientSecretRef", "create"), fmt.Sprintf("clientSecretRef create must be set when clientAuthenticatorType is %s and public is false", clientSecretType)))
			}
			// Validate EnvVarKeys is set for ClientSecretRef
			if obj.Spec.ClientSecretRef.EnvVarKeys == nil {
				allErrs = append(allErrs, field.Required(field.NewPath("spec", "clientSecretRef", "envVarKeys"), fmt.Sprintf("clientSecretRef envVarKeys must be set when clientAuthenticatorType is %s and public is false", clientSecretType)))
			}
		}
	}
	return allErrs
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
