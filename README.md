# Pliz

Pliz is a CLI tool wrapping some Docker/Docker Compose commands allowing to execute some tasks in a web project. Few examples:

* build the project: build images, start containers and execute tasks like `npm install` or `gulp`.
* start/stop the project
* execute some tasks
* backup/restore a running project (database & files)

Pliz makes some assumptions:

* Docker & Docker Compose must be used to develop and run a web project
* a project is composed of 4 Compose services: `app` (PHP by default), `db` (MySQL by default), `proxy` (nginx by default) and `srcbuild` (NodeJS stuff).
* some default tasks are defined inside Pliz: `bower`, `composer`, `db:update`, `gulp`, `npm`.
* configuration can be made using a `pliz.yml` file at project root. This configuration file can be used to:

    * override Compose service names
    * define configuration files inside the project that must be created by developers
    * define installation tasks (see below `pliz install`)
    * override default tasks behaviour
    * define new tasks
    * define some additionnal services to start with default ones
    * define backup content

Look at [pliz.example.yml](https://github.com/agence-webup/pliz/blob/master/pliz.example.yml) to see an example.

## How to use it

### Installation

Download the binary from [releases](https://github.com/agence-webup/pliz/releases) for your system and move it into your `$PATH` (for example: `/usr/local/bin/pliz`)

### Usage

Just call `pliz` and some help will be displayed (The command must be run in a Docker Compose project and a `pliz.yml` file must be present):

```bash
% pliz

Usage: pliz [OPTIONS] COMMAND [arg...]

Manage projects building

Options:
  -v, --version    Show the version and exit
  --env=""         Change the environnment of Pliz (i.e. 'prod'). The environment var 'PLIZ_ENV' can be use too.

Commands:
  start        Start (or restart) the project
  stop         Stop the project
  install      Install (or update) the project dependencies (docker containers, npm, composer...)
  bash         Display a shell inside the builder service (or the specified service)
  logs         Display logs of all services (or the specified service)
  run          Execute a single task
  backup       Perform a backup of the project
  restore      Restore a backup (Warning: files will be overrided)

Run 'pliz COMMAND --help' for more information on a command.
```

You can visit this project to see a use case of Pliz : [https://github.com/agence-webup/laravel-skeleton](https://github.com/agence-webup/laravel-skeleton)

_More documentation coming later_

## Building

### Requirements

- Go 1.6 (for vendor support)
- [glide](https://github.com/Masterminds/glide)
- _Optional_: [gox](https://github.com/mitchellh/gox) for cross compilation

### Compile Pliz

```bash
$ mkdir -p $GOPATH/src/webup
$ git clone https://github.com/agence-webup/pliz pliz
$ cd $GOPATH/src/webup/pliz
$ glide install
$ go install
```

### Cross compilation

```bash
$ cd $GOPATH/src/webup/pliz
$ gox -osarch="linux/amd64"
```
