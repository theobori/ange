# Gopher webring

[![build badge](https://github.com/theobori/ange/actions/workflows/build.yml/badge.svg)](https://github.com/theobori/ange/actions/workflows/build.yml)

[![built with nix](https://builtwithnix.org/badge.svg)](https://builtwithnix.org)

The name of the project is ange, pronounced \\ɑ̃ʒ\\, which corresponds to the word angel.

This GitHub repository is a [KISS](https://en.wikipedia.org/wiki/KISS_principle) project that contains a CLI, which is a Gopher application based on the [fleur](https://github.com/theobori/fleur) and [fleur-form](https://github.com/theobori/fleur-form) projects. It allows you to manage a persistent Gopher webring, meaning you can add, delete, and approve a Gopherspace. It also lets you retrieve a node’s neighbors and select a node at random.

## Definition

A webring is a chained list in which each node contains information about a website and its neighbors in the list. The most well-known ones are implemented using the HTTP(S) protocol, such as [baccyflap.com/noai](https://baccyflap.com/noai).

## Getting started

Before continuing to read the documentation, make sure you have at least [Go](https://go.dev/dl/) version 1.26.3.

### Installation

If you want to use the ange CLI, you can build and install the binary using the following command.

```bash
go install github.com/theobori/ange/cmd/ange
```

### Help message

Below is the CLI help message.

```text
Usage of ./ange:
  -database-path string
    	SQLite3 database path. (default "./ange-database.db")
  -directory string
    	It specifies an input directory path that will be the root of the Gopher server. (default "./")
  -domain string
    	Gopher domain. (default "localhost")
  -enable-tls
    	Enable Gopher over TLS
  -port int
    	Gopher port. (default 70)
  -tls-certificate string
    	x509 certificate path used for the TLS communication
  -tls-key string
    	Private key path used for the TLS communication
  -verbose
    	Enable verbose logs.
```

### Environment variable

You can set a value for the `ANGE_SECRET` environment variable, which will then be used for managing the webring. If it is left blank, a secret will be generated and displayed in your terminal emulator.

### Recommended client

As a user, I highly recommend [lagrange](https://gmi.skyjake.fi/lagrange/).

## How it works

The system is intentionally quite simple. The chain of Gopherspaces is stored persistently using [SQLite3](https://sqlite.org/). Users can perform operations on the webring via Gopher requests specifying specific paths.

Each node of the webring has a unique ID and a unique token. The token can be used by the users to remove their gopherspace from the webring.

### Webring routes

Below is a table showing the main routes and the associated operations on the webring.

| Path | Description |
| - | - |
| /webring/form | Page with a form to fill out to add a Gopherspace to the webring. |
| /webring/adminpanel | Administration panel for reviewing webrings. |
| /webring/adminpanel/form | Page with a form to approve or deny pending Gopherspaces. |
| /webring/delete | Page to delete your Gopherspace. |
| /webring/previous/<id> | View the previous Gopherspace relative to another. |
| /webring/next/<id> | View the next Gopherspace relative to another. |
| /webring/random | View a random Gopherspace from the webring. |
| Default route | By default, the application will serve a file named `gophermap`, which should be placed in the root of the directory specified as a CLI option. See [gophermap.example](/gophermap.example) |

### Gophermap evaluator extension

In your `gophermap` file, you can use the `>members` syntax element to display the members of the webring.

## Security recommendation

I strongly recommend that you use only Gopher over TLS, as the data exchanges may contain sensitive information such as tokens.

## Contribute

If you'd like to contribute to the project, please follow the instructions provided in the [CONTRIBUTING.md](./CONTRIBUTING.md) file.
