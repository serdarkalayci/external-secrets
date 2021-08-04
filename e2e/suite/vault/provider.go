/*
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
package vault

import (
	"context"
	"fmt"
	"net/http"

	vault "github.com/hashicorp/vault/api"

	//nolint
	. "github.com/onsi/ginkgo"

	//nolint
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	esv1alpha1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1alpha1"
	esmeta "github.com/external-secrets/external-secrets/apis/meta/v1"
	"github.com/external-secrets/external-secrets/e2e/framework"
	"github.com/external-secrets/external-secrets/e2e/framework/addon"
)

type vaultProvider struct {
	url       string
	client    *vault.Client
	framework *framework.Framework
}

const (
	certAuthProviderName    = "cert-auth-provider"
	appRoleAuthProviderName = "app-role-provider"
)

func newVaultProvider(f *framework.Framework) *vaultProvider {
	prov := &vaultProvider{
		framework: f,
	}
	BeforeEach(prov.BeforeEach)
	return prov
}

func (s *vaultProvider) CreateSecret(key, val string) {
	req := s.client.NewRequest(http.MethodPost, fmt.Sprintf("/v1/secret/data/%s", key))
	req.BodyBytes = []byte(fmt.Sprintf(`{"data": %s}`, val))
	_, err := s.client.RawRequestWithContext(context.Background(), req)
	Expect(err).ToNot(HaveOccurred())
}

func (s *vaultProvider) DeleteSecret(key string) {
	req := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/secret/data/%s", key))
	_, err := s.client.RawRequestWithContext(context.Background(), req)
	Expect(err).ToNot(HaveOccurred())
}

func (s *vaultProvider) BeforeEach() {
	ns := s.framework.Namespace.Name
	v := addon.NewVault(ns)
	s.framework.Install(v)
	s.client = v.VaultClient
	s.url = v.VaultURL

	s.createCertStore(v, ns)
	s.createTokenStore(v, ns)
	s.createAppRoleStore(v, ns)
}

func (s *vaultProvider) createCertStore(v *addon.Vault, ns string) {
	By("creating a vault secret")
	clientCert := v.ClientCert
	clientKey := v.ClientKey
	serverCA := v.VaultServerCA
	vaultCreds := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certAuthProviderName,
			Namespace: ns,
		},
		Data: map[string][]byte{
			"token":       []byte(v.RootToken),
			"client_cert": clientCert,
			"client_key":  clientKey,
		},
	}
	err := s.framework.CRClient.Create(context.Background(), vaultCreds)
	Expect(err).ToNot(HaveOccurred())

	By("creating an secret store for vault")
	secretStore := &esv1alpha1.SecretStore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certAuthProviderName,
			Namespace: ns,
		},
		Spec: esv1alpha1.SecretStoreSpec{
			Provider: &esv1alpha1.SecretStoreProvider{
				Vault: &esv1alpha1.VaultProvider{
					Version:  esv1alpha1.VaultKVStoreV2,
					Path:     "secret",
					Server:   s.url,
					CABundle: serverCA,
					Auth: esv1alpha1.VaultAuth{
						Cert: &esv1alpha1.VaultCertAuth{
							ClientCert: esmeta.SecretKeySelector{
								Name: certAuthProviderName,
								Key:  "client_cert",
							},
							SecretRef: esmeta.SecretKeySelector{
								Name: certAuthProviderName,
								Key:  "client_key",
							},
						},
					},
				},
			},
		},
	}
	err = s.framework.CRClient.Create(context.Background(), secretStore)
	Expect(err).ToNot(HaveOccurred())
}

func (s vaultProvider) createTokenStore(v *addon.Vault, ns string) {
	serverCA := v.VaultServerCA
	vaultCreds := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "token-provider",
			Namespace: ns,
		},
		Data: map[string][]byte{
			"token": []byte(v.RootToken),
		},
	}
	err := s.framework.CRClient.Create(context.Background(), vaultCreds)
	Expect(err).ToNot(HaveOccurred())
	secretStore := &esv1alpha1.SecretStore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.framework.Namespace.Name,
			Namespace: ns,
		},
		Spec: esv1alpha1.SecretStoreSpec{
			Provider: &esv1alpha1.SecretStoreProvider{
				Vault: &esv1alpha1.VaultProvider{
					Version:  esv1alpha1.VaultKVStoreV2,
					Path:     "secret",
					Server:   s.url,
					CABundle: serverCA,
					Auth: esv1alpha1.VaultAuth{
						TokenSecretRef: &esmeta.SecretKeySelector{
							Name: "token-provider",
							Key:  "token",
						},
					},
				},
			},
		},
	}
	err = s.framework.CRClient.Create(context.Background(), secretStore)
	Expect(err).ToNot(HaveOccurred())
}

func (s vaultProvider) createAppRoleStore(v *addon.Vault, ns string) {
	By("creating a vault secret")
	serverCA := v.VaultServerCA
	vaultCreds := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appRoleAuthProviderName,
			Namespace: ns,
		},
		Data: map[string][]byte{
			"approle_secret": []byte(v.AppRoleSecret),
		},
	}
	err := s.framework.CRClient.Create(context.Background(), vaultCreds)
	Expect(err).ToNot(HaveOccurred())

	By("creating an secret store for vault")
	secretStore := &esv1alpha1.SecretStore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appRoleAuthProviderName,
			Namespace: ns,
		},
		Spec: esv1alpha1.SecretStoreSpec{
			Provider: &esv1alpha1.SecretStoreProvider{
				Vault: &esv1alpha1.VaultProvider{
					Version:  esv1alpha1.VaultKVStoreV2,
					Path:     "secret",
					Server:   s.url,
					CABundle: serverCA,
					Auth: esv1alpha1.VaultAuth{
						AppRole: &esv1alpha1.VaultAppRole{
							Path:   v.AppRolePath,
							RoleID: v.AppRoleID,
							SecretRef: esmeta.SecretKeySelector{
								Name: appRoleAuthProviderName,
								Key:  "approle_secret",
							},
						},
					},
				},
			},
		},
	}
	err = s.framework.CRClient.Create(context.Background(), secretStore)
	Expect(err).ToNot(HaveOccurred())
}
