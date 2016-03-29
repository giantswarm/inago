Test that an invalid update operation raises an error message.

Setup test

  $ INITIAL_NUMBER_OF_SLICES=3
  $ MAX_GROWTH=1
  $ MIN_ALIVE=5
  $ GROUP=005-update-validation
  $ mkdir $GROUP
  $ echo "[Unit]\nDescription=Unit 1\n[Service]\nExecStart=/bin/bash -c 'while true; do echo Hello %n; sleep 10; done'\n" > $GROUP/$GROUP-unit@.service

Submit units

  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} submit $GROUP $INITIAL_NUMBER_OF_SLICES > /dev/null 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} start $GROUP > /dev/null 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP > /dev/null 2>&1
  $ sleep 10

Modify unit and perform update - invalid, as we cannot have a min-alive greater than the original number of units
  $ echo "[Unit]\nDescription=Unit 1 (CHANGED)\n[Service]\nExecStart=/bin/bash -c 'while true; do echo Hello %n; sleep 10; done'\n" > $GROUP/$GROUP-unit@.service
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} update $GROUP --max-growth=$MAX_GROWTH --min-alive=$MIN_ALIVE
  [a-zA-Z0-9\[\{\/\.\:\}\s]* update not allowed: cannot have minimum alive units greater than current number of units}] (re)
  [1]
  $ sleep 10

Shut down
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} stop $GROUP > /dev/null 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP > /dev/null 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} destroy $GROUP > /dev/null 2>&1
