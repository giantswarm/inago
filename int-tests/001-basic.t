Create a group to play around with.
  $ mkdir test-group
  $ echo "[Unit]\nDescription=Inago Test Unit\n\n[Service]\nExecStart=/bin/bash -c \"while true; do echo Hi; sleep 10; done\"" > test-group/test-group-unit.service


Validate the test group.
  $ inagoctl validate test-group
  Group 'test-group' is valid.
  Groups are valid globally.


Submit the test group.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} submit test-group
  Succeeded to submit all slices of group 'test-group'.
  $ sleep 5


Test the status of the test group, after submission.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status test-group
  Group\s*Units\s*FDState\s*FCState\s*SAState\s*IP\s*Machine (re)
  
  test-group\s*\*\s*loaded\s*loaded\s*inactive\s*[0-9.]*\s*[a-z0-9]* (re)
  

Start the test group.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} start test-group
  Succeeded to start all slices of group 'test-group'.
  $ sleep 5


Test the status of the test group, after starting.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status test-group
  Group\s*Units\s*FDState\s*FCState\s*SAState\s*IP\s*Machine (re)
  
  test-group\s*\*\s*launched\s*launched\s*active\s*[0-9.]*\s*[a-z0-9]* (re)
  

Stop the test group
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} stop test-group
  Succeeded to stop all slices of group 'test-group'.
  $ sleep 5


Test the status of the test group, after stopping.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status test-group
  Group\s*Units\s*FDState\s*FCState\s*SAState\s*IP\s*Machine (re)
  
  test-group\s*\*\s*loaded\s*loaded\s*inactive\s*[0-9.]*\s*[a-z0-9]* (re)
  

Destroy the test group
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} destroy test-group
  Succeeded to destroy all slices of group 'test-group'.
  $ sleep 5


Test the status of the test group, after destruction.
  $ inagoctl --fleet-endpoint=${FLEET_ENDPOINT} status test-group
  Failed to find all slices of group 'test-group'.
  [1]
