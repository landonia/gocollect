# gocollect

The beginnings of a simple user data collection service that (currently) uses the BoltDB library

## Overview

I required a means of simply collecting user data mainly based on an email address.
Events (coming soon) can then be saved towards a user account and they can be fetched
based on a date stream and/or event type. An event can carry other params specific
to the event. It has no Authentication layer so it should be run behind a firewall
and accessed through another application. I intend to add in support for an event queue
that can be subscribed to for certain event types. I intend this to become a data store
that builds up over time to gain an overall picture of a user. Other bits of data can
be inserted as you gather more data.

## Installation

With a healthy Go Language installed, simply run `go get github.com/landonia/gocollect`

## Run

### Program parameters:

| Parameter     | Description             | Default Value                  | Acceptable                                             |
| ------------- | ----------------------- | ------------------------------ | ------------------------------------------------------ |
| db            | The path to the DB file | "/usr/local/gocollect/bolt.db" | A path to where you want the .db file to exist         |
| addr          | The address to bind     | ":8080"                        | Any valid bind address                                 |
| loglevel      | The log level           | "info"                         | "off","fatal","error","warn","info","debug","trace"    |

### Example

`gocollect -db=/my/path/to/bolt.db -addr=:8888 -loglevel=info`

## About

golog was written by [Landon Wainwright](http://www.landotube.com) | [GitHub](https://github.com/landonia).

Follow me on [Twitter @landotube](http://www.twitter.com/landotube)! Although I don't really tweet much tbh.
