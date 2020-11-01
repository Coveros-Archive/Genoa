package controllers

import (
	"context"
	"fmt"
	coverosv1alpha1 "github.com/coveros/genoa/api/v1alpha1"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/types"
	"path/filepath"
	"time"
)

const validTestDir = "integration-test-data/valid"
const invalidTestDir = "integration-test-data/invalid"

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

		When(fmt.Sprintf("%v helm release is created",namespacedName), func() {
			It(fmt.Sprintf("%v reconcile successfully and a helm release should exist",namespacedName), func() {

				By(fmt.Sprintf("Creating a valid %v helm release", namespacedName), func() {
					Expect(k8sClient.Create(context.TODO(), testRelease)).Should(Succeed())
				})

				By(fmt.Sprintf("Verifying that %v exists and is installed", namespacedName), func() {
					Eventually(func() bool {
						releaseFromCluster := &coverosv1alpha1.Release{}

						if err := k8sClient.Get(context.TODO(), types.NamespacedName{
							Name:      testRelease.GetName(),
							Namespace: testRelease.GetNamespace()},
							releaseFromCluster); err != nil {
							return false
						}
						return releaseFromCluster.Status.Installed

					}, 30*time.Second, 10*time.Second).Should(BeTrue())
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

		When(fmt.Sprintf("%v helm release is created",namespacedName), func() {
			It(fmt.Sprintf("%v does not reconcile successfully",namespacedName), func() {

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

					}, 30*time.Second, 10*time.Second).Should(BeFalse())
				})
			})
		})
	}
})