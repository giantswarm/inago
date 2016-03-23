# Updating groups

Inago includes an `update` command, which allows updating
a group on the cluster. Updating a group includes stopping
old slices and starting new ones. The behavior of the update process
is determined by two flags: `--min-alive` and `--max-growth`.

### min-alive
The `--min-alive` flags defines the minimum number of slices,
that must be running at any time during the update process.

### max-growth
The `--max-growth` flag sets the upper limit on how many slices
may be started during the update process.

## Update strategy
