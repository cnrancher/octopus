package adaptor

import (
	"context"

	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/runtime"

	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor/validation"
	"github.com/rancher/octopus/pkg/util/log/handler"
)

func (m *manager) Register(_ context.Context, req *api.RegisterRequest) (resp *api.Void, respErr error) {
	log := log.WithValues("adaptor", req.Name)
	defer runtime.HandleCrash(handler.NewPanicsLogHandler(log))

	// TODO collect metrics

	var versionSupport bool
	for _, v := range api.SupportedVersions {
		if req.Version == v {
			versionSupport = true
			break
		}
	}
	if !versionSupport {
		var err = grpcstatus.Errorf(grpccodes.InvalidArgument, "the requested version %s is not support by Limb, supported version is %v", req.Version, api.SupportedVersions)
		log.Error(err, "rejected the register request")
		return &api.Void{}, err
	}

	if errs := validation.IsQualifiedName(req.Name); len(errs) != 0 {
		var err = grpcstatus.Errorf(grpccodes.InvalidArgument, "the requested name %s is not qualified by Limb", req.Name)
		log.Error(err, "rejected the register request")
		return &api.Void{}, err
	}

	if !validation.IsSocketFile(req.Endpoint) {
		var err = grpcstatus.Errorf(grpccodes.InvalidArgument, "the requested endpoint %s could not be recognized by Limb", req.Endpoint)
		log.Error(err, "rejected the register request")
		return &api.Void{}, err
	}

	if err := m.addAdaptor(req.Name, req.Endpoint); err != nil {
		err := grpcstatus.Errorf(grpccodes.Internal, err.Error())
		log.Error(err, "failed to add adaptor")
		return &api.Void{}, err
	}

	return &api.Void{}, nil
}
