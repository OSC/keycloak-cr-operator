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

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	keycloakv1alpha1 "github.com/OSC/keycloak-cr-operator/api/v1alpha1"
	"github.com/OSC/keycloak-cr-operator/internal/models"
)

// nolint:unused
// log is for logging in this package.
var keycloakclientlog = logf.Log.WithName("keycloakclient-resource")

// SetupKeycloakClientWebhookWithManager registers the webhook for KeycloakClient in the manager.
func SetupKeycloakClientWebhookWithManager(mgr ctrl.Manager, keycloakConfig *models.KeycloakConfig) error {
	return ctrl.NewWebhookManagedBy(mgr, &keycloakv1alpha1.KeycloakClient{}).
		WithValidator(&KeycloakClientCustomValidator{keycloakConfig: keycloakConfig}).
		WithDefaulter(&KeycloakClientCustomDefaulter{keycloakConfig: keycloakConfig}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-keycloak-osc-edu-v1alpha1-keycloakclient,mutating=true,failurePolicy=fail,sideEffects=None,groups=keycloak.osc.edu,resources=keycloakclients,verbs=create;update,versions=v1alpha1,name=mkeycloakclient-v1alpha1.kb.io,admissionReviewVersions=v1

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

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: If you want to customise the 'path', use the flags '--defaulting-path' or '--validation-path'.
// +kubebuilder:webhook:path=/validate-keycloak-osc-edu-v1alpha1-keycloakclient,mutating=false,failurePolicy=fail,sideEffects=None,groups=keycloak.osc.edu,resources=keycloakclients,verbs=create;update,versions=v1alpha1,name=vkeycloakclient-v1alpha1.kb.io,admissionReviewVersions=v1

// KeycloakClientCustomValidator struct is responsible for validating the KeycloakClient resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type KeycloakClientCustomValidator struct {
	keycloakConfig *models.KeycloakConfig
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type KeycloakClient.
func (v *KeycloakClientCustomValidator) ValidateCreate(_ context.Context, obj *keycloakv1alpha1.KeycloakClient) (admission.Warnings, error) {
	keycloakclientlog.Info("Validation for KeycloakClient upon creation", "name", obj.GetName(), "namespace", obj.GetNamespace())

	// TODO(user): fill in your validation logic upon object creation.

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type KeycloakClient.
func (v *KeycloakClientCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj *keycloakv1alpha1.KeycloakClient) (admission.Warnings, error) {
	keycloakclientlog.Info("Validation for KeycloakClient upon update", "name", newObj.GetName(), "namespace", newObj.GetNamespace())

	// TODO(user): fill in your validation logic upon object update.

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type KeycloakClient.
func (v *KeycloakClientCustomValidator) ValidateDelete(_ context.Context, obj *keycloakv1alpha1.KeycloakClient) (admission.Warnings, error) {
	keycloakclientlog.Info("Validation for KeycloakClient upon deletion", "name", obj.GetName(), "namespace", obj.GetNamespace())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
