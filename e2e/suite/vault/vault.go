/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
limitations under the License.
*/
package vault

import (

	// nolint
	. "github.com/onsi/ginkgo"
	// nolint
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/external-secrets/external-secrets/e2e/framework"
	"github.com/external-secrets/external-secrets/e2e/suite/common"
)

var _ = Describe("[vault] ", func() {
	f := framework.New("eso-vault")

	DescribeTable("sync secrets",
		framework.TableFunc(f,
			newVaultProvider(f)),
		// uses token auth
		Entry(common.JSONDataFromSync(f)),
		Entry(common.JSONDataWithProperty(f)),
		Entry(common.JSONDataWithTemplate(f)),
		Entry(common.DataPropertyDockerconfigJSON(f)),
		// use cert auth
		useCertAuth(common.JSONDataFromSync(f)),
		useCertAuth(common.JSONDataWithProperty(f)),
		useCertAuth(common.JSONDataWithTemplate(f)),
		useCertAuth(common.DataPropertyDockerconfigJSON(f)),
		// use cert auth
		useApproleAuth(common.JSONDataFromSync(f)),
		useApproleAuth(common.JSONDataWithProperty(f)),
		useApproleAuth(common.JSONDataWithTemplate(f)),
		useApproleAuth(common.DataPropertyDockerconfigJSON(f)),
	)
})

func useCertAuth(desc string, tc func(f *framework.TestCase)) TableEntry {
	return Entry(desc+" with cert auth", tc, func(tc *framework.TestCase) {
		tc.ExternalSecret.Spec.SecretStoreRef.Name = certAuthProviderName
	})
}

func useApproleAuth(desc string, tc func(f *framework.TestCase)) TableEntry {
	return Entry(desc+" with approle auth", tc, func(tc *framework.TestCase) {
		tc.ExternalSecret.Spec.SecretStoreRef.Name = appRoleAuthProviderName
	})
}
