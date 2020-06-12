# Monitor Octopus

<!-- toc -->

- [Metrics Category](#metrics-category)
    + [Exposing from Controller Runtime](#exposing-from-controller-runtime)
        - [Controller metrics](#controller-metrics)
        - [Webhook metrics](#webhook-metrics)
    + [Exposing from Kubernetes client](#exposing-from-kubernetes-client)
        - [Rest client metrics](#rest-client-metrics)
        - [Reflector metrics](#reflector-metrics)(***Deprecated***)
        - [Workqueue metrics](#workqueue-metrics)
    + [Exposing from Prometheus client](#exposing-from-prometheus-client)
        - [Go runtime metrics](#go-runtime-metrics)
        - [Running process metrics](#running-process-metrics)
    + [Exposing from Octopus](#exposing-from-octopus)
        - [Limb metrics](#limb-metrics)
- [Monitor](#monitor)
    + [Grafana Dashboard](#grafana-dashboard)
    + [Integrate with Prometheus Operator](#integrate-with-prometheus-operator)

<!-- /toc -->

Octopus is built on [sigs.k8s.io/controller-runtime](https://github.com/kubernetes-sigs/controller-runtime), so some metrics are related to controller-runtime and [client-go](https://github.com/kubernetes/client-go). At the same time, [github.com/prometheus/client_golang](https://github.com/prometheus/client_golang) provides some metrics for [Go runtime](https://golang.org/pkg/runtime/) and process state.

## Metrics Category

> In the "Type" column, use the first letter to represent the corresponding word: G - Gauge, C - Counter, H - Histogram, S - Summary.

### Exposing from Controller Runtime

#### Controller metrics

| Type | Name | Description | Usage |
|:---:|:---|:---|:---|
| C | [`controller_runtime_reconcile_total`](../../vendor/sigs.k8s.io/controller-runtime/pkg/internal/controller/metrics/metrics.go#L25-L32) | Total number of reconciliations per controller. |  |
| C | [`controller_runtime_reconcile_errors_total`](../../vendor/sigs.k8s.io/controller-runtime/pkg/internal/controller/metrics/metrics.go#L34-L39) | Total number of reconciliation errors per controller. |  |
| H | [`controller_runtime_reconcile_time_seconds`](../../vendor/sigs.k8s.io/controller-runtime/pkg/internal/controller/metrics/metrics.go#L41-L46) | Length of time per reconciliation per controller. |  |

#### Webhook metrics

| Type | Name | Description | Usage |
|:---:|:---|:---|:---|
| H | [`controller_runtime_webhook_latency_seconds`](../../vendor/sigs.k8s.io/controller-runtime/pkg/webhook/internal/metrics/metrics.go#L26-L34) | Histogram of the latency of processing admission requests. | |

### Exposing from Kubernetes client

#### Rest client metrics

| Type | Name | Description | Usage |
|:---:|:---|:---|:---|
| C | [`rest_client_requests_total`](../../vendor/sigs.k8s.io/controller-runtime/pkg/metrics/client_go_adatper.go#L44-L49) | Number of HTTP requests, partitioned by status code, method, and host. | |
| H | [`rest_client_request_latency_seconds`](../../vendor/sigs.k8s.io/controller-runtime/pkg/metrics/client_go_adatper.go#L35-L42) | Request latency in seconds. Broken down by verb and URL. | |

#### Reflector metrics

> Deprecated by [kubernetes/pull#74636](https://github.com/kubernetes/kubernetes/pull/74636) to fix the memory leak in kubelet.

| Type | Name | Description | Usage |
|:---:|:---|:---|:---|
| G | [`reflector_last_resource_version`](../../vendor/sigs.k8s.io/controller-runtime/pkg/metrics/client_go_adatper.go#L101-L105) | Last resource version seen for the reflectors. |  |
| C | [`reflector_lists_total`](../../vendor/sigs.k8s.io/controller-runtime/pkg/metrics/client_go_adatper.go#L59-L63) | Total number of API lists done by the reflectors. | |
| S | [`reflector_list_duration_seconds`](../../vendor/sigs.k8s.io/controller-runtime/pkg/metrics/client_go_adatper.go#L65-L69) | How long an API list takes to return and decode for the reflectors. | |
| S | [`reflector_items_per_list`](../../vendor/sigs.k8s.io/controller-runtime/pkg/metrics/client_go_adatper.go#L71-L75) | How many items an API list returns to the reflectors. |  |
| C | [`reflector_short_watches_total`](../../vendor/sigs.k8s.io/controller-runtime/pkg/metrics/client_go_adatper.go#L83-L87) | Total number of short API watches done by the reflectors. |  |
| C | [`reflector_watches_total`](../../vendor/sigs.k8s.io/controller-runtime/pkg/metrics/client_go_adatper.go#L77-L81) | Total number of API watches done by the reflectors. |  |
| S | [`reflector_watch_duration_seconds`](../../vendor/sigs.k8s.io/controller-runtime/pkg/metrics/client_go_adatper.go#L89-L93) | How long an API watch takes to return and decode for the reflectors. |  |
| S | [`reflector_items_per_watch`](../../vendor/sigs.k8s.io/controller-runtime/pkg/metrics/client_go_adatper.go#L95-L99) | How many items an API watch returns to the reflectors. |  |

#### Workqueue metrics

| Type | Name | Description | Usage |
|:---:|:---|:---|:---|
| G | [`workqueue_depth`](../../vendor/sigs.k8s.io/controller-runtime/pkg/metrics/workqueue.go#L44-L49) | Current depth of workqueue. | |
| G | [`workqueue_unfinished_work_seconds`](../../vendor/sigs.k8s.io/controller-runtime/pkg/metrics/workqueue.go#L90-L98) | How many seconds of work has done that is in progress and hasn't been observed by work_duration. Large values indicate stuck threads. One can deduce the number of stuck threads by observing the rate at which this increases. | |
| G | [`workqueue_longest_running_processor_seconds`](../../vendor/sigs.k8s.io/controller-runtime/pkg/metrics/workqueue.go#L104-L110) | How many seconds has the longest running processor for workqueue been running. | |
| C | [`workqueue_adds_total`](../../vendor/sigs.k8s.io/controller-runtime/pkg/metrics/workqueue.go#L55-L60) | Total number of adds handled by workqueue. | |
| C | [`workqueue_retries_total`](../../vendor/sigs.k8s.io/controller-runtime/pkg/metrics/workqueue.go#L116-L121) | Total number of retries handled by workqueue. | |
| H | [`workqueue_queue_duration_seconds`](../../vendor/sigs.k8s.io/controller-runtime/pkg/metrics/workqueue.go#L66-L72) | How long in seconds an item stays in workqueue before being requested. | |
| H | [`workqueue_work_duration_seconds`](../../vendor/sigs.k8s.io/controller-runtime/pkg/metrics/workqueue.go#L78-L84) | How long in seconds processing an item from workqueue takes. | |

### Exposing from Prometheus client

#### Go runtime metrics

| Type | Name | Description | Usage |
|:---:|:---|:---|:---|
| G | [`go_goroutines`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L66-L69) | Number of goroutines that currently exist. | |
| G | [`go_threads`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L70-L73) | Number of OS threads created. | |
| G | [`go_info`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L78-L81) | Information about the Go environment. | |
| S | [`go_gc_duration_seconds`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L74-L77) | A summary of the pause duration of garbage collection cycles. | |
| G | [`go_memstats_alloc_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L88-L94) | Number of bytes allocated and still in use. | |
| C | [`go_memstats_alloc_bytes_total`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L96-L102) | Total number of bytes allocated, even if freed. | |
| G | [`go_memstats_sys_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L104-L110) | Number of bytes obtained from system. | |
| C | [`go_memstats_lookups_total`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L112-L118) | Total number of pointer lookups. | |
| C | [`go_memstats_mallocs_total`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L120-L126) | Total number of mallocs. | |
| C | [`go_memstats_frees_total`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L128-L134) | Total number of frees. | |
| G | [`go_memstats_heap_alloc_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L136-L142) | Number of heap bytes allocated and still in use. | |
| G | [`go_memstats_heap_sys_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L144-L150) | Number of heap bytes obtained from system. | |
| G | [`go_memstats_heap_idle_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L152-L158) | Number of heap bytes waiting to be used. | |
| G | [`go_memstats_heap_inuse_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L160-L166) | Number of heap bytes that are in use. | |
| G | [`go_memstats_heap_released_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L168-L174) | Number of heap bytes released to OS. | |
| G | [`go_memstats_heap_objects`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L176-L182) | Number of allocated objects. | |
| G | [`go_memstats_stack_inuse_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L184-L190) | Number of bytes in use by the stack allocator. | |
| G | [`go_memstats_stack_sys_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L192-L198) | Number of bytes obtained from system for stack allocator. | |
| G | [`go_memstats_mspan_inuse_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L200-L206) | Number of bytes in use by mspan structures. | |
| G | [`go_memstats_mspan_sys_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L208-L214) | Number of bytes used for mspan structures obtained from system. | |
| G | [`go_memstats_mcache_inuse_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L216-L222) | Number of bytes in use by mcache structures. | |
| G | [`go_memstats_mcache_sys_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L224-L230) | Number of bytes used for mcache structures obtained from system. | |
| G | [`go_memstats_buck_hash_sys_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L232-L238) | Number of bytes used by the profiling bucket hash table. | |
| G | [`go_memstats_gc_sys_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L240-L246) | Number of bytes used for garbage collection system metadata. | |
| G | [`go_memstats_other_sys_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L248-L254) | Number of bytes used for other system allocations. | |
| G | [`go_memstats_next_gc_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L256-L262) | Number of heap bytes when next garbage collection will take place. | |
| G | [`go_memstats_last_gc_time_seconds`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L264-L270) | Number of seconds since 1970 of last garbage collection. | |
| G | [`go_memstats_gc_cpu_fraction`](../../vendor/github.com/prometheus/client_golang/prometheus/go_collector.go#L272-L278) | The fraction of this program's available CPU time used by the GC since the program started. | |

#### Running process metrics

| Type | Name | Description | Usage |
|:---:|:---|:---|:---|
| C | [`process_cpu_seconds_total`](../../vendor/github.com/prometheus/client_golang/prometheus/process_collector.go#L71-L75) | Total user and system CPU time spent in seconds. | |
| G | [`process_open_fds`](../../vendor/github.com/prometheus/client_golang/prometheus/process_collector.go#L76-L80) | Number of open file descriptors. | |
| G | [`process_max_fds`](../../vendor/github.com/prometheus/client_golang/prometheus/process_collector.go#L81-L85) | Maximum number of open file descriptors. | |
| G | [`process_virtual_memory_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/process_collector.go#L86-L90) | Virtual memory size in bytes. | |
| G | [`process_virtual_memory_max_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/process_collector.go#L91-L95) | Maximum amount of virtual memory available in bytes. | |
| G | [`process_resident_memory_bytes`](../../vendor/github.com/prometheus/client_golang/prometheus/process_collector.go#L96-L100) | Resident memory size in bytes. | |
| G | [`process_start_time_seconds`](../../vendor/github.com/prometheus/client_golang/prometheus/process_collector.go#L101-L105) | Start time of the process since unix epoch in seconds. | |

### Exposing from Octopus

#### Limb metrics

| Type | Name | Description | Usage |
|:---:|:---|:---|:---|
| G | [`limb_connect_connections`](../../pkg/metrics/limb/metrics.go#L12-L19) | How many connections are connecting adaptor. | |
| C | [`limb_connect_errors_total`](../../pkg/metrics/limb/metrics.go#L21-L28) | Total number of connecting adaptor errors. | |
| C | [`limb_send_errors_total`](../../pkg/metrics/limb/metrics.go#L30-L37) | Total number of errors of sending device desired to adaptor. | |
| H | [`limb_send_latency_seconds`](../../pkg/metrics/limb/metrics.go#L39-L46) | Histogram of the latency of sending device desired to adaptor. | |

## Monitor

By default, the metrics will be exposed on port `8080`(see [brain options](../../cmd/brain/options) and [limb options](../../cmd/limb/options)), they can be collected by [Prometheus](https://prometheus.io/) and visually analyzed through [Grafana](https://grafana.com/). Octopus provides a [ServiceMonitor definition YAML](../../deploy/e2e/integrate_with_prometheus_operator.yaml) to integrate with [Prometheus Operator](https://github.com/coreos/prometheus-operator), which is an easy tool to configure and manage Prometheus instances.

### Grafana Dashboard

For convenience, Octopus provides a [Grafana Dashboard](../../deploy/e2e/integrate_with_grafana.json) to visualize the monitoring metrics.

### Integrate with Prometheus Operator

Using [prometheus-operator HELM chart](https://github.com/helm/charts/blob/master/stable/prometheus-operator), you can easily set up a Prometheus Operator to monitor the Octopus. The following steps demonstrate how to run a Prometheus Operator on a local Kubernetes cluster:

1. Use [`cluster-k3d-spinup.sh`](../../hack/cluster-k3d-spinup.sh) to set up a local Kubernetes cluster via [k3d](https://github.com/rancher/k3d).
1. Follow the [installation guide of HELM](https://helm.sh/docs/intro/install/) to install helm tool, and then use `helm fetch --untar --untardir /tmp stable/prometheus-operator` the prometheus-operator chart to local `/tmp` directory.
1. Generate a deployment YAML from prometheus-operator chart as below, please replace the `<INGRESS_HTTP_PORT>` with the output of `cluster-k3d-spinup.sh`.
    ```shell
    INGRESS_HTTP_PORT=<the output of script>; helm template --namespace octopus-monitoring \
    --name octopus \
    --set defaultRules.create=false \
    --set global.rbac.pspEnabled=false \
    --set prometheusOperator.admissionWebhooks.patch.enabled=false \
    --set prometheusOperator.admissionWebhooks.enabled=false \
    --set prometheusOperator.kubeletService.enabled=false \
    --set prometheusOperator.tlsProxy.enabled=false \
    --set prometheusOperator.serviceMonitor.selfMonitor=false \
    --set alertmanager.enabled=false \
    --set grafana.defaultDashboardsEnabled=false \
    --set coreDns.enabled=false \
    --set kubeApiServer.enabled=false \
    --set kubeControllerManager.enabled=false \
    --set kubeEtcd.enabled=false \
    --set kubeProxy.enabled=false \
    --set kubeScheduler.enabled=false \
    --set kubeStateMetrics.enabled=false \
    --set kubelet.enabled=false \
    --set nodeExporter.enabled=false \
    --set prometheus.serviceMonitor.selfMonitor=false \
    --set prometheus.ingress.enabled=true \
    --set prometheus.ingress.hosts={localhost} \
    --set prometheus.ingress.paths={/prometheus} \
    --set prometheus.ingress.annotations.'traefik\.ingress\.kubernetes\.io\/rewrite-target'=/ \
    --set prometheus.prometheusSpec.externalUrl=http://localhost:${INGRESS_HTTP_PORT}/prometheus \
    --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false \
    --set prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues=false \
    --set prometheus.prometheusSpec.ruleSelectorNilUsesHelmValues=false \
    --set grafana.adminPassword=admin \
    --set grafana.rbac.pspUseAppArmor=false \
    --set grafana.rbac.pspEnabled=false \
    --set grafana.serviceMonitor.selfMonitor=false \
    --set grafana.testFramework.enabled=false \
    --set grafana.ingress.enabled=true \
    --set grafana.ingress.hosts={localhost} \
    --set grafana.ingress.path=/grafana \
    --set grafana.ingress.annotations.'traefik\.ingress\.kubernetes\.io\/rewrite-target'=/ \
    --set grafana.'grafana\.ini'.server.root_url=http://localhost:${INGRESS_HTTP_PORT}/grafana \
    /tmp/prometheus-operator > /tmp/prometheus-operator_all_in_one.yaml
    ```
1. Create `octopus-monitoring` Namespace via `kubectl create ns octopus-monitoring`.
1. Apply the prometheus-operator all-in-one deployment into the local cluster via `kubectl apply -f /tmp/prometheus-operator_all_in_one.yaml`.
1. Apply the Octopus all-in-one deployment via `kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus/master/deploy/e2e/all_in_one.yaml` or with admission webhooks deployment via `kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus/master/deploy/e2e/all_in_one_with_webhook.yaml`.
1. Apply the monitoring integration into the local cluster via `kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus/master/deploy/e2e/integrate_with_prometheus_operator.yaml`
1. Visit `http://localhost:${INGRESS_HTTP_PORT}/prometheus` to view the Prometheus web console through the browser, or visit `http://localhost:${INGRESS_HTTP_PORT}/grafana` to view the Grafana console(the administrator account is `admin/admin`).
1. (Optional) Import the [Octopus Overview dashboard](https://raw.githubusercontent.com/cnrancher/octopus/master/deploy/e2e/integrate_with_grafana.json) from Grafana console.
