package usability

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/rancher/octopus/adaptors/mqtt/api/v1alpha1"
	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/util/object"
	. "github.com/rancher/octopus/test/framework/envtest/dsl"
	"github.com/rancher/octopus/test/util/content"
	"github.com/rancher/octopus/test/util/exec"
	"github.com/rancher/octopus/test/util/node"
)

/*
	NB(uuuxxllj): the following cases focus on AttributedTopic pattern.
*/

var (
	testDeviceLink edgev1alpha1.DeviceLink
)

var _ = Describe("verify usability", func() {

	var targetNode string

	BeforeEach(func() {
		var err error
		targetNode, err = node.GetValidWorker(testCtx, k8sCli)
		Expect(err).ShouldNot(HaveOccurred())

		// defaults to create the kitchen light device link connected to MQTT simulator
		testDeviceLink = edgev1alpha1.DeviceLink{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    "default",
				GenerateName: "test-kitchen-light-",
			},
			Spec: edgev1alpha1.DeviceLinkSpec{
				Adaptor: edgev1alpha1.DeviceAdaptor{
					Node: targetNode,
					Name: "adaptors.edge.cattle.io/mqtt",
				},
				Model: metav1.TypeMeta{
					Kind:       "MQTTDevice",
					APIVersion: "devices.edge.cattle.io/v1alpha1",
				},
				Template: edgev1alpha1.DeviceTemplateSpec{
					Spec: content.ToRawExtension(
						map[string]interface{}{
							"protocol": map[string]interface{}{
								"pattern": "AttributedTopic",
								"client": map[string]interface{}{
									"server": "tcp://octopus-simulator-mqtt.octopus-simulator-system:1883",
								},
								"message": map[string]interface{}{
									"topic": "cattle.io/octopus/home/:operator/kitchen/light/:path",
									"operator": map[string]interface{}{
										"read":  "status",
										"write": "set",
									},
								},
							},
							"properties": []map[string]interface{}{
								{
									"name":        "switch",
									"type":        "boolean",
									"accessModes": []string{"WriteOnce"},
								},
								{
									"name": "luminance",
									"type": "int",
									"visitor": map[string]interface{}{
										"path": "parameter_luminance",
									},
								},
							},
						},
					),
				},
			},
		}
	})

	JustBeforeEach(func() {
		// create device link
		Expect(k8sCli.Create(testCtx, &testDeviceLink)).Should(Succeed())
	})

	AfterEach(func() {
		_ = k8sCli.DeleteAllOf(testCtx, &edgev1alpha1.DeviceLink{}, client.InNamespace(testDeviceLink.Namespace))
	})

	Context("modify MQTT device link spec", func() {

		Specify("if invalid node spec", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when invalid node spec", invalidNodeSpec)

			By("then node of the device link is not found", isNodeExistedFalse)

			By("when correct node spec", correctNodeSpec)

			By("then node of the device link is found", isNodeExistedTrue)

		})

		Specify("if invalid model spec", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when invalid model spec", invalidModelSpec)

			By("then model of the device link is not found", isModelExistedFalse)

			By("when correct model spec", correctModelSpec)

			By("then model of the device link is found", isModelExistedTrue)

		})

		Specify("if invalid adaptor spec", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when invalid adaptor spec", invalidAdaptorSpec)

			By("then adaptor of the device link is not found", isAdaptorExistedFalse)

			By("when correct adaptor spec", correctAdaptorSpec)

			By("then adaptor of the device link is found", isAdaptorExistedTrue)

		})

		Specify("if invalid device spec", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when invalid device spec", invalidDeviceSpec)

			By("then the device link is not created", isDeviceConnectedFalse)

			By("when correct device spec", correctDeviceSpec)

			By("then the device link is connected", isDeviceConnectedTrue)

		})

		Context("deploy device with invalid spec", func() {

			BeforeEach(func() {

				// create a device link with invalid spec
				testDeviceLink = edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:    "default",
						GenerateName: "test-kitchen-light-",
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Node: targetNode,
							Name: "adaptors.edge.cattle.io/mqtt",
						},
						Model: metav1.TypeMeta{
							Kind:       "MQTTDevice",
							APIVersion: "devices.edge.cattle.io/v1alpha1",
						},
						Template: edgev1alpha1.DeviceTemplateSpec{
							Spec: content.ToRawExtension(
								map[string]interface{}{
									"protocol": map[string]interface{}{
										// NB(thxCode): as we know, pattern is required.
										// "pattern": "AttributedTopic",
										"client": map[string]interface{}{
											"server": "tcp://octopus-simulator-mqtt.octopus-simulator-system:1883",
										},
										"message": map[string]interface{}{
											"topic": "cattle.io/octopus/home/:operator/kitchen/light/:path",
										},
										"properties": []map[string]interface{}{
											{
												"name":        "switch",
												"type":        "boolean",
												"accessModes": []string{"WriteOnce"},
											},
											{
												"name": "luminance",
												"type": "int",
												"visitor": map[string]interface{}{
													"path": "parameter_luminance",
												},
											},
										},
									},
								},
							),
						},
					},
				}

			})

			Specify("if deploy device with invalid spec", func() {

				By("given the device link is blocked in failed creation", isDeviceCreatedFalse)

				By("when correct the spec", correctDeviceSpec)

				By("then the device link is connected", isDeviceConnectedTrue)

			})

		})

	})

	Context("interfere deployment environment", func() {

		Specify("if delete MQTT adaptor pods", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when delete MQTT adaptor pods", deleteMQTTAdaptorPods)

			By("then adaptor of the device link is not found", isAdaptorExistedFalse)

		})

		Specify("if delete octopus limbs pods", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when delete limbs pods", deleteLimbsPods)

			By("then the MQTT adaptor pods become error", isMQTTAdaptorPodsError)

		})

		Specify("if delete MQTT device model", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when delete MQTT device model", deleteMQTTDeviceModel)

			By("then model of the device link is not found", isModelExistedFalse)

			By("when redeploy MQTT device model", redeployMQTTDeviceModel)

			By("then the device link is connected", isDeviceConnectedTrue)

		})

		K3dSpecify("if delete corresponding node", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when delete corresponding cluster node", deleteCorrespondingNode)

			By("then node of the device link is not found", isNodeExistedFalse)

			By("when redeploy corresponding cluster node", redeployCorrespondingNode)

			By("then the device link is connected", isDeviceConnectedTrue)

		})

	})

	Context("interact with simulation suite", func() {

		XSpecify("if connect with octopus-simulator", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when turn on the closed simulated device", func() {
				// patch switch to true
				var err = k8sCli.Get(testCtx, object.GetNamespacedName(&testDeviceLink), &testDeviceLink)
				Expect(err).Should(Succeed())
				var patch = []byte(`
 {
    "spec":{
        "template":{
            "spec":{
                "properties":[
                    {
                        "name":"switch",
                        "type":"boolean",
                        "value":true,
                        "accessModes": ["WriteOnce"],
                    },
                    {
                        "name":"luminance",
                        "type":"int",
						"visitor": {
                            "path": "parameter_luminance"
                        }
                    }
                ]
            }
        }
    }
}`)
				Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())

			})

			By("then the device is on and changing its status", func() {
				// verify the status if the switch is true and the luminance is changed
				var deviceLinkKey = object.GetNamespacedName(&testDeviceLink)

				var (
					tempVal string
					count   = 2
				)
				Eventually(func() bool {
					var device v1alpha1.MQTTDevice
					var err = k8sCli.Get(testCtx, deviceLinkKey, &device)
					if err != nil {
						GinkgoT().Log(err)
						return false
					}
					for _, prop := range device.Status.Properties {
						if prop.Name == "luminance" {
							if prop.Value != tempVal {
								count--
								tempVal = prop.Value
								if count <= 0 {
									return true
								}
							}
						}
					}
					return false
				}, 60, 3).Should(BeTrue())
			})

		})

		Context("connect with in-cluster MQTT broker", func() {

			BeforeEach(func() {

				// create the device link connected to an in-cluster MQTT broker
				testDeviceLink = edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:    "default",
						GenerateName: "test-mqtt-",
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Node: targetNode,
							Name: "adaptors.edge.cattle.io/mqtt",
						},
						Model: metav1.TypeMeta{
							Kind:       "MQTTDevice",
							APIVersion: "devices.edge.cattle.io/v1alpha1",
						},
						Template: edgev1alpha1.DeviceTemplateSpec{
							Spec: content.ToRawExtension(
								map[string]interface{}{
									"protocol": map[string]interface{}{
										"pattern": "AttributedTopic",
										"client": map[string]interface{}{
											// we connect to the in-cluster MQTT broker
											"server": "tcp://mqtt-broker.default:1883",
										},
										"message": map[string]interface{}{
											"topic": "cattle.io/octopus/home/:operator/in-cluster/:path",
											"operator": map[string]interface{}{
												"read":  "status",
												"write": "set",
											},
										},
									},
									"properties": []map[string]interface{}{
										{
											"name":        "subscribeValue",
											"type":        "string",
											"description": "subscribe from broker",
											"visitor": map[string]interface{}{
												"path": "subscribe_value",
											},
										},
										{
											"name":        "publishValue",
											"type":        "string",
											"description": "publish to broker",
											"accessModes": []string{"WriteMany"},
											"visitor": map[string]interface{}{
												"path": "publish_value",
											},
										},
									},
								},
							),
						},
					},
				}

			})

			It("should receive a message", func() {

				By("given the device link is connected", isDeviceConnectedTrue)

				By("when publish a message to a specified topic of in-cluster MQTT broker", func() {
					// run a Job connect the in-cluster MQTT broker and do as below:
					// publish message "hello" to topic "cattle.io/octopus/home/status/in-cluster/subscribe_value".
					var publishedJob = batchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:    "default",
							GenerateName: "test-mqtt-",
						},
						Spec: batchv1.JobSpec{
							TTLSecondsAfterFinished: pointer.Int32Ptr(10),
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									RestartPolicy: corev1.RestartPolicyNever,
									Containers: []corev1.Container{
										{
											Name:  "publish",
											Image: "eclipse-mosquitto:1.6.12",
											Command: []string{
												"mosquitto_pub",
											},
											Args: []string{
												"-h",
												"mqtt-broker.default",
												"-t",
												"cattle.io/octopus/home/status/in-cluster/subscribe_value",
												"-m",
												"hello",
												"-q",
												"1",
												"-r",
											},
										},
									},
								},
							},
						},
					}
					Expect(k8sCli.Create(testCtx, &publishedJob)).Should(Succeed())

					var publishedJobKey = object.GetNamespacedName(&publishedJob)
					Eventually(func() bool {
						var err = k8sCli.Get(testCtx, publishedJobKey, &publishedJob)
						if err != nil {
							GinkgoT().Log(err)
							return false
						}
						for _, cond := range publishedJob.Status.Conditions {
							if cond.Type == batchv1.JobComplete {
								return cond.Status == corev1.ConditionTrue
							}
						}
						return false
					}, 30, 1).Should(BeTrue())
				})

				By("then the device received the message", func() {
					var deviceLinkKey = object.GetNamespacedName(&testDeviceLink)
					Eventually(func() bool {
						var device v1alpha1.MQTTDevice
						var err = k8sCli.Get(testCtx, deviceLinkKey, &device)
						if err != nil {
							GinkgoT().Log(err)
							return false
						}

						for _, prop := range device.Status.Properties {
							if prop.Name == "subscribeValue" {
								return prop.Value == "hello"
							}
						}
						return false
					}, 30, 1).Should(BeTrue())
				})

			})

			It("should publish a message", func() {

				By("given the device link is connected", isDeviceConnectedTrue)

				By("when set value to a writable property", func() {
					var err = k8sCli.Get(testCtx, object.GetNamespacedName(&testDeviceLink), &testDeviceLink)
					Expect(err).Should(Succeed())
					var patch = []byte(`
{
    "spec":{
        "template":{
            "spec":{
                "properties":[
                    {
                        "name":"publishValue",
                        "description":"publish to broker",
                        "type":"string",
                        "value":"hello",
                        "accessModes": ["WriteOnce"],
                        "visitor": {
                            "path": "publish_value"
                        }
                    },
                    {
                        "name":"subscribeValue",
                        "type":"string",
                        "description":"subscribe from broker",
                        "visitor": {
                            "path": "subscribe_value"
                        }
                    }
                ]
            }
        }
    }
}`)
					Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
				})

				By("then the value can be received by subscribers", func() {
					// run a Job connect to the in-cluster MQTT broker and do as below:
					// subscribe topic "cattle.io/octopus/home/set/in-cluster/publish_value",
					// and verify the message is "hello".
					var subscribedJob = batchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:    "default",
							GenerateName: "test-mqtt-",
						},
						Spec: batchv1.JobSpec{
							TTLSecondsAfterFinished: pointer.Int32Ptr(10),
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									InitContainers: []corev1.Container{
										{
											Name:  "subscribe",
											Image: "eclipse-mosquitto:1.6.12",
											Command: []string{
												"/bin/sh",
											},
											Args: []string{
												"-c",
												"mosquitto_sub -h mqtt-broker.default -t cattle.io/octopus/home/set/in-cluster/publish_value -C 1 >/storage/test",
											},
											VolumeMounts: []corev1.VolumeMount{
												{
													Name:      "temp-volume",
													MountPath: "/storage",
												},
											},
										},
									},
									RestartPolicy: corev1.RestartPolicyOnFailure,
									Containers: []corev1.Container{
										{
											Name:  "test",
											Image: "eclipse-mosquitto:1.6.12",
											Command: []string{
												"/bin/sh",
											},
											Args: []string{
												"-c",
												"grep -w hello /storage/test",
											},
											VolumeMounts: []corev1.VolumeMount{
												{
													Name:      "temp-volume",
													MountPath: "/storage",
												},
											},
										},
									},
									Volumes: []corev1.Volume{
										{
											Name: "temp-volume",
											VolumeSource: corev1.VolumeSource{
												EmptyDir: &corev1.EmptyDirVolumeSource{},
											},
										},
									},
								},
							},
						},
					}
					Expect(k8sCli.Create(testCtx, &subscribedJob)).Should(Succeed())

					var subscribedJobKey = object.GetNamespacedName(&subscribedJob)
					Eventually(func() bool {
						var err = k8sCli.Get(testCtx, subscribedJobKey, &subscribedJob)
						if err != nil {
							GinkgoT().Log(err)
							return false
						}
						for _, cond := range subscribedJob.Status.Conditions {
							if cond.Type == batchv1.JobComplete {
								return cond.Status == corev1.ConditionTrue
							}
						}
						return false
					}, 30, 1).Should(BeTrue())
				})

			})

		})

	})

})

type judgeFunc func(edgev1alpha1.DeviceLink) bool

func doDeviceLinkJudgment(judge judgeFunc) {
	var deviceLinkKey = object.GetNamespacedName(&testDeviceLink)
	Eventually(func() bool {
		if err := k8sCli.Get(testCtx, deviceLinkKey, &testDeviceLink); err != nil {
			Fail(err.Error())
		}
		return judge(testDeviceLink)
	}, 300, 3).Should(BeTrue())
}

func correctNodeSpec() {
	var targetNode, err = node.GetValidWorker(testCtx, k8sCli)
	Expect(err).ShouldNot(HaveOccurred())
	err = k8sCli.Get(testCtx, object.GetNamespacedName(&testDeviceLink), &testDeviceLink)
	Expect(err).Should(Succeed())
	var patch = []byte(fmt.Sprintf(`{"spec":{"adaptor":{"node":"%s"}}}`, targetNode))
	Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func invalidNodeSpec() {
	var err = k8sCli.Get(testCtx, object.GetNamespacedName(&testDeviceLink), &testDeviceLink)
	Expect(err).Should(Succeed())
	var patch = []byte(`{"spec":{"adaptor":{"node":"wrong-node"}}}`)
	Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func isNodeExistedTrue() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetNodeExistedStatus() == metav1.ConditionTrue
	}
	doDeviceLinkJudgment(judge)
}

func isNodeExistedFalse() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetNodeExistedStatus() == metav1.ConditionFalse
	}
	doDeviceLinkJudgment(judge)
}

func correctModelSpec() {
	var err = k8sCli.Get(testCtx, object.GetNamespacedName(&testDeviceLink), &testDeviceLink)
	Expect(err).Should(Succeed())
	var patch = []byte(`{"spec":{"model":{"apiVersion":"devices.edge.cattle.io/v1alpha1"}}}`)
	Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func invalidModelSpec() {
	var err = k8sCli.Get(testCtx, object.GetNamespacedName(&testDeviceLink), &testDeviceLink)
	Expect(err).Should(Succeed())
	var patch = []byte(`{"spec":{"model":{"apiVersion":"wrong-apiVersion"}}}`)
	Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func isModelExistedTrue() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetModelExistedStatus() == metav1.ConditionTrue
	}
	doDeviceLinkJudgment(judge)
}

func isModelExistedFalse() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetModelExistedStatus() == metav1.ConditionFalse
	}
	doDeviceLinkJudgment(judge)
}

func correctAdaptorSpec() {
	var err = k8sCli.Get(testCtx, object.GetNamespacedName(&testDeviceLink), &testDeviceLink)
	Expect(err).Should(Succeed())
	var patch = []byte(`{"spec":{"adaptor":{"name":"adaptors.edge.cattle.io/mqtt"}}}`)
	Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func invalidAdaptorSpec() {
	var err = k8sCli.Get(testCtx, object.GetNamespacedName(&testDeviceLink), &testDeviceLink)
	Expect(err).Should(Succeed())
	var patch = []byte(`{"spec":{"adaptor":{"name":"wrong-adaptor-name"}}}`)
	Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func isAdaptorExistedTrue() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetAdaptorExistedStatus() == metav1.ConditionTrue
	}
	doDeviceLinkJudgment(judge)
}

func isAdaptorExistedFalse() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetAdaptorExistedStatus() == metav1.ConditionFalse
	}
	doDeviceLinkJudgment(judge)
}

func correctDeviceSpec() {
	var err = k8sCli.Get(testCtx, object.GetNamespacedName(&testDeviceLink), &testDeviceLink)
	Expect(err).Should(Succeed())
	var patch = []byte(`{"spec":{"template":{"spec":{"protocol":{"pattern":"AttributedTopic"}}}}}`)
	Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func invalidDeviceSpec() {
	var err = k8sCli.Get(testCtx, object.GetNamespacedName(&testDeviceLink), &testDeviceLink)
	Expect(err).Should(Succeed())
	var patch = []byte(`{"spec":{"template":{"spec":{"protocol":{"pattern":"wrong-pattern"}}}}}`)
	Expect(k8sCli.Patch(testCtx, &testDeviceLink, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func isDeviceConnectedTrue() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetDeviceConnectedStatus() == metav1.ConditionTrue
	}
	doDeviceLinkJudgment(judge)
}

func isDeviceConnectedFalse() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetDeviceConnectedStatus() == metav1.ConditionFalse
	}
	doDeviceLinkJudgment(judge)
}

func isDeviceCreatedFalse() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetDeviceCreatedStatus() == metav1.ConditionFalse || deviceLink.GetDeviceConnectedStatus() == metav1.ConditionFalse
	}
	doDeviceLinkJudgment(judge)
}

func deleteMQTTAdaptorPods() {
	Expect(k8sCli.DeleteAllOf(testCtx, &corev1.Pod{}, client.InNamespace("octopus-system"), client.MatchingLabels{"app.kubernetes.io/name": "octopus-adaptor-mqtt"})).
		Should(Succeed())
}

func deleteLimbsPods() {
	Expect(k8sCli.DeleteAllOf(testCtx, &corev1.Pod{}, client.InNamespace("octopus-system"), client.MatchingLabels{"app.kubernetes.io/component": "limb"})).
		Should(Succeed())
}

func isMQTTAdaptorPodsError() {
	var podList corev1.PodList
	Eventually(func() bool {
		if err := k8sCli.List(testCtx, &podList, client.InNamespace("octopus-system"), client.MatchingLabels{"app.kubernetes.io/name": "octopus-adaptor-mqtt"}); err != nil {
			Fail(err.Error())
		}
		for _, pod := range podList.Items {
			for _, condition := range pod.Status.Conditions {
				if condition.Type == "Ready" && condition.Status == "False" {
					return true
				}
			}
		}
		return false
	}, 300, 1).Should(BeTrue())
}

func deleteCorrespondingNode() {
	var correspondingNode = corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: testDeviceLink.Spec.Adaptor.Node,
		},
	}
	Expect(k8sCli.Delete(testCtx, &correspondingNode)).Should(Succeed())
}

func redeployCorrespondingNode() {
	Expect(testEnv.AddWorker(testDeviceLink.Spec.Adaptor.Node)).Should(Succeed())
}

func deleteMQTTDeviceModel() {
	var crd = v1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mqttdevices.devices.edge.cattle.io",
		},
	}
	Expect(k8sCli.Delete(testCtx, &crd)).Should(Succeed())
}

func redeployMQTTDeviceModel() {
	Expect(exec.RunKubectl(nil, GinkgoWriter, "apply", "-f", filepath.Join(testCurrDir, "deploy", "manifests", "crd", "base", "devices.edge.cattle.io_mqttdevices.yaml"))).
		Should(Succeed())
}
