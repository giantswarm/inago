# Getting Started

## What is the goal of Inago?

Inago is a deployment tool that manages groups of unit files to deploy them to
a fleet cluster similar to `fleetctl`. Inago
aims to abstract units away so you can handle groups containing large numbers
of unit files.
Additionally, it provides more sugar on top like rolling updates with different
strategies.

## Prerequisites

Inago requires a certain directory structure and unit file names to make the
tool work. There must be a group folder with unit files in it. For details see the [Unit File Structure](structure.md) chapter.

## Working with Inago

The basic commands you can use with Inago are `submit`, `start`, `stop`, `destroy`, and `status`. The subject of your commands is always a unit group as defined by a group folder.

### Submit

Submitting groups enables deploying multiple slices using one command.

```nohighlight
inagoctl submit myapp 3
```

This assumes there is a directory `myapp` and that the unit files in that directory cary an `@` and thus are to be sliced. The argument `3` tells Inago to create 3 slices of the units found in the group directoy. Inago looks up the
directory `myapp`. So lets imagine there are two unit files within
`myapp`.

```nohighlight
myapp_some_unit_name@.service
myapp_some_other_unit_name@.service
```

Given the scaling argument `3` and the unit file names the resulting unit file
names that are going to be created for our `submit` command should look
something like the following.

```shell
# first slice
myapp_some_unit_name@s8k.service
myapp_some_other_unit_name@s8k.service

# second slice
myapp_some_unit_name@0ds.service
myapp_some_other_unit_name@0ds.service

# third slice
myapp_some_unit_name@h38.service
myapp_some_other_unit_name@h38.service
```

### Start, Stop, Destroy

Once you have submitted a group like explained above, you can then use Inago to start, stop, or destroy that group with a single command each.

```nohighlight
inagoctl start myapp

inagoctl stop myapp

inagoctl destroy myapp
```

### Status

Using the `status` command you can view the current status of your group and compare desired and actual states of each slice. By default the substates of the units of each group slice are aggregated as long as they are consistent across the slice.

```shell
$ inagoctl status myapp
Slice     Unit      DState    State     IP            Active
myapp@s8k    *     active    active    10.0.0.100    running
myapp@0ds    *     active    active    10.0.0.101    running
myapp@h38    *     inactive  inactive  10.0.0.102    stopped
```

In case of inconsistency `status` will give you a detailed view of the inconsistent slice:

```shell
$ inagoctl status myapp
Slice     Unit                          DState    State     IP            Active
myapp@s8k    *                             active    active    10.0.0.100    running
myapp@0ds    myapp_some_unit_name@0ds.service          active    failed    10.0.0.101    failed
myapp@0ds    myapp_some_other_unit_name@0ds.service          active    inactive  10.0.0.101    dead
myapp@h38    *                             active    active    10.0.0.102    running
```

You can also use the `-v` flag to always show details of each unit as well as a hash for each unit deployed, so that you can check if all units are running the same version.