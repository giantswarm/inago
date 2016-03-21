# kubernetes

With Inago you can easily deploy kubernetes on a coreos cluster.
The `example` folder contains a `k8s` example. Kubernetes is split into
two groups: `k8s-master` and `k8s-node`.

## note
The deployment sets up a kubernetes cluster, which uses insecure communication
between the components. Use this only as a Demo or POC. If you want to deploy a
kubernetes cluster for production please modify the example, so a secure
communication is used. [Cluster TLS using OpenSSL](https://coreos.com/kubernetes/docs/latest/openssl.html)

## deployment
If your coreos cluster is up and running you can deploy kubernetes in 4 steps:

1. submit the k8s-master by running: `inagoctl submit k8s-master`
2. submit the k8s-node by running: `inagoctl submit k8s-node`
3. start the k8s-master by running: `inagoctl start k8s-master`
4. start the k8s-node by running: `inagoctl start k8s-node`

And you're done!

The k8s-node contains the `Global=true` flag and will be scheduled on all machines.

## start your pod
_Make sure you installed `kubectl`. If you are on osx run `brew install kubectl`.
Otherwise get it here: [Download](https://coreos.com/kubernetes/docs/latest/configure-kubectl.html)_

First we need the IP of the machine running the api-server. Login to a coreos machine
and run

```
$ fleetctl list-units -fields=unit,machine --full --no-legend 2>/dev/null | grep ^k8s-master-apiserver.service | cut -d/ -f2 | paste -d, -s
172.17.8.102
```

Now we can start a nginx example by running:
`kubectl -s 172.17.8.102:9090 run my-nginx --image=nginx --replicas=2 --port=80`

To validate that the pods got indeed started, we can take a look at the pods.
```
$ kubectl -s 172.17.8.102:9090 get pods
NAME             READY     STATUS    RESTARTS   AGE
my-nginx-0uhad   0/1       Pending   0          13s
my-nginx-zuztz   0/1       Pending   0          13s
nginx            1/1       Running   1          1h
```
