# Submit
Submitting groups enables deploying multiple slices using one command.

```
inagoctl submit myapp 3
```

This assumes there is a directory `myapp`. The argument `3` tells Inago to
create 3 slices of the units found in the group directoy. Inago looks up the
directory `myapp`. So lets imagine there are this two unit files within
`myapp`.

```
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
