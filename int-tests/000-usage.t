  $ inagoctl
  orchestrate groups of unit files on Fleet clusters
  
  Usage:
    inagoctl [flags]
    inagoctl [command]
  
  Available Commands:
    submit      submit a group
    status      status of a group
    start       Starts the specified group or slices
    stop        Stops the specified group or slices
    destroy     Destroys the specified group or slices
    update      update a group
    validate    validate groups
  
  Flags:
        --fleet-endpoint string   endpoint used to connect to fleet (default "unix:///var/run/fleet.sock")
    -h, --help                    help for inagoctl
        --no-block                block on synchronous actions or not
    -v, --verbose                 verbose output or not
  
  Use "inagoctl [command] --help" for more information about a command.
