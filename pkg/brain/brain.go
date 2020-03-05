package brain

import (
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
	var setupLog = ctrl.Log.WithName(name).WithName("setup")
	defer runtime.HandleCrash(handler.NewPanicsLogHandler(setupLog))

	setupLog.V(0).Info("registering APIs scheme")
	var scheme = k8sruntime.NewScheme()
	if err := RegisterScheme(scheme); err != nil {
		setupLog.Error(err, "unable to register APIs scheme")
		return err
	}

	setupLog.V(0).Info("creating controller manager")
	var controllerMgr, err = ctrl.NewManager(
		ctrl.GetConfigOrDie(),
		ctrl.Options{
			Scheme:             scheme,
			MetricsBindAddress: fmt.Sprintf(":%d", opts.MetricsAddr),
			Port:               opts.AdmissionWebhookAddr,
			LeaderElection:     opts.EnableLeaderElection,
		},
	)
	if err != nil {
		setupLog.Error(err, "unable to start controller manager")
		return err
	}

	setupLog.V(0).Info("creating controllers")
	if err = (&controller.DeviceLinkReconciler{
		Client:        controllerMgr.GetClient(),
		EventRecorder: controllerMgr.GetEventRecorderFor(name),
		Log:           ctrl.Log.WithName("controller").WithName("DeviceLink"),
	}).SetupWithManager(name, controllerMgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DeviceLink")
		return err
	}
	if err = (&controller.NodeReconciler{
		Client: controllerMgr.GetClient(),
		Log:    ctrl.Log.WithName("controller").WithName("Node"),
	}).SetupWithManager(name, controllerMgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Node")
		return err
	}
	if err = (&controller.ModelReconciler{
		Client: controllerMgr.GetClient(),
		Log:    ctrl.Log.WithName("controller").WithName("Model"),
	}).SetupWithManager(name, controllerMgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Model")
		return err
	}

	if !opts.DisableAdmissionWebhook {
		setupLog.V(0).Info("creating admission webhooks")
		if err = (&edgev1alpha1.DeviceLink{}).SetupWebhookWithManager(controllerMgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "DeviceLink")
			return err
		}
	}

	setupLog.Info("starting")
	if err = controllerMgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running")
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
