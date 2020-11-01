package controllers

import (
	"context"
	"fmt"
	coverosv1alpha1 "github.com/coveros/genoa/api/v1alpha1"
	v3 "github.com/coveros/genoa/pkg/helm/v3"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/types"
	"path/filepath"
	"strings"
	"time"
)

const validTestDir = "reconcile-test-data/valid"
const invalidTestDir = "reconcile-test-data/invalid"

var _ = Describe("Valid test Reconciles", func() {

	// test-data prep
	tests, errReadingDir := ioutil.ReadDir(validTestDir)
	if errReadingDir != nil {
		Fail("Failed to read valid test dir")
	}

	for _, test := range tests {
		rawTestData, errReading := ioutil.ReadFile(filepath.Join(validTestDir, test.Name()))
		if errReading != nil {
			Fail("Failed to read valid test file")
		}

		testRelease := &coverosv1alpha1.Release{}
		err := yaml.Unmarshal(rawTestData, testRelease)
		if err != nil {
			Fail("Failed to parse test file")
		}
		namespacedName := fmt.Sprintf("%s/%s", testRelease.GetNamespace(), testRelease.GetName())

		When(fmt.Sprintf("%v helm release is created", namespacedName), func() {
			It(fmt.Sprintf("%v reconcile successfully and a helm release should exist", namespacedName), func() {

				helmV3, errCreatingActionConfig := v3.NewActionConfig(testRelease.GetNamespace(), cfg)
				if errCreatingActionConfig != nil {
					Fail("Failed to set up a helm client")
				}

				By(fmt.Sprintf("Creating a valid %v helm release", namespacedName), func() {
					Expect(k8sClient.Create(context.TODO(), testRelease)).Should(Succeed())
				})
				releaseFromCluster := &coverosv1alpha1.Release{}
				By(fmt.Sprintf("Verifying that %v CR exists and status is installed", namespacedName), func() {
					Eventually(func() bool {

						if err := k8sClient.Get(context.TODO(), types.NamespacedName{
							Name:      testRelease.GetName(),
							Namespace: testRelease.GetNamespace()},
							releaseFromCluster); err != nil {
							return false
						}
						return releaseFromCluster.Status.Installed

					}, 30*time.Second, 10*time.Second).Should(BeTrue())
				})

				releaseInfoFromCluster, errGettingRelease := helmV3.GetRelease(testRelease.GetName())
				if errGettingRelease != nil {
					Fail(fmt.Sprintf("%v helm release not found: %v", namespacedName, errGettingRelease))
				}

				By(fmt.Sprintf("Ensuring %v helm release info matches as per CR specs", namespacedName), func() {
					Eventually(func() bool {

						wantReleaseName := testRelease.GetName()
						wantReleaseNamespace := testRelease.GetNamespace()
						gotReleaseName := releaseInfoFromCluster.Name
						gotReleaseNamespace := releaseInfoFromCluster.Namespace
						wantChartName := strings.Split(testRelease.Spec.Chart, "/")[1]
						wantChartVersion := testRelease.Spec.Version
						gotChartName := releaseInfoFromCluster.Chart.Metadata.Name
						gotChartVers := releaseInfoFromCluster.Chart.Metadata.Version

						if wantReleaseName != gotReleaseName ||
							wantReleaseNamespace != gotReleaseNamespace ||
							wantChartName != gotChartName ||
							wantChartVersion != gotChartVers {
							Fail(fmt.Sprintf("Expected %v@%v-%v "+
								"from cluster, but got: %v/%v@%v-%v", namespacedName, wantChartName, wantChartVersion,
								gotReleaseNamespace, gotReleaseName, gotChartName, gotChartVers))
							return false
						}
						return true
					}).Should(BeTrue())
				})

			})
		})
	}
})

var _ = Describe("Invalid test Reconciles", func() {

	// test-data prep
	tests, errReadingDir := ioutil.ReadDir(invalidTestDir)
	if errReadingDir != nil {
		Fail("Failed to read valid test dir")
	}

	for _, test := range tests {
		rawTestData, errReading := ioutil.ReadFile(filepath.Join(invalidTestDir, test.Name()))
		if errReading != nil {
			Fail("Failed to read valid test file")
		}

		testRelease := &coverosv1alpha1.Release{}
		err := yaml.Unmarshal(rawTestData, testRelease)
		if err != nil {
			Fail("Failed to parse test file")
		}
		namespacedName := fmt.Sprintf("%s/%s", testRelease.GetNamespace(), testRelease.GetName())

		When(fmt.Sprintf("%v helm release is created", namespacedName), func() {
			It(fmt.Sprintf("%v does not reconcile successfully", namespacedName), func() {

				By(fmt.Sprintf("Creating a schema-valid %v helm release", namespacedName), func() {
					Expect(k8sClient.Create(context.TODO(), testRelease)).Should(Succeed())
				})

				By(fmt.Sprintf("Verifying that %v does not exist in cluster", namespacedName), func() {
					Eventually(func() bool {
						releaseFromCluster := &coverosv1alpha1.Release{}

						if err := k8sClient.Get(context.TODO(), types.NamespacedName{
							Name:      testRelease.GetName(),
							Namespace: testRelease.GetNamespace()},
							releaseFromCluster); err != nil {
							return false
						}
						return releaseFromCluster.Status.Installed

					}, 10*time.Second, 5*time.Second).Should(BeFalse())
				})
			})
		})
	}
})
