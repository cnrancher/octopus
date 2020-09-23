package limb

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/rancher/octopus/pkg/brain"
	"github.com/rancher/octopus/pkg/limb"
	"github.com/rancher/octopus/pkg/limb/controller"
	"github.com/rancher/octopus/pkg/suctioncup"
	"github.com/rancher/octopus/pkg/suctioncup/adaptor"
	"github.com/rancher/octopus/pkg/suctioncup/event"
	"github.com/rancher/octopus/pkg/util/log/zap"
	"github.com/rancher/octopus/test/framework/envtest"
	"github.com/rancher/octopus/test/framework/envtest/printer"
	"github.com/rancher/octopus/test/util/crd"
	"github.com/rancher/octopus/test/util/node"
)

var (
	testCtx       context.Context
	testCtxCancel context.CancelFunc
	testCurrDir   string
	testRootDir   string
	testEnv       *envtest.Environment

	k8sCfg *rest.Config
	k8sCli client.Client

	testNodeName   string
	testModel      metav1.TypeMeta
	testAdaptors   adaptor.Adaptors
	testEventQueue event.Queue
)

func TestLimb(t *testing.T) {
	defer GinkgoRecover()

	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"brain suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	defer close(done)

	testCtx, testCtxCancel = context.WithCancel(context.Background())

	// sets the log of controller-runtime as dev mode
	logf.SetLogger(zap.WrapAsLogr(zap.NewDevelopmentLogger()))

	var err error

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDInstallOptions: envtest.CRDInstallOptions{
			Paths: []string{
				filepath.Join(testRootDir, "deploy", "manifests", "crd", "base"),
			},
		},
	}

	// NB(thxCode) use the native client to avoid that the cache is not started
	By("creating kubernetes client")
	var k8sSchema = clientsetscheme.Scheme
	err = brain.RegisterScheme(k8sSchema)
	Expect(err).NotTo(HaveOccurred())

	k8sCfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sCfg).ToNot(BeNil())

	k8sCli, err = client.New(k8sCfg, client.Options{Scheme: k8sSchema})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sCli).ToNot(BeNil())

	By("creating controller manager")
	var ctrlScheme = runtime.NewScheme()
	err = limb.RegisterScheme(ctrlScheme)
	Expect(err).NotTo(HaveOccurred())

	controllerMgr, err := ctrl.NewManager(k8sCfg, ctrl.Options{Scheme: ctrlScheme, LeaderElection: false})
	Expect(err).ToNot(HaveOccurred())
	Expect(controllerMgr).ToNot(BeNil())

	By("getting a valid node")
	testNodeName, err = node.GetValidWorker(testCtx, k8sCli)
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
		Ctx:           testCtx,
		Log:           ctrl.Log.WithName("controller").WithName("deviceLink"),
		NodeName:      testNodeName,
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

	By("creating global testing resources")
	testModel = metav1.TypeMeta{
		Kind:       "IntegrationLimbDLValidateDevice",
		APIVersion: "devices.edge.cattle.io/v1alpha1",
	}
	_ = k8sCli.Create(testCtx, crd.MakeOfTypeMeta(testModel))
}, 600)

var _ = AfterSuite(func(done Done) {
	defer close(done)

	By("deleting global testing resources")
	_ = k8sCli.Delete(testCtx, crd.MakeOfTypeMeta(testModel))

	By("tearing down test environment")
	var err = testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())

	if testCtxCancel != nil {
		testCtxCancel()
	}
}, 600)

func init() {
	// calculate the project dir of ${GOPATH}/github.com/rancher/octopus/test/integration/limb
	testCurrDir, _ = filepath.Abs(filepath.Join(filepath.Dir("."), "..", "..", ".."))
	// calculate the project root dir of ${GOPATH}/github.com/rancher/octopus/test/integration/limb
	testRootDir = testCurrDir
}
