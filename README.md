# Inago

[![Build Status](https://api.travis-ci.org/giantswarm/inago.svg)](https://travis-ci.org/giantswarm/inago) [![codecov.io](https://codecov.io/github/giantswarm/inago/coverage.svg?branch=master)](https://codecov.io/github/giantswarm/inago?branch=master)

Inago is a deployment tool that manages groups of unit files to deploy them to
a fleet cluster similar to `fleetctl`. Since `fleetctl` is quite limited, Inago
aims to abstract units away and provide more sugar on top like update
strategies. That way the user can manage unit files more easily.

## Getting Inago

Clone the git repository: `git@github.com:giantswarm/inago.git`

## Running Inago

Simply run the binary like `fleetctl`. See help usage for more information.

```
inagoctl -h
```

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
