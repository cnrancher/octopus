package brain

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/rancher/octopus/pkg/brain"
	"github.com/rancher/octopus/pkg/brain/controller"
	"github.com/rancher/octopus/test/framework"
)

var (
	rootDir    string
	ctx        context.Context
	cancelFunc context.CancelFunc
	k8sCfg     *rest.Config
	k8sCli     client.Client
	ctrlMgr    ctrl.Manager
	testEnv    *envtest.Environment
)

func TestAPIs(t *testing.T) {
	defer GinkgoRecover()

	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"brain suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	ctx, cancelFunc = context.WithCancel(context.Background())

	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	var err error

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		UseExistingCluster: pointer.BoolPtr(true),
		CRDDirectoryPaths: []string{
			filepath.Join(rootDir, "deploy", "manifests", "crd", "base"),
			filepath.Join(rootDir, "adaptors", "dummy", "deploy", "manifests", "crd", "base"),
		},
	}

	k8sCfg, err = framework.StartEnv(rootDir, testEnv, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sCfg).ToNot(BeNil())

	By("creating kubernetes client")
	var k8sSchema = scheme.Scheme
	err = brain.RegisterScheme(k8sSchema)
	Expect(err).NotTo(HaveOccurred())

	// NB(thxCode) use the native client to avoid that the cache is not started
	k8sCli, err = client.New(k8sCfg, client.Options{Scheme: k8sSchema})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sCli).ToNot(BeNil())

	By("creating controller manager")
	var ctrlScheme = runtime.NewScheme()
	err = brain.RegisterScheme(ctrlScheme)
	Expect(err).NotTo(HaveOccurred())

	ctrlMgr, err = ctrl.NewManager(k8sCfg, ctrl.Options{Scheme: ctrlScheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(ctrlMgr).ToNot(BeNil())

	By("creating controllers")
	var name = "brain"

	err = (&controller.DeviceLinkReconciler{
		Client:        ctrlMgr.GetClient(),
		EventRecorder: ctrlMgr.GetEventRecorderFor(name),
		Log:           ctrl.Log.WithName("controller").WithName("DeviceLink"),
	}).SetupWithManager(name, ctrlMgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controller.NodeReconciler{
		Client: ctrlMgr.GetClient(),
		Log:    ctrl.Log.WithName("controller").WithName("Node"),
	}).SetupWithManager(name, ctrlMgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controller.ModelReconciler{
		Client: ctrlMgr.GetClient(),
		Log:    ctrl.Log.WithName("controller").WithName("Model"),
	}).SetupWithManager(name, ctrlMgr)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = ctrlMgr.Start(ctx.Done())
		Expect(err).ToNot(HaveOccurred())
	}()

	close(done)
}, 600)

var _ = AfterSuite(func() {
	By("tearing down test environment")
	var err = framework.StopEnv(rootDir, testEnv, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	if cancelFunc != nil {
		cancelFunc()
	}
}, 600)

func init() {
	var currDir = filepath.Dir(".")
	// calculate the project root dir of ${GOPATH}/github.com/rancher/octopus/test/integration/brain
	rootDir, _ = filepath.Abs(filepath.Join(currDir, "..", "..", ".."))
}
