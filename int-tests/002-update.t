Create a group to play around with.
  $ GROUP=update-group
  $ mkdir $GROUP 
  $ printf "[Unit]\nDescription=$GROUP Inago Update Test Unit\n\n[Service]\nExecStart=/bin/bash -c \"while true; do echo Hi; sleep 10; done\"\n" > $GROUP/$GROUP-foo@.service
  $ printf "[Unit]\nDescription=$GROUP Inago Update Test Unit\n\n[Service]\nExecStart=/bin/bash -c \"while true; do echo Hi; sleep 10; done\"\n" > $GROUP/$GROUP-bar@.service

Cleanup cluster first

  $ fleetctl list-unit-files --fields=unit --no-legend | grep ^$GROUP | xargs fleetctl destroy
Validate the test group.
  $ inagoctl validate $GROUP 
  Group 'update-group' is valid.
  Groups are valid globally.

Submit 2 slices of update group.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} submit $GROUP 2
  .*\|\scontext.Background: Succeeded to submit group 'update-group'. (re)
  $ sleep 5

Start update group.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} start $GROUP
  .*\|\scontext.Background: Succeeded to start 2 slices for group 'update-group': \[([a-z0-9]{3}\s?){2}\]. (re)
  $ sleep 5

Test the status of the update group, after starting
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP 
  Group\s*Units\s*FDState\s*FCState\s*SAState\s*IP\s*Machine (re)
  
  update-group@[a-z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  update-group@[a-z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  
Update update-group without changing the unit file first.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} update --max-growth 0 --min-alive 3 $GROUP 
  .*\|\scontext.Background: Failed to update 2 slices for group 'update-group': \[([a-z0-9]{3}\s?){2}\]. \(update not allowed: units already up to date\) (re)
  [1]

Changing content of update-group-bar unit file.
  $ echo "[Unit]\nDescription=Inago Update Test Unit CHANGED\n\n[Service]\nExecStart=/bin/bash -c \"while true; do echo Hi; sleep 10; done\"" > update-group/update-group-bar@.service
Update update-group.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} update --max-growth 2 --min-alive 1 $GROUP 
  .*\|\scontext.Background: Succeeded to update 2 slices for group 'update-group': \[([a-z0-9]{3}\s?){2}\]. (re)

Test the status of the updated group.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP 
  Group\s*Units\s*FDState\s*FCState\s*SAState\s*IP\s*Machine (re)
  
  update-group@[a-z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  update-group@[a-z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  
