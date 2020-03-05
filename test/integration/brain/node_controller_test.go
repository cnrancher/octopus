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
	"github.com/rancher/octopus/pkg/brain/controller"
	"github.com/rancher/octopus/pkg/status/devicelink"
	"github.com/rancher/octopus/pkg/util/collection"
	"github.com/rancher/octopus/pkg/util/object"
	"github.com/rancher/octopus/test/util/content"
	"github.com/rancher/octopus/test/util/node"
)

var _ = Describe("Node controller", func() {
	var namespace = "default"

	AfterEach(func() {
		_ = k8sCli.DeleteAllOf(ctx, &edgev1alpha1.DeviceLink{}, client.InNamespace(namespace))
	})

	Context("Node instance", func() {

		It("should have finalizer", func() {
			// confirmed
			Eventually(func() (bool, error) {
				var list corev1.NodeList
				if err := k8sCli.List(ctx, &list); err != nil {
					return false, err
				}
				for _, node := range list.Items {
					if !collection.StringSliceContain(node.Finalizers, controller.ReconcilingNode) {
						return false, errors.Errorf("could not find corresponding finalizer from %s node", node.Name)
					}
				}
				return true, nil
			}, 30, 1).Should(BeTrue())
		})

	})

	Context("DeviceLink instance", func() {

		It("should be changed if deleted the node", func() {
			var existedNode, err = node.GetValidWorker(ctx, k8sCli)
			Expect(err).ShouldNot(HaveOccurred())

			var item = edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    namespace,
					GenerateName: "test-",
				},
				Spec: edgev1alpha1.DeviceLinkSpec{
					Adaptor: edgev1alpha1.DeviceAdaptor{
						Node: existedNode,
						Name: "adaptors.edge.cattle.io/dummy",
						Parameters: content.ToRawExtension(
							map[string]string{
								"ip": "1.2.3.4",
							},
						),
					},
					Model: metav1.TypeMeta{
						Kind:       "DummyDevice",
						APIVersion: "devices.edge.cattle.io/v1alpha1",
					},
					Template: edgev1alpha1.DeviceTemplateSpec{
						DeviceMeta: edgev1alpha1.DeviceMeta{
							Labels: map[string]string{
								"l1": "v1",
							},
						},
						Spec: content.ToRawExtension(
							map[string]interface{}{
								"gear": "slow",
								`"on"`: true,
							},
						),
					},
				},
			}

			// created
			Expect(k8sCli.Create(ctx, &item)).Should(Succeed())

			var key = types.NamespacedName{
				Namespace: item.Namespace,
				Name:      item.Name,
			}

			// confirmed
			Eventually(func() error {
				if err := k8sCli.Get(ctx, key, &item); err != nil {
					if !apierrs.IsNotFound(err) {
						return err
					}
				}
				if !object.IsActivating(&item) {
					return errors.Errorf("%s link isn't activated", key)
				}
				if devicelink.GetNodeExistedStatus(&item.Status) != metav1.ConditionTrue {
					return errors.Errorf("could not find the corresponding node of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())

			// deleted the node
			Eventually(func() error {
				var node corev1.Node
				if err := k8sCli.Get(ctx, types.NamespacedName{Name: existedNode}, &node); err != nil {
					if apierrs.IsNotFound(err) {
						return nil
					}
					return err
				}
				return k8sCli.Delete(ctx, &node)
			}, 30, 1).Should(Succeed())

			// confirmed
			Eventually(func() error {
				if err := k8sCli.Get(ctx, key, &item); err != nil {
					if !apierrs.IsNotFound(err) {
						return err
					}
				}
				if !object.IsActivating(&item) {
					return errors.Errorf("%s link isn't activated", key)
				}
				if devicelink.GetNodeExistedStatus(&item.Status) != metav1.ConditionFalse {
					return errors.Errorf("should not find the corresponding node of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())
		})

		XIt("should be changed if added the node back", func() {
			// TODO
		})
	})

})
