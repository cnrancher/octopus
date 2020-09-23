package adaptor

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rancher/octopus/test/framework/envtest/printer"
)

var (
	testCtx       context.Context
	testCtxCancel context.CancelFunc
)

func TestAdaptor(t *testing.T) {
	defer GinkgoRecover()

	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"adaptor suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	defer close(done)

	testCtx, testCtxCancel = context.WithCancel(context.Background())
}, 600)

var _ = AfterSuite(func(done Done) {
	defer close(done)

	if testCtxCancel != nil {
		testCtxCancel()
	}
}, 600)
