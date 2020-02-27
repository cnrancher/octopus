package limb

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/cmd/limb/options"
	"github.com/rancher/octopus/pkg/adaptor"
	"github.com/rancher/octopus/pkg/limb/controller"
	"github.com/rancher/octopus/pkg/util/log/handler"

	_ "github.com/rancher/octopus/pkg/util/log/handler"
	_ "github.com/rancher/octopus/pkg/util/version/metric"
)

func Run(name string, opts *options.Options) error {
	var setupLog = ctrl.Log.WithName(name).WithName("setup")
	defer runtime.HandleCrash(handler.NewPanicsLogHandler(setupLog))

	setupLog.V(0).Info("registering APIs scheme")
	var scheme = k8sruntime.NewScheme()
	if err := RegisterScheme(scheme); err != nil {
		setupLog.Error(err, "unable to register APIs scheme")
		return err
	}

	setupLog.V(0).Info("parsing arguments")
	// processing options
	var nodeName = os.Getenv("NODE_NAME")
	if nodeName == "" {
		nodeName = opts.NodeName
	}
	if nodeName == "" {
		return errors.New("node name could not be blank")
	}

	setupLog.V(0).Info("creating controller manager")
	var controllerMgr, err = ctrl.NewManager(
		ctrl.GetConfigOrDie(),
		ctrl.Options{
			Scheme:             scheme,
			MetricsBindAddress: fmt.Sprintf(":%d", opts.MetricsAddr),
		},
	)
	if err != nil {
		setupLog.Error(err, "unable to start controller manager")
		return err
	}

	setupLog.V(0).Info("creating adaptor manager")
	adaptorMgr, err := adaptor.NewManager(
		adaptor.Options{},
	)
	if err != nil {
		setupLog.Error(err, "unable to start adaptor manager")
		return err
	}

	setupLog.V(0).Info("creating controllers")
	if err = (&controller.DeviceLinkReconciler{
		Client:        controllerMgr.GetClient(),
		EventRecorder: controllerMgr.GetEventRecorderFor(name),
		Scheme:        controllerMgr.GetScheme(),
		Log:           ctrl.Log.WithName("controller").WithName("DeviceLink"),
		Adaptors:      adaptorMgr.GetPool(),
		NodeName:      nodeName,
	}).SetupWithManager(name, controllerMgr, adaptorMgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DeviceLink")
		return err
	}

	setupLog.Info("starting")
	var ctx = spliceContext(ctrl.SetupSignalHandler(), nil)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return controllerMgr.Start(ctx.Done())
	})
	eg.Go(func() error {
		return adaptorMgr.Start(ctx.Done())
	})
	if err = eg.Wait(); err != nil {
		setupLog.Error(err, "problem running")
		return err
	}
	return nil
}

func RegisterScheme(scheme *k8sruntime.Scheme) error {
	if err := edgev1alpha1.AddToScheme(scheme); err != nil {
		return err
	}
	return nil
}

func spliceContext(previousC <-chan struct{}, previousFn func()) context.Context {
	var ctx, cancel = context.WithCancel(context.Background())
	closeContext := func() {
		if previousFn != nil {
			previousFn()
		}
		cancel()
	}
	go func() {
		for {
			select {
			case <-previousC:
				closeContext()
				return
			}
		}
	}()
	return ctx
}
