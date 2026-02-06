# WaterDuty.org)

## Setup (Windows)

Copy `set_env_vars.ps1.template` to `set_env_vars.ps1`. Then set the environment variables. [Obtain a mapbox access token](https://account.mapbox.com/access-tokens/). Next run

```powershell
. .\set_env_vars.ps1
```

Run locally:

```bash
go mod tidy
go run main.go
```

Open http://localhost:8080, click `Login (dev)` to quickly sign in for development. Then click the map to pick a location and press Save.
