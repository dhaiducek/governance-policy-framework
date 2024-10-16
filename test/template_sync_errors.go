// Copyright (c) 2022 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/stolostron/governance-policy-framework/test/common"
	policiesv1 "open-cluster-management.io/governance-policy-propagator/api/v1"
)

func TemplateSyncErrors(labels ...string) bool {
	const (
		nonexistentPolicyKindYaml string = "../resources/template-sync-errors/pretend-policy-template.yaml"
		nonexistentPolicyKindName string = "pretend-policy-template"
		invalidCRPolicyYaml       string = "../resources/template-sync-errors/invalid-cr-template.yaml"
		invalidCRPolicyName       string = "invalid-cr-template"
	)

	Describe("GRC: [P1][Sev1][policy-grc] Test handling template-sync errors", Label(labels...), func() {
		Describe("Test using a template with a non-existent CRD", Ordered, func() {
			AfterAll(func() {
				OcHub("delete", "-f", nonexistentPolicyKindYaml, "-n", UserNamespace)
			})
			It("Should be noncompliant with a mapping not found status", func() {
				clientManagedDynamic := NewKubeClientDynamic("", KubeconfigManaged, "")
				clientHubDynamic := NewKubeClientDynamic("", KubeconfigHub, "")

				DoCreatePolicyTest(clientHubDynamic, clientManagedDynamic, nonexistentPolicyKindYaml)
				DoRootComplianceTest(clientHubDynamic, nonexistentPolicyKindName, policiesv1.NonCompliant)

				Eventually(
					GetLatestStatusMessage(clientManagedDynamic, nonexistentPolicyKindName, 0),
					DefaultTimeoutSeconds, 1,
				).Should(MatchRegexp(".*Mapping not found.*"))
			})
			It("Should become compliant when the kind is fixed", func() {
				clientManagedDynamic := NewKubeClientDynamic("", KubeconfigManaged, "")
				clientHubDynamic := NewKubeClientDynamic("", KubeconfigHub, "")

				OcHub("patch", "policies.policy.open-cluster-management.io", nonexistentPolicyKindName,
					"-n", UserNamespace, "--type=json", "-p", `[{
						"op":"replace",
						"path":"/spec/policy-templates/0/objectDefinition/kind",
						"value":"ConfigurationPolicy"
					}]`)

				DoRootComplianceTest(clientHubDynamic, nonexistentPolicyKindName, policiesv1.Compliant)

				Eventually(
					GetLatestStatusMessage(clientManagedDynamic, nonexistentPolicyKindName, 0),
					DefaultTimeoutSeconds, 1,
				).ShouldNot(MatchRegexp(".*Mapping not found.*"))
			})
			It("Should become noncompliant when the original policy is restored", func() {
				clientManagedDynamic := NewKubeClientDynamic("", KubeconfigManaged, "")
				clientHubDynamic := NewKubeClientDynamic("", KubeconfigHub, "")

				OcHub("apply", "-f", nonexistentPolicyKindYaml, "-n", UserNamespace)

				DoRootComplianceTest(clientHubDynamic, nonexistentPolicyKindName, policiesv1.NonCompliant)

				Eventually(
					GetLatestStatusMessage(clientManagedDynamic, nonexistentPolicyKindName, 0),
					DefaultTimeoutSeconds, 1,
				).Should(MatchRegexp(".*Mapping not found.*"))
			})
		})
		Describe("Test using a template with an invalid CR", Ordered, func() {
			AfterAll(func() {
				OcHub("delete", "-f", invalidCRPolicyYaml, "-n", UserNamespace)
			})
			It("Should be noncompliant and report the reason the CR is invalid", func() {
				clientManagedDynamic := NewKubeClientDynamic("", KubeconfigManaged, "")
				clientHubDynamic := NewKubeClientDynamic("", KubeconfigHub, "")

				DoCreatePolicyTest(clientHubDynamic, clientManagedDynamic, invalidCRPolicyYaml)
				DoRootComplianceTest(clientHubDynamic, invalidCRPolicyName, policiesv1.NonCompliant)

				Eventually(
					GetLatestStatusMessage(clientManagedDynamic, invalidCRPolicyName, 0),
					DefaultTimeoutSeconds, 1,
				).Should(MatchRegexp(".*Failed to create.*Unsupported value.*"))
			})
			It("Should become compliant when the spec is fixed", func() {
				clientManagedDynamic := NewKubeClientDynamic("", KubeconfigManaged, "")
				clientHubDynamic := NewKubeClientDynamic("", KubeconfigHub, "")

				OcHub("patch", "policies.policy.open-cluster-management.io", invalidCRPolicyName,
					"-n", UserNamespace, "--type=json", "-p", `[{
						"op":"replace",
						"path":"/spec/policy-templates/0/objectDefinition/spec/pruneObjectBehavior",
						"value":"None"
					}]`)

				DoRootComplianceTest(clientHubDynamic, invalidCRPolicyName, policiesv1.Compliant)

				Eventually(
					GetLatestStatusMessage(clientManagedDynamic, invalidCRPolicyName, 0),
					DefaultTimeoutSeconds, 1,
				).ShouldNot(MatchRegexp(".*Failed to create.*Unsupported value.*"))
			})
			It("Should become noncompliant when the original policy is restored", func() {
				clientManagedDynamic := NewKubeClientDynamic("", KubeconfigManaged, "")
				clientHubDynamic := NewKubeClientDynamic("", KubeconfigHub, "")

				OcHub("apply", "-f", invalidCRPolicyYaml, "-n", UserNamespace)

				DoRootComplianceTest(clientHubDynamic, invalidCRPolicyName, policiesv1.NonCompliant)

				Eventually(
					GetLatestStatusMessage(clientManagedDynamic, invalidCRPolicyName, 0),
					DefaultTimeoutSeconds, 1,
				).Should(MatchRegexp(".*Failed to update.*Unsupported value.*"))
			})
		})
	})

	return true
}
