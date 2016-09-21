```
$ cat example.yml
--- 
units: 
  ambassador.service: 
    includes: 
      - ambassador.json
      - restart_policy.json
    overrides: 
      Unit: 
        Wants: 
          - user-app.service
  lb-register.service: 
    includes: 
      - restart_policy.json
    systemd: 
      Service: 
        ExecStart: "docker run..."
      Unit: 
        After: 
          - user-app.service
        Description: description
        Wants: 
          - user-app.service
      X-Fleet: 
        Global: true
  user-app.service: 
    includes: 
      - restart_policy.json
    systemd: 
      Service: 
        ExecStart: "docker run..."
      Unit: 
        After: 
          - ambassador.service
        Description: description
        Wants: 
          - ambassador.service
          - lb-register.service
      X-Fleet: 
        Global: true

$ cat ambassador.yml
--- 
systemd: 
  Service: 
    ExecStart: "docker run..."
  Unit: 
    Description: description
    Wants: []
  X-Fleet: 
    Global: true

$ cat restart_policy.yml
---
  systemd: 
    Service: 
      Restart: "on-failure"
      RestartSec: "1s"
      StartLimitInterval: "300s"
      StartLimitBurst: "3"
      TimeoutStartSec: "0"
```