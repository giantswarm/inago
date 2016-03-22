This cram test performs the basic operations on an instance group, meaning no slices.

Setup unit files

  $ GROUP=050-simple-group
  $ mkdir $GROUP
  $ echo "[Unit]\nDescription=Unit 1\n[Service]\nExecStart=/bin/bash -c 'while true; do echo Hello %n; sleep 10; done'\n" > $GROUP/$GROUP-1-unit.service
  $ echo "[Unit]\nDescription=Unit 1\n[Service]\nExecStart=/bin/bash -c 'while true; do echo Hello %n; sleep 10; done'\n[X-Fleet]\nMachineOf=$GROUP-1-unit.service\n" > $GROUP/$GROUP-2-unit.service

Submit units

  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} submit $GROUP >010.out 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP -v >021.out 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP >025.out 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} start $GROUP >030.out 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP -v >041.out 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP > 045.out  2>&1
  $ egrep --quiet '\s+active\s+' 045.out 
  $ sleep 10

All up, now shut down again and cleanup

  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} stop $GROUP >050.out 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP -v >061.out 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP >065.out 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} destroy $GROUP >070.out 2>&1
