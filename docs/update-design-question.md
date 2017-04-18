Guys, we have a question for Formica where everyone is invited to participate deciding.

So, formica currently is a tool for managing groups of units. Example:

```
$ formicactl submit elasticsearch@{1,2}
$ formciactl start elasticsearch@{1,2}
$ formicactl status elasticsearch
Group           Unit  <more headers>
elasticsearch@1 *     active  active  running 10.0.4.100
elasticsearch@2 *     active  active  running 10.0.4.101
```

As you see, it is very similar to `fleetctl`. Next we want to implement an update
functionality, similar (but better ;) than what `releaseit` currently does.

## Question 1
In the example above we have used `1` and `2` for the slices. When updating, we
want the users to have the option to create new instances first. This raises where
the names for those slices come from.

* Option 1: Just increment the found ids. So if we find `7,8,9` we continue with
  `10, 11, ..`
* Option 2: Use random IDs. This would imply changes to the idea that user dictates
  these names as shown above.
* Option 3: Extend the slice with a timestamp so we can reuse the old slice but still differentiate
  them from existing slices.

## Question 2

The second question is how the update command it self looks like.

1. `$ formicactl update --sleep-time=1m --max-unavailable=50% --max-surge=100% elasticsearch`
   This would try to replace all slices that are currently deployed with the new version,
   while never falling below 50% down (of the initial number) and being allowed
   to start 100% new instances to achieve that. If `max-surge` would be `0`, it
   would stop slices first.

   This is inspired / similar to how Kubernetes / ECS / Nomad allows their updates.

2. `formicactl update --confirm --sleep=1mins --strategy=add-first 50% elasticsearch`
   A more 1:1 copy of the current releaseit logic. It picks 50% of the existing
   slices and replaces them by adding new slices first, then destroying the old
   ones. Then it sleeps for some time (LB updates) and asks for confirmation for
   the next slices.
