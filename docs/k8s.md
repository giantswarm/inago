# Kubernetes

With Inago you can easily deploy Kubernetes on a CoreOS cluster.
The [example folder](https://github.com/giantswarm/inago/tree/master/example) contains a simple [k8s example](https://github.com/giantswarm/inago/tree/master/example/k8s). The Kubernetes deployment is split into
two groups: `k8s-master` and `k8s-node`.

## Note
The deployment sets up a Kubernetes cluster, which uses insecure communication
between the components. Use this only as a Demo or POC. If you want to deploy a
Kubernetes cluster for production please modify the example, so secure
communication is used: [Cluster TLS using OpenSSL](https://coreos.com/Kubernetes/docs/latest/openssl.html).
Further, this deployment uses a single Kubernetes master, so the API is not highly available. For HA deployments of Kubernetes check out: [Building High-Availability Clusters](http://kubernetes.io/docs/admin/high-availability/).

## Deployment

You can use this deployment on any CoreOS fleet cluster. Below we work with a 3 node cluster. For a simple way to setup a CoreOS cluster on your laptop try the official [Vagrant setup](https://coreos.com/os/docs/latest/booting-on-vagrant.html).
Once your CoreOS cluster is up and running you can deploy Kubernetes in two easy steps:

First, we need to setup the network environment. This, needs to be done only once, but on all nodes, which is why we set this up to be a global unit (running on all machines).

```
$ inagoctl up k8s-network
2016-04-14 15:31:05.273 | INFO     | context.Background: Succeeded to submit group 'k8s-master'.
2016-04-14 15:32:01.532 | INFO     | context.Background: Succeeded to start group 'k8s-network'.
```

Then, we can schedule the k8s-master group.

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

And you're done! You can check if your cluster is running with [`kubectl`](https://coreos.com/kubernetes/docs/latest/configure-kubectl.html):

```
$ kubectl cluster-info
Kubernetes master is running at http://localhost:8080

$ kubectl version
Client Version: version.Info{Major:"1", Minor:"2", GitVersion:"v1.2.0", GitCommit:"5cb86ee022267586db386f62781338b0483733b3", GitTreeState:"clean"}
Server Version: version.Info{Major:"1", Minor:"2", GitVersion:"v1.2.0", GitCommit:"5cb86ee022267586db386f62781338b0483733b3", GitTreeState:"clean"}
```

Note: We assume running `kubectl` on the same node as the API server here. If you're running on a different node, you can use the `-s` flag to connect to the right one (for more details the the last chapter of this page).

## Testing your Kubernetes cluster



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
$ inagoctl update k8s-node --max-growth=0 --min-alive=1 --ready-secs=60
2016-04-14 15:45:22.988 | INFO     | context.Background: Succeeded to update 3 slices for group 'k8s-node': [045 3a9 79b].
```

And we're done! Another `kubectl describe nodes` will show that we're now running version 1.2.2 on our nodes.

Looking closely at the update command above we can inspect some of its options. `--max-growth=0` tells Inago to not start any additional instances, which means it has to remove an instance to be able to start a new one. In this example this is due to the fact that we don't have any free machine left. `--min-alive=1` tells Inago to keep at least one instance of the group alive during the update process, we could also use `2` here, to have more availability during the update. Last but not least `--ready-secs=60` tells Inago to wait 60 seconds between updates. This gives the new instances some time to start replicating or do some other bootstrapping activities. Note that this might need to be higher, once you're actually running pods on these nodes, as the pods might need to get rescheduled.

### Updating the Master

Now sadly, we didn't use an HA deployment of the Kubernetes master, so Inago's update command isn't of help much here. Accordingly, we will have a short downtime of our API while we update this group. Luckily, Kubernetes nodes and their pods don't rely on the API server that much and will usually not be impacted by this downtime.

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

## Start your first pod

First, we need the IP of the machine running the API server. Login to a CoreOS machine
and run

```
$ fleetctl list-units -fields=unit,machine --full --no-legend 2>/dev/null | grep ^k8s-master-api-server.service | cut -d/ -f2 | paste -d, -s
172.17.8.102
```

Now we can start an nginx example by running:
`kubectl -s 172.17.8.102:8080 run my-nginx --image=nginx --replicas=2 --port=80`

To validate that the pods indeed got started, we can take a look at the pods.

```
$ kubectl -s 172.17.8.102:8080 get pods
NAME             READY     STATUS    RESTARTS   AGE
my-nginx-0uhad   0/1       Pending   0          13s
my-nginx-zuztz   0/1       Pending   0          13s
nginx            1/1       Running   1          1h
```
