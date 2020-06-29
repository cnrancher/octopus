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
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
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
		mockingAdaptor fakeAdaptor

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
					Name: mockingAdaptor.GetName(),
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
			targetItem.SuccessOnNodeExisted(nil)
			targetItem.SuccessOnModelExisted()
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
				if targetItem.GetAdaptorExistedStatus() != metav1.ConditionFalse {
					return errors.Errorf("should not find the corresponding adaptor of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())

			// simulated that added the corresponding adaptor
			testAdaptors.Put(mockingAdaptor)
			testEventQueue.GetAdaptorNotifier().NoticeAdaptorRegistered(mockingAdaptor.GetName())

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
				if targetItem.GetAdaptorExistedStatus() != metav1.ConditionTrue {
					return errors.Errorf("could not find the corresponding adaptor of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())
		})

		It("should be invalidated if delete the adaptor", func() {
			// created
			Expect(k8sCli.Create(testCtx, &targetItem)).Should(Succeed())

			// simulated that has completed the validation of node and model
			targetItem.SuccessOnNodeExisted(nil)
			targetItem.SuccessOnModelExisted()
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
				if targetItem.GetAdaptorExistedStatus() != metav1.ConditionTrue {
					return errors.Errorf("could not find the corresponding adaptor of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())

			// simulated that added the corresponding adaptor
			testAdaptors.Delete(mockingAdaptor.GetName())
			testEventQueue.GetAdaptorNotifier().NoticeAdaptorUnregistered(mockingAdaptor.GetName())

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
				if targetItem.GetAdaptorExistedStatus() != metav1.ConditionFalse {
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
			testAdaptors.Put(mockingAdaptor)

			// simulated that has completed the validation of node and model
			targetItem.SuccessOnNodeExisted(nil)
			targetItem.SuccessOnModelExisted()
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
				if targetItem.GetDeviceCreatedStatus() != metav1.ConditionTrue {
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

type fakeAdaptor string

func (a fakeAdaptor) GetName() string {
	return "adaptors.edge.cattle.io/dummy"
}

func (a fakeAdaptor) GetEndpoint() string {
	return "dummy.sock"
}

func (a fakeAdaptor) Stop() error {
	return nil
}

func (a fakeAdaptor) CreateConnection(name types.NamespacedName) (overwritten bool, conn connection.Connection, err error) {
	return false, fakeConnection(name), nil
}

func (a fakeAdaptor) DeleteConnection(name types.NamespacedName) bool {
	return false
}

type fakeConnection types.NamespacedName

func (c fakeConnection) GetAdaptorName() string {
	return "adaptors.edge.cattle.io/dummy"
}

func (c fakeConnection) GetName() types.NamespacedName {
	return types.NamespacedName(c)
}

func (c fakeConnection) Stop() error {
	return nil
}

func (c fakeConnection) IsStop() bool {
	return false
}

func (c fakeConnection) Send(*metav1.TypeMeta, []byte, map[string]*api.ConnectRequestReferenceEntry) error {
	return nil
}
