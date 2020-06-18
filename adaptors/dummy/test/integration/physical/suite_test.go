package adaptor

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
)

var (
	testCtx       context.Context
	testCtxCancel context.CancelFunc
)

func TestPhysical(t *testing.T) {
	defer GinkgoRecover()

	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"physical suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	testCtx, testCtxCancel = context.WithCancel(context.Background())

	close(done)
}, 600)

var _ = AfterSuite(func() {
	if testCtxCancel != nil {
		testCtxCancel()
	}
}, 600)
