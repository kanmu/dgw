# dgw

[![Build Status](https://travis-ci.org/achiku/dgw.svg?branch=master)](https://travis-ci.org/achiku/dgw)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/achiku/dgw/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/achiku/dgw)](https://goreportcard.com/report/github.com/achiku/dgw)

## Description

`dgw` generates Table/Row Data Gateway Golang struct from PostgreSQL tables. Heavily inspired by [xo](https://github.com/knq/xo).


## Installation

```
go get -u github.com/achiku/dgw
```


## How to use

```
usage: dgw [<flags>] <conn>

Flags:
  --help             Show context-sensitive help (also try --help-long and --help-man).
  --schema="public"  PostgreSQL schema name
  --package="main"   package name
  --typemap=TYPEMAP  column type and go type map file path

Args:
  <conn>  PostgreSQL connection string in URL format
```

```
dgw postgres://dbuser@localhost/dbname?sslmode=disable 
```
