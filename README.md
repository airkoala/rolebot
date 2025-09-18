# Rolebot

a super simple bot for setting roles on discord. meant to be self hosted and used on directly managed servers only.

## Build and install
build the go module with `go build` and then install the binary to `/usr/local/bin/rolebot` for the included systemd service (`rolebot.service`) to work. the service should be installed into `/etc/systemd/system/rolebot.service` or wherever and then enabled with `systemctl enable --now rolebot`.

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
per-guild configuration as a list of role groups. example config:

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
