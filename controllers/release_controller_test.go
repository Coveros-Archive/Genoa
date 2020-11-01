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
	"time"
)

const validTestDir = "integration-test-data/valid"

var _ = Describe("Successful Reconcile", func() {
	When("A helm release is created", func() {
		It("Should reconcile successfully and a helm release should exist", func() {

			var tests, errReadingDir = ioutil.ReadDir(validTestDir)
			Expect(errReadingDir).ToNot(HaveOccurred())

			for _, test := range tests {

				rawTestData, errReading := ioutil.ReadFile(validTestDir + "/" + test.Name())
				Expect(errReading).ToNot(HaveOccurred())
				Expect(rawTestData).ToNot(BeNil())

				testRelease := &coverosv1alpha1.Release{}
				err := yaml.Unmarshal(rawTestData, testRelease)
				Expect(err).ToNot(HaveOccurred())
				Expect(testRelease).ToNot(BeNil())

				By(fmt.Sprintf("Creating a valid %v/%v helm release", testRelease.GetNamespace(), testRelease.GetName()), func() {
					Expect(k8sClient.Create(context.TODO(), testRelease)).Should(Succeed())
				})

				By(fmt.Sprintf("Verifying that %v/%v exists and is installed", testRelease.GetNamespace(), testRelease.GetName()), func() {
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
			}

		})
	})
})
