package adaptor

import (
	"io"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rancher/octopus/adaptors/dummy/pkg/adaptor"
	"github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	mock_v1alpha1 "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1/mock"
)

// testing scenarios:
// 	+ Server
//		- validate if the connection stop when it closes
//		- validate the process of input model
//      - validate the recognition of the input model
// 		- validate the process of input device
var _ = Describe("Connection", func() {
	var (
		err error

		mockCtrl *gomock.Controller
		service  *adaptor.Service
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		service = adaptor.NewService()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("Server", func() {

		var mockServer *mock_v1alpha1.MockConnection_ConnectServer

		BeforeEach(func() {
			mockServer = mock_v1alpha1.NewMockConnection_ConnectServer(mockCtrl)
		})

		It("should be stopped if closed", func() {
			// io.EOF
			mockServer.EXPECT().Recv().Return(nil, io.EOF)
			err = service.Connect(mockServer)
			Expect(err).ToNot(HaveOccurred())

			// canceled by context
			mockServer.EXPECT().Recv().Return(nil, status.Error(grpccodes.Canceled, "context canceled"))
			err = service.Connect(mockServer)
			Expect(err).ToNot(HaveOccurred())

			// other canceled reason
			mockServer.EXPECT().Recv().Return(nil, status.Error(grpccodes.Canceled, "other"))
			err = service.Connect(mockServer)
			Expect(err).To(HaveOccurred())

			// transport is closing
			mockServer.EXPECT().Recv().Return(nil, status.Error(grpccodes.Unavailable, "transport is closing"))
			err = service.Connect(mockServer)
			Expect(err).ToNot(HaveOccurred())

			// other unavailable reason
			mockServer.EXPECT().Recv().Return(nil, status.Error(grpccodes.Unavailable, "other"))
			err = service.Connect(mockServer)
			Expect(err).To(HaveOccurred())
		})

		It("should process the input model", func() {
			// failed as model is nil
			mockServer.EXPECT().Recv().Return(&v1alpha1.ConnectRequest{
				Model: nil,
			}, nil)
			err = service.Connect(mockServer)
			var sts = status.Convert(err)
			Expect(sts.Code()).To(Equal(grpccodes.InvalidArgument))
			Expect(sts.Message()).To(HavePrefix("invalid empty model"))

			// failed as invalidate group
			mockServer.EXPECT().Recv().Return(&v1alpha1.ConnectRequest{
				Model: &metav1.TypeMeta{
					APIVersion: "invalidate.group/v1alpha1",
					Kind:       "DummySpecialDevice",
				},
			}, nil)
			err = service.Connect(mockServer)
			sts = status.Convert(err)
			Expect(sts.Code()).To(Equal(grpccodes.InvalidArgument))
			Expect(sts.Message()).To(Equal("invalid model group: invalidate.group"))

			// failed as invalidate kind
			mockServer.EXPECT().Recv().Return(&v1alpha1.ConnectRequest{
				Model: &metav1.TypeMeta{
					APIVersion: "devices.edge.cattle.io/v1alpha1",
					Kind:       "InvalidateSpecialDevice",
				},
			}, nil)
			err = service.Connect(mockServer)
			sts = status.Convert(err)
			Expect(sts.Code()).To(Equal(grpccodes.InvalidArgument))
			Expect(sts.Message()).To(Equal("invalid model kind: InvalidateSpecialDevice"))
		})

		It("should distinguish the input model", func() {
			// distinguish the devices.edge.cattle.io/v1alpha1/DummySpecialDevice model
			mockServer.EXPECT().Recv().Return(&v1alpha1.ConnectRequest{
				Model: &metav1.TypeMeta{
					APIVersion: "devices.edge.cattle.io/v1alpha1",
					Kind:       "DummySpecialDevice",
				},
				Device: []byte(`{}`),
			}, nil)
			err = service.Connect(mockServer)
			var sts = status.Convert(err)
			Expect(sts.Code()).To(Equal(grpccodes.InvalidArgument))
			Expect(sts.Message()).To(Equal("failed to recognize the empty device as the namespace/name is blank"))

			// distinguish the devices.edge.cattle.io/v1alpha1/DummyProtocolDevice model
			mockServer.EXPECT().Recv().Return(&v1alpha1.ConnectRequest{
				Model: &metav1.TypeMeta{
					APIVersion: "devices.edge.cattle.io/v1alpha1",
					Kind:       "DummyProtocolDevice",
				},
				Device: []byte(`{}`),
			}, nil)
			err = service.Connect(mockServer)
			sts = status.Convert(err)
			Expect(sts.Code()).To(Equal(grpccodes.InvalidArgument))
			Expect(sts.Message()).To(Equal("failed to recognize the empty device as the namespace/name is blank"))
		})

		It("should process the input device", func() {
			// failed unmarshal
			mockServer.EXPECT().Recv().Return(&v1alpha1.ConnectRequest{
				Model: &metav1.TypeMeta{
					APIVersion: "devices.edge.cattle.io/v1alpha1",
					Kind:       "DummySpecialDevice",
				},
				Device: []byte(`{this is an illegal json}`),
			}, nil)
			err = service.Connect(mockServer)
			var sts = status.Convert(err)
			Expect(sts.Code()).To(Equal(grpccodes.InvalidArgument))
			Expect(sts.Message()).To(HavePrefix("failed to unmarshal device"))

			// correct logic
			mockServer.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()
			mockServer.EXPECT().Recv().Return(&v1alpha1.ConnectRequest{
				Model: &metav1.TypeMeta{
					APIVersion: "devices.edge.cattle.io/v1alpha1",
					Kind:       "DummySpecialDevice",
				},
				Device: []byte(`{"apiVersion":"devices.edge.cattle.io/v1alpha1","kind":"DummySpecialDevice","metadata":{"name":"living-room-fan","namespace":"default"},"spec":{"protocol":{"location":"living-room"},"gear":"slow","on":true}}`),
			}, nil)
			mockServer.EXPECT().Recv().Return(nil, io.EOF) // simulate to close the connection
			err = service.Connect(mockServer)
			Expect(err).ToNot(HaveOccurred())
		})

	})

})
