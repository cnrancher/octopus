package adaptor

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rancher/octopus/test/framework"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	testCtx       context.Context
	testCtxCancel context.CancelFunc
	testEnv       *envtest.Environment
	testRootDir   string

	k8sCfg *rest.Config
)

func TestAPIs(t *testing.T) {
	defer GinkgoRecover()

	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"adaptor suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	testCtx, testCtxCancel = context.WithCancel(context.Background())

	logf.SetLogger(zap.New(zap.UseDevMode(true)))
	testEnv = &envtest.Environment{
		UseExistingCluster: pointer.BoolPtr(true),
		CRDDirectoryPaths: []string{
			filepath.Join(testRootDir, "deploy", "manifests", "crd", "base"),
			filepath.Join(testRootDir, "adaptors", "agent-device", "deploy", "manifests", "crd", "base"),
		},
	}
	var err error

	k8sCfg, err = framework.StartEnv(testRootDir, testEnv, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sCfg).ToNot(BeNil())

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
	// calculate the project root dir
	testRootDir, _ = filepath.Abs(filepath.Join(currDir, "..", "..", "..", "..", ".."))
}
