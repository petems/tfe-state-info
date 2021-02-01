# tfe-state-info

[![Build Status](https://travis-ci.com/petems/tfe-state-info.svg?branch=master)](https://travis-ci.com/petems/tfe-state-info)

A simple cli app to return certificates from a Vault PKI mount

## Install 

```
# install it into ./bin/
curl -sSfL https://raw.githubusercontent.com/petems/tfe-state-info/master/install.sh | sh -s v0.1.0
```

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

#### all-statefiles-size

Lists the total size of all statefiles for a workspace

Currently this is done by mass-downloading all of the listed statefiles for the workspace in question and then totalling the size of all downloads. As of 0.2.0, it cleans up files after download, this behaviour is configurable with the `--cleanup` flag, which defaults to true.

```
 $ export TFE_HOSTNAME=app.terraform.io
 $ export TFE_TOKEN=<REDACTED>
 $ tfe-state-info all-statefiles-size
 $ export TFE_ORG_NAME=psouter-hashicorp
 $ tfe-state-info all-statefiles-size
Total of all state file sizes for arbitrary-terraform-code-with-vcs was 0B (Statefile Count: 0)
Total of all state file sizes for arbitary-terraform-code was 1.4K (Statefile Count: 1)
Total of all state file sizes for folder-triggers-terraform was 0B (Statefile Count: 0)
Total of all state file sizes for aws-single-instance was 15.3K (Statefile Count: 3)
Total of all state file sizes for testing-output-changing was 1.9K (Statefile Count: 2)
Total of all state file sizes for terraform_tfvars_import was 0B (Statefile Count: 0)
Total of all state file sizes for testing-tfvar-export was 0B (Statefile Count: 0)
Total of all state file sizes for cfgmgmtcamp-2020-cost-restrict was 6K (Statefile Count: 1)
Total of all state file sizes for infracoding-with-terraform-testcon-2019 was 3K (Statefile Count: 2)
Total of all state file sizes for aws-single-instance-with-provisioner was 0B (Statefile Count: 0)
Total of all state file sizes for petersouterxyz-s3-website was 9.8K (Statefile Count: 1)
Total of all state file sizes for petersouterxyz-circle-ci-credentials was 7.2K (Statefile Count: 6)
Total of all state file sizes for tfe-saas-remote-data-example was 4.1K (Statefile Count: 2)
Total of all state file sizes for aws-single-micro was 91.9K (Statefile Count: 14)
```