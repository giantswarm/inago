## Running Integration Tests (locally)

Runnning `make int-test` will execute the integration test suite in a docker container.
The tests need to run against fleet, so we manage a vagrant box. Starting and destroying
this test machine is done via the make target `int-test`.

### Integration Test Machine Configuration

Since the integration tests run in your docker machine (not the fleet machine), we need to
provide it with an IP. We use port forwarding on your host for this.
Set the `FLEET_ENDPOINT` environment variable to `http://ip:491563`, with the IP listed
under the vboxnet interface of your docker-machine.
The integration test suite also tests the `--tunnel` flag. Set the
`INAGO_TUNNEL_ENDPOINT` to `ip:2202`, with the same ip as the `FLEET_ENDPOINT`.

```
$ ifconfig
[..snip..]
vboxnet0: flags=8842<BROADCAST,RUNNING,SIMPLEX,MULTICAST> mtu 1500
    ether 0a:00:27:00:00:00
vboxnet1: flags=8842<BROADCAST,RUNNING,SIMPLEX,MULTICAST> mtu 1500
    ether 0a:00:27:00:00:01
vboxnet2: flags=8842<BROADCAST,RUNNING,SIMPLEX,MULTICAST> mtu 1500
    ether 0a:00:27:00:00:02
vboxnet3: flags=8842<BROADCAST,RUNNING,SIMPLEX,MULTICAST> mtu 1500
    ether 0a:00:27:00:00:03
vboxnet4: flags=8842<BROADCAST,RUNNING,SIMPLEX,MULTICAST> mtu 1500
    ether 0a:00:27:00:00:04
vboxnet5: flags=8843<UP,BROADCAST,RUNNING,SIMPLEX,MULTICAST> mtu 1500
    ether 0a:00:27:00:00:05
    inet 172.17.8.1 netmask 0xffffff00 broadcast 172.17.8.255
vboxnet6: flags=8843<UP,BROADCAST,RUNNING,SIMPLEX,MULTICAST> mtu 1500
    ether 0a:00:27:00:00:06
    inet 192.168.59.3 netmask 0xffffff00 broadcast 192.168.59.255
```

On the above machine the docker machine is in the `vboxnet6` network and the coreos VMs
are in `vboxnet5`, so you would execute `FLEET_ENDPOINT=http://172.17.8.1:49153 make int-test`

## Integration Test Server Setup

For running integration tests online you need setup an integration test server.

On AWS use a `t2.micro` running in `eu-central-1`, with `ami-15190379` (CoreOS 835.13.0).

The server should have both `etcd` and `fleet` running - use:

```
sudo systemctl start fleet
sudo systemctl start etcd2
```

to start both services.
