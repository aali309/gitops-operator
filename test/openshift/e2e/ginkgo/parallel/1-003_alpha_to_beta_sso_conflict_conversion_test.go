/*
Copyright 2025.

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

package parallel

import (
	"context"

	argov1alpha1api "github.com/argoproj-labs/argocd-operator/api/v1alpha1"
	argov1beta1api "github.com/argoproj-labs/argocd-operator/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/gitops-operator/test/openshift/e2e/ginkgo/fixture"
	argocdFixture "github.com/redhat-developer/gitops-operator/test/openshift/e2e/ginkgo/fixture/argocd"
	k8sFixture "github.com/redhat-developer/gitops-operator/test/openshift/e2e/ginkgo/fixture/k8s"
	fixtureUtils "github.com/redhat-developer/gitops-operator/test/openshift/e2e/ginkgo/fixture/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("GitOps Operator Parallel E2E Tests", func() {

	Context("1-003_alpha_to_beta_sso_conflict_conversion", func() {

		var (
			ctx       context.Context
			k8sClient client.Client
		)

		BeforeEach(func() {
			fixture.EnsureParallelCleanSlate()
			k8sClient, _ = fixtureUtils.GetE2ETestKubeClient()
			ctx = context.Background()
		})

		It("verifies expected behaviour of ArgoCD CR when dex and keycloak are both specified in v1alpha1 API", func() {

			if fixture.EnvLocalRun() {
				Skip("Conversion via webhook requires the operator to be running on the cluster, which is not the case for a local run")
				return
			}

			ns, nsCleanup := fixture.CreateRandomE2ETestNamespaceWithCleanupFunc()
			defer nsCleanup()

			By("creating Argo CD with both Dex and Keycloak values set in v1alpha1 API")

			argoCDalpha1 := &argov1alpha1api.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: ns.Name,
				},
				Spec: argov1alpha1api.ArgoCDSpec{
					Dex: &argov1alpha1api.ArgoCDDexSpec{
						OpenShiftOAuth: true,
					},
					SSO: &argov1alpha1api.ArgoCDSSOSpec{
						Provider: argov1alpha1api.SSOProviderTypeKeycloak,
						Keycloak: &argov1alpha1api.ArgoCDKeycloakSpec{
							RootCA: "\"---BEGIN---END---\"",
						},
						VerifyTLS: ptr.To(false),
					},
					ExtraConfig: map[string]string{
						"oidc.tls.insecure.skip.verify": "true",
					},
				},
			}
			Expect(k8sClient.Create(ctx, argoCDalpha1)).To(Succeed())

			argoCDbeta1 := &argov1beta1api.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: ns.Name,
				},
			}
			Expect(argoCDbeta1).Should(k8sFixture.ExistByName())

			By("verifying expected Argo CD is running, that SSO status has failed, and has the expected values via the v1beta1 API")

			Eventually(argoCDbeta1, "2m", "5s").Should(
				And(argocdFixture.HaveSSOStatus("Failed")))

			Expect(argoCDbeta1.Spec.SSO.Provider).To(Equal(argov1beta1api.SSOProviderTypeDex))
			Expect(argoCDbeta1.Spec.SSO.Dex.OpenShiftOAuth).To(Equal(true))
			Expect(argoCDbeta1.Spec.SSO.Keycloak.RootCA).To(Equal("\"---BEGIN---END---\""))
			Expect(*argoCDbeta1.Spec.SSO.Keycloak.VerifyTLS).To(Equal(false))
			Expect(argoCDbeta1.Spec.ExtraConfig["oidc.tls.insecure.skip.verify"]).To(Equal("true"))

			By("deleting ArgoCD CR via v1alpha1 API")
			Expect(k8sClient.Delete(ctx, argoCDalpha1)).To(Succeed())

			By("verifying ArgoCD CR no longer exists via v1beta1 API")
			Eventually(argoCDbeta1).Should(k8sFixture.NotExistByName())

		})

	})
})
