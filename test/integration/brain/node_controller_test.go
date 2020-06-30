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
	. "github.com/rancher/octopus/test/framework"
	. "github.com/rancher/octopus/test/framework/ginkgo"
	"github.com/rancher/octopus/test/util/node"
)

var _ = Describe("verify Node controller", func() {
	var (
		testNamespace corev1.Namespace
		testNodeName  string

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
		targetItem = edgev1alpha1.DeviceLink{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    testNamespace.Name,
				GenerateName: "test-",
			},
			Spec: edgev1alpha1.DeviceLinkSpec{
				Adaptor: edgev1alpha1.DeviceAdaptor{
					Name: "adaptors.edge.cattle.io/fake",
					Node: testNodeName,
				},
			},
		}
	})

	Context("if the node is not existed", func() {

		BeforeEach(func() {
			var err error
			testNodeName, err = node.GetInvalidWorker(testCtx, k8sCli)
			Expect(err).ToNot(HaveOccurred())
		})

		K3dIt("should succeed after the node is added", func() {

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

			By("when add the target node", func() {
				Expect(GetCluster().AddWorker(testRootDir, GinkgoWriter, targetItem.Spec.Adaptor.Node)).Should(Succeed())
			})

			By("then it succeeded on node verification", func() {
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

	Context("if the node is existed", func() {

		BeforeEach(func() {
			var err error
			testNodeName, err = node.GetValidWorker(testCtx, k8sCli)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail after the node is deleted", func() {

			By("given a new link which succeed on node verification", func() {
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

			By("when delete the target node", func() {
				Expect(k8sCli.Delete(testCtx, &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: targetItem.Status.NodeName,
					},
				})).Should(Succeed())
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
	})

})
