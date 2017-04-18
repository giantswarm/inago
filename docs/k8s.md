# Kubernetes

With Inago you can easily deploy Kubernetes on a CoreOS cluster.
The [examples folder](https://github.com/giantswarm/inago/tree/master/examples) contains a simple [k8s example](https://github.com/giantswarm/inago/tree/master/examples/k8s). The Kubernetes deployment is split into
two groups: `k8s-master` and `k8s-node`.

## Note
The deployment sets up a Kubernetes cluster with insecure communication 
between the components. Use this only as a Demo or POC. If you want to deploy a
Kubernetes cluster for production please modify the example, so secure
communication is used: [Cluster TLS using OpenSSL](https://coreos.com/Kubernetes/docs/latest/openssl.html).
Further, this deployment uses a single Kubernetes master, so the API is not highly available. For HA deployments of Kubernetes check out [Building High-Availability Clusters](http://kubernetes.io/docs/admin/high-availability/).

## Deployment

You can use this deployment on any CoreOS fleet cluster. Below we work with a 3 node cluster. For a simple way to setup a CoreOS cluster on your laptop try the official [Vagrant setup](https://coreos.com/os/docs/latest/booting-on-vagrant.html).
Once your CoreOS cluster is up and running you can deploy Kubernetes in two easy steps:

First, we need to setup the network environment. This, needs to be done only once, but on all nodes, which is why we set this up to be a global unit (running on all machines).

```
$ inagoctl up k8s-network
2016-04-14 15:31:05.273 | INFO     | context.Background: Succeeded to submit group 'k8s-master'.
2016-04-14 15:32:01.532 | INFO     | context.Background: Succeeded to start group 'k8s-network'.
```

Then we can schedule the k8s-master group.

```
$ inagoctl up k8s-master
2016-04-14 15:33:09.396 | INFO     | context.Background: Succeeded to submit group 'k8s-master'.
2016-04-14 15:34:01.558 | INFO     | context.Background: Succeeded to start group 'k8s-master'.
```

You can check if it is up with `inagoctl status k8s-master`.

Now, we can schedule the nodes. As we have a 3 node cluster we'll be running 3 slices of nodes:

```
$ inagoctl up k8s-node 3
2016-04-14 15:41:26.824 | INFO     | context.Background: Succeeded to submit group 'k8s-node'.
2016-04-14 15:41:36.847 | INFO     | context.Background: Succeeded to start 3 slices for group 'k8s-node'
```

And we're done!

## Testing your Kubernetes cluster

You can check if your cluster is running with [`kubectl`](https://coreos.com/kubernetes/docs/latest/configure-kubectl.html).  If you download the statically compiled version directly from kubernetes you will need to symlink the hosts file into /etc with `sudo ln -s /usr/share/baselayout/hosts /etc/hosts`:

```
$ kubectl cluster-info
Kubernetes master is running at http://localhost:8080

$ kubectl version
Client Version: version.Info{Major:"1", Minor:"2", GitVersion:"v1.2.0", GitCommit:"5cb86ee022267586db386f62781338b0483733b3", GitTreeState:"clean"}
Server Version: version.Info{Major:"1", Minor:"2", GitVersion:"v1.2.0", GitCommit:"5cb86ee022267586db386f62781338b0483733b3", GitTreeState:"clean"}
```

A further `kubectl describe nodes` will show that we’re running 3 nodes with kubelet and proxy version 1.2.0 deployed.

Note: We assume running `kubectl` on the same node as the API server here. If you're running on a different node, you can use the `-s` flag to connect to the right one.

For finding the right node, login to a CoreOS machine and run

```
$ fleetctl list-units -fields=unit,machine --full --no-legend 2>/dev/null | grep ^k8s-master-api-server.service | cut -d/ -f2 | paste -d, -s
172.17.8.102
```

By now, everything should be ready to deploy our first pod. We use a prepared a little image based on the [Kubernetes hello world example](http://kubernetes.io/docs/hellonode/) that we can start with

```
kubectl run hello-node --image=puja/k8s-hello-node:v1 --port=8080
```

This shouldn’t take much, and we can check if it is running.

```
$ kubectl get deployments
NAME         DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
hello-node   1         1         1            1           1m
$ kubectl get pods
NAME                          READY     STATUS    RESTARTS   AGE
hello-node-3488712740-zponq   1/1       Running   0          1m
```

As we have 3 nodes let’s scale the pod up to 3 replicas.

```
$ kubectl scale deployment hello-node --replicas=3
```

Wait a while and check again.

```
$ kubectl get deployments
NAME         DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
hello-node   3         3         3            3           6m
```

Voilà, we have 3 replicas of our hello world pod available.

## Updating your Kubernetes cluster

So now that we have a running Kubernetes cluster, we might want to update it at some point in time. The `update` command of Inago makes this quite easy.

### Updating nodes

First, we check which version of the nodes is currently deployed with `kubectl describe nodes`. You'll see that we have 3 nodes running kubelet and proxy versions 1.2.0 each. As of this writing, the latest stable version is 1.2.2, so let's update those nodes! The update command is pretty simple and very similar to how kubectl manages rolling updates.

First, we need to edit a line in each unit file of the group to make the unit use the latest tag:

`Environment="IMAGE=giantswarm/k8s-kubelet:1.2.2"` in `k8s-node-kubelet@.service`

and

`Environment="IMAGE=giantswarm/k8s-proxy:1.2.2"` in `k8s-node-proxy@.service`

Now, a single command lets us update the whole k8s-node group

```
$ inagoctl update k8s-node --max-growth=0 --min-alive=2 --ready-secs=60
2016-04-14 15:45:22.988 | INFO     | context.Background: Succeeded to update 3 slices for group 'k8s-node': [045 3a9 79b].
```

And we're done! After a while, all nodes will be updated and another `kubectl describe nodes` will show that we’re now running 3 nodes with kubelet and proxy version 1.2.2 deployed.

Looking closely at the update command above we can inspect some of its options. `--max-growth=0` tells Inago to not start any additional instances, which means it has to remove an instance to be able to start a new one. In this example this is due to the fact that we don't have any free machine left. `--min-alive=2` tells Inago to keep at least two instances of the group alive during the update process. the `ready-secs` flag determines how long Inago waits between rounds of updating. This gives the new node a bit of time to start replicating that pod we scheduled.

Watching `kubectl get deployments` during the update process, we would see that available instances of our hello-node pod shortly go down to 2, while a Kubernetes node is being updated and then quickly come up again to 3.

### Updating the Master

Now sadly, we didn't use an HA deployment of the Kubernetes master, so Inago's update command currently won't work here. Accordingly, we will have a short downtime of our API while we update this group. Luckily, Kubernetes nodes and their pods don't rely on the API server that much and will usually stay running while we update the master.

Again we need to update our unit files to use the newest image. After that a short series of commands will update the group.

```
$ inagoctl stop k8s-master
2016-04-14 16:32:08.465 | INFO     | context.Background: Succeeded to stop group 'k8s-master'.

$ inagoctl destroy k8s-master
2016-04-14 16:32:27.384 | INFO     | context.Background: Succeeded to destroy group 'k8s-master'.

$ inagoctl up k8s-master
2016-04-14 16:35:23.503 | INFO     | context.Background: Succeeded to submit group 'k8s-master'.
2016-04-14 16:36:30.600 | INFO     | context.Background: Succeeded to start group 'k8s-master'.
```

A quick look at `kubectl version` will tell us that we're now running a 1.2.2 server.

## That's all, folks!

We have managed to start up a Kubernetes cluster, start a replicated pod on it, and perform a rolling update of the Kubernetes nodes without downtimes of both Kubernetes and the running pod, all thanks to Inago.
