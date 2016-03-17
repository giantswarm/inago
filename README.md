# Inago

[![Build Status](https://api.travis-ci.org/giantswarm/inago.svg)](https://travis-ci.org/giantswarm/inago) [![codecov.io](https://codecov.io/github/giantswarm/inago/coverage.svg?branch=master)](https://codecov.io/github/giantswarm/inago?branch=master)

Inago is a deployment tool that manages groups of unit files to deploy them to
a fleet cluster similar to `fleetctl`. Since `fleetctl` is quite limited, Inago
aims to abstract units away and provide more sugar on top like update
strategies. That way the user can manage unit files more easily.

## Getting Inago

Download binaries: https://github.com/giantswarm/inago/releases

Clone the git repository: `git@github.com:giantswarm/inago.git`

## Running Inago

Simply run the binary like `fleetctl`. See help usage for more information.

```
inagoctl -h
```

## Running integration tests

Runnning `make int-test` will execute the integration test suite in a docker container.
The tests need to run against fleet, so we manage a vagrant box. Starting and destroying
this test machine is done via the make target `int-test`.

### Integration Test Machine Configuration

Since the integration tests run in your docker machine (not the fleet machine), we need to
provide it with an IP. We use port fowarding on your host for this.
Set the `FLEET_ENDPOINT` environment variable to `http://ip:491563`, with the IP listed
under the vboxnet interface of your docker-machine.

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

On Stephan's machine the docker machine is in the `vboxnet6` network and the coreos VMs
are in `vboxnet5`, so you would execute `FLEET_ENDPOINT=http://172.17.8.1:49153 make int-test`

## Releasing

We're using Giant Swarm's [builder](https://github.com/giantswarm/builder) for releases.
You will need to have [GitHub releases support](https://github.com/giantswarm/builder#github-releases) set up.

Releasing is done via:
```
builder release major|minor|patch
```
This command will perform the necessary steps for release, including uploading a tarball with built binaries to GitHub.

## Further Steps

Check more detailed documentation: [docs](docs)

Check code documentation: [godoc](https://godoc.org/github.com/giantswarm/inago)

## Contact

- Mailing list: [giantswarm](https://groups.google.com/forum/#!forum/giantswarm)
- IRC: #[giantswarm](irc://irc.freenode.org:6667/#giantswarm) on freenode.org
- Bugs: [issues](https://github.com/giantswarm/inago/issues)

## Contributing & Reporting Bugs

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches, the
contribution workflow as well as reporting bugs.

## License

Inago is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.

## Origin of the Name

`inago` (いなご [稲子] pronounced "inago") is Japanese for grasshopper.
