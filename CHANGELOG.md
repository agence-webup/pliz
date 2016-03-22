# rev 2-dev
WIP

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
