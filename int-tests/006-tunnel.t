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
  $ inagoctl --tunnel=localhost submit $GROUP 2
  .*\|\scontext.Background: Succeeded to submit group '002-tunnel-group'. (re)
  $ sleep 5


Clean Up:

    $ inagoctl --tunnel=localhost destroy $GROUP > /dev/null 2>&1
