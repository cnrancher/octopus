package brain

import (
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

var _ = Describe("verify Model controller", func() {
	var (
		testNamespace corev1.Namespace
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
	})

	AfterEach(func() {
		_ = k8sCli.DeleteAllOf(testCtx, &edgev1alpha1.DeviceLink{}, client.InNamespace(testNamespace.Name))
		_ = k8sCli.Delete(testCtx, &testNamespace)
	})

	JustBeforeEach(func() {
		var targetNodeName, _ = node.GetValidWorker(testCtx, k8sCli)

		targetItem = edgev1alpha1.DeviceLink{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    testNamespace.Name,
				GenerateName: "test-",
			},
			Spec: edgev1alpha1.DeviceLinkSpec{
				Adaptor: edgev1alpha1.DeviceAdaptor{
					Node: targetNodeName,
					Name: "adaptors.edge.cattle.io/fake",
				},
				Model: testModel,
			},
		}
	})

	Context("if the node spec is valid", func() {

		Context("and the model is not existed", func() {

			BeforeEach(func() {
				testModel = metav1.TypeMeta{
					Kind:       "IntegrationBrainModelInvalidFirstDevice",
					APIVersion: "devices.edge.cattle.io/v1alpha1",
				}

				_ = k8sCli.Delete(testCtx, crd.MakeOfTypeMeta(testModel))
			})

			It("should succeed after the model is added", func() {

				By("given a new link which failed on model verification", func() {
					// creates
					Expect(k8sCli.Create(testCtx, &targetItem)).Should(Succeed())

					// confirmed
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

				By("when add the target model", func() {
					Expect(k8sCli.Create(testCtx, crd.MakeOfTypeMeta(targetItem.Spec.Model))).Should(Succeed())
				})

				By("then it succeeded on model verification", func() {
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

		Context("and the model is existed", func() {

			BeforeEach(func() {
				testModel = metav1.TypeMeta{
					Kind:       "IntegrationBrainModelValidFirstDevice",
					APIVersion: "devices.edge.cattle.io/v1alpha1",
				}

				_ = k8sCli.Create(testCtx, crd.MakeOfTypeMeta(testModel))
			})

			It("should fail after the model is deleted", func() {

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

				By("when delete the target model", func() {
					Expect(k8sCli.Delete(testCtx, crd.MakeOfTypeMeta(*targetItem.Status.Model))).Should(Succeed())
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
