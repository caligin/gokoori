# gokoori

[![Build Status](https://travis-ci.org/caligin/gokoori.svg?branch=master)](https://travis-ci.org/caligin/gokoori)

A small `find`-like cli to mass pause/unpause [gocd](https://www.gocd.io/) pipelines.

## Usage

`./gokoori [options]`

Will query all available pipelines from `https://localhost:8154` with no filter and print their names.

### Authentication

Gokoori will read credentials form a file named `~/.gokoori/credentials` and use them automatically for basic auth, if the file is present.

The file is in JSON format and should contain an object with the fields `username` and `password`, like this:

```
{"username":"me","password":"my_pass"}
```

### Options

#### Actions
- `--pause`: will pause all matching pipelines with the reason speacified in the `--reason` option (or the default one). Cannot be used in conjunction with `--unpause`.
- `--unpause`: will unpause all matching pipelines. Cannot be used in conjunction with `--pause`.

#### Filters
- `--name <regex>`: applies a regex filter by pipeline name.

#### Config
- `--reason <reason>`: specifies a reason to use for the `--pause` action. Ignored when `--pause` is not specified.
- `--host`: the hostname or ip where the gocd server is running. Defaults to localhost.
- `--port`: the port on which the gocd server is running on. Defaults to 8154, or 8153 when the `--insecure` option is specified.
- `--insecure`: uses HTTP plain instead of https. Causes the port to default to 8153 if not explicitly set.

## Build

`make`

## Run the playground gocd server

`make docker`

## Contributing

- PRs welcome!
- I deliberately wrote code without studying golang properly to try and have a crash-course, if you have any comments, suggestions or insults open an issue or ping me on twitter!

## TODOs

- tests. doing none for now was 100% a conscious decision
