# structure
Inago requires a certain directory structure and unit file names to make the
tool actually work. There must be a group folder. See also the description of
[the term group](terms.md#Group).

There must be files within the this group folder. The files must start with the
folder name. That means the unit file names need to be prefixed with the group
name. See also documentation about [slice expansion](slice_expansion.md).

The file MUST be a systemd unitfile. Optionally there can be fleet statements
defined. See
https://www.freedesktop.org/software/systemd/man/systemd.html#Concepts and
https://coreos.com/fleet/docs/latest/unit-files-and-scheduling.html.

If these requirements are not given, Formica will not work properly
