package usability

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/rancher/octopus/pkg/brain"
	"github.com/rancher/octopus/pkg/limb"
	"github.com/rancher/octopus/test/framework/envtest"
	"github.com/rancher/octopus/test/framework/envtest/printer"
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

func TestOctopus(t *testing.T) {
	defer GinkgoRecover()

	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"usability suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	defer close(done)

	testCtx, testCtxCancel = context.WithCancel(context.Background())

	var err error

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{}

	By("creating kubernetes client")
	var k8sSchema = clientsetscheme.Scheme
	err = brain.RegisterScheme(k8sSchema)
	Expect(err).NotTo(HaveOccurred())
	err = limb.RegisterScheme(k8sSchema)
	Expect(err).NotTo(HaveOccurred())

	k8sCfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sCfg).ToNot(BeNil())

	k8sCli, err = client.New(k8sCfg, client.Options{Scheme: k8sSchema})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sCli).ToNot(BeNil())
}, 600)

var _ = AfterSuite(func(done Done) {
	defer close(done)

	By("tearing down test environment")
	var err = testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())

	if testCtxCancel != nil {
		testCtxCancel()
	}
}, 600)

func init() {
	// calculate the project dir of ${GOPATH}/github.com/rancher/octopus/test/e2e/usability
	testCurrDir, _ = filepath.Abs(filepath.Join(filepath.Dir("."), "..", "..", ".."))
	// calculate the project root dir of ${GOPATH}/github.com/rancher/octopus/test/e2e/usability
	testRootDir = testCurrDir
}
