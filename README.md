# Rolebot

a super simple bot for setting roles on discord. meant to be self hosted and used on directly managed servers only.

## Build and install
build the go module (ie. `go build`). a systemd service is included (`rolebot.service`). install this into `/etc/systemd/system/` or wherever you'd like to place it then enable it for autostarting and restarting on fail. the `rolebot` binary must be installed to `/usr/local/bin/rolebot`.

## Config
the default config directory is `/etc/rolebot/`. the environment variable `CONFIG_DIR` can be set for an alternate directory.

config directory structure:
```
/etc/rolebot/
     |- token (contains the discord bot token)
     |- appid (contains the application ID)
     -- guilds/ (contains per-guild config)
        - <guild_id>.toml (detailed below)
```

### `guilds/<guild_id>.toml`
per-guild configuration file of a list of role groups. sample config:

```toml
# /etc/rolebot/guilds/1234.toml

[[RoleGroups]]
Heading = "Group 1"
Description = "Some group of roles"
Multiple = false  # allow multiple roles from this group?
Roles = [
    "1234",  # role IDs must be strings
    "5678",
]

[[RoleGroups]]
Heading = "Group 2"
Description = "Some other group"
Multiple = true
Roles = [
    "7890",
    "0987",
]
```

<!--
vim:linebreak
-->
