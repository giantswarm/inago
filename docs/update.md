# Updating Groups

Inago includes an `update` command, which allows updating
a group on the cluster.

__Note__: The `update` command currently only works on sliced groups.

## Usage

Updating a group includes stopping old slices and starting new ones.
The behavior of the update process is determined by two flags: `--min-alive` and `--max-growth`.

### min-alive
The `--min-alive` flag defines the minimum number of slices
that must be running at any time during the update process.

### max-growth
The `--max-growth` flag sets the upper limit on how many additional 
slices may be started during the update process.

### Update Strategies

Using the above mentioned flags you can enforce various update strategies. We will show this using the `myapp` example from [Getting Started](getting_started.md) using `n=3` slices.

A one-by-one update strategy would use a `--min-alive` of `n-1` and a `--max-growth` of `0`:

`inagoctl update --min-alive=2 --max-growth=0`

A all-at-once update strategy would use a `--min-alive` of `n` and a `--max-growth` of `n`:

`inagoctl update --min-alive=3 --max-growth=3`

A hot-swap update strategy would use a `--min-alive` of `n` and a `--max-growth` of `1`:

`inagoctl update --min-alive=3 --max-growth=1`

There's many more variations that you could think of depending on your specific needs in terms of downtimes and workloads.