package installation

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/rancher/octopus/pkg/util/object"
	"github.com/rancher/octopus/test/util/exec"
)

var _ = Describe("verify installation", func() {

	Context("through kubectl", func() {

		Specify("kubectl apply -f all_in_one.yaml", func() {

			By("given a clean cluster", isClusterClean)

			By("when apply via all_in_one.yaml", applyAllInOneYAML)

			By(" then the Octopus is ready", isOctopusReady)

		})

		Specify("kubectl delete -f all_in_one.yaml", func() {

			By("given the Octopus is ready", isOctopusReady)

			By("when delete via all_in_one.yaml", deleteAllInOneYAML)

			By("then the cluster is clean", isClusterClean)

		})

	})

	XContext("through helm", func() {

		Specify("helm install chart", func() {

			By("given the cluster is clean", isClusterClean)

			By("when install via chart", installChart)

			By("then the Octopus is ready", isOctopusReady)

		})

		Specify("helm uninstall chart", func() {

			By("given the Octopus is ready", isOctopusReady)

			By("when uninstall via chart", uninstallChart)

			By("then the cluster is clean", isClusterClean)

		})

	})

})

func applyAllInOneYAML() {
	Expect(exec.RunKubectl(nil, GinkgoWriter, "apply", "-f", filepath.Join(testRootDir, "deploy", "e2e", "all_in_one.yaml"))).
		Should(Succeed())
}

func deleteAllInOneYAML() {
	Expect(exec.RunKubectl(nil, GinkgoWriter, "delete", "-f", filepath.Join(testRootDir, "deploy", "e2e", "all_in_one.yaml"))).
		Should(Succeed())
}

func installChart() {
	// TODO
}

func uninstallChart() {
	// TODO
}

func isClusterClean() {
	// namespace should not exist
	Eventually(func() (bool, error) {
		var ns corev1.Namespace
		var err = k8sCli.Get(testCtx, types.NamespacedName{Name: "octopus-system"}, &ns)
		if err != nil {
			GinkgoT().Log(err)
			if !apierrs.IsNotFound(err) {
				return false, err
			}
		}
		return !object.IsActivating(&ns), nil
	}, 30, 1).Should(BeTrue())

	// crd should not exist
	Eventually(func() (bool, error) {
		var crd apiextensionsv1.CustomResourceDefinition
		var err = k8sCli.Get(testCtx, types.NamespacedName{Name: "devicelinks.edge.cattle.io"}, &crd)
		if err != nil {
			GinkgoT().Log(err)
			if !apierrs.IsNotFound(err) {
				return false, err
			}
		}
		return !object.IsActivating(&crd), nil
	}, 30, 1).Should(BeTrue())
}

func isOctopusReady() {
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
}
