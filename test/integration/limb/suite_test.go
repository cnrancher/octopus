package limb

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
	"github.com/rancher/octopus/pkg/limb"
	"github.com/rancher/octopus/pkg/limb/controller"
	"github.com/rancher/octopus/pkg/suctioncup"
	"github.com/rancher/octopus/pkg/suctioncup/adaptor"
	"github.com/rancher/octopus/pkg/suctioncup/event"
	"github.com/rancher/octopus/test/framework"
	"github.com/rancher/octopus/test/util/node"
)

var (
	testCtx       context.Context
	testCtxCancel context.CancelFunc
	testRootDir   string
	testEnv       *envtest.Environment

	k8sCfg *rest.Config
	k8sCli client.Client

	targetNode     string
	testAdaptors   adaptor.Adaptors
	testEventQueue event.Queue
)

func TestAPIs(t *testing.T) {
	defer GinkgoRecover()

	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"brain suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	testCtx, testCtxCancel = context.WithCancel(context.Background())

	// sets the log of controller-runtime as dev mode
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	var err error

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		UseExistingCluster: pointer.BoolPtr(true),
		CRDDirectoryPaths: []string{
			filepath.Join(testRootDir, "deploy", "manifests", "crd", "base"),
			filepath.Join(testRootDir, "adaptors", "dummy", "deploy", "manifests", "crd", "base"),
		},
	}

	k8sCfg, err = framework.StartEnv(testRootDir, testEnv, GinkgoWriter)
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
	err = limb.RegisterScheme(ctrlScheme)
	Expect(err).NotTo(HaveOccurred())

	controllerMgr, err := ctrl.NewManager(k8sCfg, ctrl.Options{Scheme: ctrlScheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(controllerMgr).ToNot(BeNil())

	By("getting a valid node")
	targetNode, err = node.GetValidWorker(testCtx, k8sCli)
	Expect(err).ToNot(HaveOccurred())

	By("starting suctioncup manager")
	testAdaptors = adaptor.NewAdaptors()
	testEventQueue = event.NewQueue()
	suctionCupMgr, err := suctioncup.NewManagerWith(testAdaptors, testEventQueue)
	Expect(err).ToNot(HaveOccurred())

	By("creating controllers")
	err = (&controller.DeviceLinkReconciler{
		Client:        controllerMgr.GetClient(),
		EventRecorder: controllerMgr.GetEventRecorderFor("limb"),
		Scheme:        controllerMgr.GetScheme(),
		Log:           ctrl.Log.WithName("controller").WithName("deviceLink"),
		NodeName:      targetNode,
		SuctionCup:    suctionCupMgr.GetNeurons(),
	}).SetupWithManager(controllerMgr, suctionCupMgr)
	Expect(err).ToNot(HaveOccurred())

	var stopCh = testCtx.Done()
	go func() {
		err = controllerMgr.Start(stopCh)
		Expect(err).ToNot(HaveOccurred())
	}()
	go func() {
		err = suctionCupMgr.Start(stopCh)
		Expect(err).ToNot(HaveOccurred())
	}()

	close(done)
}, 600)

var _ = AfterSuite(func() {
	By("tearing down test environment")
	var err = framework.StopEnv(testRootDir, testEnv, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	if testCtxCancel != nil {
		testCtxCancel()
	}
}, 600)

func init() {
	var currDir = filepath.Dir(".")
	// calculate the project root dir of ${GOPATH}/github.com/rancher/octopus/test/integration/limb
	testRootDir, _ = filepath.Abs(filepath.Join(currDir, "..", "..", ".."))
}
