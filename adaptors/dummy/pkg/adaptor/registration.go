package adaptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
)

func Register(ctx context.Context, conn *grpc.ClientConn) error {
	var request = &api.RegisterRequest{
		Name:     Name,
		Version:  Version,
		Endpoint: Endpoint,
	}

	var _, err = api.NewRegistrationServiceClient(conn).Register(ctx, request)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to register to Limb: %v", err)
	}
	return nil
}
