# gokoori

A small `find`-like cli to mass pause/unpause [gocd](https://www.gocd.io/) pipelines.

## Usage

`./gokoori [options]`

Will query all available pipelines from `localhost:8153` with no filter and print their names.

### Options

#### Actions
- `--pause`: will pause all matching pipelines with the reason speacified in the `--reason` option (or the default one). Cannot be used in conjunction with `--unpause`.
- `--unpause`: will unpause all matching pipelines. Cannot be used in conjunction with `--pause`.

#### Filters
- `--name <regex>`: applies a regex filter by pipeline name.

#### Config
- `--reason <reason>`: specifies a reason to use for the `--pause` action. Ignored when `--pause` is not specified.
- `--host`: the hostname or ip where the gocd server is running. Defaults to localhost.
- `--port`: the port on which the gocd server is running on. Defaults to 8153.

## Build

`make`

## TODOs

- auth is not supported yet
- tls is not supported yet
