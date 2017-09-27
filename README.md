# go-servd
go-servd is standard API for service like daemon.

## Installation
```
$ go get github.com/KurioApp/go-servd
```

## Overview
Standard way to run
```golang
    httpd := NewHTTPService() // construct your Servd

    // Standard way to run
    go func() {
        if err := httpd.Run(); err != nil {
        // fail to run
    }
    }()
```

Standard way to stop
```golang
    if ok := httpd.Stop(); !ok {
        // already been stop
    }
```

Standard way to wait for service to be shutdown gracefully
```golang
    // wait and get the status
    stat := httpd.WaitForStatus(servd.Stopped)
```
