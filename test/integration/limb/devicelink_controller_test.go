package limb

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/status"
	"github.com/rancher/octopus/pkg/util/object"
	"github.com/rancher/octopus/test/util/content"
)

var _ = Describe("DeviceLink controller", func() {
	var namespace = "default"

	Context("Corresponding adaptor", func() {
		var adaptorName = "adaptors.edge.cattle.io/dummy"

		It("should be validated if add the adaptor", func() {
			var item = edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    namespace,
					GenerateName: "test-",
				},
				Spec: edgev1alpha1.DeviceLinkSpec{
					Adaptor: edgev1alpha1.DeviceAdaptor{
						Node: adaptorOnNode,
						Name: adaptorName,
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
						ObjectMeta: metav1.ObjectMeta{
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

			// simulated that has completed the validation of node and model
			status.SuccessOnNodeExisted(&item.Status)
			status.SuccessOnModelExisted(&item.Status)
			Expect(k8sCli.Status().Update(ctx, &item)).Should(Succeed())

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
				if status.GetAdaptorExistedStatus(&item.Status) != metav1.ConditionFalse {
					return errors.Errorf("should not find the corresponding adaptor of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())

			// simulated that added the corresponding adaptor
			adaptorMgr.AddAdaptor(adaptorName)

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
				if status.GetAdaptorExistedStatus(&item.Status) != metav1.ConditionTrue {
					return errors.Errorf("could not find the corresponding adaptor of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())
		})

		It("should be invalidated if delete the adaptor", func() {
			var item = edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    namespace,
					GenerateName: "test-",
				},
				Spec: edgev1alpha1.DeviceLinkSpec{
					Adaptor: edgev1alpha1.DeviceAdaptor{
						Node: adaptorOnNode,
						Name: adaptorName,
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
						ObjectMeta: metav1.ObjectMeta{
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

			// simulated that has completed the validation of node and model
			status.SuccessOnNodeExisted(&item.Status)
			status.SuccessOnModelExisted(&item.Status)
			Expect(k8sCli.Status().Update(ctx, &item)).Should(Succeed())

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
				if status.GetAdaptorExistedStatus(&item.Status) != metav1.ConditionTrue {
					return errors.Errorf("could not find the corresponding adaptor of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())

			// simulated that added the corresponding adaptor
			adaptorMgr.DeleteAdaptor(adaptorName)

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
				if status.GetAdaptorExistedStatus(&item.Status) != metav1.ConditionFalse {
					return errors.Errorf("should not find the corresponding adaptor of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())
		})

	})

	XContext("Corresponding device", func() {
		// TODO
	})

})
