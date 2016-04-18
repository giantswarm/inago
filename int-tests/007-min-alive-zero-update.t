This test checks the behaviourof an update with max-growth zero

Setup unit file

  $ GROUP=007-min-alive-zero-update
  $ mkdir $GROUP
  $ echo "[Unit]\nDescription=Unit 1\n[Service]\nExecStart=/bin/bash -c 'while true; do echo Hello %n; sleep 10; done'\n" > $GROUP/$GROUP-unit@.service

Submit and start group

  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} submit $GROUP 2 > /dev/null 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} start $GROUP > /dev/null 2>&1
  $ sleep 3

Modify unit

  $ echo "[Unit]\nDescription=Unit 1 (CHANGED)\n[Service]\nExecStart=/bin/bash -c 'while true; do echo Hello %n; sleep 10; done'\n" > $GROUP/$GROUP-unit@.service

Update unit

  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} update $GROUP --max-growth=0 --min-alive=0
  .*\|\scontext.Background: Succeeded to update 2 slices for group '007-min-alive-zero-update': \[[a-z0-9]{3} [a-z0-9]{3}\]. (re)
  $ sleep 3

Tear down

  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} destroy $GROUP
  .*\|\scontext.Background: Succeeded to destroy 2 slices for group '007-min-alive-zero-update': \[[a-z0-9]{3} [a-z0-9]{3}\]. (re)
