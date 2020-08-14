package usability_test

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"

	mqttv1alpha1 "github.com/rancher/octopus/adaptors/mqtt/api/v1alpha1"
	"github.com/rancher/octopus/pkg/brain"
	"github.com/rancher/octopus/pkg/limb"
	"github.com/rancher/octopus/pkg/util/object"
	"github.com/rancher/octopus/test/framework"
	"github.com/rancher/octopus/test/util/exec"
)

var (
	testCtx       context.Context
	testCtxCancel context.CancelFunc
	testCurrDir   string
	testRootDir   string
	testEnv       *envtest.Environment

	k8sCfg *rest.Config
	k8sCli client.Client
)

func TestMQTTAdaptor(t *testing.T) {
	defer GinkgoRecover()

	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"usability suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	testCtx, testCtxCancel = context.WithCancel(context.Background())

	var err error

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		UseExistingCluster: pointer.BoolPtr(true),
	}

	By("creating kubernetes client")
	var k8sSchema = clientsetscheme.Scheme
	err = brain.RegisterScheme(k8sSchema)
	Expect(err).NotTo(HaveOccurred())
	err = limb.RegisterScheme(k8sSchema)
	Expect(err).NotTo(HaveOccurred())

	err = registerScheme(k8sSchema)
	Expect(err).NotTo(HaveOccurred())

	k8sCfg, err = framework.StartEnv(testRootDir, testEnv, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sCfg).ToNot(BeNil())

	k8sCli, err = client.New(k8sCfg, client.Options{Scheme: k8sSchema})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sCli).ToNot(BeNil())

	installOctopus()

	installMQTTSimulationSuite()

	close(done)
}, 600)

var _ = AfterSuite(func(done Done) {
	uninstallMQTTSimulationSuite()

	uninstallOctopus()

	By("tearing down test environment")
	var err = framework.StopEnv(testRootDir, testEnv, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	if testCtxCancel != nil {
		testCtxCancel()
	}

	close(done)
}, 600)

func init() {
	// calculate the project dir of ${GOPATH}/github.com/rancher/octopus/adaptors/mqtt
	testCurrDir, _ = filepath.Abs(filepath.Join(filepath.Dir("."), "..", "..", ".."))
	// calculate the project root dir of ${GOPATH}/github.com/rancher/octopus
	testRootDir, _ = filepath.Abs(filepath.Join(testCurrDir, "..", ".."))
}

func registerScheme(scheme *runtime.Scheme) error {
	return mqttv1alpha1.AddToScheme(scheme)
}

func installOctopus() {
	// install octopus
	Expect(exec.RunKubectl(nil, GinkgoWriter, "apply", "-f", filepath.Join(testRootDir, "deploy", "e2e", "all_in_one.yaml"))).
		Should(Succeed())

	// install MQTT adaptor
	Expect(exec.RunKubectl(nil, GinkgoWriter, "apply", "-f", filepath.Join(testCurrDir, "deploy", "e2e", "all_in_one.yaml"))).
		Should(Succeed())

	isOctopusAvailable()

	isMQTTAdaptorAvailable()
}

func uninstallOctopus() {
	// uninstall MQTT adaptor
	Expect(exec.RunKubectl(nil, GinkgoWriter, "delete", "-f", filepath.Join(testCurrDir, "deploy", "e2e", "all_in_one.yaml"))).
		Should(Succeed())

	// uninstall octopus
	Expect(exec.RunKubectl(nil, GinkgoWriter, "delete", "-f", filepath.Join(testRootDir, "deploy", "e2e", "all_in_one.yaml"))).
		Should(Succeed())
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
}

func isMQTTAdaptorAvailable() {
	// confirm MQTT adaptor if exist
	Eventually(func() (bool, error) {
		var podList corev1.PodList
		if err := k8sCli.List(testCtx, &podList, client.InNamespace("octopus-system"), client.MatchingLabels{"app.kubernetes.io/name": "octopus-adaptor-mqtt"}); err != nil {
			GinkgoT().Log(err)
			if !apierrs.IsNotFound(err) {
				return false, err
			}
		}
		for _, pod := range podList.Items {
			if !object.IsActivating(&pod) {
				return false, nil
			}
			for _, condition := range pod.Status.Conditions {
				if condition.Type == "Ready" && condition.Status == "False" {
					return false, nil
				}
			}
		}

		var daemonset appsv1.DaemonSet
		err := k8sCli.Get(testCtx, types.NamespacedName{Namespace: "octopus-system", Name: "octopus-adaptor-mqtt-adaptor"}, &daemonset)
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

func installMQTTSimulationSuite() {
	// install MQTT broker
	Expect(exec.RunKubectl(nil, GinkgoWriter, "apply", "-f", filepath.Join(testCurrDir, "test", "e2e", "usability", "testdata", "mqtt-broker.yaml"))).
		Should(Succeed())

	// install MQTT simulator
	Expect(exec.RunKubectl(nil, GinkgoWriter, "apply", "-f", filepath.Join(testCurrDir, "deploy", "e2e", "simulator.yaml"))).
		Should(Succeed())

	isMQTTServerAvailable()

	isMQTTSimulatorAvailable()
}

func uninstallMQTTSimulationSuite() {
	// uninstall MQTT broker
	Expect(exec.RunKubectl(nil, GinkgoWriter, "delete", "-f", filepath.Join(testCurrDir, "test", "e2e", "usability", "testdata", "mqtt-broker.yaml"))).
		Should(Succeed())

	// uninstall MQTT simulator
	Expect(exec.RunKubectl(nil, GinkgoWriter, "delete", "-f", filepath.Join(testCurrDir, "deploy", "e2e", "simulator.yaml"))).
		Should(Succeed())
}

func isMQTTServerAvailable() {
	var endpoints corev1.Endpoints
	var key = types.NamespacedName{
		Name:      "mqtt-broker",
		Namespace: "default",
	}
	Eventually(func() bool {
		if err := k8sCli.Get(testCtx, key, &endpoints); err != nil {
			GinkgoT().Log(err)
			return false
		}
		if len(endpoints.Subsets) != 0 {
			var subset = endpoints.Subsets[0]
			if len(subset.Addresses) != 0 && len(subset.Ports) != 0 {
				for _, port := range subset.Ports {
					return port.Name == "unencrypted" && port.Port == 1883
				}
			}
		}
		return false
	}, 300, 1).Should(BeTrue())
}

func isMQTTSimulatorAvailable() {
	var endpoints corev1.Endpoints
	var key = types.NamespacedName{
		Name:      "octopus-simulator-mqtt",
		Namespace: "octopus-simulator-system",
	}
	Eventually(func() bool {
		if err := k8sCli.Get(testCtx, key, &endpoints); err != nil {
			GinkgoT().Log(err)
			return false
		}
		if len(endpoints.Subsets) != 0 {
			var subset = endpoints.Subsets[0]
			if len(subset.Addresses) != 0 && len(subset.Ports) != 0 {
				for _, port := range subset.Ports {
					return port.Name == "tcp" && port.Port == 1883
				}
			}
		}
		return false
	}, 300, 1).Should(BeTrue())
}
