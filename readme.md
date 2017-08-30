# gopherb2
GopherB2 is a basic library for interacting with Backblaze B2 Storage in Golang

- [Overview](#overview)
- [Command Line App](#command-line-app)
  - [Install](#cli-install)
  - [Usage](#cli-usage)
- [Planned Features](#planned-features)

![Go Report Card](https://goreportcard.com/badge/github.com/dwin/gopherb2)
[![license](https://img.shields.io/github/license/dwin/gopherb2.svg?style=flat-square)](https://github.com/dwin/gopherb2/blob/Development/LICENSE)

---

## Overview

---

## Backblaze B2 Credentials

Must set in Configuration file or with OS Environment variables

You can environment variables for gopherb2 to authorize with the Backblaze B2 API. The application will expect the following:

- ```B2AcctID```
- ```B2AppID```
- ```B2APIURL```

To set on macOS & *nix in Terminal:

```bash
export B2AcctID=123464abc
export B2AppID=456789ddffgghhii
export B2APIURL=https://api.backblazeb2.com/b2api/v1/
```

You can also set the necessary credentials in ```$GOPATH/src/github.com/dwin/gopherb2/config``` :

```bash
nano $GOPATH/src/github.com/dwin/gopherb2/config
```

You should see a file that appears like the example below that you will need to edit with your credentials:

```toml
# BackBlaze B2 API credentials
# You will need to get these from the BackBlaze API Dashboard
[Account1]
  AcctID = "3a1234567b89"
  AppID = "001f38150dfsdgfdsgdfsg80c23b9"
  APIURL = "https://api.backblazeb2.com/b2api/v1/"
```

> The URL above is the correct URL as of writing. You will need to obtain other credentials from the B2 Dashboard ([https://secure.backblaze.com/b2_buckets.htm](https://secure.backblaze.com/b2_buckets.htm)), then click 'Show Account ID and Application Key".

---

## Command Line App

### CLI Install

To install command line app:

```bash
go get github.com/dwin/gopherb2
```

Make sure your ```PATH``` includes the ```$GOPATH/bin``` directory so your commands can be easily used:

```bash
export PATH=$PATH:$GOPATH/bin
```

Then:

```bash
cd $GOPATH/src/github.com/dwin/gopherb2/gb2 && go install
```

### CLI Usage

```bash
gb2 [global options] command [command options] [arguments...]
```

Show Help

```bash
gb2 help
```

#### CLI Example

```bash
NAME:
   gb2 - [global options] command [command options] [arguments...]

USAGE:
   gb2 [global options] command [command options] [arguments...]

VERSION:
   0.1.0

DESCRIPTION:
   Application for managing and interacting with Backblaze B2

COMMANDS:
     bucket, buckets  [global] bucket [command] [arguments...]
     upload, put      [global] upload [bucket id] [path or file]
     file, files      [global] file [command] [arguments..]
     version, v       Display version
     help, h          Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --log gopher.log                 gb2 -log gopher.log
   --debug -debug|-d, -d -debug|-d  -debug|-d [command]
   --help, -h                       show help
   --version, -v                    print the version
```

---

## Library Usage

---

## Planned Features

- Directory Sync
- Compress and Upload
- File Encryption
- Basic GUI
