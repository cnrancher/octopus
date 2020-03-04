package brain

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/status/devicelink"
	"github.com/rancher/octopus/pkg/util/object"
	"github.com/rancher/octopus/test/util/content"
	"github.com/rancher/octopus/test/util/node"
)

var _ = Describe("DeviceLink controller", func() {
	var (
		namespace = "default"
		validNode string
	)

	BeforeEach(func() {
		var err error
		validNode, err = node.GetValidWorker(ctx, k8sCli)
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("DeviceLink instance", func() {

		It("should be managed", func() {
			var item = edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    namespace,
					GenerateName: "test-",
				},
				Spec: edgev1alpha1.DeviceLinkSpec{
					Adaptor: edgev1alpha1.DeviceAdaptor{
						Node: validNode,
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
				return nil
			}, 30, 1).Should(Succeed())

			// updated
			Eventually(func() error {
				if err := k8sCli.Get(ctx, key, &item); err != nil {
					if !apierrs.IsNotFound(err) {
						return err
					}
				}

				item.Spec.Template.Labels = map[string]string{
					"l2": "v2",
				}
				return k8sCli.Update(ctx, &item)
			}, 30, 1).Should(Succeed())

			// confirmed
			Eventually(func() error {
				if err := k8sCli.Get(ctx, key, &item); err != nil {
					return err
				}
				if !object.IsActivating(&item) {
					return errors.Errorf("%s link isn't activated", key)
				}
				if !reflect.DeepEqual(item.Spec.Template.Labels, map[string]string{
					"l2": "v2",
				}) {
					return errors.Errorf("%s link isn't updated", key)
				}
				return nil
			}, 30, 1).Should(Succeed())

			// deleted
			Expect(k8sCli.Delete(ctx, &item)).Should(Succeed())

			// confirmed
			Eventually(func() error {
				var err = k8sCli.Get(ctx, key, &item)
				if !apierrs.IsNotFound(err) {
					return errors.Wrapf(err, "link is existed")
				}
				return nil
			}, 30, 1).Should(Succeed())
		})

	})

	Context("Corresponding node", func() {

		It("should be invalidated if the node isn't existed", func() {
			var invalidNode, err = node.GetInvalidWorker(ctx, k8sCli)
			Expect(err).ShouldNot(HaveOccurred())

			var item = edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    namespace,
					GenerateName: "test-",
				},
				Spec: edgev1alpha1.DeviceLinkSpec{
					Adaptor: edgev1alpha1.DeviceAdaptor{
						Node: invalidNode,
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
				if devicelink.GetNodeExistedStatus(&item.Status) != metav1.ConditionFalse {
					return errors.Errorf("should not find the corresponding node of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())
		})

	})

	Context("Corresponding model", func() {

		It("should be invalidated if the model isn't existed", func() {
			var invalidModel = metav1.TypeMeta{
				Kind:       "Missed",
				APIVersion: "devices.edge.cattle.io/v1alpha1",
			}

			var item = edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    namespace,
					GenerateName: "test-",
				},
				Spec: edgev1alpha1.DeviceLinkSpec{
					Adaptor: edgev1alpha1.DeviceAdaptor{
						Node: validNode,
						Name: "adaptors.edge.cattle.io/dummy",
						Parameters: content.ToRawExtension(
							map[string]string{
								"ip": "1.2.3.4",
							},
						),
					},
					Model: invalidModel,
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
				if devicelink.GetModelExistedStatus(&item.Status) != metav1.ConditionFalse {
					return errors.Errorf("should not find the corresponding model of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())
		})

	})

})
