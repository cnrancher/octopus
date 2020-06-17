package brain

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/cmd/brain/options"
	"github.com/rancher/octopus/pkg/brain/controller"
	"github.com/rancher/octopus/pkg/util/log/handler"
)

func Run(name string, opts *options.Options) error {
	var log = ctrl.Log.WithName(name).WithName("setup")
	defer runtime.HandleCrash(handler.NewPanicsLogHandler(log))

	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	log.V(0).Info("Registering APIs scheme")
	var scheme = k8sruntime.NewScheme()
	if err := RegisterScheme(scheme); err != nil {
		log.Error(err, "Unable to register APIs scheme")
		return err
	}

	log.V(0).Info("Creating controller manager")
	var controllerMgr, err = ctrl.NewManager(
		ctrl.GetConfigOrDie(),
		ctrl.Options{
			Scheme:             scheme,
			MetricsBindAddress: fmt.Sprintf(":%d", opts.MetricsAddr),
			LeaderElection:     opts.EnableLeaderElection,
			LeaderElectionID:   "octopus-brain-leader-election-id",
		},
	)
	if err != nil {
		log.Error(err, "Unable to start controller manager")
		return err
	}

	log.V(0).Info("Creating controllers")
	if err = (&controller.DeviceLinkReconciler{
		Client: controllerMgr.GetClient(),
		Ctx:    ctx,
		Log:    ctrl.Log.WithName("controller").WithName("deviceLink"),
	}).SetupWithManager(controllerMgr); err != nil {
		log.Error(err, "Unable to create controller", "controller", "DeviceLink")
		return err
	}
	if err = (&controller.NodeReconciler{
		Client: controllerMgr.GetClient(),
		Ctx:    ctx,
		Log:    ctrl.Log.WithName("controller").WithName("node"),
	}).SetupWithManager(controllerMgr); err != nil {
		log.Error(err, "Unable to create controller", "controller", "Node")
		return err
	}
	if err = (&controller.ModelReconciler{
		Client: controllerMgr.GetClient(),
		Ctx:    ctx,
		Log:    ctrl.Log.WithName("controller").WithName("crd"),
	}).SetupWithManager(controllerMgr); err != nil {
		log.Error(err, "Unable to create controller", "controller", "CRD")
		return err
	}

	log.Info("Starting")
	if err = controllerMgr.Start(ctrl.SetupSignalHandler()); err != nil {
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
	if err := apiextensionsv1.AddToScheme(scheme); err != nil {
		return err
	}
	return nil
}
