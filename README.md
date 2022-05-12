# WLM-operator

The wlm-operator project is developed to explore interaction between the Kubernetes and HPC worlds.

---

**WLM operator** is a Kubernetes operator implementation, capable of submitting and
monitoring WLM jobs, while using all of Kubernetes features, such as smart scheduling and volumes.

WLM operator connects Kubernetes node with a whole WLM cluster, which enables multi-cluster scheduling.
In other words, Kubernetes integrates with WLM as one to many.

Each WLM partition(queue) is represented as a dedicated virtual node in Kubernetes. WLM operator
can automatically discover WLM partition resources(CPUs, memory, nodes, wall-time) and propagates them
to Kubernetes by labeling virtual node. Those node labels will be respected during Slurm job scheduling so that a
job will appear only on a suitable partition with enough resources.

Right now WLM-operator supports only SLURM clusters. But it's easy to add a support for another WLM. For it you need to implement a [GRPc server](https://github.com/dptech-corp/wlm-operator/blob/master/pkg/workload/api/workload.proto). You can use [current SLURM implementation](https://github.com/dptech-corp/wlm-operator/blob/master/internal/red-box/api/slurm.go) as a reference.

<p align="center">
  <img style="width:100%;" height="600" src="./docs/integration.svg">
</p>

## Installation

Since wlm-operator is now built with [go modules](https://github.com/golang/go/wiki/Modules)
there is no need to create standard [go workspace](https://golang.org/doc/code.html). If you still
prefer keeping source code under `GOPATH` make sure `GO111MODULE` is set. 

### Prerequisites

- Go 1.11+

### Installation steps

Installation process is required to connect Kubernetes with Slurm cluster.

*NOTE*: further described installation process for a single Slurm cluster,
the same steps should be performed for each cluster to be connected.

1. Login the Slurm cluster as a user, all submitted Slurm jobs will be executed on behalf
of that user. Make sure the user has execute permissions for the following Slurm binaries:`sbatch`,
`scancel`, `sacct` and `scontol`.

2. Clone the repo.
```bash
git clone https://github.com/dptech-corp/wlm-operator
```

3. Build and start *red-box* – a gRPC proxy between Kubernetes and a Slurm cluster.
```bash
cd wlm-operator && make
```
Run `./bin/red-box` in the background.
By default red-box listens on `/var/run/syslurm/red-box.sock`, you can specify the socket path by `-socket`, e.g.
```
./bin/red-box -socket /var/run/user/$(id -u)/syslurm/red-box.sock
```

4. Forward the red-box socket on the Slurm cluster
to a local socket on a Kubernetes node through SSH
```
ssh -nNT -L /var/run/syslurm/red-box.sock:/var/run/syslurm/red-box.sock username@cluster-ip
```

5. Set up Slurm operator in Kubernetes.
```bash
kubectl apply -f deploy/crds/slurm_v1alpha1_slurmjob.yaml
kubectl apply -f deploy/crds/wlm_v1alpha1_wlmjob.yaml
kubectl apply -f deploy/operator-rbac.yaml
kubectl apply -f deploy/operator.yaml
```
This will create new [CRD](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) that
introduces `SlurmJob` to Kubernetes. After that, Kubernetes controller for `SlurmJob` CRD is set up as a Deployment.

6. Start up configurator that will bring up a virtual node for each partition in the Slurm cluster.
```bash
kubectl apply -f deploy/configurator.yaml
```

Make sure the configurator pod is scheduled to the node in the Step 4.
After all those steps Kubernetes cluster is ready to run SlurmJobs. 
```
$ kubectl get nodes
NAME                            STATUS   ROLES                  AGE    VERSION
minikube                        Ready    control-plane,master   49d    v1.22.3
slurm-minikube-cpu              Ready    agent                  131m   v1.13.1-vk-N/A
slurm-minikube-dplc-ai-v100x8   Ready    agent                  131m   v1.13.1-vk-N/A
slurm-minikube-v100             Ready    agent                  131m   v1.13.1-vk-N/A
```

## Usage

The most convenient way to submit them is using YAML files, take a look at [basic examples](/examples).

```yaml
apiVersion: wlm.sylabs.io/v1alpha1
kind: SlurmJob
metadata:
  name: prepare-and-results
spec:
  batch: |
    #!/bin/sh
    #SBATCH --nodes=1
    echo Hello
    mkdir bar
    cp inputs/foo.txt bar/output.txt
    echo slurm >> bar/output.txt
  nodeSelector:
    kubernetes.io/hostname: slurm-minikube-cpu
  prepare:
    to: .
    mount:
      name: inputs
      hostPath:
        path: /home/docker/inputs
        type: DirectoryOrCreate
  results:
    from: bar
    mount:
      name: outputs
      hostPath:
        path: /home/docker/outputs
        type: DirectoryOrCreate
```

In the example above we will upload data in `/home/docker/inputs` located on a k8s node where job has been scheduled to Slurm, submit a Slurm job,
and collect the results to `/home/docker/outputs` located on the k8s node. Generally, job results
can be collected to any supported [k8s volume](https://kubernetes.io/docs/concepts/storage/volumes/).

Slurm job specification will be processed by operator and a dummy pod will be scheduled in order to transfer job
specification to a specific queue. That dummy pod will not have actual physical process under that hood, but instead 
its specification will be used to schedule slurm job directly on a connected cluster. To prepare data and collect results, another two pods
will be created with UID and GID 1000 (default values), so you should make sure it has a write access to 
a volume where you want to store the results (host directory `/home/docker/outputs` in the example above).
The UID and GID are inherited from virtual kubelet that spawns the pod, and virtual kubelet inherits them
from configurator (see `runAsUser` in [configurator.yaml](./deploy/configurator.yaml)).

After preparing the input data on the k8s node
```
$ ls /home/docker/inputs
foo.txt
```

you can submit the job:
```bash
$ kubectl apply -f examples/prepare-and-results.yaml 
slurmjob.wlm.sylabs.io "prepare-and-results" created

$ kubectl get slurmjob
NAME                   AGE   STATUS
prepare-and-results    66s   Succeeded

$ kubectl get pod
NAME                                             READY   STATUS         RESTARTS   AGE
prepare-and-results-job-prepare                  0/1     Completed      0          26s
prepare-and-results-job                          0/1     Job finished   0          17s
prepare-and-results-job-collect                  0/1     Completed      0          9s
```

Validate job results appeared on the node:
```bash
$ ls /home/docker/outputs
bar

$ kubectl logs prepare-and-results-job
Hello
```


### Data preparation and results collection

Slurm operator supports file transfer between [k8s volume](https://kubernetes.io/docs/concepts/storage/volumes/) and Slurm cluster
so that a user won't need to have access Slurm cluster manually.

_NOTE_: file transfer is a network and IO consuming task, so transfer large files (e.g. 1Gb of data) may not be a great idea.


## Configuring red-box

By default red-box performs automatic resources discovery for all partitions.
However, it's possible to setup available resources for a partition manually with in the config file.
The following resources can be specified: `nodes`, `cpu_per_node`, `mem_per_node` and `wall_time`. 
Additionally you can specify partition features there, e.g. available software or hardware. 
Config path should be passed to red-box with the `--config` flag.

Config example:
```yaml
patition1:
  nodes: 10
  mem_per_node: 2048 # in MBs
  cpu_per_node: 8
  wall_time: 10h 
partition2:
  nodes: 10
  # mem, cpu and wall_time will be automatic discovered
partition3:
  additional_feautres:
    - name: singularity
      version: 3.2.0
    - name: nvidia-gpu
      version: 2080ti-cuda-7.0
      quantity: 20
```


## Vagrant

If you want to try wlm-operator locally before updating your production cluster, use vagrant that will automatically
install and configure all necessary software:

```bash
cd vagrant
vagrant up && vagrant ssh k8s-master
```
_NOTE_: `vagrant up` may take about 15 minutes to start as k8s cluster will be installed from scratch.

Vagrant will spin up two VMs: a k8s master and a k8s worker node with Slurm installed.
If you wish to set up more workers, fell free to modify `N` parameter in [Vagrantfile](./vagrant/Vagrantfile).
