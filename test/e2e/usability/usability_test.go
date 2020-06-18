package usability

import (
	"math/rand"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/status/devicelink"
	"github.com/rancher/octopus/pkg/util/object"
	"github.com/rancher/octopus/test/util/content"
	"github.com/rancher/octopus/test/util/exec"
	"github.com/rancher/octopus/test/util/node"
)

var _ = Describe("verify usability", func() {

	BeforeEach(installOctopus)

	AfterEach(uninstallOctopus, 120)

	Specify("if deleted the node that hosted limb", func() {

		By("given the Octopus is available", isOctopusAvailable)

		By("when delete the node that hosted limb", deleteNodeHostingLimb)

		By("then the Octopus is still available", isOctopusAvailable)

	})

	Specify("if deleted the node that hosted brain", func() {

		By("given the Octopus is available", isOctopusAvailable)

		By("when delete the node that hosted brain", deleteNodeHostingBrain)

		By("then the Octopus is still available", isOctopusAvailable)

	})

})

func installOctopus() {
	// install octopus
	Expect(exec.RunKubectl(nil, GinkgoWriter, "apply", "-f", filepath.Join(testRootDir, "deploy", "e2e", "all_in_one.yaml"))).
		Should(Succeed())

	// install dummy adaptor
	Expect(exec.RunKubectl(nil, GinkgoWriter, "apply", "-f", filepath.Join(testRootDir, "adaptors", "dummy", "deploy", "e2e", "all_in_one.yaml"))).
		Should(Succeed())
}

func uninstallOctopus(done chan<- interface{}) {
	// uninstall dummy adaptor
	Expect(exec.RunKubectl(nil, GinkgoWriter, "delete", "-f", filepath.Join(testRootDir, "adaptors", "dummy", "deploy", "e2e", "all_in_one.yaml"))).
		Should(Succeed())

	// uninstall octopus
	Expect(exec.RunKubectl(nil, GinkgoWriter, "delete", "-f", filepath.Join(testRootDir, "deploy", "e2e", "all_in_one.yaml"))).
		Should(Succeed())

	close(done)
}

func deleteNodeHostingLimb() {
	// get limb daemonset
	var daemonset appsv1.DaemonSet
	Expect(k8sCli.Get(testCtx, types.NamespacedName{Namespace: "octopus-system", Name: "octopus-limb"}, &daemonset)).Should(Succeed())

	// get limb pods
	var pods corev1.PodList
	Expect(k8sCli.List(testCtx, &pods, client.MatchingLabels(daemonset.Spec.Selector.MatchLabels))).Should(Succeed())

	// delete node randomly
	Eventually(func() bool {
		var pod = pods.Items[rand.Intn(len(pods.Items))]
		var podName = pod.Spec.NodeName
		var n corev1.Node
		if err := k8sCli.Get(testCtx, types.NamespacedName{Name: podName}, &n); err != nil {
			GinkgoT().Log(err)
		}
		if node.IsOnlyWorker(&n) {
			return k8sCli.Delete(testCtx, &n) == nil
		}
		return false
	}, 120, 1).Should(BeTrue())
}

func deleteNodeHostingBrain() {
	// get brain deployment
	var deployment appsv1.Deployment
	Expect(k8sCli.Get(testCtx, types.NamespacedName{Namespace: "octopus-system", Name: "octopus-brain"}, &deployment)).Should(Succeed())

	// get brain pods
	var pods corev1.PodList
	Expect(k8sCli.List(testCtx, &pods, client.MatchingLabels(deployment.Spec.Selector.MatchLabels))).Should(Succeed())

	// delete node randomly
	Eventually(func() bool {
		var pod = pods.Items[rand.Intn(len(pods.Items))]
		var podName = pod.Spec.NodeName
		var n corev1.Node
		if err := k8sCli.Get(testCtx, types.NamespacedName{Name: podName}, &n); err != nil {
			GinkgoT().Log(err)
		}
		if node.IsOnlyWorker(&n) {
			return k8sCli.Delete(testCtx, &n) == nil
		}
		return false
	}, 120, 1).Should(BeTrue())
}

func isOctopusAvailable() {
	// confirm brain if exist
	Eventually(func() (bool, error) {
		var svc corev1.Service
		var err = k8sCli.Get(testCtx, types.NamespacedName{Namespace: "octopus-system", Name: "octopus-brain"}, &svc)
		if err != nil {
			GinkgoT().Log(err)
			if !apierrs.IsNotFound(err) {
				return false, err
			}
		}
		if !object.IsActivating(&svc) {
			return false, nil
		}

		var deployment appsv1.Deployment
		err = k8sCli.Get(testCtx, types.NamespacedName{Namespace: "octopus-system", Name: "octopus-brain"}, &deployment)
		if err != nil {
			GinkgoT().Log(err)
			if !apierrs.IsNotFound(err) {
				return false, err
			}
		}
		if !object.IsActivating(&deployment) {
			return false, nil
		}

		return deployment.Status.Replicas > 0 &&
			deployment.Status.Replicas == deployment.Status.AvailableReplicas, nil
	}, 300, 1).Should(BeTrue())

	// confirm limb if exist
	Eventually(func() (bool, error) {
		var svc corev1.Service
		var err = k8sCli.Get(testCtx, types.NamespacedName{Namespace: "octopus-system", Name: "octopus-limb"}, &svc)
		if err != nil {
			GinkgoT().Log(err)
			if !apierrs.IsNotFound(err) {
				return false, err
			}
		}
		if !object.IsActivating(&svc) {
			return false, nil
		}

		var daemonset appsv1.DaemonSet
		err = k8sCli.Get(testCtx, types.NamespacedName{Namespace: "octopus-system", Name: "octopus-limb"}, &daemonset)
		if err != nil {
			GinkgoT().Log(err)
			if !apierrs.IsNotFound(err) {
				return false, err
			}
		}
		if !object.IsActivating(&daemonset) {
			return false, nil
		}

		return daemonset.Status.NumberAvailable > 0 &&
			daemonset.Status.DesiredNumberScheduled == daemonset.Status.NumberReady, nil
	}, 300, 1).Should(BeTrue())

	verifyDeviceLink()
}

func verifyDeviceLink() {
	// get valid node
	var targetNode, err = node.GetValidWorker(testCtx, k8sCli)
	Expect(err).ShouldNot(HaveOccurred())

	// create devicelink
	var targetItem = edgev1alpha1.DeviceLink{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    "default",
			GenerateName: "test-", // NB(thxCode) create name randomly
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
	Expect(k8sCli.Create(testCtx, &targetItem)).Should(Succeed())

	// confirm the devicelink is on DeviceCreated=True
	var targetKey = types.NamespacedName{
		Namespace: targetItem.Namespace,
		Name:      targetItem.Name,
	}
	Eventually(func() bool {
		if err := k8sCli.Get(testCtx, targetKey, &targetItem); err != nil {
			GinkgoT().Log(err)
			return false
		}
		return devicelink.GetDeviceConnectedStatus(&targetItem.Status) == metav1.ConditionTrue
	}, 300, 1).Should(BeTrue())
}
