package physical

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/rancher/octopus/adaptors/agent-device/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	apps "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	TokenPath   = "/var/lib/rancher/k3s/server/node-token"
	AdaptorName = "agent-device-adaptor"

	AdaptorAnnotationName = "edge.cattle.io/adaptor-name"
	GroupLabelName        = "devices.edge.cattle.io/group"
)

type Device interface {
	Configure(spec v1alpha1.AgentDeviceGroupSpec)
	On()
	Shutdown()
}

func NewDevice(log logr.Logger, name types.NamespacedName, clientSet *kubernetes.Clientset, handler DataHandler) Device {
	d := &device{
		log:       log,
		name:      name,
		clientSet: clientSet,
		handler:   handler,
	}

	d.nodeClient = clientSet.CoreV1().Nodes()
	d.appClient = clientSet.AppsV1()

	return d
}

type device struct {
	sync.Mutex

	stop chan struct{}

	log  logr.Logger
	name types.NamespacedName

	handler DataHandler

	clientSet *kubernetes.Clientset

	nodeClient corev1.NodeInterface
	appClient  apps.AppsV1Interface

	status v1alpha1.AgentDeviceGroupStatus
	spec   v1alpha1.AgentDeviceGroupSpec
}

func (d *device) Configure(spec v1alpha1.AgentDeviceGroupSpec) {
	d.Lock()
	defer d.Unlock()

	d.spec = spec
	// update k3s server's register url if specified by user
	url := spec.ServerURL
	if url != "" {
		if !strings.Contains(url, ":") {
			url += ":6443"
		}
		d.status.RegisterCommand = d.registerCommand(url)
	}

	// delete old apps found in status and update apps found in spec
	d.deleteOldApps(spec)
	d.updateApps(spec)

	d.handler(d.name, d.status)
}

func (d *device) deleteOldApps(spec v1alpha1.AgentDeviceGroupSpec) {
Status:
	for _, appStatus := range d.status.Apps {
		for _, app := range spec.Apps {
			// found same app in spec, not going to delete this app
			if app.Name == appStatus.Name && app.Namespace == appStatus.Namespace {
				continue Status
			}
		}
		daemonSets := d.appClient.DaemonSets(appStatus.Namespace)
		if err := daemonSets.Delete(appStatus.Name, &metav1.DeleteOptions{}); err != nil {
			d.log.Error(err, "Fail to delete daemonSet", "name", appStatus.Name)
		}
	}
}

func (d *device) updateApps(spec v1alpha1.AgentDeviceGroupSpec) {
	for _, app := range spec.Apps {
		name := app.Name
		template := app.Template

		template.Spec.NodeSelector = map[string]string{"devices.edge.cattle.io/group": d.name.Name}
		newDs := ConstructDaemonSet(name, app.Namespace, d.name.Name, template)

		daemonSets := d.appClient.DaemonSets(app.Namespace)
		daemonSet, err := daemonSets.Get(name, metav1.GetOptions{})
		if err != nil {
			// daemonSet does not exist, create new one
			if errors.IsNotFound(err) {
				daemonSet, err = daemonSets.Create(newDs)
				if err != nil {
					d.log.Error(err, "Fail to create daemonSet", "name", name)
				}
			} else {
				d.log.Error(err, "Fail to get daemonSet", "name", name)
			}
		} else {
			// daemonSet exists, update the old one
			dsCopy := daemonSet.DeepCopy()
			dsCopy.Spec = newDs.Spec

			daemonSet, err = daemonSets.Update(dsCopy)
			if err != nil {
				d.log.Error(err, "Fail to update daemonSet", "name", name)
			}
		}
		d.updateAppStatus(daemonSet)
	}
}

func (d *device) updateAppStatus(daemonSet *appsv1.DaemonSet) {
	appStatus := v1alpha1.AppStatus{
		Name:            daemonSet.Name,
		Namespace:       daemonSet.Namespace,
		DaemonSetStatus: daemonSet.Status,
		UpdatedAt:       metav1.Time{Time: time.Now()},
	}
	found := false
	for i, property := range d.status.Apps {
		if property.Name == appStatus.Name {
			d.status.Apps[i] = appStatus
			found = true
			break
		}
	}
	if !found {
		d.status.Apps = append(d.status.Apps, appStatus)
	}
}

func (d *device) On() {
	if d.stop != nil {
		close(d.stop)
	}
	d.stop = make(chan struct{})

	// update register command when available server endpoints change
	go d.watchServerEndpoints()
	// create agent device when node join
	go d.watchNodes()
	// watch daemonSets to update app status
	go d.watchDaemonSets()
}

func (d *device) watchDaemonSets() {
	label := fmt.Sprintf("%s=%s", GroupLabelName, d.name.Name)

	daemonSets, err := d.appClient.DaemonSets("").Watch(metav1.ListOptions{LabelSelector: label})
	if err != nil {
		d.log.Error(err, "Error watch daemonSets", "group", d.name)
		return
	}

	d.log.Info("watching daemonSets")
	for {
		select {
		case <-d.stop:
			return
		case event := <-daemonSets.ResultChan():
			daemonSet := event.Object.(*appsv1.DaemonSet)
			d.updateAppStatus(daemonSet)

			d.handler(d.name, d.status)
		}
	}
}

func (d *device) watchServerEndpoints() {
	endpoints, err := d.clientSet.CoreV1().Endpoints("default").Watch(metav1.ListOptions{})
	if err != nil {
		d.log.Error(err, "Error watch endpoints", "group", d.name)
		return
	}

	d.log.Info("watching endpoints")
	for {
		select {
		case <-d.stop:
			return
		case event := <-endpoints.ResultChan():
			// if k3s server's register url is not specified by user, fetch from endpoint
			if d.spec.ServerURL == "" {
				d.handleEndpointsEvents(event)
			}
		}
	}
}

func (d *device) handleEndpointsEvents(event watch.Event) {
	d.Lock()
	defer d.Unlock()
	endpoints := event.Object.(*v1.Endpoints)
	if endpoints.Name == "kubernetes" {
		subset := endpoints.Subsets[0]
		d.log.Info("Endpoint changes", "endpoint", subset)

		ip := subset.Addresses[0].IP
		port := subset.Ports[0]

		d.status.RegisterCommand = d.registerCommand(fmt.Sprintf("%s://%s:%d", port.Name, ip, port.Port))
		d.handler(d.name, d.status)
	}
}

func (d *device) registerCommand(endpoint string) string {
	// read token from mounted server token file
	content, err := ioutil.ReadFile(TokenPath)
	if err != nil {
		d.log.Error(err, "Fail to read token file")
	}
	token := strings.TrimSpace(string(content))

	label := GroupLabelName + "=" + d.name.Name
	return fmt.Sprintf(
		"curl -sfL https://get.k3s.io | K3S_URL=%s K3S_TOKEN=%s sh -s - agent --node-label %s", endpoint, token, label)
}

func (d *device) watchNodes() {
	name := d.name.Name
	label := fmt.Sprintf("%s=%s", GroupLabelName, name)

	nodes, err := d.nodeClient.Watch(metav1.ListOptions{LabelSelector: label})
	if err != nil {
		d.log.Error(err, "Error watch nodes", "group", name)
		return
	}
	d.log.Info("Watching nodes", "group", name)

	for {
		select {
		case <-d.stop:
			return
		case <-nodes.ResultChan():
			d.handleNodeEvents(label)
		}
	}
}

func (d *device) handleNodeEvents(label string) {
	d.Lock()
	defer d.Unlock()

	nodes, err := d.nodeClient.List(metav1.ListOptions{LabelSelector: label})
	if err != nil {
		d.log.Error(err, "Error list nodes", "label", label)
		return
	}

	d.log.Info("Updating nodes list in agent device group status", "label", label)
	d.status.Nodes = make([]string, len(nodes.Items))
	for i, node := range nodes.Items {
		d.status.Nodes[i] = node.Name
	}
	d.handler(d.name, d.status)
}

func (d *device) Shutdown() {
	// delete daemonSet apps
	d.deleteDaemonSets()

	// if specified should delete nodes, delete the added nodes when delete this AgentDeviceGroup
	if d.spec.DeleteNodes {
		if err := d.deleteNodes(); err != nil {
			d.log.Error(err, "Fail to delete nodes", "group", d.name)
		}
	}

	if d.stop != nil {
		close(d.stop)
	}
	d.log.Info("Closed connection")
}

func (d *device) deleteDaemonSets() {
	for _, app := range d.status.Apps {
		daemonSets := d.appClient.DaemonSets(app.Namespace)
		if err := daemonSets.Delete(app.Name, &metav1.DeleteOptions{}); err != nil {
			d.log.Error(err, "Fail to delete daemonSet", "name", app.Name)
		}
	}
}

func (d *device) deleteNodes() error {
	name := d.name.Name
	label := fmt.Sprintf("%s=%s", GroupLabelName, name)

	if err := d.nodeClient.DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: label}); err != nil {
		return err
	}
	return nil
}

func ConstructDaemonSet(name, namespace, creator string, template v1alpha1.PodTemplate) *appsv1.DaemonSet {
	daemonSet := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				AdaptorAnnotationName: AdaptorName,
			},
			Labels: map[string]string{
				GroupLabelName: creator,
			},
		},
		Spec: appsv1.DaemonSetSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: template.Labels,
				},
				Spec: template.Spec,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: template.Labels,
			},
		},
	}
	return daemonSet
}
