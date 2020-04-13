package brain

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/status/devicelink"
	"github.com/rancher/octopus/pkg/util/object"
	"github.com/rancher/octopus/test/util/content"
	"github.com/rancher/octopus/test/util/node"
)

// testing scenarios:
//	+ DeviceLink instance
//		- validate if target link can be managed
//	+ Corresponding node
//		- validate if the target link unavailable when assigning to an invalid Node
//	+ Corresponding model
//		- validate if the target link unavailable when requesting an invalid model CRD
var _ = Describe("DeviceLink controller", func() {
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

	Context("DeviceLink instance", func() {

		It("should be managed", func() {
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
				return nil
			}, 30, 1).Should(Succeed())

			// updated
			Eventually(func() error {
				if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
					if !apierrs.IsNotFound(err) {
						return err
					}
				}

				targetItem.Spec.Template.Labels = map[string]string{
					"l2": "v2",
				}
				return k8sCli.Update(testCtx, &targetItem)
			}, 30, 1).Should(Succeed())

			// confirmed
			Eventually(func() error {
				if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
					return err
				}
				if !object.IsActivating(&targetItem) {
					return errors.Errorf("%s link isn't activated", key)
				}
				if !reflect.DeepEqual(targetItem.Spec.Template.Labels, map[string]string{
					"l2": "v2",
				}) {
					return errors.Errorf("%s link isn't updated", key)
				}
				return nil
			}, 30, 1).Should(Succeed())

			// deleted
			Expect(k8sCli.Delete(testCtx, &targetItem)).Should(Succeed())

			// confirmed
			Eventually(func() error {
				var err = k8sCli.Get(testCtx, key, &targetItem)
				if !apierrs.IsNotFound(err) {
					return errors.Wrapf(err, "link is existed")
				}
				return nil
			}, 30, 1).Should(Succeed())
		})

	})

	Context("Corresponding node", func() {

		BeforeEach(func() {
			targetNode, err = node.GetInvalidWorker(testCtx, k8sCli)
			Expect(err).ToNot(HaveOccurred())

			targetItem.Spec.Adaptor.Node = targetNode
		})

		It("should be invalidated if the node isn't existed", func() {
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
		})

	})

	Context("Corresponding model", func() {

		BeforeEach(func() {
			targetModel = metav1.TypeMeta{
				Kind:       "Missed",
				APIVersion: "devices.edge.cattle.io/v1alpha1",
			}

			targetItem.Spec.Model = targetModel
		})

		It("should be invalidated if the model isn't existed", func() {
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
				if devicelink.GetModelExistedStatus(&targetItem.Status) != metav1.ConditionFalse {
					return errors.Errorf("should not find the corresponding model of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())
		})

	})

})
