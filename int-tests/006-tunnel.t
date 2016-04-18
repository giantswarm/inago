Create a group to play around with.
  $ GROUP=002-tunnel-group
  $ mkdir $GROUP
  $ printf "[Unit]\nDescription=$GROUP Inago Update Test Unit\n\n[Service]\nExecStart=/bin/bash -c \"while true; do echo Hi; sleep 10; done\"\n" > $GROUP/$GROUP-foo@.service
  $ printf "[Unit]\nDescription=$GROUP Inago Update Test Unit\n\n[Service]\nExecStart=/bin/bash -c \"while true; do echo Hi; sleep 10; done\"\n" > $GROUP/$GROUP-bar@.service

Validate the test group.
  $ inagoctl validate $GROUP
  Group '002-tunnel-group' is valid.
  Groups are valid globally.

Submit 2 slices of update group, using the tunnel flag.
  $ inagoctl --tunnel=${INAGO_TUNNEL_ENDPOINT} --ssh-strict-host-key-checking=false submit $GROUP 2
  .*\|\scontext.Background: Succeeded to submit group '002-tunnel-group'. (re)
  $ sleep 5

Start update group.
  $ inagoctl --tunnel=${INAGO_TUNNEL_ENDPOINT} --ssh-strict-host-key-checking=false start $GROUP
  .*\|\scontext.Background: Succeeded to start 2 slices for group '002-tunnel-group': \[([a-z0-9]{3}\s?){2}\]. (re)
  $ sleep 5

Test the status of the update group, after starting
  $ inagoctl --tunnel=${INAGO_TUNNEL_ENDPOINT} --ssh-strict-host-key-checking=false status $GROUP
  Group\s*Units\s*FDState\s*FCState\s*SAState\s*IP\s*Machine (re)
  
  002-tunnel-group@[a-z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  002-tunnel-group@[a-z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  
Update update-group without changing the unit file first.
  $ inagoctl --tunnel=${INAGO_TUNNEL_ENDPOINT} --ssh-strict-host-key-checking=false update --max-growth 1 --min-alive 1 $GROUP
  .*\|\scontext.Background: Not updating group '002-tunnel-group'. \(units already up to date\) (re)

Changing content of update-group-bar unit file.
  $ echo "[Unit]\nDescription=Inago Update Test Unit CHANGED\n\n[Service]\nExecStart=/bin/bash -c \"while true; do echo Hi; sleep 10; done\"" > $GROUP/$GROUP-bar@.service

Update update-group.
  $ inagoctl --tunnel=${INAGO_TUNNEL_ENDPOINT} --ssh-strict-host-key-checking=false update --max-growth 2 --min-alive 1 $GROUP
  .*controller: adding units (re)
  .*controller: adding units (re)
  .*controller: removing units (re)
  .*controller: removing units (re)
  .*\|\scontext.Background: Succeeded to update 2 slices for group '002-tunnel-group': \[([a-z0-9]{3}\s?){2}\]. (re)

Test the status of the updated group.
  $ inagoctl --tunnel=${INAGO_TUNNEL_ENDPOINT} --ssh-strict-host-key-checking=false status $GROUP
  Group\s*Units\s*FDState\s*FCState\s*SAState\s*IP\s*Machine (re)
  
  002-tunnel-group@[a-z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  002-tunnel-group@[a-z\d]{3}\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  

Kill it:

  $ inagoctl --tunnel=${INAGO_TUNNEL_ENDPOINT} destroy $GROUP > /dev/null 2>&1
