# rev 3
WIP

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
