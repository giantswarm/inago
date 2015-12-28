# Workflow Examples

- Detail a workflow for the deployment tool

## Commands
### endpoint
The deployment tool communicates with a cluster fleet endpoint.
This command sets which endpoint to communicate with. 
Only one endpoint can be used at a time. 
```
$ formica endpoint http://127.0.0.1:4001
$ formica endpoint
http://127.0.0.1:4001
```

### ping
For sanity's sake, let's add an easy way to test if we can actually reach the cluster (exit codes for automation).
```
$ formica ping
Cluster is reachable
$ echo $?
0
$ formica ping --endpoint http://127.0.0.1:4002
Cluster is not reachable
$ echo $?
1
```

### apply
Configuration is available locally, and then applied to the cluster.
The tool determines the delta between the state detailed locally and the cluster state, and then executes the changes required to set the cluster state to be equal to the local state.
```
$ ls
[CONFIGURATION DETAILING HELLO_WORLD SERVICE]
$ formica apply
Found service definition "hello_world"
Service "hello_world" not found in cluster
Submitting service definition "hello_world" to cluster
Starting service "hello_world"
$ fleetctl --endpoint http://127.0.0.1:4001 list-units
UNIT                  HASH    DSTATE   STATE    TMACHINE
hello-world.service   e55c0ae launched launched 113f16a7.../127.0.0.1
```

### fetch
The cluster should be the source of truth concerning state - having the state stored externally can lead to drift between the cluster state and the external state. It is still possible to commit the configuration to source control.
To this end, we should be able to fetch the configuration from the cluster itself.
```
$ fleetctl --endpoint http://127.0.0.1:4001 list-units
UNIT                  HASH    DSTATE   STATE    TMACHINE
hello-world.service   e55c0ae launched launched 113f16a7.../127.0.0.1
$ ls
$ formica fetch
Service "hello-world" found in cluster
Fetching service definition "hello-world" from cluster
$ ls
[CONFIGURATION DETAILING HELLO_WORLD SERVICE]
$ formica fetch
No changes detected between cluster and local state
```

## Workflows

### Without source control
- Fetch latest configuration from cluster
- Make changes to configuration
- Apply changes to cluster

### With source control
- Checkout latest configuration from source control
- Test no change between cluster and source control states
- Make changes to configuration
- Apply changes to cluster
- Commit changes to source control

### Continuous deployment
On changes to services in the stack:
- Determine latest configuration (with external tool)
- Apply configuration to test cluster