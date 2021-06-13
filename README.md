# goapp [![Build Status](https://api.travis-ci.com/xgfone/goapp.svg?branch=master)](https://travis-ci.com/github/xgfone/goapp) [![GoDoc](https://pkg.go.dev/badge/github.com/xgfone/goapp)](https://pkg.go.dev/github.com/xgfone/goapp) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/goapp/master/LICENSE)

The package is used to initialize an application to simply the creation.

## Install
```shell
$ go get -u github.com/xgfone/goapp
```

## Usage
```go
package main

import (
	"github.com/xgfone/gconf/v5"
	"github.com/xgfone/gconf/v5/field"
	"github.com/xgfone/goapp"
	"github.com/xgfone/goapp/router"
	"github.com/xgfone/ship/v4"
)

// Define some options.
var (
	conf config
	opts = []gconf.Opt{
		gconf.StrOpt("opt1", "help doc"),
		gconf.IntOpt("opt2", "help doc"),
	}
)

type config struct {
	Addr field.StringOptField `default:":80" help:"The address to listen to."`
}

func main() {
	// Initialize the app configuration
	goapp.Init("app", &conf, opts)

	// Initialize and start the app router.
	app := router.InitRouter()
	app.Route("/path1").GET(ship.OkHandler())
	app.Route("/path2").GET(func(c *ship.Context) error { return c.Text(200, "OK") })
	app.Start(conf.Addr.Get()).Wait()
}
```

Build the application above by using the script `build.sh` like this,
```shell
$ ./build.sh
```

```shell
$ ./app --help
  --addr string
        The address to listen to. (default ":80")
  --config-file string
        the config file path.
  --logfile string
        The file path of the log. The default is stdout.
  --loglevel string
        The level of the log, such as debug, info (default "info")
  --opt1 string
        help doc
  --opt2 int
        help doc
  --version bool
        Print the version and exit.
```
