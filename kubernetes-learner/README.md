
# Kubernetes

Kubernetes, also known as K8s, is an open source system for automating deployment, scaling, and management of containerized applications.

Kubernetes is an open-source container orchestration system for automating software deployment, scaling, and management. Originally designed by Google, the project is now maintained by a worldwide community of contributors, and the trademark is held by the Cloud Native Computing Foundation. 

## Core components of K8s cluster

![k8s-architecture](https://kubernetes.io/images/docs/components-of-kubernetes.svg)

### Controller Plane components
* **API Server** - The API server is a component of the Kubernetes control plane that exposes the Kubernetes API. The API server is the front end for the Kubernetes control plane.

* **ETCD** - Consistent and highly-available key value store used as Kubernetes' backing store for all cluster data. If your Kubernetes cluster uses etcd as its backing store,
* **Scheduler** - Control plane component that watches for newly created Pods with no assigned node, and selects a node for them to run on.
* **Controller Manager** - Control plane component that runs controller processes.

    Logically, each controller is a separate process, but to reduce complexity, they are all compiled into a single binary and run in a single process.


### Worker Node components
* **Kubelet** - An agent that runs on each node in the cluster. It makes sure that containers are running in a Pod.
* **KubeProxy** - kube-proxy is a network proxy that runs on each node in your cluster, implementing part of the Kubernetes Service concept. kube-proxy maintains network rules on nodes. These network rules allow network communication to your Pods from network sessions inside or outside of your cluster.
* **Container Runtime** - A fundamental component that empowers Kubernetes to run containers effectively. It is responsible for managing the execution and lifecycle of containers within the Kubernetes environment.

    Kubernetes supports container runtimes such as containerd, CRI-O, and any other implementation of the Kubernetes CRI (Container Runtime Interface).



### Cluster Architecture
The architectural concepts behind Kubernetes.

A Kubernetes cluster consists of a control plane plus a set of worker machines, called nodes, that run containerized applications. Every cluster needs at least one worker node in order to run Pods.

![K8s-cluster-architecture-docs](https://kubernetes.io/images/docs/kubernetes-cluster-architecture.svg)

The worker node(s) host the Pods that are the components of the application workload. The control plane manages the worker nodes and the Pods in the cluster. In production environments, the control plane usually runs across multiple computers and a cluster usually runs multiple nodes, providing fault-tolerance and high availability.

This document outlines the various components you need to have for a complete and working Kubernetes cluster.


## Kubectl

Kubernetes provides a command line tool for communicating with a Kubernetes cluster's control plane, using the Kubernetes API.

### How to use

```bash
kubectl [command] [TYPE] [NAME] [flags]
```

where command, TYPE, NAME, and flags are:

command: Specifies the operation that you want to perform on one or more resources, for example create, get, describe, delete.

**Example**:

```bash
kubectl get pod pod1
kubectl get pods pod1
kubectl get po pod1
```

TYPE: Specifies the resource type. Resource types are case-insensitive and you can specify the singular, plural, or abbreviated forms. For example, the following commands produce the same output:


### Kubectl - Installation

* **Installtion Guide:** https://kubernetes.io/docs/tasks/tools/

* **MiniKube Installation:** https://minikube.sigs.k8s.io/docs/start/?arch=%2Fmacos%2Farm64%2Fstable%2Fhomebrew#Service