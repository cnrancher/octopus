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

	"github.com/rancher/octopus/pkg/limb"
	"github.com/rancher/octopus/pkg/limb/controller"
	"github.com/rancher/octopus/test/framework"
	fakeadaptor "github.com/rancher/octopus/test/framework/adaptor"
	"github.com/rancher/octopus/test/util/node"
	"github.com/rancher/octopus/test/util/rootdir"
)

var (
	ctx           context.Context
	cancelFunc    context.CancelFunc
	k8sCfg        *rest.Config
	k8sCli        client.Client
	ctrlMgr       ctrl.Manager
	adaptorMgr    *fakeadaptor.Manager
	adaptorOnNode string
	testEnv       *envtest.Environment
	rootDir       = rootdir.Get()
)

func TestAPIs(t *testing.T) {
	defer GinkgoRecover()

	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"limb suite",
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
			filepath.Join(rootDir, "deploy", "manifests", "crd"),
			filepath.Join(rootDir, "adaptors", "dummy", "deploy", "manifests", "crd"),
		},
	}

	k8sCfg, err = framework.StartEnv(testEnv, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sCfg).ToNot(BeNil())

	By("creating kubernetes client")
	var k8sSchema = scheme.Scheme
	err = limb.RegisterScheme(k8sSchema)
	Expect(err).NotTo(HaveOccurred())

	// NB(thxCode) use the native client to avoid that the cache is not started
	k8sCli, err = client.New(k8sCfg, client.Options{Scheme: k8sSchema})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sCli).ToNot(BeNil())

	By("creating controller manager")
	var ctrlScheme = runtime.NewScheme()
	err = limb.RegisterScheme(ctrlScheme)
	Expect(err).NotTo(HaveOccurred())

	ctrlMgr, err = ctrl.NewManager(k8sCfg, ctrl.Options{Scheme: ctrlScheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(ctrlMgr).ToNot(BeNil())

	By("creating adaptor manager")
	adaptorMgr = fakeadaptor.NewManager()
	Expect(adaptorMgr).ToNot(BeNil())

	By("getting a valid node")
	adaptorOnNode, err = node.GetValidWorker(ctx, k8sCli)
	Expect(err).ToNot(HaveOccurred())

	By("creating controllers")
	var name = "limb"

	err = (&controller.DeviceLinkReconciler{
		Client:        ctrlMgr.GetClient(),
		EventRecorder: ctrlMgr.GetEventRecorderFor(name),
		Scheme:        ctrlMgr.GetScheme(),
		Log:           ctrl.Log.WithName("controller").WithName("DeviceLink"),
		Adaptors:      adaptorMgr.GetPool(),
		NodeName:      adaptorOnNode,
	}).SetupWithManager(name, ctrlMgr, adaptorMgr)

	go func() {
		err = ctrlMgr.Start(ctx.Done())
		Expect(err).ToNot(HaveOccurred())
	}()

	close(done)
}, 600)

var _ = AfterSuite(func() {
	By("tearing down test environment")
	var err = framework.StopEnv(testEnv, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	if cancelFunc != nil {
		cancelFunc()
	}
}, 600)
