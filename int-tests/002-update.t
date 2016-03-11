Create a group to play around with.
  $ mkdir update-group
  $ touch update-group/test-group-foo@.service
  $ echo "[Unit]\nDescription=Inago Test Unit\n\n[Service]\nExecStart=/bin/bash -c \"while true; do echo Hi; sleep 10; done\"" > update-group/test-group-foo@.service
  $ touch update-group/test-group-bar@.service
  $ echo "[Unit]\nDescription=Inago Test Unit\n\n[Service]\nExecStart=/bin/bash -c \"while true; do echo Hi; sleep 10; done\"" > update-group/test-group-bar@.service

Validate the test group.
  $ inagoctl validate update-group
  Group 'update-group' is valid.
  Groups are valid globally.


Submit 2 slices of update group.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} submit update-group
  Succeeded to submit all slices of group 'update-group': \[[a-zA-Z\d]{3}\s[a-zA-Z\d]{3}\]. (re)
  $ sleep 5
  
Test the status of the test group, after submission.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status update-group
  Group\s*Units\s*FDState\s*FCState\s*SAState\s*IP\s*Machine (re)
  
  update-group@[a-zA-Z\d]{3}\s*\*\s*loaded\s*loaded\s*inactive\s*[0-9.]*\s*[a-z0-9]* (re)
  update-group@[a-zA-Z\d]{3}\s*\*\s*loaded\s*loaded\s*inactive\s*[0-9.]*\s*[a-z0-9]* (re)

Start update group.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} start update-group
  Succeeded to start 2 slices for group 'update-group': \[[a-zA-Z\d]{3}\s[a-zA-Z\d]{3}\]. (re)
  $ sleep 5

Test the status of the update group, after starting.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status update-group
  Group\s*Units\s*FDState\s*FCState\s*SAState\s*IP\s*Machine (re)
  
  update-group@[a-zA-Z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  update-group@[a-zA-Z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  
Update update-group without chaning the unit file first.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} update --max-growth 0 --min-alive 3 update-group
  Failed to update 2 slices for group 'update-group': \[[a-zA-Z\d]{3}\s[a-zA-Z\d]{3}\]. (update not allowed: units already up to date) (re)

Changing content of test-group-bar unit file.
  $ echo "[Unit]\nDescription=Inago Test Unit CHANGED\n\n[Service]\nExecStart=/bin/bash -c \"while true; do echo Hi; sleep 10; done\"" > update-group/test-group-bar@.service
  
Update update-group.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} update --max-growth 2 --min-alive 1 update-group
  Succeeded to update 2 slices for group 'mygroup': \[[a-zA-Z\d]{3}\s[a-zA-Z\d]{3}\].

Test the status of the update group, after stopping.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status update-group
  Group\s*Units\s*FDState\s*FCState\s*SAState\s*IP\s*Machine (re)
  
  update-group@[a-zA-Z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  update-group@[a-zA-Z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
