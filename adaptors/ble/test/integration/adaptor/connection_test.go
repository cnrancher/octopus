package adaptor

import (
	"fmt"
	"io"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rancher/octopus/adaptors/ble/pkg/adaptor"
	"github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	mock_v1alpha1 "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1/mock"
)

var _ = Describe("verify Connection", func() {
	var (
		err error

		mockCtrl *gomock.Controller
		service  *adaptor.Service
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		service, err = adaptor.NewService()
		if err != nil {
			// NB(thxCode) this skip equals do-not-testing, it will be fixed until
			// we found a good way to test bluetooth inside a container on the CI platform.
			Skip(fmt.Sprintf("failed to start service: %v", err))
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
		if service != nil {
			err = service.Close()
			if err != nil {
				GinkgoT().Logf("failed to stop service: %v", err)
			}
		}
	})

	Context("on Connect server", func() {

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

		It("should process the input device", func() {
			// failed unmarshal
			mockServer.EXPECT().Recv().Return(&v1alpha1.ConnectRequest{
				Model: &metav1.TypeMeta{
					APIVersion: "devices.edge.cattle.io/v1alpha1",
					Kind:       "BluetoothDevice",
				},
				Device: []byte(`{this is an illegal json}`),
			}, nil)
			err = service.Connect(mockServer)
			var sts = status.Convert(err)
			Expect(sts.Code()).To(Equal(grpccodes.InvalidArgument))
			Expect(sts.Message()).To(HavePrefix("failed to unmarshal device"))

			// failed to connect a device
			mockServer.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()
			mockServer.EXPECT().Recv().Return(&v1alpha1.ConnectRequest{
				Model: &metav1.TypeMeta{
					APIVersion: "devices.edge.cattle.io/v1alpha1",
					Kind:       "BluetoothDevice",
				},
				Device: []byte(`
				{
					"apiVersion":"devices.edge.cattle.io/v1alpha1",
					"kind":"BluetoothDevice",
					"metadata":{
						"name":"correct",
						"namespace":"default"
					},
					"spec":{
						"protocol":{
							"endpoint":"FAKED_TH"
						},
						"properties":[
							{
								"name":"temperature_and_humidity",
								"accessModes":["Notify"],
								"visitor":{
									"service":"226c0000-6476-4566-7562-66734470666d",
									"characteristic":"226caa55-6476-4566-7562-66734470666d"
								}
							}
						]
					}
				}`),
			}, nil)
			err = service.Connect(mockServer)
			sts = status.Convert(err)
			Expect(sts.Code()).To(Equal(grpccodes.InvalidArgument))
			Expect(sts.Message()).To(Equal("failed to connect to Bluetooth device: failed to scan Bluetooth device FAKED_TH: GATT peripheral is not found"))
		})

	})

})
