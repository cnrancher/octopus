---
name: Report Bug
about: Report a bug encountered while operating Octopus
labels: bug
---

<!-- [1] Please search for existing issues first. If this is a derived issue, 
please link the existing issues(see below).
-->

<!-- [2] Notice:
Long-term unresponsive and unreproducible issue will be closed.
-->

<!-- [3] Fill below content please. -->
**Environment information:**

- Octopus version <!-- master, branch name, tag name, image tag -->: master
- Installer <!-- *, yaml(all-in-one, without-webhook), helm -->: *

**Cluster information:**

- Cluster vendor <!-- k3s, kind, or others -->: k3s
- Machine type <!-- cloud(gce, aws, azure, ...), vm(virtual box, multipass, vmware, ...), metal -->: vm
- Kubernetes version (use `kubectl version`): 
- Docker version (use `docker version`): 

**Machine specification:**

| Role | Number | Specification (CPU/RAM) |
|:---:|:---:|:---:|
| <!-- master, control plane, etcd --> master | 1 | 4U8G |
| agent | 2 | 4U8G | 

**Steps to reproduce (least amount of steps as possible):**

1.
1.
1.

**Result:**


**Other details that may be helpful:**


**Except:**


**Related Issues:**

