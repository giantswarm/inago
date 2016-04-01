# Inago
[![GoDoc](https://godoc.org/github.com/giantswarm/inago?status.svg)](http://godoc.org/github.com/giantswarm/inago)
[![Build Status](https://api.travis-ci.org/giantswarm/inago.svg)](https://travis-ci.org/giantswarm/inago)
[![Go Report Card](http://goreportcard.com/badge/giantswarm/inago)](http://goreportcard.com/report/giantswarm/inago)
[![codecov.io](https://codecov.io/github/giantswarm/inago/coverage.svg?branch=master)](https://codecov.io/github/giantswarm/inago?branch=master)

Inago is a deployment tool that manages groups of unit files to deploy them to a fleet cluster similar to `fleetctl`. Inago aims to abstract units away so you can handle groups containing large numbers of unit files. Additionally, it provides more sugar on top like rolling updates with different strategies.

_Note:_ Currently only `.service` files are fully supported. Other unit types like `.timer` are not supported in [RC1](https://github.com/giantswarm/inago/releases).

## Getting Inago
#### Brew
Install inagoctl via brew.
```
$ brew tap giantswarm/giantswarm
$ brew install inagoctl
```

#### Binaries
Download binaries for Linux and Mac OSX: https://github.com/giantswarm/inago/releases

#### Source
Clone the git repository: `git@github.com:giantswarm/inago.git`

## Running Inago

Simply run the binary like `fleetctl`. See help usage for more information.

```
inagoctl -h
```

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
