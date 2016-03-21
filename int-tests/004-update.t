This cram test performs the basic operations on an instance group, meaning no slices.

Setup unit files

  $ GROUP=004-update
  $ mkdir $GROUP
  $ echo "[Unit]\nDescription=Unit 1\n[Service]\nExecStart=/bin/bash -c 'while true; do echo Hello %n; sleep 10; done'\n" > $GROUP/$GROUP-1-unit@.service
  $ echo "[Unit]\nDescription=Unit 2\n[Service]\nExecStart=/bin/bash -c 'while true; do echo Hello %n; sleep 10; done'\n[X-Fleet]\nMachineOf=$GROUP-1-unit@%i.service\n" > $GROUP/$GROUP-2-unit@.service

Submit units

  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} submit $GROUP 2 >010.out 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP -v >021.out 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP >025.out 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} start $GROUP >030.out 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP -v >041.out 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP > 045.out  2>&1
  $ sleep 10

Modify unit and perform update
  $ echo "[Unit]\nDescription=Unit 1 (CHANGED)\n[Service]\nExecStart=/bin/bash -c 'while true; do echo Hello %n; sleep 10; done'\n" > $GROUP/$GROUP-1-unit@.service
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} update $GROUP --max-growth=1 --min-alive=1 > 48.out 2>&1
  $ sleep 10

Shut down
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} stop $GROUP >050.out 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP -v >061.out 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP >065.out 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} destroy $GROUP >070.out 2>&1
