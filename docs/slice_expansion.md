# slice expansion
Slice expansion is a feature to provide multiple slices using one command line
argument in your shell. In fact the argument expansion is a feature of common
shells. So you can do things like this.

```
formica submit myapp@{1..3}
```

This assumes there is a directory `myapp`. The expression `@{1..3}` tells
formica to create 3 slices. Your shell expands the argument `myapp@{1..3}` to
the following arguments that formica receives.
```
myapp@1
myapp@2
myapp@3
```

Formica takes these arguments and looks up the directory `myapp`. Then the
slice IDs `1`, `2` and `3` are used to create the unit file names accordingly.
So lets imagine there are this two unit files within `myapp`.
```
some_unit_name@.service
some_other_unit_name@.service
```

Given the 3 slice IDs and the unit file names the resulting unit file names
that are going to be created for our `submit` command are the following.
```
some_unit_name@1.service
some_other_unit_name@1.service
some_unit_name@2.service
some_other_unit_name@2.service
some_unit_name@3.service
some_other_unit_name@3.service
```
