# Paketo Buildpack for Dist Zip

## Buildpack ID: `paketo-buildpacks/dist-zip`
## Registry URLs: `docker.io/paketobuildpacks/dist-zip`

The Paketo Buildpack for DistZip is a Cloud Native Buildpack that contributes a Process Type for DistZip-style applications.

## Behavior

This buildpack will participate all the following conditions are met

* Exactly one file matching `<APPLICATION_ROOT>/$BP_APPLICATION_SCRIPT` exists

The buildpack will do the following:

* Requests that a JRE be installed
* Contributes `dist-zip`, `task`, and `web` process types

When `$BP_LIVE_RELOAD_ENABLE` is true:
* Requests that `watchexec` be installed
* Contributes `reload` process type

## Configuration

| Environment Variable      | Description                                                                                       |
| ------------------------- | ------------------------------------------------------------------------------------------------- |
| `$BP_APPLICATION_SCRIPT`  | Configures the application start script, using [Bash Pattern Matching][b]. Defaults to `*/bin/*`. |
| `$BP_LIVE_RELOAD_ENABLED` | Enable live process reloading. Defaults to false.                                                 |

## License

This buildpack is released under version 2.0 of the [Apache License][a].

[a]: http://www.apache.org/licenses/LICENSE-2.0
[b]: https://www.gnu.org/software/bash/manual/html_node/Pattern-Matching.html

