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

	dummyv1alpha1 "github.com/rancher/octopus/adaptors/dummy/api/v1alpha1"
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

func TestDummyAdaptor(t *testing.T) {
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

	close(done)
}, 600)

var _ = AfterSuite(func(done Done) {
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
	// calculate the project dir of ${GOPATH}/github.com/rancher/octopus/adaptors/dummy
	testCurrDir, _ = filepath.Abs(filepath.Join(filepath.Dir("."), "..", "..", ".."))
	// calculate the project root dir of ${GOPATH}/github.com/rancher/octopus
	testRootDir, _ = filepath.Abs(filepath.Join(testCurrDir, "..", ".."))
}

func registerScheme(scheme *runtime.Scheme) error {
	return dummyv1alpha1.AddToScheme(scheme)
}

func installOctopus() {
	// install octopus
	Expect(exec.RunKubectl(nil, GinkgoWriter, "apply", "-f", filepath.Join(testRootDir, "deploy", "e2e", "all_in_one.yaml"))).
		Should(Succeed())

	// install dummy adaptor
	Expect(exec.RunKubectl(nil, GinkgoWriter, "apply", "-f", filepath.Join(testCurrDir, "deploy", "e2e", "all_in_one.yaml"))).
		Should(Succeed())

	isOctopusAvailable()
}

func uninstallOctopus() {
	// uninstall dummy adaptor
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
