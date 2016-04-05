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
    up          Bring a group up
    update      Update a group
    validate    Validate groups
    version     Print version
  
  Flags:
        --fleet-endpoint string          endpoint used to connect to fleet (default "unix:///var/run/fleet.sock")
    -h, --help                           help for inagoctl
        --no-block                       block on synchronous actions
        --ssh-known-hosts-file string    file used to store remote machine fingerprints (default "~/.fleetctl/known_hosts")
        --ssh-strict-host-key-checking   verify host keys presented by remote machines before initiating SSH connections (default true)
        --ssh-timeout duration           timeout in seconds when establishing the connection via SSH (default 10s)
        --ssh-username string            username to use when connecting to CoreOS machine (default "core")
        --tunnel string                  use a tunnel to communicate with fleet
    -v, --verbose                        verbose output
  
  Use "inagoctl [command] --help" for more information about a command.
