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

package externalsecret

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	esv1alpha1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1alpha1"
)

const ExternalSecretNamespace = "default"

var _ = Describe("ExternalsecretController", func() {
	var (
		ExternalSecretName          = "externalsecret-operator-test"
		ExternalSecretKey           = "test-key"
		ExternalSecretVersion       = "test-version"
		ExternalSecret2Key          = "test-key-2"
		ExternalSecret2Version      = "test-version-2"
		ExternalSecret3Key          = "test-key-3"
		ExternalSecret3Version      = "test-version-3"
		ExternalSecretKeyUpdate     = "test-key-update"
		ExternalSecretVersionUpdate = "test-version-update"
		SecretStoreName             = "test-externalsecret-store"
		StoreControllerName         = "test-externalsecret-ctrl"
		CredentialSecretName        = "credential-secret-external-secret"
		TargetName                  = "test-secret-target"

		timeout  = time.Second * 30
		interval = time.Millisecond * 250
	)

	BeforeEach(func() {})

	AfterEach(func() {})

	Context("Given an ExternalSecret", func() {
		It("Should handle ExternalSecret correctly", func() {
			By("Creating a new ExternalSecret")
			ctx := context.Background()

			credentialsSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      CredentialSecretName,
					Namespace: ExternalSecretNamespace,
				},
				StringData: map[string]string{
					"credentials.json": `{
						"Credential": "-dummyvalue"
					}`,
				},
			}

			Expect(k8sClient.Create(ctx, credentialsSecret)).Should(Succeed())

			credentialsSecretLookupKey := types.NamespacedName{Name: CredentialSecretName, Namespace: ExternalSecretNamespace}
			createdCredentialsSecret := &corev1.Secret{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, credentialsSecretLookupKey, createdCredentialsSecret)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			secretStore := &esv1alpha1.SecretStore{
				ObjectMeta: metav1.ObjectMeta{
					Name:      SecretStoreName,
					Namespace: ExternalSecretNamespace,
				},

				Spec: esv1alpha1.SecretStoreSpec{
					Controller: StoreControllerName,
				},
			}

			Expect(k8sClient.Create(ctx, secretStore)).Should(Succeed())

			secretStoreLookupKey := types.NamespacedName{Name: SecretStoreName, Namespace: ExternalSecretNamespace}
			createdSecretStore := &esv1alpha1.SecretStore{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretStoreLookupKey, createdSecretStore)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			externalSecret := &esv1alpha1.ExternalSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ExternalSecretName,
					Namespace: ExternalSecretNamespace,
				},
				Spec: esv1alpha1.ExternalSecretSpec{
					SecretStoreRef: esv1alpha1.SecretStoreRef{
						Name: SecretStoreName,
					},
					Target: esv1alpha1.ExternalSecretTarget{
						Name: TargetName,
					},
					Data: []esv1alpha1.ExternalSecretData{
						{
							SecretKey: ExternalSecretKey,
							RemoteRef: esv1alpha1.ExternalSecretDataRemoteRef{
								Key:     ExternalSecretKey,
								Version: ExternalSecretVersion,
							},
						},

						{
							SecretKey: ExternalSecret2Key,
							RemoteRef: esv1alpha1.ExternalSecretDataRemoteRef{
								Key:     ExternalSecret2Key,
								Version: ExternalSecret2Version,
							},
						},

						{
							SecretKey: ExternalSecret3Key,
							RemoteRef: esv1alpha1.ExternalSecretDataRemoteRef{
								Key:     ExternalSecret3Key,
								Version: ExternalSecret3Version,
							},
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, externalSecret)).Should(Succeed())

			externalSecretLookupKey := types.NamespacedName{Name: ExternalSecretName, Namespace: ExternalSecretNamespace}
			createdExternalSecret := &esv1alpha1.ExternalSecret{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, externalSecretLookupKey, createdExternalSecret)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			Expect(len(createdExternalSecret.Spec.Data)).Should(BeNumerically("==", 3))

			Expect(createdExternalSecret.Spec.Data[0].SecretKey).Should(Equal(ExternalSecretKey))
			Expect(createdExternalSecret.Spec.Data[0].RemoteRef.Version).Should(Equal(ExternalSecretVersion))

			Expect(createdExternalSecret.Spec.Data[1].SecretKey).Should(Equal(ExternalSecret2Key))
			Expect(createdExternalSecret.Spec.Data[1].RemoteRef.Version).Should(Equal(ExternalSecret2Version))

			Expect(createdExternalSecret.Spec.Data[2].SecretKey).Should(Equal(ExternalSecret3Key))
			Expect(createdExternalSecret.Spec.Data[2].RemoteRef.Version).Should(Equal(ExternalSecret3Version))

			By("Creating a new secret with correct values")
			secret := &corev1.Secret{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, externalSecretLookupKey, secret)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			secretValue := string(secret.Data[ExternalSecretKey])
			Expect(secretValue).Should(Equal("test-keytest-versionTestParameter"))

			secretValue2 := string(secret.Data[ExternalSecret2Key])
			Expect(secretValue2).Should(Equal("test-key-2test-version-2TestParameter"))

			secretValue3 := string(secret.Data[ExternalSecret3Key])
			Expect(secretValue3).Should(Equal("test-key-3test-version-3TestParameter"))

			By("Updating the Secret if it already exists")
			updatedSecrets := []esv1alpha1.ExternalSecretData{
				{
					SecretKey: ExternalSecretKeyUpdate,
					RemoteRef: esv1alpha1.ExternalSecretDataRemoteRef{
						Version: ExternalSecretVersionUpdate,
					},
				},
			}

			createdExternalSecret.Spec.Data = updatedSecrets

			Expect(k8sClient.Update(ctx, createdExternalSecret)).Should(Succeed())

			updatedExternalSecret := &esv1alpha1.ExternalSecret{}
			Eventually(func() []esv1alpha1.ExternalSecretData {
				err := k8sClient.Get(ctx, externalSecretLookupKey, updatedExternalSecret)
				if err != nil {
					return []esv1alpha1.ExternalSecretData{}
				}
				return updatedExternalSecret.Spec.Data
			}, timeout, interval).Should(Equal(updatedSecrets))

			updatedSecret := &corev1.Secret{}
			Eventually(func() string {
				err := k8sClient.Get(ctx, externalSecretLookupKey, updatedSecret)
				if err != nil {
					return ""
				}
				return string(updatedSecret.Data[ExternalSecretKeyUpdate])
			}, timeout, interval).Should(Equal("test-key-updatetest-version-updateTestParameter"))

			By("Deleting the External Secret")
			Eventually(func() error {
				es := &esv1alpha1.ExternalSecret{}
				k8sClient.Get(context.Background(), externalSecretLookupKey, es)
				return k8sClient.Delete(ctx, es)
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				es := &esv1alpha1.ExternalSecret{}
				return k8sClient.Get(ctx, externalSecretLookupKey, es)
			}, timeout, interval).ShouldNot(Succeed())
		})

	})

})
