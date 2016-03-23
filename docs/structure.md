# Unit File Structure

Inago requires a certain directory structure and unit file names to make the
tool actually work. There must be a group folder. See also the description of
[the term group](terminology.md#Group).

There must be files within the this group folder. The files must start with the
folder name. That means the unit file names need to be prefixed with the group
name.
Furthermore, sliceable and instance units may not be mixed. See also documentation
about [slice](terminology.md#Slice) and [instance](terminology.md#Instance).

The files MUST be systemd unit files. Optionally there can be fleet statements
defined. See
https://www.freedesktop.org/software/systemd/man/systemd.html#Concepts and
https://coreos.com/fleet/docs/latest/unit-files-and-scheduling.html.

If these requirements are not given, Inago will not work properly.
