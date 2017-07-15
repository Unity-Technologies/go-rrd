# go-rrd [![Go Report Card](https://goreportcard.com/badge/github.com/multiplay/go-rrd)](https://goreportcard.com/report/github.com/multiplay/go-rrd) [![License](https://img.shields.io/badge/license-BSD-blue.svg)](https://github.com/multiplay/go-rrd/blob/master/LICENSE) [![GoDoc](https://godoc.org/github.com/multiplay/go-rrd?status.svg)](https://godoc.org/github.com/multiplay/go-rrd) [![Build Status](https://travis-ci.org/multiplay/go-rrd.svg?branch=master)](https://travis-ci.org/multiplay/go-rrd)

go-rrd is a [Go](http://golang.org/) client for talking to [rrdtool's](https://oss.oetiker.ch/rrdtool/index.en.html) [rrdcached](https://oss.oetiker.ch/rrdtool/doc/rrdcached.en.html).

Features
--------
* Full [rrdcached](https://oss.oetiker.ch/rrdtool/doc/rrdcached.en.html) Support.

Installation
------------
```sh
go get -u github.com/multiplay/go-rrd
```

Examples
--------

Using go-rrd is simple just create a client and then send commands e.g.
```go
package main

import (
	"log"
	"time"

	"github.com/multiplay/go-rrd"
)

func main() {
	c, err := rrd.NewClient("192.168.1.102:10011")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	if err := c.Create(
		"test.rrd",
		[]DS{NewGauge("watts", time.Minute*5, 0, 24000)},
		[]RRA{NewAverage(0.5, 1, 864000)},
	); err != nil {
		log.Fatal(err)
	}
}
```

Documentation
-------------
- [GoDoc API Reference](http://godoc.org/github.com/multiplay/go-rrd).

License
-------
go-rrd is available under the [BSD 2-Clause License](https://opensource.org/licenses/BSD-2-Clause).
```
