# Terminology

The following explains what we mean by and how we use certain terms in this documentation.

### Group
A group is a collection of unit files tied together. This can be seen as
equivalent to kubernetes pods. Units within an Inago group are scheduled
together on the same host by convention.

Since a group represents many units, Inago aims to abstract units away, so
the user can manage a group more easily. Having this given, a group can be
treated like a single unit. All actions one can apply to a unit using
`fleetctl`, can also be applied towards a group using `inagoctl`.

A group is named after the directory its unit files live in. Having two units
`mygroup-foo@.service` and `mygroup-bar@.timer` in the directory `mygroup`,
leads to the Inago group `mygroup`. So creating a directory and putting units
into it, creates a new group as Ingao is able to understand it.

### Slice
A slice is a scalable group instance. The kubernetes equivalent would be a
replica. Slices are a scaled version of a group. So you can scale up your group
using slice expansion. See also [What is slice expansion?](slice_expansion.md).

### Instance
A group instance is a group in a fleet cluster. Having a group definition on
your local file system and deploying this using Inago to a fleet cluster,
creates a group instance. This group instance may be one of multiple slices.
