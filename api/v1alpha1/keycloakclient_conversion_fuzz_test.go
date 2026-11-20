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
	"testing"

	keycloakcommon "github.com/OSC/keycloak-cr-operator/api/common"
	keycloakv1alpha2 "github.com/OSC/keycloak-cr-operator/api/v1alpha2"
	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	utilconversion "sigs.k8s.io/cluster-api/util/conversion"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	"sigs.k8s.io/randfill"
)

func customFuzzerFuncs(codecs runtimeserializer.CodecFactory) []any {
	return []any{
		func(obj *KeycloakClient, c randfill.Continue) {
			c.FillNoCustom(obj) // Fill standard fields first

			// Ignore Optional as not something this operator uses but
			// pulled in with corev1.SecretKeySelector
			if obj.Spec.ClientSecretRef != nil {
				obj.Spec.ClientSecretRef.Optional = nil
			}
			// Avoid "&nil" and "nil" differences for slice fields
			if obj.Spec.DefaultClientScopes != nil && *obj.Spec.DefaultClientScopes == nil {
				obj.Spec.DefaultClientScopes = nil
			}
			if obj.Spec.DefaultRoles != nil && *obj.Spec.DefaultRoles == nil {
				obj.Spec.DefaultRoles = nil
			}
			if obj.Spec.OptionalClientScopes != nil && *obj.Spec.OptionalClientScopes == nil {
				obj.Spec.OptionalClientScopes = nil
			}
			if obj.Spec.RedirectURIs != nil && *obj.Spec.RedirectURIs == nil {
				obj.Spec.RedirectURIs = nil
			}
			if obj.Spec.WebOrigins != nil && *obj.Spec.WebOrigins == nil {
				obj.Spec.WebOrigins = nil
			}
		},
		func(obj *keycloakv1alpha2.KeycloakClient, c randfill.Continue) {
			c.FillNoCustom(obj) // Fill standard fields first

			// Avoid "&nil" and "nil" differences for slice fields
			if obj.Spec.DefaultClientScopes != nil && *obj.Spec.DefaultClientScopes == nil {
				obj.Spec.DefaultClientScopes = nil
			}
			if obj.Spec.DefaultRoles != nil && *obj.Spec.DefaultRoles == nil {
				obj.Spec.DefaultRoles = nil
			}
			if obj.Spec.OptionalClientScopes != nil && *obj.Spec.OptionalClientScopes == nil {
				obj.Spec.OptionalClientScopes = nil
			}
			if obj.Spec.RedirectURIs != nil && *obj.Spec.RedirectURIs == nil {
				obj.Spec.RedirectURIs = nil
			}
			if obj.Spec.WebOrigins != nil && *obj.Spec.WebOrigins == nil {
				obj.Spec.WebOrigins = nil
			}
		},
	}
}

func customProtocolMapper(_ runtimeserializer.CodecFactory) []any {
	return []any{
		func(s *keycloakcommon.KeycloakClientProtocolMapper, c randfill.Continue) {
			c.FillNoCustom(s)
			if s.Config != nil && *s.Config == nil {
				s.Config = nil
			}
		},
	}
}

func TestFuzzyConversion(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = AddToScheme(scheme)
	_ = keycloakv1alpha2.AddToScheme(scheme)

	t.Run("for KeycloakClient", utilconversion.FuzzTestFunc(utilconversion.FuzzTestFuncInput{
		Scheme: scheme,
		Hub:    &keycloakv1alpha2.KeycloakClient{},
		Spoke:  &KeycloakClient{},
		// Ignore a field on the Spoke object (Spoke -> Hub -> Spoke test)
		SpokeAfterMutation: func(spoke conversion.Convertible) {
			obj := spoke.(*KeycloakClient)
			if obj.Spec.ClientSecretRef != nil {
				obj.Spec.ClientSecretRef.Optional = nil
			}
		},
		FuzzerFuncs: []fuzzer.FuzzerFuncs{customFuzzerFuncs, customProtocolMapper},
	}))
}
