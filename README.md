# `paketo-buildpacks/dist-zip`
The Paketo DistZip Buildpack is a Cloud Native Buildpack that contributes a Process Type for DistZip-style applications.

## Behavior
This buildpack will participate all the following conditions are met

* `<APPLICATION_ROOT>/*/bin/*` exists

The buildpack will do the following:

* Requests that a JRE be installed
* Contributes `dist-zip`, `task`, and `web` process types

## License
This buildpack is released under version 2.0 of the [Apache License][a].

[a]: http://www.apache.org/licenses/LICENSE-2.0
