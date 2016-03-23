# Terminology

The following explains what we mean by and how we use certain terms in this documentation

## Group
A group is a collection of unit files tied together. This can be seen as
equivalent to kubernetes pods. Units within an Inago group are scheduled
together on the same host by convention.

Since a group represents many units, Inago aims to abstract units away, so
the user can manage a group more easily. Having this given, a group can be
treated like a single unit. All actions one can apply to a unit using
`fleetctl`, can also be applied towards a group using `inagoctl`.

There are two types of groups: slice groups and instance groups. Unit files,
which contain an `@` sign are considered sliceable. If the unit files does **not**
contain an `@` sign the unit is recognized an instance unit.

Examples:

### Sliceable group

```nohighlight
└── mygroup
    ├── mygroup-bar@.service
    └── mygroup-foo@.service
```

### Instance group

```nohighlight
└── timer
    ├── timer-example.service
    └── timer-example.timer
```

For more examples take a look at the example in the
[example folder](https://github.com/giantswarm/inago/tree/master/example).

A group is named after the directory its unit files live in. Having two units
`mygroup-foo@.service` and `mygroup-bar@.timer` in the directory `mygroup`,
leads to the Inago group `mygroup`. So creating a directory and putting units
into it, creates a new group as Ingao is able to understand it.

## Slice
A slice is a scalable group instance. The kubernetes equivalent would be a
replica. Slices are a scaled version of a group. So you can scale up your group
using slice expansion. See also [What is slice expansion?](slice_expansion.md).

## Instance
A group instance is a group, which can **not** be scaled. Unit files which
don't contain an `@` sign are considered unscalable.
Example: `prefix-unit-name.service`
