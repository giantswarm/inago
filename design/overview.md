# Design Overview

Details the overall design of Formica - a management layer over fleet

- _Formica_ is a genus of ants, see https://en.wikipedia.org/wiki/Formica
- See overview.png for a more graphical view of this material

## General Design Overview
- Service running inside cluster named ```formicad```
- Command line tool to interact with ```formicad``` named ```formicactl```

### ```formicad```
- Handles generating unit files from unit file definitions
- Handles submitting and updating unit files
- Supports submitting unit file definitions
- Supports querying against the state of the unit file definition being processed
- Supports users viewing the generated unit files
- Fleet is source of truth - no state is held in ```formicad``` that can not be regenerated from fleet
- Uses a task system with swappable backends
- No user management, authentication or authorization

### ```formicactl```
- Supports sending unit file definitions to ```formicad```
- Supports querying state of unit file definition being processed

## Example Flow

### Using command line tool
- User submits unit file definition to service using command line tool
- Service validates unit file definition, returns 200
- Service generates unit files from unit file definition
- Service submits all unit files to fleet cluster
- Meanwhile, command line tool periodically queries service for unit file definition status, displays status to user

## Shipping
- Docker image as main deployment artifact
- Continuous integration with automated unit tests
- Continuous deployment to build Docker image and push image to Docker Hub

## Unit File Definition
This shows an example of and describes the unit file definition.
```
$ cat example.json
{
  "units": {
    "ambassador.service": {
      "includes": ["ambassador.json", "restart_policy.json"],
      "overrides": {
        "Unit": {
          "Wants": ["user-app.service]
        },
      }
    },
    "user-app.service": {
      "includes": ["restart_policy.json"],
      "systemd": {
        "Unit": {
          "Description": "description",
          "Wants": [
            "ambassador.service",
            "lb-register.service"
          ],
          "After": [
            "ambassador.service"
          ]
        },
        "Service": {
          "ExecStart": "docker run..."
        },
        "X-Fleet": {
          "Global": true
        }
      }
    },
    "lb-register.service": {
      "includes": ["restart_policy.json"],
      "systemd": {
        "Unit": {
          "Description": "description",
          "Wants": [
            "user-app.service"
          ],
          "After": [
            "user-app.service"
          ]
        },
        "Service": {
          "ExecStart": "docker run..."
        },
        "X-Fleet": {
          "Global": true
        }
      }
    }
  }
}

$ cat ambassador.json
{
  "systemd": {
    "Unit": {
      "Description": "description",
      "Wants": []
    },
    "Service": {
      "ExecStart": "docker run..."
    },
    "X-Fleet": {
      "Global": true
    }
  }
}

$ cat restart_policy.json
{
  "Restart": "on-failure",
  "RestartSec": "1s",
  "StartLimitInterval": "300s",
  "StartLimitBurst": "3",
  "TimeoutStartSec": "0"
}
```
Units are described by the "systemd" key, whose keys map directly to systemd unit values, allowing new systemd unit values without re-releasing

The "includes" key allows for values from other dictionaries to be included.
The "overrides" key allows for keys to be overriden from included dictionaries.

With regards to the service API, the files would be joined together into one larger dictionary, for example:

```
{
    "example.json": {
        "units": {
        ...
    "ambassador.json": {
        "systemd": {
        ...
    },
    "restart_policy.json": {
        "Restart": "on-failure",
        ...
    }
}
```
The service API requires more fleshing out, this is for exemplar purposes