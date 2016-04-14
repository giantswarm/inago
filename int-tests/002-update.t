Create a group to play around with.
  $ GROUP=002-update-group
  $ mkdir $GROUP
  $ printf "[Unit]\nDescription=$GROUP Inago Update Test Unit\n\n[Service]\nExecStart=/bin/bash -c \"while true; do echo Hi; sleep 10; done\"\n" > $GROUP/$GROUP-foo@.service
  $ printf "[Unit]\nDescription=$GROUP Inago Update Test Unit\n\n[Service]\nExecStart=/bin/bash -c \"while true; do echo Hi; sleep 10; done\"\n" > $GROUP/$GROUP-bar@.service

Validate the test group.
  $ inagoctl validate $GROUP
  Group '002-update-group' is valid.
  Groups are valid globally.

Submit 2 slices of update group.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} submit $GROUP 2
  .*\|\scontext.Background: Succeeded to submit group '002-update-group'. (re)
  $ sleep 5

Start update group.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} start $GROUP
  .*\|\scontext.Background: Succeeded to start 2 slices for group '002-update-group': \[([a-z0-9]{3}\s?){2}\]. (re)
  $ sleep 5

Test the status of the update group, after starting
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP
  Group\s*Units\s*FDState\s*FCState\s*SAState\s*IP\s*Machine (re)
  
  002-update-group@[a-z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  002-update-group@[a-z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  
Update update-group without changing the unit file first.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} update --max-growth 1 --min-alive 1 $GROUP
  .*\|\scontext.Background: Not updating group '002-update-group'. \(units already up to date\) (re)

Changing content of update-group-bar unit file.
  $ echo "[Unit]\nDescription=Inago Update Test Unit CHANGED\n\n[Service]\nExecStart=/bin/bash -c \"while true; do echo Hi; sleep 10; done\"" > $GROUP/$GROUP-bar@.service

Update update-group.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} update --max-growth 2 --min-alive 1 $GROUP
  .*\|\scontroller: adding units \[([a-z0-9]{3}\s?){2}\]. (re)
  .*\|\scontroller: adding units \[([a-z0-9]{3}\s?){2}\]. (re)
  .*\|\scontroller: removing units \[([a-z0-9]{3}\s?){2}\]. (re)
  .*\|\scontroller: removing units \[([a-z0-9]{3}\s?){2}\]. (re)
  .*\|\scontext.Background: Succeeded to update 2 slices for group '002-update-group': \[([a-z0-9]{3}\s?){2}\]. (re)

Test the status of the updated group.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP
  Group\s*Units\s*FDState\s*FCState\s*SAState\s*IP\s*Machine (re)
  
  002-update-group@[a-z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  002-update-group@[a-z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  

Kill it:

  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} destroy $GROUP > /dev/null 2>&1
