# backup2glacier
A CLI-Tool for uploading (encrypted) backups to AWS-Glacier

## Installation


## Usage example


## Documentation

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

* 0.0.1
    * CLI Command for uploading backups
    * CLI Command for list backups
    * CLI COmmand for downloding backups

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
