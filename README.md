# Golang application metrics
> NATS application metrics package

[![Build Status](https://travis-ci.org/cryptopay-dev/go-metrics.svg?branch=master)](https://travis-ci.org/cryptopay-dev/go-metrics)
[![codecov](https://codecov.io/gh/cryptopay-dev/go-metrics/branch/master/graph/badge.svg)](https://codecov.io/gh/cryptopay-dev/go-metrics)
[![GoDoc](https://godoc.org/github.com/cryptopay-dev/go-metrics?status.svg)](https://godoc.org/github.com/cryptopay-dev/go-metrics)
[![Go Report Card](https://goreportcard.com/badge/github.com/cryptopay-dev/go-metrics)](https://goreportcard.com/report/github.com/cryptopay-dev/go-metrics)

## Installation
```bash
go get github.com/cryptopay-dev/go-metrics
```

## Prerequisites
There should be 2 `env` variables defined:
- `INFLUX_METRICS_ENABLED` - to enable metrics at all
- `APPLICATION_NAME` - application name should be send

## Default metrics tags
```
hostname - application host
app - application name
```

## Usage
```go
package main

import (
    "log"

    "github.com/cryptopay.dev/go-metrics"
)

func main() {
    err := metrics.Setup("nats://localhost:4222", "metrics")
    if err != nil {
        log.Fatal(err)
    }

    for i:=0; i<10; i++ {
        err = metrics.SendAndWait(metrics.M{
            "counter": i,
        })

        if err != nil {
            log.Fatal(err)
        }
    }
}
```