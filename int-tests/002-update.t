Create a group to play around with.
  $ mkdir update-group
  $ echo "[Unit]\nDescription=Inago Update Test Unit\n\n[Service]\nExecStart=/bin/bash -c \"while true; do echo Hi; sleep 10; done\"" > update-group/update-group-foo@.service
  $ echo "[Unit]\nDescription=Inago Update Test Unit\n\n[Service]\nExecStart=/bin/bash -c \"while true; do echo Hi; sleep 10; done\"" > update-group/update-group-bar@.service

Validate the test group.
  $ inagoctl validate update-group
  Group 'update-group' is valid.
  Groups are valid globally.

Submit 2 slices of update group.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} submit update-group 2
  Succeeded to submit 2 slices for group 'update-group': \[([a-z0-9]{3}\s?){2}\]. (re)
  $ sleep 5

Start update group.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} start update-group
  Succeeded to start 2 slices for group 'update-group': \[([a-z0-9]{3}\s?){2}\]. (re)
  $ sleep 5

Test the status of the update group, after starting.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status update-group
  Group\s*Units\s*FDState\s*FCState\s*SAState\s*IP\s*Machine (re)
  
  update-group@[a-z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  update-group@[a-z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  
Update update-group without changing the unit file first.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} update --max-growth 0 --min-alive 3 update-group
  Failed to update 2 slices for group 'update-group': \[([a-z0-9]{3}\s?){2}\]. \(update not allowed: units already up to date\) (re)
  [1]

Changing content of update-group-bar unit file.
  $ echo "[Unit]\nDescription=Inago Update Test Unit CHANGED\n\n[Service]\nExecStart=/bin/bash -c \"while true; do echo Hi; sleep 10; done\"" > update-group/update-group-bar@.service
Update update-group.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} update --max-growth 2 --min-alive 1 update-group
  Succeeded to update 2 slices for group 'update-group': \[([a-z0-9]{3}\s?){2}\]. (re)

Test the status of the updated group.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status update-group
  Group\s*Units\s*FDState\s*FCState\s*SAState\s*IP\s*Machine (re)
  
  update-group@[a-z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  update-group@[a-z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  
