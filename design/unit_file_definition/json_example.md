```
$ cat example.json
{
	"units": {
		"ambassador.service": {
			"includes": ["ambassador.json", "restart_policy.json"],
			"overrides": {
				"Unit": {
					"Wants": ["user-app.service"]
				}
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
	"systemd": {
		"Service": {
			"Restart": "on-failure",
			"RestartSec": "1s",
			"StartLimitInterval": "300s",
			"StartLimitBurst": "3",
			"TimeoutStartSec": "0"
		}
	}
}
```