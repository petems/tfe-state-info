# tfe-state-info

[![Build Status](https://travis-ci.com/petems/tfe-state-info.svg?branch=master)](https://travis-ci.com/petems/tfe-state-info)

A simple cli app to return certificates from a Vault PKI mount

## Install 

```
# install it into ./bin/
curl -sSfL https://raw.githubusercontent.com/petems/tfe-state-info/master/install.sh | sh -s v0.1.0
```

## Example

![Example](./svg_demo.svg)

### Help

```
NAME:
  tfe-state-info - A simple cli app to return state information from TFE
USAGE:
  tfe-state-info [global options] command [command options] [arguments...]

COMMANDS:
   list-workspaces        List all workspaces for an Organization
   latest-statefile-size  Get latest statefile size for all workspaces
   all-statefiles-size    Get total size of all statefiles of all workspaces
   validate               Validate your current credentials
   help, h                Shows a list of commands or help for one command

TFE CONFIGURATION:
  TFE configuration is set by the common Vault environmental variables:
    TFE_HOSTNAME: The address for the TFE server (Required)
    TFE_TOKEN: The token for the TFE server (Required)
    ORG_NAME: The org you want to use

GLOBAL OPTIONS:
  --format value  The format you want them returned in, valid values are: table, json, pretty_json (default: "pretty_json")
  --silent        Do not output anything other than errors or returned data (default: true)
  --debug         Show debug information, with full http logs (default: false)
  --help, -h      show help (default: false)
  --version, -v   print the version (default: false)

VERSION:
  0.1.0-ba9d68a

```