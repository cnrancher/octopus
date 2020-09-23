package adaptor

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/rancher/octopus/test/framework/envtest/printer"
	"github.com/rancher/octopus/test/util/exec"
	"github.com/rancher/octopus/test/util/fuzz"
)

var (
	testCtx       context.Context
	testCtxCancel context.CancelFunc

	testMQTTBrokerContainer = exec.NewContainer()
	testMQTTBrokerAddress   string
)

func TestPhysical(t *testing.T) {
	defer GinkgoRecover()

	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"physical suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	defer close(done)

	testCtx, testCtxCancel = context.WithCancel(context.Background())

	By("starting test MQTT broker")
	Expect(startMQTTBrokerContainer()).Should(Succeed())
}, 600)

var _ = AfterSuite(func(done Done) {
	defer close(done)

	By("tearing down test MQTT broker")
	if err := stopMQTTBrokerContainer(); err != nil {
		GinkgoT().Logf("failed to stop test MQTT broker, %v", err)
	}

	if testCtxCancel != nil {
		testCtxCancel()
	}
}, 600)

func startMQTTBrokerContainer() error {
	// generate a random port
	var ports, err = fuzz.FreePorts(1)
	if err != nil {
		return errors.Wrap(err, "failed to get listening port of MQTT broker service")
	}
	testMQTTBrokerAddress = fmt.Sprintf("tcp://127.0.0.1:%d", ports[0])

	err = testMQTTBrokerContainer.Start(testCtx, exec.DockerContainerStartConfiguration{
		Config: &container.Config{
			Image: "eclipse-mosquitto:1.6.12",
			ExposedPorts: nat.PortSet{
				"1883/tcp": struct{}{},
			},
			Healthcheck: &container.HealthConfig{
				Test:        []string{"CMD-SHELL", "mosquitto_sub -t '$SYS/#' -C 1 | grep -v Error || exit 1"},
				Interval:    10 * time.Second,
				Timeout:     5 * time.Second,
				StartPeriod: 5 * time.Second,
				Retries:     3,
			},
		},
		HostConfig: &container.HostConfig{
			AutoRemove: true,
			PortBindings: nat.PortMap{
				"1883/tcp": []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: strconv.Itoa(ports[0]),
					},
				},
			},
		},
	})
	return err
}

func stopMQTTBrokerContainer() error {
	return testMQTTBrokerContainer.Stop(testCtx, exec.DockerContainerStopConfiguration{})
}
