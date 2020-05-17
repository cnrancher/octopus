package limb

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	status "github.com/rancher/octopus/pkg/status/devicelink"
	"github.com/rancher/octopus/pkg/suctioncup/connection"
	"github.com/rancher/octopus/pkg/util/model"
	"github.com/rancher/octopus/pkg/util/object"
	"github.com/rancher/octopus/test/util/content"
)

// testing scenarios:
//	+ Corresponding adaptor
//		- validate if target link available when adding the corresponding adaptor
//		- validate if target link unavailable when deleting the corresponding adaptor
//	+ Corresponding device
//     	- validate if the device of target link create when adding the corresponding adaptor
var _ = Describe("DeviceLink controller", func() {
	var (
		mockingDummyAdaptor fakeDummyAdaptor

		targetModel     metav1.TypeMeta
		targetNamespace string
		targetItem      edgev1alpha1.DeviceLink
	)

	AfterEach(func() {
		_ = k8sCli.DeleteAllOf(testCtx, &edgev1alpha1.DeviceLink{}, client.InNamespace(targetNamespace))
	})

	BeforeEach(func() {
		targetModel = metav1.TypeMeta{
			Kind:       "DummySpecialDevice",
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
					Name: mockingDummyAdaptor.GetName(),
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
							"protocol": map[string]interface{}{
								"location": "living-room-fan",
							},
							"gear": "slow",
							"on":   true,
						},
					),
				},
			},
		}
	})

	Context("Corresponding adaptor", func() {

		It("should be validated if add the adaptor", func() {
			// created
			Expect(k8sCli.Create(testCtx, &targetItem)).Should(Succeed())

			// simulated that has completed the validation of node and model
			targetItem.Status.Adaptor = targetItem.Spec.Adaptor
			status.SuccessOnNodeExisted(&targetItem.Status)
			targetItem.Status.Model = targetItem.Spec.Model
			status.SuccessOnModelExisted(&targetItem.Status)
			Expect(k8sCli.Status().Update(testCtx, &targetItem)).Should(Succeed())

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
				if status.GetAdaptorExistedStatus(&targetItem.Status) != metav1.ConditionFalse {
					return errors.Errorf("should not find the corresponding adaptor of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())

			// simulated that added the corresponding adaptor
			testAdaptors.Put(mockingDummyAdaptor)
			testEventQueue.GetAdaptorNotifier().NoticeAdaptorRegistered(mockingDummyAdaptor.GetName())

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
				if status.GetAdaptorExistedStatus(&targetItem.Status) != metav1.ConditionTrue {
					return errors.Errorf("could not find the corresponding adaptor of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())
		})

		It("should be invalidated if delete the adaptor", func() {
			// created
			Expect(k8sCli.Create(testCtx, &targetItem)).Should(Succeed())

			// simulated that has completed the validation of node and model
			targetItem.Status.Adaptor = targetItem.Spec.Adaptor
			status.SuccessOnNodeExisted(&targetItem.Status)
			targetItem.Status.Model = targetItem.Spec.Model
			status.SuccessOnModelExisted(&targetItem.Status)
			Expect(k8sCli.Status().Update(testCtx, &targetItem)).Should(Succeed())

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
				if status.GetAdaptorExistedStatus(&targetItem.Status) != metav1.ConditionTrue {
					return errors.Errorf("could not find the corresponding adaptor of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())

			// simulated that added the corresponding adaptor
			testAdaptors.Delete(mockingDummyAdaptor.GetName())
			testEventQueue.GetAdaptorNotifier().NoticeAdaptorUnregistered(mockingDummyAdaptor.GetName())

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
				if status.GetAdaptorExistedStatus(&targetItem.Status) != metav1.ConditionFalse {
					return errors.Errorf("should not find the corresponding adaptor of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())
		})

	})

	Context("Corresponding device", func() {

		It("should be created if add the adaptor", func() {
			// created
			Expect(k8sCli.Create(testCtx, &targetItem)).Should(Succeed())

			// simulated that added the corresponding adaptor
			testAdaptors.Put(mockingDummyAdaptor)

			// simulated that has completed the validation of node and model
			targetItem.Status.Adaptor = targetItem.Spec.Adaptor
			status.SuccessOnNodeExisted(&targetItem.Status)
			targetItem.Status.Model = targetItem.Spec.Model
			status.SuccessOnModelExisted(&targetItem.Status)
			Expect(k8sCli.Status().Update(testCtx, &targetItem)).Should(Succeed())

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
				if status.GetDeviceCreatedStatus(&targetItem.Status) != metav1.ConditionTrue {
					return errors.Errorf("could not find the corresponding adaptor of %s link", key)
				}

				var targetDevice, _ = model.NewInstanceOfTypeMeta(targetModel)
				if err := k8sCli.Get(testCtx, key, &targetDevice); err != nil {
					return err
				}
				if !object.IsActivating(&targetDevice) {
					return errors.Errorf("%s %s device isn't activated", key, targetModel)
				}
				return nil
			}, 30, 1).Should(Succeed())
		})

	})

})

type fakeDummyAdaptor string

func (a fakeDummyAdaptor) GetName() string {
	return "adaptors.edge.cattle.io/dummy"
}

func (a fakeDummyAdaptor) GetEndpoint() string {
	return "dummy.socket"
}

func (a fakeDummyAdaptor) Stop() error {
	return nil
}

func (a fakeDummyAdaptor) CreateConnection(name types.NamespacedName) (bool, error) {
	return false, nil
}

func (a fakeDummyAdaptor) DeleteConnection(name types.NamespacedName) bool {
	return false
}

func (a fakeDummyAdaptor) GetConnection(name types.NamespacedName) connection.Connection {
	return fakeDummyConnection(name)
}

type fakeDummyConnection types.NamespacedName

func (c fakeDummyConnection) GetAdaptorName() string {
	return "adaptors.edge.cattle.io/dummy"
}

func (c fakeDummyConnection) GetName() types.NamespacedName {
	return types.NamespacedName(c)
}

func (c fakeDummyConnection) Stop() error {
	return nil
}

func (c fakeDummyConnection) Send(parameters []byte, model *metav1.TypeMeta, device []byte) error {
	return nil
}
