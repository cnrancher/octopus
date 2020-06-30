package limb

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/suctioncup/connection"
	modelutil "github.com/rancher/octopus/pkg/util/model"
	"github.com/rancher/octopus/pkg/util/object"
)

var _ = Describe("verify DeviceLink controller", func() {
	var (
		testNamespace   corev1.Namespace
		testAdaptor     fakeAdaptor
		testAdaptorName string

		targetItem edgev1alpha1.DeviceLink
	)

	BeforeEach(func() {
		testNamespace = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-",
			},
		}
		_ = k8sCli.Create(testCtx, &testNamespace)

		testAdaptorName = testAdaptor.GetName()
	})

	AfterEach(func() {
		_ = k8sCli.DeleteAllOf(testCtx, &edgev1alpha1.DeviceLink{}, client.InNamespace(testNamespace.Name))
		_ = k8sCli.Delete(testCtx, &testNamespace)
	})

	JustBeforeEach(func() {
		targetItem = edgev1alpha1.DeviceLink{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    testNamespace.Name,
				GenerateName: "test-",
			},
			Spec: edgev1alpha1.DeviceLinkSpec{
				Adaptor: edgev1alpha1.DeviceAdaptor{
					Node: testNodeName,
					Name: testAdaptorName,
				},
				Model: testModel,
			},
		}
	})

	Context("when the adaptor is unregistered", func() {

		BeforeEach(func() {
			testAdaptors.Delete(testAdaptor.GetName())
		})

		It("should succeed after registered the adaptor", func() {

			By("given a new link which failed on adaptor verification", func() {
				// creates
				Expect(k8sCli.Create(testCtx, &targetItem)).Should(Succeed())

				// simulates that reconciled by brain
				targetItem.SucceedOnNodeExisted(nil)
				targetItem.SucceedOnModelExisted()
				Expect(k8sCli.Status().Update(testCtx, &targetItem)).Should(Succeed())

				// confirms
				var key = object.GetNamespacedName(&targetItem)
				Eventually(func() error {
					if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
						return err
					}
					if targetItem.GetAdaptorExistedStatus() != metav1.ConditionFalse {
						return errors.New("should not find the corresponding adaptor of link")
					}
					return nil
				}, 30, 1).Should(Succeed())
			})

			By("when register the adaptor", func() {
				testAdaptors.Put(testAdaptor)
				testEventQueue.GetAdaptorNotifier().NoticeAdaptorRegistered(testAdaptor.GetName())
			})

			By("then it succeed on adaptor verification", func() {
				var key = object.GetNamespacedName(&targetItem)
				Eventually(func() error {
					if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
						return err
					}
					if targetItem.GetAdaptorExistedStatus() != metav1.ConditionTrue {
						return errors.New("could not find the corresponding adaptor of link")
					}
					return nil
				}, 30, 1).Should(Succeed())
			})

		})

	})

	Context("when the adaptor is registered", func() {

		BeforeEach(func() {
			testAdaptors.Put(testAdaptor)
		})

		Context("and the adaptor spec is invalid", func() {

			BeforeEach(func() {
				testAdaptorName = fmt.Sprintf("x%s", testAdaptor.GetName())
			})

			It("should succeed if modified to valid adaptor", func() {

				By("given a new link which failed on adaptor verification", func() {
					// creates
					Expect(k8sCli.Create(testCtx, &targetItem)).Should(Succeed())

					// simulates that reconciled by brain
					targetItem.SucceedOnNodeExisted(nil)
					targetItem.SucceedOnModelExisted()
					Expect(k8sCli.Status().Update(testCtx, &targetItem)).Should(Succeed())

					// confirms
					var key = object.GetNamespacedName(&targetItem)
					Eventually(func() error {
						if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
							return err
						}
						if targetItem.GetAdaptorExistedStatus() != metav1.ConditionFalse {
							return errors.New("should not find the corresponding adaptor of link")
						}
						return nil
					}, 30, 1).Should(Succeed())
				})

				By("when modify to valid adaptor", func() {
					Expect(k8sCli.Get(testCtx, object.GetNamespacedName(&targetItem), &targetItem)).Should(Succeed())

					targetItem.Spec.Adaptor.Name = testAdaptor.GetName()
					Expect(k8sCli.Update(testCtx, &targetItem)).Should(Succeed())
				})

				By("then it succeeded on adaptor verification", func() {
					var key = object.GetNamespacedName(&targetItem)
					Eventually(func() error {
						if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
							return err
						}
						if targetItem.GetAdaptorExistedStatus() != metav1.ConditionTrue {
							return errors.New("could not find the corresponding adaptor of link")
						}
						return nil
					}, 30, 1).Should(Succeed())
				})

			})

		})

		Context("and the adaptor spec is valid", func() {

			It("should fail if unregistered adaptor", func() {

				By("given a new link which succeeded on adaptor verification", func() {
					// creates
					Expect(k8sCli.Create(testCtx, &targetItem)).Should(Succeed())

					// simulates that reconciled by brain
					targetItem.SucceedOnNodeExisted(nil)
					targetItem.SucceedOnModelExisted()
					Expect(k8sCli.Status().Update(testCtx, &targetItem)).Should(Succeed())

					// confirms
					var key = object.GetNamespacedName(&targetItem)
					Eventually(func() error {
						if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
							return err
						}
						if targetItem.GetAdaptorExistedStatus() != metav1.ConditionTrue {
							return errors.New("could not find the corresponding adaptor of link")
						}
						return nil
					}, 30, 1).Should(Succeed())
				})

				By("when unregister the adaptor", func() {
					testAdaptors.Delete(targetItem.Status.AdaptorName)
					testEventQueue.GetAdaptorNotifier().NoticeAdaptorUnregistered(testAdaptor.GetName())
				})

				By("then it failed on adaptor verification", func() {
					var key = object.GetNamespacedName(&targetItem)
					Eventually(func() error {
						if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
							return err
						}
						if targetItem.GetAdaptorExistedStatus() != metav1.ConditionFalse {
							return errors.New("should not find the corresponding adaptor of link")
						}
						return nil
					}, 30, 1).Should(Succeed())
				})

			})

			It("should fail if modified to invalid adaptor", func() {

				By("given a new link which succeeded on adaptor verification", func() {
					// creates
					Expect(k8sCli.Create(testCtx, &targetItem)).Should(Succeed())

					// simulates that reconciled by brain
					targetItem.SucceedOnNodeExisted(nil)
					targetItem.SucceedOnModelExisted()
					Expect(k8sCli.Status().Update(testCtx, &targetItem)).Should(Succeed())

					// confirms
					var key = object.GetNamespacedName(&targetItem)
					Eventually(func() error {
						if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
							return err
						}
						if targetItem.GetAdaptorExistedStatus() != metav1.ConditionTrue {
							return errors.New("could not find the corresponding adaptor of link")
						}
						return nil
					}, 30, 1).Should(Succeed())
				})

				By("when modify to invalid adaptor", func() {
					Expect(k8sCli.Get(testCtx, object.GetNamespacedName(&targetItem), &targetItem)).Should(Succeed())

					targetItem.Spec.Adaptor.Name = fmt.Sprintf("x%s", targetItem.Status.AdaptorName)
					Expect(k8sCli.Update(testCtx, &targetItem)).Should(Succeed())
				})

				By("then it failed on adaptor verification", func() {
					var key = object.GetNamespacedName(&targetItem)
					Eventually(func() error {
						if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
							return err
						}
						if targetItem.GetAdaptorExistedStatus() != metav1.ConditionFalse {
							return errors.New("should not find the corresponding adaptor of link")
						}
						return nil
					}, 30, 1).Should(Succeed())
				})

			})

			It("should create the corresponding device", func() {

				By("given a new link", func() {
					// creates
					Expect(k8sCli.Create(testCtx, &targetItem)).Should(Succeed())

					// simulates that reconciled by brain
					targetItem.SucceedOnNodeExisted(nil)
					targetItem.SucceedOnModelExisted()
					Expect(k8sCli.Status().Update(testCtx, &targetItem)).Should(Succeed())
				})

				By("then it succeeded on adaptor verification", func() {
					var key = object.GetNamespacedName(&targetItem)
					Eventually(func() error {
						if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
							return err
						}
						if targetItem.GetAdaptorExistedStatus() != metav1.ConditionTrue {
							return errors.New("could not find the corresponding adaptor of link")
						}
						return nil
					}, 30, 1).Should(Succeed())
				})

				By("and it succeeded on device creation", func() {
					var key = object.GetNamespacedName(&targetItem)
					Eventually(func() error {
						if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
							return err
						}
						if targetItem.GetDeviceCreatedStatus() != metav1.ConditionTrue {
							return errors.New("could not find the corresponding device of link")
						}
						return nil
					}, 30, 1).Should(Succeed())

					var device, err = modelutil.NewInstanceOfTypeMeta(*targetItem.Status.Model)
					Expect(err).ToNot(HaveOccurred())
					Expect(k8sCli.Get(testCtx, key, &device)).Should(Succeed())
				})

				By("and it succeeded on device connection", func() {
					var key = object.GetNamespacedName(&targetItem)
					Eventually(func() error {
						if err := k8sCli.Get(testCtx, key, &targetItem); err != nil {
							return err
						}
						if targetItem.GetDeviceConnectedStatus() != metav1.ConditionTrue {
							return errors.New("could not find the corresponding connection of link")
						}
						return nil
					}, 30, 1).Should(Succeed())
				})

			})

		})

	})

})

type fakeAdaptor string

func (a fakeAdaptor) GetName() string {
	return "adaptors.edge.cattle.io/fake"
}

func (a fakeAdaptor) GetEndpoint() string {
	return "fake.sock"
}

func (a fakeAdaptor) Stop() error {
	return nil
}

func (a fakeAdaptor) CreateConnection(name types.NamespacedName) (overwritten bool, conn connection.Connection, err error) {
	return false, fakeConnection(name), nil
}

func (a fakeAdaptor) DeleteConnection(name types.NamespacedName) bool {
	return false
}

type fakeConnection types.NamespacedName

func (c fakeConnection) GetAdaptorName() string {
	return "adaptors.edge.cattle.io/fake"
}

func (c fakeConnection) GetName() types.NamespacedName {
	return types.NamespacedName(c)
}

func (c fakeConnection) Stop() error {
	return nil
}

func (c fakeConnection) IsStop() bool {
	return false
}

func (c fakeConnection) Send(*metav1.TypeMeta, []byte, map[string]*api.ConnectRequestReferenceEntry) error {
	return nil
}
