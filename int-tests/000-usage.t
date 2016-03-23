  $ inagoctl
  Inago orchestrates groups of units on Fleet clusters
  
  Usage:
    inagoctl [flags]
    inagoctl [command]
  
  Available Commands:
    submit      Submit a group
    status      Get group status
    start       Start a group
    stop        Stop a group
    destroy     Destroy a group
    update      Update a group
    validate    Validate groups
    version     Print version
  
  Flags:
        --fleet-endpoint string   endpoint used to connect to fleet (default "unix:///var/run/fleet.sock")
    -h, --help                    help for inagoctl
        --no-block                block on synchronous actions
    -v, --verbose                 verbose output
  
  Use "inagoctl [command] --help" for more information about a command.
