package usability_test

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1alpha1 "github.com/rancher/octopus/adaptors/dummy/api/v1alpha1"
	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/test/util/content"
	"github.com/rancher/octopus/test/util/exec"
	"github.com/rancher/octopus/test/util/node"
)

/*
	NB(uuuxxllj): the following cases focus on DummySpecialDevice.
*/

var (
	testDeviceLink edgev1alpha1.DeviceLink
)

var _ = Describe("verify usability", func() {

	BeforeEach(func() {
		// create device link
		deployDummyDeviceLink()
	})

	AfterEach(func() {
		// delete device link, ignore error
		_ = k8sCli.DeleteAllOf(testCtx, &edgev1alpha1.DeviceLink{}, client.InNamespace(testDeviceLink.Namespace))
	})

	Context("modify dummy device link spec", func() {

		Specify("if invalid node spec", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when invalid node spec", invalidNodeSpec)

			By("then node of the device link is not found", isNodeExistedFalse)

			By("when correct node spec", correctNodeSpec)

			By("then node of the device link is found", isNodeExistedTrue)

		})

		Specify("if invalid model spec", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when invalid model spec", invalidModelSpec)

			By("then model of the device link is not found", isModelExistedFalse)

			By("when correct model spec", correctModelSpec)

			By("then model of the device link is found", isModelExistedTrue)

		})

		Specify("if invalid adaptor spec", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when invalid adaptor spec", invalidAdaptorSpec)

			By("then adaptor of the device link is not found", isAdaptorExistedFalse)

			By("when correct adaptor spec", correctAdaptorSpec)

			By("then adaptor of the device link is found", isAdaptorExistedTrue)

		})

		Specify("if invalid device spec", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when invalid device spec", invalidDeviceSpec)

			By("then the device link is not connected", isDeviceConnectedFalse)

			By("when correct device spec", correctDeviceSpec)

			By("then the device link is connected", isDeviceConnectedTrue)

		})

		Specify("if switch device gear to fast", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when switch device gear to fast", switchGearToFast)

			By("then the device gear is fast", isDeviceGearFast)

		})

		Specify("if deploy device link without turn-on spec", func() {

			By("when deploy device link without turn-on spec", deployDeviceLinkWithoutTurnOnSpec)

			By("then the device link fail to be created", isDeviceCreatedFalse)

			By("when add turn-on spec", addTurnOnSpec)

			By("then the device link is connected", isDeviceConnectedTrue)

		})
	})

	Context("restart limbs/adaptors pods", func() {

		Specify("if delete dummy adaptor pods", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when delete dummy adaptor pods", deleteDummyAdaptorPods)

			By("then adaptor of the device link is not found", isAdaptorExistedFalse)

		})

		Specify("if delete octopus limbs pods", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when delete octopus limbs pod", deleteLimbsPods)

			By("then the dummy adaptor pods become error", isDummyAdaptorPodsError)

		})
	})

	Specify("if delete dummy device model", func() {

		By("given the device link is connected", isDeviceConnectedTrue)

		By("when delete dummy device model", deleteDummyDeviceModel)

		By("then model of the device link is not found", isModelExistedFalse)

		By("when redeploy dummy device model", redeployDummyDeviceModel)

		By("then the device link is connected", isDeviceConnectedTrue)

	})

	Specify("if delete cluster node", func() {

		By("given the device link is connected", isDeviceConnectedTrue)

		By("when delete corresponding cluster node", deleteCorrespondingNode)

		By("then node of the device link is not found", isNodeExistedFalse)

	})
})

type judgeFunc func(edgev1alpha1.DeviceLink) bool

func doDeviceLinkJudgment(judge judgeFunc) {
	Eventually(func() bool {
		var deviceLinkKey = types.NamespacedName{
			Name:      testDeviceLink.Name,
			Namespace: testDeviceLink.Namespace,
		}
		if err := k8sCli.Get(testCtx, deviceLinkKey, &testDeviceLink); err != nil {
			Fail(err.Error())
		}
		return judge(testDeviceLink)
	}, 300, 1).Should(BeTrue())
}

func deployDummyDeviceLink() {
	var targetNode, err = node.GetValidWorker(testCtx, k8sCli)
	Expect(err).ShouldNot(HaveOccurred())

	testDeviceLink = edgev1alpha1.DeviceLink{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    "default",
			GenerateName: "test-fan-",
		},
		Spec: edgev1alpha1.DeviceLinkSpec{
			Adaptor: edgev1alpha1.DeviceAdaptor{
				Node: targetNode,
				Name: "adaptors.edge.cattle.io/dummy",
			},
			Model: metav1.TypeMeta{
				Kind:       "DummySpecialDevice",
				APIVersion: "devices.edge.cattle.io/v1alpha1",
			},
			Template: edgev1alpha1.DeviceTemplateSpec{
				Spec: content.ToRawExtension(
					map[string]interface{}{
						"protocol": map[string]interface{}{
							"location": "living-room",
						},
						"gear": "slow",
						"on":   true,
					},
				),
			},
		},
	}
	Expect(k8sCli.Create(testCtx, &testDeviceLink)).Should(Succeed())
}

func correctNodeSpec() {
	var targetNode, err = node.GetValidWorker(testCtx, k8sCli)
	Expect(err).ShouldNot(HaveOccurred())
	var deviceLinkKey = types.NamespacedName{
		Namespace: testDeviceLink.Namespace,
		Name:      testDeviceLink.Name,
	}
	if err := k8sCli.Get(testCtx, deviceLinkKey, &testDeviceLink); err != nil {
		Fail(err.Error())
	}
	patch := []byte(fmt.Sprintf(`{"spec":{"adaptor":{"node":"%s"}}}`, targetNode))
	Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func invalidNodeSpec() {
	var deviceLinkKey = types.NamespacedName{
		Namespace: testDeviceLink.Namespace,
		Name:      testDeviceLink.Name,
	}
	if err := k8sCli.Get(testCtx, deviceLinkKey, &testDeviceLink); err != nil {
		Fail(err.Error())
	}
	patch := []byte(`{"spec":{"adaptor":{"node":"wrong-node"}}}`)
	Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func isNodeExistedTrue() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetNodeExistedStatus() == metav1.ConditionTrue
	}
	doDeviceLinkJudgment(judge)
}

func isNodeExistedFalse() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetNodeExistedStatus() == metav1.ConditionFalse
	}
	doDeviceLinkJudgment(judge)
}

func correctModelSpec() {
	var deviceLinkKey = types.NamespacedName{
		Namespace: testDeviceLink.Namespace,
		Name:      testDeviceLink.Name,
	}
	if err := k8sCli.Get(testCtx, deviceLinkKey, &testDeviceLink); err != nil {
		Fail(err.Error())
	}
	patch := []byte(`{"spec":{"model":{"apiVersion":"devices.edge.cattle.io/v1alpha1"}}}`)
	Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func invalidModelSpec() {
	var deviceLinkKey = types.NamespacedName{
		Namespace: testDeviceLink.Namespace,
		Name:      testDeviceLink.Name,
	}
	if err := k8sCli.Get(testCtx, deviceLinkKey, &testDeviceLink); err != nil {
		Fail(err.Error())
	}
	patch := []byte(`{"spec":{"model":{"apiVersion":"wrong-apiVersion"}}}`)
	Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func isModelExistedTrue() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetModelExistedStatus() == metav1.ConditionTrue
	}
	doDeviceLinkJudgment(judge)
}

func isModelExistedFalse() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetModelExistedStatus() == metav1.ConditionFalse
	}
	doDeviceLinkJudgment(judge)
}

func correctAdaptorSpec() {
	var deviceLinkKey = types.NamespacedName{
		Namespace: testDeviceLink.Namespace,
		Name:      testDeviceLink.Name,
	}
	if err := k8sCli.Get(testCtx, deviceLinkKey, &testDeviceLink); err != nil {
		Fail(err.Error())
	}
	patch := []byte(`{"spec":{"adaptor":{"name":"adaptors.edge.cattle.io/dummy"}}}`)
	Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func invalidAdaptorSpec() {
	var deviceLinkKey = types.NamespacedName{
		Namespace: testDeviceLink.Namespace,
		Name:      testDeviceLink.Name,
	}
	if err := k8sCli.Get(testCtx, deviceLinkKey, &testDeviceLink); err != nil {
		Fail(err.Error())
	}
	patch := []byte(`{"spec":{"adaptor":{"name":"wrong-adaptor-name"}}}`)
	Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func isAdaptorExistedTrue() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetAdaptorExistedStatus() == metav1.ConditionTrue
	}
	doDeviceLinkJudgment(judge)
}

func isAdaptorExistedFalse() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetAdaptorExistedStatus() == metav1.ConditionFalse
	}
	doDeviceLinkJudgment(judge)
}

func correctDeviceSpec() {
	var deviceLinkKey = types.NamespacedName{
		Namespace: testDeviceLink.Namespace,
		Name:      testDeviceLink.Name,
	}
	if err := k8sCli.Get(testCtx, deviceLinkKey, &testDeviceLink); err != nil {
		Fail(err.Error())
	}
	patch := []byte(`{"spec":{"template":{"spec":{"gear":"slow"}}}}`)
	Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func invalidDeviceSpec() {
	var deviceLinkKey = types.NamespacedName{
		Namespace: testDeviceLink.Namespace,
		Name:      testDeviceLink.Name,
	}
	if err := k8sCli.Get(testCtx, deviceLinkKey, &testDeviceLink); err != nil {
		Fail(err.Error())
	}
	patch := []byte(`{"spec":{"template":{"spec":{"gear":"wrong-gear"}}}}`)
	Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func isDeviceConnectedTrue() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetDeviceConnectedStatus() == metav1.ConditionTrue
	}
	doDeviceLinkJudgment(judge)
}

func isDeviceConnectedFalse() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetDeviceConnectedStatus() == metav1.ConditionFalse
	}
	doDeviceLinkJudgment(judge)
}

func switchGearToFast() {
	var deviceLinkKey = types.NamespacedName{
		Namespace: testDeviceLink.Namespace,
		Name:      testDeviceLink.Name,
	}
	if err := k8sCli.Get(testCtx, deviceLinkKey, &testDeviceLink); err != nil {
		Fail(err.Error())
	}
	patch := []byte(`{"spec":{"template":{"spec":{"gear":"fast"}}}}`)
	Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func isDeviceGearFast() {
	var deviceList apiv1alpha1.DummySpecialDeviceList
	Eventually(func() bool {
		if err := k8sCli.List(testCtx, &deviceList, client.InNamespace("default")); err != nil {
			Fail(err.Error())
		}
		for _, device := range deviceList.Items {
			return device.Status.Gear == "fast"
		}
		return false
	}, 300, 1).Should(BeTrue())
}

func deployDeviceLinkWithoutTurnOnSpec() {
	var targetNode, err = node.GetValidWorker(testCtx, k8sCli)
	Expect(err).ShouldNot(HaveOccurred())

	testDeviceLink = edgev1alpha1.DeviceLink{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    "default",
			GenerateName: "test-fan-",
		},
		Spec: edgev1alpha1.DeviceLinkSpec{
			Adaptor: edgev1alpha1.DeviceAdaptor{
				Node: targetNode,
				Name: "adaptors.edge.cattle.io/dummy",
			},
			Model: metav1.TypeMeta{
				Kind:       "DummySpecialDevice",
				APIVersion: "devices.edge.cattle.io/v1alpha1",
			},
			Template: edgev1alpha1.DeviceTemplateSpec{
				Spec: content.ToRawExtension(
					map[string]interface{}{
						"protocol": map[string]interface{}{
							"location": "living-room",
						},
						// "on": true,
					},
				),
			},
		},
	}
	Expect(k8sCli.Create(testCtx, &testDeviceLink)).Should(Succeed())
}

func isDeviceCreatedFalse() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetDeviceCreatedStatus() == metav1.ConditionFalse
	}
	doDeviceLinkJudgment(judge)
}

func addTurnOnSpec() {
	var deviceLinkKey = types.NamespacedName{
		Namespace: testDeviceLink.Namespace,
		Name:      testDeviceLink.Name,
	}
	if err := k8sCli.Get(testCtx, deviceLinkKey, &testDeviceLink); err != nil {
		Fail(err.Error())
	}
	patch := []byte(`{"spec":{"template":{"spec":{"on":true}}}}`)
	Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func deleteDummyAdaptorPods() {
	Expect(k8sCli.DeleteAllOf(testCtx, &corev1.Pod{}, client.InNamespace("octopus-system"), client.MatchingLabels{"app.kubernetes.io/name": "octopus-adaptor-dummy"})).
		Should(Succeed())
}

func deleteLimbsPods() {
	Expect(k8sCli.DeleteAllOf(testCtx, &corev1.Pod{}, client.InNamespace("octopus-system"), client.MatchingLabels{"app.kubernetes.io/component": "limb"})).
		Should(Succeed())
}

func isDummyAdaptorPodsError() {
	var podList corev1.PodList
	Eventually(func() bool {
		if err := k8sCli.List(testCtx, &podList, client.InNamespace("octopus-system"), client.MatchingLabels{"app.kubernetes.io/name": "octopus-adaptor-dummy"}); err != nil {
			Fail(err.Error())
		}
		for _, pod := range podList.Items {
			for _, condition := range pod.Status.Conditions {
				if condition.Type == "Ready" && condition.Status == "False" {
					return true
				}
			}
		}
		return false
	}, 300, 1).Should(BeTrue())
}

func deleteCorrespondingNode() {
	var correspondingNode = corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: testDeviceLink.Spec.Adaptor.Node,
		},
	}
	Expect(k8sCli.Delete(testCtx, &correspondingNode)).Should(Succeed())
}

func deleteDummyDeviceModel() {
	var crd = v1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dummyspecialdevices.devices.edge.cattle.io",
		},
	}
	Expect(k8sCli.Delete(testCtx, &crd)).Should(Succeed())
}

func redeployDummyDeviceModel() {
	Expect(exec.RunKubectl(nil, GinkgoWriter, "apply", "-f", filepath.Join(testCurrDir, "deploy", "manifests", "crd", "base", "devices.edge.cattle.io_dummyspecialdevices.yaml"))).
		Should(Succeed())
}
