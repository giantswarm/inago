# workflow examples

## hello world

check we can reach cluster
```
$ formica ping --endpoint http://127.0.0.1:4001
OK!
$ formica ping --endpoint http://127.0.0.1:4002
Cluster is not reachable
```

starting a _very_ basic service
```
$ ls
hello_world.service
$ cat hello_world.service
service "hello_world" {
    description = "a basic hello world example"
    start_pre = [
        "docker pull hello-world"
    ]
    start = [
        "docker run hello-world"
    ]
}
$ formica apply --endpoint http://127.0.0.1:4001
Found service definition "hello_world"
Service "hello_world" not found in cluster
Submitting service definition "hello_world" to cluster
Starting service "hello_world"
OK!
$ fleetctl --endpoint http://127.0.0.1:4001 list-units
UNIT                  HASH    DSTATE   STATE    TMACHINE
hello-world.service   e55c0ae launched launched 113f16a7.../127.0.0.1
$ fleetctl --endpoint http://127.0.0.1:4001 cat hello-world.service
[Unit]
Description=a basic hello world example

[Service]
ExecStartPre=docker pull hello-world
ExecStart=docker run hello-world
```
