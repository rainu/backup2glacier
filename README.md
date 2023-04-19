# backup2glacier
A CLI-Tool for uploading (encrypted) backups to AWS-Glacier

## Installation

0. First you need a AWS user for programmatic access. How it goes you can find here: [Creating an IAM User in Your AWS Account](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html).
    0. The user needs full glacier rights. Just gave him the role **AmazonGlacierFullAccess**.
0. Create a vault. How it goes see: [Create a Vault in Amazon Glacier](https://docs.aws.amazon.com/amazonglacier/latest/dev/getting-started-create-vault.html)
0. You have to create your AWS config files like described here: [Configuration and Credential Files](https://docs.aws.amazon.com/cli/latest/userguide/cli-config-files.html)
0. Download the binary of [backup2glacier](#) or [build on your own](#Development setup)

## Usage

Upload a backup to glacier
```bash
./backup2glacier CREATE <vault name> [<file or dir to backup>, ...]
```

Show Backups
```bash
./backup2glacier LIST
```

Download a backup to glacier
```bash
./backup2glacier GET <BackupID> <target file on your desk>
```

Delete Backups older than 30 days
```bash
./backup2glacier CURATOR <vaultname> --max-age 30
```

More information
```bash
./backup2glacier -h
./backup2glacier CREATE -h
./backup2glacier LIST -h
./backup2glacier SHOW -h
./backup2glacier GET -h
./backup2glacier CURATOR -h
```

## Development setup

The following scriptlet shows how to setup the project and build from source code.

```sh
mkdir -p ./workspace/src
export GOPATH=$(pwd)/workspace

cd ./workspace/src
git clone git@github.com:rainu/backup2glacier.git

cd backup2glacier
go get ./...
go build
```

## Release History

* 0.2.5
    * add whitelist functionality for CREATE command
* 0.2.4
    * delete the archive even if the archive id is empty
* 0.2.3
    * ask for the password if no one is in the database and no corresponding argument is giving
* 0.2.2
    * Fix nil pointer error if there is a glacier job of type != "ArchiveRetrieval" 
    * Fix nil pointer error if upload failed
* 0.2.1
    * CLI curator option for keeping n backups
* 0.2.0
    * Fix CLI help
    * add option for length factor in LIST command
    * CLI Command for curator backups
* 0.1.1
    * CLI option for blacklisting files
    * Correct default value (1) of option for upload part size
* 0.1.0
    * Backup will not exit on files which does not exists (dead links)
* 0.0.2
    * CLI Command for delete backups
* 0.0.1
    * CLI Command for uploading backups
    * CLI Command for downloading backups
    * CLI Command for list backups

## Meta

Distributed under the MIT license. See ``LICENSE`` for more information.

### Intention

I want to use [AWS Glacier](https://aws.amazon.com/glacier/) for my backups. But i dont want to upload it uncompressed
an plain. The other thing: i needed a place to put some meta information for my glacier archives (such like the archive-id,
description and the CONTENT). So i decide to implement this tool to make it easier to use it for backups.

## Contributing

1. Fork it (<https://github.com/rainu/backup2glacier/fork>)
2. Create your feature branch (`git checkout -b feature/fooBar`)
3. Commit your changes (`git commit -am 'Add some fooBar'`)
4. Push to the branch (`git push origin feature/fooBar`)
5. Create a new Pull Request
