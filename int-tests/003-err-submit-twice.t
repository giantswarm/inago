We had a bug where submit would block forever, because it would check on the wrong unit files.

  $ GROUP="test-group-003"
  $ mkdir $GROUP 
  $ printf "[Unit]\nDescription=Inago Test Unit $GROUP\n\n[Service]\nExecStart=/bin/bash -c \"while true; do echo Hi $GROUP; sleep 10; done\"\n" > $GROUP/$GROUP-unit.service

  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} submit $GROUP >/dev/null 2>&1 
  $ sleep 5
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} start $GROUP >/dev/null 2>&1
  $ sleep 5
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} submit $GROUP >/dev/null 2>&1

Cleanup
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status $GROUP >/dev/null 2>&1
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} destroy $GROUP >/dev/null 2>&1
