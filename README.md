![Baton Logo](./baton-logo.png)

# `baton-freshbooks` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-freshbooks.svg)](https://pkg.go.dev/github.com/conductorone/baton-freshbooks) ![main ci](https://github.com/conductorone/baton-freshbooks/actions/workflows/main.yaml/badge.svg)

`baton-freshbooks` is a connector for [FreshBooks](https://www.freshbooks.com/) built using the [Baton SDK](https://github.com/conductorone/baton-sdk).
This connector allows you to interact with the platform and to view the list of users and the permissions that each one has. However, the modification of the permits isn't available, since the platform does not allow modifications of this type to be made from the API.
FreshBooks uses OAuth 2.0 with the Authorization Code grant type.

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.

You can run this connector in two different modes:
1. Run it with an Access Token using the argument `--token`
or
2. Run it with a Refresh Token, Client ID and Client Secret of the Freshbooks account. Arguments: `--refresh-token`, `--fb-client-id` and `--fb-client-secret`

This second mode was added in case this connector recieves the adjustments needed to run as a service.

# Getting Started

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-freshbooks
baton-freshbooks
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_DOMAIN_URL=domain_url -e BATON_API_KEY=apiKey -e BATON_USERNAME=username ghcr.io/conductorone/baton-freshbooks:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-freshbooks/cmd/baton-freshbooks@main

baton-freshbooks

baton resources
```

# Data Model

`baton-freshbooks` will pull down information about the following resources:
- Users
- Roles

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually
building spreadsheets. We welcome contributions, and ideas, no matter how
small&mdash;our goal is to make identity and permissions sprawl less painful for
everyone. If you have questions, problems, or ideas: Please open a GitHub Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-freshbooks` Command Line Usage

```
baton-freshbooks

Usage:
  baton-freshbooks [flags]
  baton-freshbooks [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
      --client-id string             The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string         The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
  -f, --file string                  The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                         help for baton-freshbooks
      --log-format string            The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string             The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
  -p, --provisioning                 If this connector supports provisioning, this must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --ticketing                    This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                      version for baton-freshbooks

      --token string                 Access Token to connect to the platform. Basic functioning mode
      --refresh-token string         The Refresh Token that should be used to request a new Access Token when expired
      --fb-client-id string          The client ID used to authenticate with FreshBooks
      --fb-client-secret string      The client secret used to authenticate with FreshBooks

Use "baton-freshbooks [command] --help" for more information about a command.
```
