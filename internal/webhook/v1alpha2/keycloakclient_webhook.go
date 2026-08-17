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

package v1alpha2

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

	keycloakv1alpha2 "github.com/OSC/keycloak-cr-operator/api/v1alpha2"
	"github.com/OSC/keycloak-cr-operator/internal/models"
	"github.com/stoewer/go-strcase"
)

var (
	keycloakclientlog = logf.Log.WithName("keycloakclient-resource")
	defaultEnvVarKeys = true
)

// SetupKeycloakClientWebhookWithManager registers the webhook for KeycloakClient in the manager.
func SetupKeycloakClientWebhookWithManager(mgr ctrl.Manager, keycloakConfig *models.KeycloakConfig) error {
	return ctrl.NewWebhookManagedBy(mgr, &keycloakv1alpha2.KeycloakClient{}).
		WithValidator(&KeycloakClientCustomValidator{keycloakConfig: keycloakConfig}).
		WithDefaulter(&KeycloakClientCustomDefaulter{keycloakConfig: keycloakConfig}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-keycloak-osc-edu-v1alpha2-keycloakclient,mutating=true,failurePolicy=fail,sideEffects=None,groups=keycloak.osc.edu,resources=keycloakclients,verbs=create;update;delete,versions=v1alpha2,name=mkeycloakclient-v1alpha2.kb.io,admissionReviewVersions=v1,servicePort=9443

// KeycloakClientCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind KeycloakClient when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type KeycloakClientCustomDefaulter struct {
	keycloakConfig *models.KeycloakConfig
}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind KeycloakClient.
func (d *KeycloakClientCustomDefaulter) Default(_ context.Context, obj *keycloakv1alpha2.KeycloakClient) error {
	keycloakclientlog.Info("Defaulting for KeycloakClient v1alpha2", "name", obj.GetName(), "namespace", obj.GetNamespace())

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

	if d.keycloakConfig.ClientIDRequired != nil {
		requiredClientID, err := keycloakv1alpha2.RequiredClientID(d.keycloakConfig, obj)
		if err != nil {
			keycloakclientlog.Error(err, "Failed to get required ClientID")
			return err
		}
		if requiredClientID != "" {
			obj.Spec.ClientID = &requiredClientID
		}
	}

	// Set default Realm if not set
	if obj.Spec.Realm == nil || *obj.Spec.Realm == "" {
		defaultRealm := d.keycloakConfig.DefaultRealm
		obj.Spec.Realm = &defaultRealm
	}

	d.defaultSecret(obj)
	d.defaultConfigMap(obj)

	// Apply defaulting to ProtocolMappers
	if obj.Spec.ProtocolMappers != nil {
		// Create a new slice with default values applied
		for _, mapper := range obj.Spec.ProtocolMappers {
			// Protocol defaults to "openid-connect"
			if mapper.Protocol == nil {
				protocol := "openid-connect"
				mapper.Protocol = &protocol
			}

			// IDTokenClaim defaults to true
			if mapper.IDTokenClaim == nil {
				idTokenClaim := true
				mapper.IDTokenClaim = &idTokenClaim
			}

			// AccessTokenClaim defaults to true
			if mapper.AccessTokenClaim == nil {
				accessTokenClaim := true
				mapper.AccessTokenClaim = &accessTokenClaim
			}

			// ConsentRequired defaults to false
			if mapper.ConsentRequired == nil {
				consentRequired := false
				mapper.ConsentRequired = &consentRequired
			}
		}
	}

	return nil
}

func (d *KeycloakClientCustomDefaulter) defaultSecret(obj *keycloakv1alpha2.KeycloakClient) {
	if obj.Spec.Secret == nil {
		obj.Spec.Secret = &keycloakv1alpha2.KeycloakClientSecret{}
	}
	if obj.Spec.Secret.Name == nil || *obj.Spec.Secret.Name == "" {
		defaultSecretName := obj.SecretName()
		obj.Spec.Secret.Name = &defaultSecretName
	}
	if obj.Spec.Secret.ClientSecretKey == nil || *obj.Spec.Secret.ClientSecretKey == "" {
		defaultClientSecretKey := obj.SecretClientSecretKey()
		obj.Spec.Secret.ClientSecretKey = &defaultClientSecretKey
	}
	if obj.Spec.Secret.ClientIdKey == nil || *obj.Spec.Secret.ClientIdKey == "" {
		defaultClientIdKey := obj.SecretClientIdKey()
		obj.Spec.Secret.ClientIdKey = &defaultClientIdKey
	}
	if obj.Spec.Secret.IssuerUrlKey == nil || *obj.Spec.Secret.IssuerUrlKey == "" {
		defaultIssuerUrlKey := obj.SecretIssuerUrlKey()
		obj.Spec.Secret.IssuerUrlKey = &defaultIssuerUrlKey
	}
	if obj.Spec.Secret.CookieSecretKey == nil || *obj.Spec.Secret.CookieSecretKey == "" {
		defaultCookieSecretKey := obj.SecretCookieSecretKey()
		obj.Spec.Secret.CookieSecretKey = &defaultCookieSecretKey
	}
	if obj.Spec.Secret.Create == nil {
		create := true
		obj.Spec.Secret.Create = &create
	}
	if obj.Spec.Secret.EnvVarKeys == nil {
		envVarKeys := obj.SecretEnvVarKeys()
		obj.Spec.Secret.EnvVarKeys = &envVarKeys
	}
}

func (d *KeycloakClientCustomDefaulter) defaultConfigMap(obj *keycloakv1alpha2.KeycloakClient) {
	if obj.Spec.ConfigMap == nil {
		obj.Spec.ConfigMap = &keycloakv1alpha2.KeycloakClientConfigMap{}
	}
	if obj.Spec.ConfigMap.Name == nil || *obj.Spec.ConfigMap.Name == "" {
		defaultConfigMapName := obj.ConfigMapName()
		obj.Spec.ConfigMap.Name = &defaultConfigMapName
	}
	if obj.Spec.ConfigMap.EnvVarKeys == nil {
		obj.Spec.ConfigMap.EnvVarKeys = &defaultEnvVarKeys
	}
	if obj.Spec.ConfigMap.KeycloakUrlKey == nil || *obj.Spec.ConfigMap.KeycloakUrlKey == "" {
		defaultKeycloakUrlKey := obj.ConfigMapKeycloakUrlKey()
		obj.Spec.ConfigMap.KeycloakUrlKey = &defaultKeycloakUrlKey
	}
	if obj.Spec.ConfigMap.KeycloakHostKey == nil || *obj.Spec.ConfigMap.KeycloakHostKey == "" {
		defaultKeycloakHostKey := obj.ConfigMapKeycloakHostKey()
		obj.Spec.ConfigMap.KeycloakHostKey = &defaultKeycloakHostKey
	}
	if obj.Spec.ConfigMap.IssuerUrlKey == nil || *obj.Spec.ConfigMap.IssuerUrlKey == "" {
		defaultIssuerUrlKey := obj.ConfigMapIssuerUrlKey()
		obj.Spec.ConfigMap.IssuerUrlKey = &defaultIssuerUrlKey
	}
	if obj.Spec.ConfigMap.ProviderUrlKey == nil || *obj.Spec.ConfigMap.ProviderUrlKey == "" {
		defaultProviderUrlKey := obj.ConfigMapProviderUrlKey()
		obj.Spec.ConfigMap.ProviderUrlKey = &defaultProviderUrlKey
	}
}

// +kubebuilder:webhook:path=/validate-keycloak-osc-edu-v1alpha2-keycloakclient,mutating=false,failurePolicy=fail,sideEffects=None,groups=keycloak.osc.edu,resources=keycloakclients,verbs=create;update;delete,versions=v1alpha2,name=vkeycloakclient-v1alpha2.kb.io,admissionReviewVersions=v1,servicePort=9443

// KeycloakClientCustomValidator struct is responsible for validating the KeycloakClient resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type KeycloakClientCustomValidator struct {
	keycloakConfig *models.KeycloakConfig
}

// validateKeycloakClient validates a KeycloakClient resource based on the specified rules.
func (v *KeycloakClientCustomValidator) validateKeycloakClient(obj *keycloakv1alpha2.KeycloakClient) error {
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
		requiredClientID, err := keycloakv1alpha2.RequiredClientID(v.keycloakConfig, obj)
		if err != nil {
			keycloakclientlog.Error(err, "Failed to execute ClientIDRequired template during validation")
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "clientID"), *obj.Spec.ClientID, fmt.Sprintf("failed to apply ClientIDRequired template: %s", err)))
		} else {
			if *obj.Spec.ClientID != requiredClientID {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "clientID"), *obj.Spec.ClientID, fmt.Sprintf("clientID must match the required template: %s", requiredClientID)))
			}
		}
	}

	secretErrs := v.validateSecret(obj)
	if secretErrs != nil {
		allErrs = append(allErrs, secretErrs...)
	}

	configMapErrs := v.validateConfigMap(obj)
	if configMapErrs != nil {
		allErrs = append(allErrs, configMapErrs...)
	}

	protocolMapperErrs := v.validateProtocolMappers(obj)
	if protocolMapperErrs != nil {
		allErrs = append(allErrs, protocolMapperErrs...)
	}

	if len(allErrs) > 0 {
		return errors.NewInvalid(keycloakv1alpha2.GroupVersion.WithKind("KeycloakClient").GroupKind(), obj.Name, allErrs)
	}

	return nil
}

func (v *KeycloakClientCustomValidator) validateSecret(obj *keycloakv1alpha2.KeycloakClient) field.ErrorList {
	var allErrs field.ErrorList
	if obj.Spec.Secret == nil {
		allErrs = append(allErrs, field.Required(field.NewPath("spec", "secret"), "secret must be set"))
	} else {
		if obj.Spec.Secret.Name == nil || *obj.Spec.Secret.Name == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "secret", "name"), "secret name must be set"))
		}
		if obj.Spec.Secret.ClientSecretKey == nil || *obj.Spec.Secret.ClientSecretKey == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "secret", "clientSecretKey"), "secret clientSecretKey must be set"))
		} else if obj.SecretEnvVarKeys() {
			envClientSecretKey := strcase.UpperSnakeCase(*obj.Spec.Secret.ClientSecretKey)
			if envClientSecretKey != *obj.Spec.Secret.ClientSecretKey {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "secret", "clientSecretKey"), obj.Spec.Secret.ClientSecretKey, fmt.Sprintf("secret clientSecretKey must be upper snake case when envVarKeys is true, expected: %s", envClientSecretKey)))
			}
		}
		if obj.Spec.Secret.ClientIdKey == nil || *obj.Spec.Secret.ClientIdKey == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "secret", "clientIdKey"), "secret clientIdKey must be set"))
		} else if obj.SecretEnvVarKeys() {
			envClientIdKey := strcase.UpperSnakeCase(*obj.Spec.Secret.ClientIdKey)
			if envClientIdKey != *obj.Spec.Secret.ClientIdKey {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "secret", "clientIdKey"), obj.Spec.Secret.ClientIdKey, fmt.Sprintf("secret clientIdKey must be upper snake case when envVarKeys is true, expected: %s", envClientIdKey)))
			}
		}
		if obj.Spec.Secret.IssuerUrlKey == nil || *obj.Spec.Secret.IssuerUrlKey == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "secret", "issuerUrlKey"), "secret issuerUrlKey must be set"))
		} else if obj.SecretEnvVarKeys() {
			envIssuerUrlKey := strcase.UpperSnakeCase(*obj.Spec.Secret.IssuerUrlKey)
			if envIssuerUrlKey != *obj.Spec.Secret.IssuerUrlKey {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "secret", "issuerUrlKey"), obj.Spec.Secret.IssuerUrlKey, fmt.Sprintf("secret issuerUrlKey must be upper snake case when envVarKeys is true, expected: %s", envIssuerUrlKey)))
			}
		}
		if obj.Spec.Secret.CookieSecretKey == nil || *obj.Spec.Secret.CookieSecretKey == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "secret", "cookieSecretKey"), "secret cookieSecretKey must be set"))
		} else if obj.SecretEnvVarKeys() {
			envCookieSecretKey := strcase.UpperSnakeCase(*obj.Spec.Secret.CookieSecretKey)
			if envCookieSecretKey != *obj.Spec.Secret.CookieSecretKey {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "secret", "cookieSecretKey"), obj.Spec.Secret.CookieSecretKey, fmt.Sprintf("secret cookieSecretKey must be upper snake case when envVarKeys is true, expected: %s", envCookieSecretKey)))
			}
		}
		if obj.Spec.Secret.Create == nil {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "secret", "create"), "secret create must be set"))
		}
		if obj.Spec.Secret.EnvVarKeys == nil {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "secret", "envVarKeys"), "secret envVarKeys must be set"))
		}
	}
	return allErrs
}

func (v *KeycloakClientCustomValidator) validateConfigMap(obj *keycloakv1alpha2.KeycloakClient) field.ErrorList {
	var allErrs field.ErrorList
	if obj.Spec.ConfigMap == nil {
		allErrs = append(allErrs, field.Required(field.NewPath("spec", "configMap"), "configMap data must be provided"))
	} else {
		if obj.Spec.ConfigMap.Name == nil || *obj.Spec.ConfigMap.Name == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "configMap", "name"), "configMap.name must be set"))
		}
		if obj.Spec.ConfigMap.EnvVarKeys == nil {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "configMap", "envVarKeys"), "configMap.envVarKeys must be set"))
		}
		if obj.Spec.ConfigMap.KeycloakUrlKey == nil || *obj.Spec.ConfigMap.KeycloakUrlKey == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "configMap", "keycloakUrlKey"), "configMap keycloakUrlKey must be set"))
		} else if obj.ConfigMapEnvVarKeys() {
			envKeycloakUrlKey := strcase.UpperSnakeCase(*obj.Spec.ConfigMap.KeycloakUrlKey)
			if envKeycloakUrlKey != *obj.Spec.ConfigMap.KeycloakUrlKey {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "configMap", "keycloakUrlKey"), obj.Spec.ConfigMap.KeycloakUrlKey, fmt.Sprintf("configMap keycloakUrlKey must be upper snake case when envVarKeys is true, expected: %s", envKeycloakUrlKey)))
			}
		}
		if obj.Spec.ConfigMap.KeycloakHostKey == nil || *obj.Spec.ConfigMap.KeycloakHostKey == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "configMap", "keycloakHostKey"), "configMap keycloakHostKey must be set"))
		} else if obj.ConfigMapEnvVarKeys() {
			envKeycloakHostKey := strcase.UpperSnakeCase(*obj.Spec.ConfigMap.KeycloakHostKey)
			if envKeycloakHostKey != *obj.Spec.ConfigMap.KeycloakHostKey {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "configMap", "keycloakHostKey"), obj.Spec.ConfigMap.KeycloakHostKey, fmt.Sprintf("configMap keycloakHostKey must be upper snake case when envVarKeys is true, expected: %s", envKeycloakHostKey)))
			}
		}
		if obj.Spec.ConfigMap.IssuerUrlKey == nil || *obj.Spec.ConfigMap.IssuerUrlKey == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "configMap", "issuerUrlKey"), "configMap issuerUrlKey must be set"))
		} else if obj.ConfigMapEnvVarKeys() {
			envIssuerUrlKey := strcase.UpperSnakeCase(*obj.Spec.ConfigMap.IssuerUrlKey)
			if envIssuerUrlKey != *obj.Spec.ConfigMap.IssuerUrlKey {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "configMap", "issuerUrlKey"), obj.Spec.ConfigMap.IssuerUrlKey, fmt.Sprintf("configMap issuerUrlKey must be upper snake case when envVarKeys is true, expected: %s", envIssuerUrlKey)))
			}
		}
		if obj.Spec.ConfigMap.ProviderUrlKey == nil || *obj.Spec.ConfigMap.ProviderUrlKey == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "configMap", "providerUrlKey"), "configMap providerUrlKey must be set"))
		} else if obj.ConfigMapEnvVarKeys() {
			envProviderUrlKey := strcase.UpperSnakeCase(*obj.Spec.ConfigMap.ProviderUrlKey)
			if envProviderUrlKey != *obj.Spec.ConfigMap.ProviderUrlKey {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "configMap", "providerUrlKey"), obj.Spec.ConfigMap.ProviderUrlKey, fmt.Sprintf("configMap providerUrlKey must be upper snake case when envVarKeys is true, expected: %s", envProviderUrlKey)))
			}
		}
	}
	return allErrs
}

// validateProtocolMappers validates the ProtocolMappers field of KeycloakClient
func (v *KeycloakClientCustomValidator) validateProtocolMappers(obj *keycloakv1alpha2.KeycloakClient) field.ErrorList {
	var allErrs field.ErrorList

	// Validate ProtocolMappers if they exist
	if obj.Spec.ProtocolMappers != nil {
		for i, mapper := range obj.Spec.ProtocolMappers {
			path := field.NewPath("spec", "protocolMappers").Index(i)

			// Name is required
			if mapper.Name == nil || *mapper.Name == "" {
				allErrs = append(allErrs, field.Required(path.Child("name"), "name must be set"))
			}

			// Protocol is required
			if mapper.Protocol == nil || *mapper.Protocol == "" {
				allErrs = append(allErrs, field.Required(path.Child("protocol"), "protocol must be set"))
			}

			// Type is required
			if mapper.Type == nil || *mapper.Type == "" {
				allErrs = append(allErrs, field.Required(path.Child("type"), "type must be set"))
			}

			// If type is oidc-audience-mapper, IncludedClientAudience is required
			if mapper.Type != nil && *mapper.Type == "oidc-audience-mapper" {
				if mapper.IncludedClientAudience == nil || *mapper.IncludedClientAudience == "" {
					allErrs = append(allErrs, field.Required(path.Child("includedClientAudience"), "includedClientAudience must be set when type is oidc-audience-mapper"))
				}
			}
		}
	}

	return allErrs
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type KeycloakClient.
func (v *KeycloakClientCustomValidator) ValidateCreate(_ context.Context, obj *keycloakv1alpha2.KeycloakClient) (admission.Warnings, error) {
	keycloakclientlog.Info("Validation for KeycloakClient v1alpha2 upon creation", "name", obj.GetName(), "namespace", obj.GetNamespace())

	// Validate the KeycloakClient resource
	if err := v.validateKeycloakClient(obj); err != nil {
		return nil, err
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type KeycloakClient.
func (v *KeycloakClientCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj *keycloakv1alpha2.KeycloakClient) (admission.Warnings, error) {
	keycloakclientlog.Info("Validation for KeycloakClient v1alpha2 upon update", "name", newObj.GetName(), "namespace", newObj.GetNamespace())

	// Validate the KeycloakClient resource
	if err := v.validateKeycloakClient(newObj); err != nil {
		return nil, err
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type KeycloakClient.
func (v *KeycloakClientCustomValidator) ValidateDelete(_ context.Context, obj *keycloakv1alpha2.KeycloakClient) (admission.Warnings, error) {
	keycloakclientlog.Info("Validation for KeycloakClient v1alpha2 upon deletion", "name", obj.GetName(), "namespace", obj.GetNamespace())

	// For deletion, we don't perform any validation as the resource is being deleted
	// but we can add validation logic here if needed in the future

	return nil, nil
}
