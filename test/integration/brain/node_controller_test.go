package brain

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/status/devicelink"
	"github.com/rancher/octopus/pkg/util/collection"
	"github.com/rancher/octopus/pkg/util/object"
	. "github.com/rancher/octopus/test/framework"
	"github.com/rancher/octopus/test/util/content"
	"github.com/rancher/octopus/test/util/node"
)

// testing scenarios:
//	+ Node instance
//		- validate if all instances have `edge.cattle.io/octopus-brain` finalizer
//	+ DeviceLink instance
//		- validate if target link change when deleting the Node instance
//		- validate if target link change when adding the lost Node instance back (working on local k3d cluster)
var _ = Describe("Node controller", func() {
	var (
		err error

		targetNode      string
		targetAdaptor   string
		targetModel     metav1.TypeMeta
		targetNamespace string
		targetItem      edgev1alpha1.DeviceLink
	)

	AfterEach(func() {
		_ = k8sCli.DeleteAllOf(testCtx, &edgev1alpha1.DeviceLink{}, client.InNamespace(targetNamespace))
	})

	BeforeEach(func() {
		targetNode, err = node.GetValidWorker(testCtx, k8sCli)
		Expect(err).ToNot(HaveOccurred())

		targetAdaptor = "adaptors.edge.cattle.io/dummy"
		targetModel = metav1.TypeMeta{
			Kind:       "DummyDevice",
			APIVersion: "devices.edge.cattle.io/v1alpha1",
		}
		targetNamespace = "default"

		targetItem = edgev1alpha1.DeviceLink{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    targetNamespace,
				GenerateName: "test-",
			},
			Spec: edgev1alpha1.DeviceLinkSpec{
				Adaptor: edgev1alpha1.DeviceAdaptor{
					Node: targetNode,
					Name: targetAdaptor,
					Parameters: content.ToRawExtension(
						map[string]string{
							"ip": "1.2.3.4",
						},
					),
				},
				Model: targetModel,
				Template: edgev1alpha1.DeviceTemplateSpec{
					DeviceMeta: edgev1alpha1.DeviceMeta{
						Labels: map[string]string{
							"l1": "v1",
						},
					},
					Spec: content.ToRawExtension(
						map[string]interface{}{
							"gear": "slow",
							"on":   true,
						},
					),
				},
			},
		}
	})

	Context("Node instance", func() {

		It("should have finalizer", func() {
			// confirmed
			Eventually(func() (bool, error) {
				var list corev1.NodeList
				if err := k8sCli.List(testCtx, &list); err != nil {
					return false, err
				}
				for _, node := range list.Items {
					if !collection.StringSliceContain(node.Finalizers, "edge.cattle.io/octopus-brain") {
						return false, errors.Errorf("could not find corresponding finalizer from %s node", node.Name)
					}
				}
				return true, nil
			}, 30, 1).Should(BeTrue())
		})

	})

	Context("DeviceLink instance", func() {
		var targetNodeRecord string

		It("should be changed if deleted the node", func() {
			// created
			Expect(k8sCli.Create(testCtx, &targetItem)).Should(Succeed())

			var key = types.NamespacedName{
				Namespace: targetItem.Namespace,
				Name:      targetItem.Name,
			}

			// confirmed
			Eventually(func() error {
				if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
					if !apierrs.IsNotFound(err) {
						return err
					}
				}
				if !object.IsActivating(&targetItem) {
					return errors.Errorf("%s link isn't activated", key)
				}
				if devicelink.GetNodeExistedStatus(&targetItem.Status) != metav1.ConditionTrue {
					return errors.Errorf("could not find the corresponding node of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())

			// deleted the node
			Eventually(func() error {
				var node corev1.Node
				if err := k8sCli.Get(testCtx, types.NamespacedName{Name: targetNode}, &node); err != nil {
					if apierrs.IsNotFound(err) {
						return nil
					}
					return err
				}
				return k8sCli.Delete(testCtx, &node)
			}, 30, 1).Should(Succeed())

			// confirmed
			Eventually(func() error {
				if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
					if !apierrs.IsNotFound(err) {
						return err
					}
				}
				if !object.IsActivating(&targetItem) {
					return errors.Errorf("%s link isn't activated", key)
				}
				if devicelink.GetNodeExistedStatus(&targetItem.Status) != metav1.ConditionFalse {
					return errors.Errorf("should not find the corresponding node of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())

			// record
			targetNodeRecord = targetNode
		})

		K3dIt("should be changed if added the node back", func() {
			if targetNodeRecord == "" {
				Skip("cannot test because no recorder was found")
			}
			targetItem.Spec.Adaptor.Node = targetNodeRecord

			// created
			Expect(k8sCli.Create(testCtx, &targetItem)).Should(Succeed())

			var key = types.NamespacedName{
				Namespace: targetItem.Namespace,
				Name:      targetItem.Name,
			}

			// confirmed
			Eventually(func() error {
				if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
					if !apierrs.IsNotFound(err) {
						return err
					}
				}
				if !object.IsActivating(&targetItem) {
					return errors.Errorf("%s link isn't activated", key)
				}
				if devicelink.GetNodeExistedStatus(&targetItem.Status) != metav1.ConditionFalse {
					return errors.Errorf("should not find the corresponding node of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())

			// add node back
			err = GetLocalCluster().AddWorker(testRootDir, GinkgoWriter, targetNodeRecord)
			Expect(err).ToNot(HaveOccurred())

			// confirmed
			Eventually(func() error {
				if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
					if !apierrs.IsNotFound(err) {
						return err
					}
				}
				if !object.IsActivating(&targetItem) {
					return errors.Errorf("%s link isn't activated", key)
				}
				if devicelink.GetNodeExistedStatus(&targetItem.Status) != metav1.ConditionTrue {
					return errors.Errorf("could not find the corresponding node of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())
		})
	})

})
