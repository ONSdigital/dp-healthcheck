dp-healthcheck
================

A health check git repository for DP

### Getting started

Read the [Health Check Specification](https://github.com/ONSdigital/dp/blob/master/standards/HEALTH_CHECK_SPECIFICATION.md) for details.

#### How to use
1. Add Health Check library to an app

2. Create an array of Health Check clients by calling `NewClient()` passing in the following:

- An optional RCHTTP clienter; if none is provided one will be created
- A function that implements the `Checker` interface

3. Call `Create()` passing in the following:

- Versioning Information
- Critical time duration; time to wait for dependent apps critical unhealthy status to make current app unhealthy- Time Interval to run health checks on dependencies
- Clients; An array of clients created in the previous step, like so:

```
package main

import (
    ...
    health "github.com/ONSdigital/dp-healthcheck/healthcheck"
    ...
)
...
var BuildTime, GitCommit, Version string
...

func main() {
    ...

    criticalTimeout := time.Minute
    interval := 10 * time.Second

    versionInfo := health.CreateVersionInfo(
        time.Unix(BuildTime, 0),
        GitCommit,
        Version,
    )

    # Initialise your clients
    cli1 := client1.NewAPIClient()
    cli2 := client2.NewDataStoreClient()

    hc := health.Create(versionInfo criticalTimeout, interval, &cli1.Checker, &cli2.Checker)

    ...
}
```

4. Optionally call `AddClient` on the healthcheck to add additional clients, note this can only be done prior to `Start()` being called

```
    ...
    mongoClient := <mongo health client>

    if err = hc.AddCheck(&mongoClient.Checker); err != nil {
        ...
    }
    ...

```

5. Setting the BuildTime, GitCommit and Version during compile time, using the following commands:

```
BUILD_TIME=$(date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)

go build -ldflags="-X 'main.BuildTime=$BUILD_TIME' -X 'main.GitCommit=$GIT_COMMIT' -X 'main.Version=$VERSION'"`
```

Makefile example:

```
...
BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)
...

build:
        @mkdir -p $(BUILD_ARCH)/$(BIN_DIR)
        go build -o $(BUILD_ARCH)/$(BIN_DIR)/$(MAIN) -ldflags "-X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT -X main.Version=$VERSION" cmd/$(MAIN)/main.go
debug:
        HUMAN_LOG=1 go run -race -ldflags "-X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT -X main.Version=$VERSION" cmd/$(MAIN)/main.go
...
```

6. Call `Start()` on the healthcheck

```
    ...
    ctx := context.Context(context.Background())

    hc.Start(ctx)
    ...
```

7. Call `Stop()` on healthcheck to gracefully shutdown application

```
    ...
    hc.Stop()
    ...
```

### Configuration

Configuration of the health check takes place via arguments passed to the `Create()` function, this includes a variable of `VersionInfo` which can be created by passing arguments to the `CreateVersionInfo()` function, see below example of setup:

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2019, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
