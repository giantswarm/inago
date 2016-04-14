# Tunneling

Inago provides ssh tunneling to remote hosts. That way we are able to operate
on units running on remote machines from our local environment. Simply add the
`--tunnel` flag like with `fleetctl`.

Fetching the status of `mygroup` running at `my.remote.host` could work like
that.
```
inagoctl --tunnel=my.remote.host status mygroup
```

Updating `mygroup` running at `my.remote.host` could work like that. Note that
the updated unit files need to be located where `inagoctl update` is executed.
```
inagoctl --tunnel=my.remote.host update mygroup
```
