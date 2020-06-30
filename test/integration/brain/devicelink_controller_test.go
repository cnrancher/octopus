package brain

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/util/object"
	"github.com/rancher/octopus/test/util/crd"
	"github.com/rancher/octopus/test/util/node"
)

var _ = Describe("verify DeviceLink controller", func() {
	var (
		testNamespace corev1.Namespace
		testNodeName  string
		testModel     metav1.TypeMeta

		targetItem edgev1alpha1.DeviceLink
	)

	BeforeEach(func() {
		testNamespace = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-",
			},
		}
		_ = k8sCli.Create(testCtx, &testNamespace)

		testNodeName, _ = node.GetValidWorker(testCtx, k8sCli)

		testModel = metav1.TypeMeta{
			Kind:       "IntegrationBrainDLValidateDevice",
			APIVersion: "devices.edge.cattle.io/v1alpha1",
		}
		_ = k8sCli.Create(testCtx, crd.MakeOfTypeMeta(testModel))
	})

	AfterEach(func() {
		_ = k8sCli.DeleteAllOf(testCtx, &edgev1alpha1.DeviceLink{}, client.InNamespace(testNamespace.Name))
		_ = k8sCli.Delete(testCtx, &testNamespace)
	})

	JustBeforeEach(func() {
		targetItem = edgev1alpha1.DeviceLink{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    testNamespace.Name,
				GenerateName: "test-",
			},
			Spec: edgev1alpha1.DeviceLinkSpec{
				Adaptor: edgev1alpha1.DeviceAdaptor{
					Node: testNodeName,
					Name: "adaptors.edge.cattle.io/fake",
				},
				Model: testModel,
			},
		}
	})

	Context("if the node spec is invalid", func() {

		BeforeEach(func() {
			var err error
			testNodeName, err = node.GetInvalidWorker(testCtx, k8sCli)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should succeed after modified to valid node", func() {

			By("given a new link which failed on node verification", func() {
				// creates
				Expect(k8sCli.Create(testCtx, &targetItem)).Should(Succeed())

				// confirms
				var key = object.GetNamespacedName(&targetItem)
				Eventually(func() error {
					if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
						return err
					}
					if targetItem.GetNodeExistedStatus() != metav1.ConditionFalse {
						return errors.New("should not find the corresponding node of link")
					}
					return nil
				}, 30, 1).Should(Succeed())
			})

			By("when modify to valid node", func() {
				Expect(k8sCli.Get(testCtx, object.GetNamespacedName(&targetItem), &targetItem)).Should(Succeed())

				var targetNodeName, err = node.GetValidWorker(testCtx, k8sCli)
				Expect(err).ToNot(HaveOccurred())

				targetItem.Spec.Adaptor.Node = targetNodeName
				Expect(k8sCli.Update(testCtx, &targetItem)).Should(Succeed())
			})

			By("then it succeed on node verification", func() {
				var key = object.GetNamespacedName(&targetItem)
				Eventually(func() error {
					if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
						return err
					}
					if targetItem.GetNodeExistedStatus() != metav1.ConditionTrue {
						return errors.New("could not find the corresponding node of link")
					}
					return nil
				}, 30, 1).Should(Succeed())
			})

		})

	})

	Context("if the node spec is valid", func() {

		Context("and the model spec is invalid", func() {

			BeforeEach(func() {
				testModel = metav1.TypeMeta{
					Kind:       fmt.Sprintf("X%s", testModel.Kind),
					APIVersion: testModel.APIVersion,
				}
			})

			It("should succeed after modified to valid model", func() {

				By("given a new link which failed on model verification", func() {
					// creates
					Expect(k8sCli.Create(testCtx, &targetItem)).Should(Succeed())

					// confirms
					var key = object.GetNamespacedName(&targetItem)
					Eventually(func() error {
						if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
							return err
						}
						if targetItem.GetModelExistedStatus() != metav1.ConditionFalse {
							return errors.New("should not find the corresponding model of link")
						}
						return nil
					}, 30, 1).Should(Succeed())
				})

				By("when modify to valid model", func() {
					Expect(k8sCli.Get(testCtx, object.GetNamespacedName(&targetItem), &targetItem)).Should(Succeed())

					targetItem.Spec.Model = metav1.TypeMeta{
						Kind:       strings.TrimPrefix(targetItem.Spec.Model.Kind, "X"),
						APIVersion: targetItem.Spec.Model.APIVersion,
					}
					Expect(k8sCli.Update(testCtx, &targetItem)).Should(Succeed())
				})

				By("then it succeed on model verification", func() {
					var key = object.GetNamespacedName(&targetItem)
					Eventually(func() error {
						if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
							return err
						}
						if targetItem.GetModelExistedStatus() != metav1.ConditionTrue {
							return errors.New("could not find the corresponding model of link")
						}
						return nil
					}, 30, 1).Should(Succeed())
				})

			})

		})

		Context("and the model spec is valid", func() {

			It("should fail after modified to invalid node", func() {

				By("given a new link which succeeded on node verification", func() {
					// creates
					Expect(k8sCli.Create(testCtx, &targetItem)).Should(Succeed())

					// confirms
					var key = object.GetNamespacedName(&targetItem)
					Eventually(func() error {
						if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
							return err
						}
						if targetItem.GetNodeExistedStatus() != metav1.ConditionTrue {
							return errors.New("could not find the corresponding node of link")
						}
						return nil
					}, 30, 1).Should(Succeed())
				})

				By("when modify to invalid node", func() {
					Expect(k8sCli.Get(testCtx, object.GetNamespacedName(&targetItem), &targetItem)).Should(Succeed())

					var targetNodeName, err = node.GetInvalidWorker(testCtx, k8sCli)
					Expect(err).ToNot(HaveOccurred())

					targetItem.Spec.Adaptor.Node = targetNodeName
					Expect(k8sCli.Update(testCtx, &targetItem)).Should(Succeed())
				})

				By("then it failed on node verification", func() {
					var key = object.GetNamespacedName(&targetItem)
					Eventually(func() error {
						if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
							return err
						}
						if targetItem.GetNodeExistedStatus() != metav1.ConditionFalse {
							return errors.New("should not find the corresponding node of link")
						}
						return nil
					}, 30, 1).Should(Succeed())
				})

			})

			It("should fail after modified to invalid model", func() {

				By("given a new link which succeed on model verification", func() {
					// creates
					Expect(k8sCli.Create(testCtx, &targetItem)).Should(Succeed())

					// confirms
					var key = object.GetNamespacedName(&targetItem)
					Eventually(func() error {
						if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
							return err
						}
						if targetItem.GetModelExistedStatus() != metav1.ConditionTrue {
							return errors.New("could not find the corresponding model of link")
						}
						return nil
					}, 30, 1).Should(Succeed())
				})

				By("when modify to invalid model", func() {
					Expect(k8sCli.Get(testCtx, object.GetNamespacedName(&targetItem), &targetItem)).Should(Succeed())

					targetItem.Spec.Model = metav1.TypeMeta{
						Kind:       fmt.Sprintf("X%s", targetItem.Status.Model.Kind),
						APIVersion: targetItem.Status.Model.APIVersion,
					}
					Expect(k8sCli.Update(testCtx, &targetItem)).Should(Succeed())
				})

				By("then it failed on model verification", func() {
					var key = object.GetNamespacedName(&targetItem)
					Eventually(func() error {
						if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
							return err
						}
						if targetItem.GetModelExistedStatus() != metav1.ConditionFalse {
							return errors.New("should not find the corresponding model of link")
						}
						return nil
					}, 30, 1).Should(Succeed())
				})

			})

		})

	})

})
