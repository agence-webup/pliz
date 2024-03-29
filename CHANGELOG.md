# rev 11

- Add support for MariaDB database
- Add encryption on backup

# rev 10

- Add the task `git:hook`

# rev 9

- Drop database items before restoring a Postgres DB

# rev 8

- Add support for PostgreSQL database
- add the ability to skip some tasks during `pliz install`

# rev 7

- Add databases configuration for MySQL backup

# rev 6
10/23/2017

- Add a `quiet` option for `pliz restore`
- Display a warning when using pliz in prod environment without a `docker-compose.prod.yml`

# rev 5
8/21/2016

- Improve error handling for `pliz backup`

# rev 4
6/8/2016

- Don't display access url after `pliz start` for production environment
- Fix #3: `pliz restore` prompts to `false` by default

# rev 3
5/12/2016

- Rename the task `db-update` to `db:update`. Every custom task should respect this convention.
- Add a new config entry `additional_startup_containers` to start some containers with `pliz start`
- Add a new config entry `startup_container`. Allows to override the container used by `pliz start` to launch project
- Add some options to `pliz backup` to use the command in scripts (disable prompt)
- Improve UX using some colors and some informations like the accessible URL of the proxy
- Start the containers just after the build step during a `pliz install`, fixing issue with db-update which cannot be executed.
- Add `pliz backup` and `pliz restore` to manage backup: gzip archive containing config files, list of specified files and DB dumps

# rev 2
4/5/2016

- __BREAKING CHANGE__: Refactoring of the config file
    - allow to override default tasks even if they are not in `enabled_tasks`
    - new section: `tasks`
    - renamed section: `enabled_tasks` => `install_tasks`
    - overrided default tasks: `override` keyword is removed

    Check the `pliz.example.yml` to review changes

- Add the option `-p` for `pliz bash` allowing to publish some ports (i.e. exposing ports needed by BrowserSync)
- Remove the `restart` command (always use `start`)
- Handle the production environment with the option `--env`: use the `docker-compose.prod.yml` instead of `docker-compose.override.yml`
- Don't use a hard link for config files anymore. Classic file's content copy instead.
- Fix issue with `pliz bash`: the config was not parsed before attempting to use it


# 1.0.1
3/11/2016

- Fix issue with the container for the custom tasks (#1)


# 1.0
3/08/2016

First release
