# Deploy Octopus 

There are two ways to deploy Octopus, one is [Helm chart](https://helm.sh/), another one bases on [Kustomize](https://github.com/kubernetes-sigs/kustomize).

<!-- toc -->

- [Helm chart](#helm-chart)
- [Bases on Kustomize](#bases-on-kustomize)
    - [Animated quick demo](#animated-quick-demo)

<!-- /toc -->

## Helm chart

TODO

## Bases on Kustomize

Kustomize is an interesting tool, in solving Kubernetes application management, it uses a different idea then Helm, which calls [Declarative Application Management](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/architecture/declarative-application-management.md). 

The kustomize layout of Octopus looks like as follows:

```
deploy/manifests
    - crd/base - stores raw CRD files 
    - overlays - stores the topmost overlays
    - rbac - stores RBAC files
    - workload - stores workload files
```

During [Octopus generation stage](./develop.md), the topmost overlays will be rendered accordingly:

- `default` overlay -> `deploy/e2e/all_in_one.yaml`

### Animated quick demo

[![asciicast](https://asciinema.org/a/338649.svg)](https://asciinema.org/a/338649)

<details>
  <summary>process instruction</summary>
  <code>
  
    # deploy octopus without webhook
    kubectl apply -f deploy/e2e/all_in_one.yaml
    
    # confirm the octopus deployment
    kubectl get all -n octopus-system
    kubectl get crd | grep devicelinks
    
    # deploy a devicelink
    cat adaptors/dummy/deploy/e2e/dl_specialdevice.yaml
    kubectl apply -f adaptors/dummy/deploy/e2e/dl_specialdevice.yaml
    
    # confirm the state of devicelink
    kubectl get dl living-room-fan -n default
    
    # deploy dummy adaptor and model
    kubectl apply -f adaptors/dummy/deploy/e2e/all_in_one.yaml
    
    # confirm the dummy adaptor deployment
    kubectl get daemonset octopus-adaptor-dummy-adaptor -n octopus-system
    kubectl get crd | grep dummyspecialdevice
    
    # confirm the state of devicelink
    kubectl get dl living-room-fan -n default
    
    # watch the device instance
    kubectl get dummyspecialdevice living-room-fan -n default -w
    
  </code>
</details>
