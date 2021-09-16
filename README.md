# goapp [![Build Status](https://github.com/xgfone/goapp/actions/workflows/go.yml/badge.svg)](https://github.com/xgfone/goapp/actions/workflows/go.yml) [![GoDoc](https://pkg.go.dev/badge/github.com/xgfone/goapp)](https://pkg.go.dev/github.com/xgfone/goapp) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/goapp/master/LICENSE)

The package is used to initialize an application to simply the creation. Support `Go1.12+`.

## Install
```shell
$ go get -u github.com/xgfone/goapp
```

## Usage
```go
package main

import (
	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-log"
	"github.com/xgfone/goapp"
	"github.com/xgfone/goapp/router"
	"github.com/xgfone/ship/v5"
)

var opts = []gconf.Opt{
	gconf.StrOpt("opt1", "help doc"),
	gconf.IntOpt("opt2", "help doc"),
}

func main() {
	addr := gconf.NewString("addr", ":80", "The address to listen to.")
	goapp.Init("app", opts...)

	// Initialize the app router.
	app := router.InitRouter(nil)
	app.Logger = log.DefalutLogger
	app.Use(router.Logger(true), router.Recover)
	app.Route("/path1").GET(ship.OkHandler())
	app.Route("/path2").GET(func(c *ship.Context) error { return c.Text(200, "OK") })

	// Start the HTTP server.
	runner := ship.NewRunner(app)
	runner.RegisterOnShutdown(atexit.Execute)
	runner.Start(addr.Get())
}
```

Build the application above by `Makefile` like this,
```shell
$ make
```

```shell
$ ./app --help
  --addr string
        The address to listen to. (default: ":80")
  --config-file string
        the config file path. (default: "")
  --logfile string
        The file path of the log. The default is stdout. (default: "")
  --loglevel string
        The level of the log, such as debug, info (default: "info")
  --opt1 string
        help doc (default: "")
  --opt2 int
        help doc (default: "0")
  --version bool
        Print the version and exit. (default: "false")
```
