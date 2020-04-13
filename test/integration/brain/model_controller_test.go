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
	"github.com/rancher/octopus/pkg/status/devicelink"
	"github.com/rancher/octopus/pkg/util/collection"
	"github.com/rancher/octopus/pkg/util/object"
	"github.com/rancher/octopus/test/util/content"
	"github.com/rancher/octopus/test/util/node"
)

// testing scenarios:
//	+ CRD instance
//		- validate if all instances have `edge.cattle.io/octopus-brain` finalizer
//	+ DeviceLink instance
//		- validate if target link change when deleting the model CRD instance
//		- validate if target link change when adding the lost model CRD instance back
var _ = Describe("Model controller", func() {
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

	Context("CRD instance", func() {

		It("should have finalizer", func() {
			// confirmed
			Eventually(func() error {
				var list apiextensionsv1.CustomResourceDefinitionList
				if err := k8sCli.List(testCtx, &list); err != nil {
					return err
				}
				for _, crd := range list.Items {
					if !collection.StringSliceContain(crd.Finalizers, "edge.cattle.io/octopus-brain") {
						return errors.Errorf("could not find corresponding finalizer from %s CRD", crd.Name)
					}
				}
				return nil
			}, 30, 1).Should(Succeed())
		})

	})

	Context("DeviceLink instance", func() {

		It("should be changed if deleting the model", func() {
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
				if devicelink.GetModelExistedStatus(&targetItem.Status) != metav1.ConditionTrue {
					return errors.Errorf("could not find the corresponding model of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())

			// deleted the model
			Eventually(func() error {
				return envtest.UninstallCRDs(k8sCfg, envtest.CRDInstallOptions{
					Paths: []string{
						filepath.Join(testRootDir, "adaptors", "dummy", "deploy", "manifests", "crd", "base"),
					},
				})
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
				if devicelink.GetModelExistedStatus(&targetItem.Status) != metav1.ConditionFalse {
					return errors.Errorf("should not find the corresponding model of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())
		})

		It("should be changed if adding the model back", func() {
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

			// added the model
			Eventually(func() error {
				var _, err = envtest.InstallCRDs(k8sCfg, envtest.CRDInstallOptions{
					Paths: []string{
						filepath.Join(testRootDir, "adaptors", "dummy", "deploy", "manifests", "crd", "base"),
					},
				})
				return err
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
				if devicelink.GetModelExistedStatus(&targetItem.Status) != metav1.ConditionTrue {
					return errors.Errorf("could not find the corresponding model of %s link", key)
				}
				return nil
			}, 30, 1).Should(Succeed())
		})

	})

})
