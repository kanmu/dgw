# dgw

[![Build Status](https://travis-ci.org/achiku/dgw.svg?branch=master)](https://travis-ci.org/achiku/dgw)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/achiku/dgw/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/achiku/dgw)](https://goreportcard.com/report/github.com/achiku/dgw)

## Description

This tool generates Table/Row Data Gateway Golang struct from PostgreSQL tables.


## Installation

```
go get -u github.com/achiku/dgw/cmd/dgw
```


## How to use

```
usage: dgw [<flags>] <conn>

Flags:
  --help             Show context-sensitive help (also try --help-long and --help-man).
  --schema="public"  PostgreSQL schema name
  --typemap=TYPEMAP  type map file path

Args:
  <conn>  PostgreSQL connection string in URL format
```

```
dgw postgres://dbuser@localhost/dbname?sslmode=disable 
```
