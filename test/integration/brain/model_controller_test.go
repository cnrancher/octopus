package brain

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/brain/controller"
	"github.com/rancher/octopus/pkg/status/devicelink"
	"github.com/rancher/octopus/pkg/util/collection"
	"github.com/rancher/octopus/pkg/util/object"
	"github.com/rancher/octopus/test/util/content"
	"github.com/rancher/octopus/test/util/node"
)

var _ = Describe("Model controller", func() {
	var (
		namespace = "default"
		validNode string
	)

	BeforeEach(func() {
		var err error
		validNode, err = node.GetValidWorker(ctx, k8sCli)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		_ = k8sCli.DeleteAllOf(ctx, &edgev1alpha1.DeviceLink{}, client.InNamespace(namespace))
	})

	Context("CRD instance", func() {

		It("should have finalizer", func() {
			// confirmed
			Eventually(func() error {
				var list apiextensionsv1.CustomResourceDefinitionList
				if err := k8sCli.List(ctx, &list); err != nil {
					return err
				}
				for _, crd := range list.Items {
					if !collection.StringSliceContain(crd.Finalizers, controller.ReconcilingModel) {
						return errors.Errorf("could not find corresponding finalizer from %s CRD", crd.Name)
					}
				}
				return nil
			}, 30, 1).Should(Succeed())
		})

	})

	Context("DeviceLink instance", func() {

		It("should be changed if deleted the model", func() {
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
				if devicelink.GetModelExistedStatus(&item.Status) != metav1.ConditionTrue {
					return errors.Errorf("could not find the corresponding model of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())

			// deleted the model
			Eventually(func() error {
				return envtest.UninstallCRDs(k8sCfg, envtest.CRDInstallOptions{
					Paths: []string{
						filepath.Join(rootDir, "adaptors", "dummy", "deploy", "manifests", "crd", "base"),
					},
				})
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
				if devicelink.GetModelExistedStatus(&item.Status) != metav1.ConditionFalse {
					return errors.Errorf("should not find the corresponding model of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())
		})

		It("should be changed if added the model back", func() {
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
				if devicelink.GetModelExistedStatus(&item.Status) != metav1.ConditionFalse {
					return errors.Errorf("should not find the corresponding model of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())

			// added the model
			Eventually(func() error {
				var _, err = envtest.InstallCRDs(k8sCfg, envtest.CRDInstallOptions{
					Paths: []string{
						filepath.Join(rootDir, "adaptors", "dummy", "deploy", "manifests", "crd", "base"),
					},
				})
				return err
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
				if devicelink.GetModelExistedStatus(&item.Status) != metav1.ConditionTrue {
					return errors.Errorf("could not find the corresponding model of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())
		})

	})

})
