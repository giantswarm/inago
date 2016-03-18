# Integration Test Server Setup

This document describes the set up for the integration test server.

`t2.micro` running in `eu-central-1`, with `ami-15190379` (CoreOS 835.13.0).

Server has both `etcd` and `fleet` running - use:
```
sudo systemctl start fleet
sudo systemctl start etcd2
```
to start both services.
