package limb

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/cmd/limb/options"
	"github.com/rancher/octopus/pkg/limb/controller"
	"github.com/rancher/octopus/pkg/metrics"
	"github.com/rancher/octopus/pkg/suctioncup"
	"github.com/rancher/octopus/pkg/util/critical"
	"github.com/rancher/octopus/pkg/util/log/handler"
)

func Run(name string, opts *options.Options) error {
	var log = ctrl.Log.WithName(name).WithName("setup")
	defer runtime.HandleCrash(handler.NewPanicsLogHandler(log))

	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	log.V(0).Info("Registering metrics")
	if err := RegisterMetrics(ctrlmetrics.Registry); err != nil {
		log.Error(err, "Unable to register metrics")
		return err
	}

	log.V(0).Info("Registering APIs scheme")
	var scheme = k8sruntime.NewScheme()
	if err := RegisterScheme(scheme); err != nil {
		log.Error(err, "Unable to register APIs scheme")
		return err
	}

	log.V(0).Info("Parsing arguments")
	// processing options
	var nodeName = os.Getenv("NODE_NAME")
	if nodeName == "" {
		nodeName = opts.NodeName
	}
	if nodeName == "" {
		return errors.New("node name could not be blank")
	}
	log.Info("Landing on", "node", nodeName)

	log.V(0).Info("Creating controller manager")
	var controllerMgr, err = ctrl.NewManager(
		ctrl.GetConfigOrDie(),
		ctrl.Options{
			Scheme:             scheme,
			LeaderElection:     false,
			MetricsBindAddress: fmt.Sprintf(":%d", opts.MetricsAddr),
		},
	)
	if err != nil {
		log.Error(err, "Unable to start controller manager")
		return err
	}

	log.V(0).Info("Creating suction cup manager")
	suctionCupMgr, err := suctioncup.NewManager()
	if err != nil {
		log.Error(err, "Unable to start suction cup manager")
		return err
	}

	log.V(0).Info("Creating controllers")
	if err = (&controller.DeviceLinkReconciler{
		Client:        controllerMgr.GetClient(),
		EventRecorder: controllerMgr.GetEventRecorderFor(name),
		Ctx:           ctx,
		Log:           ctrl.Log.WithName("controller").WithName("deviceLink"),
		SuctionCup:    suctionCupMgr.GetNeurons(),
		NodeName:      nodeName,
	}).SetupWithManager(controllerMgr, suctionCupMgr); err != nil {
		log.Error(err, "Unable to create controller", "controller", "DeviceLink")
		return err
	}

	log.Info("Starting")
	var stop = ctrl.SetupSignalHandler()
	var eg, egCtx = errgroup.WithContext(critical.Context(stop))
	stop = egCtx.Done()
	eg.Go(func() error {
		return suctionCupMgr.Start(stop)
	})
	eg.Go(func() error {
		return controllerMgr.Start(stop)
	})
	if err = eg.Wait(); err != nil {
		log.Error(err, "Problem running")
		return err
	}
	return nil
}

func RegisterScheme(scheme *k8sruntime.Scheme) error {
	if err := edgev1alpha1.AddToScheme(scheme); err != nil {
		return err
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		return err
	}
	return nil
}

func RegisterMetrics(registry prometheus.Registerer) error {
	return metrics.RegisterLimbMetrics(registry)
}
