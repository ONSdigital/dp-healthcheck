dp-healthcheck
==============

A health check git repository for Digital Publishing that implements the [Health Check Specification](https://github.com/ONSdigital/dp/blob/master/standards/HEALTH_CHECK_SPECIFICATION.md).  All Digital Publishing apps must implement a health check using this library.  Functions that implement the `Checker` type are registered with the library which will check internal and external measures of the apps health.  The library will then call these functions periodically to determine the overall health of the app and report this back using the included handler.

Getting started
---------------

* [Add health check to an app](#adding-a-health-check-to-an-app)
* [Implementing a `Checker` function](#implementing-a-checker)

Adding a health check to an app
-------------------------------

1. Import the library and your HTTP server dependencies:

    ```
    package main

    import (
        ...
        health "github.com/ONSdigital/dp-healthcheck/healthcheck"
        "github.com/ONSdigital/go-ns/server"
        "github.com/gorilla/mux"
        ...
    )
    ```

2. Create a version object defining the version of your app:

    ```
    ...

    var BuildTime, GitCommit, Version string

    ...

    func main() {
        ctx := context.Context(context.Background())

        ...

        // Likely these values would come from your app's config
        criticalTimeout := time.Minute
        interval := 10 * time.Second

        ...

        // Likely these values would come from your app's config
        criticalTimeout := time.Minute
        interval := 10 * time.Second

        ...

        versionInfo := health.CreateVersionInfo(
            BuildTime,
            GitCommit,
            Version,
        )

        ...
    ```

2. Initialise any clients that have `Checker` type functions you wish to use

3. Instantiate the health check library:

    If you have only have a few `Checker` functions to register you can pass them in to Create:

    ```
        ...

        hc, err := health.Create(versionInfo criticalTimeout, interval, CheckFunc1, CheckFunc2, someClient.Check)
        if err != nil {
            ...
        }

        ...
    ```

    If you don't have any `Checker` functions to register or you have too many to register inline then:

    ```
        ...

        hc, err := health.Create(versionInfo criticalTimeout, interval)
        if err != nil {
            ...
        }
        if err = hc.AddCheck(CheckFunc1); err != nil {
            ...
        }
        if err = hc.AddCheck(&mongoClient.Checker); err != nil {
            ...
        }

        ...
    ```

    Or you can use any combination of the above.

4. Register the health handler:

    ```
        ...

        r := mux.NewRouter()
        r.HandleFunc("/health", hc.Handler)

        ...
    ```

5. Start the health check library:

    ```
        ...

        hc.Start(ctx)

        ...
    ```

6. Start the HTTP server:

    ```
        ...

        s := server.New(":8080", r
        if err := s.ListenAndServe(); err != nil {
            ...
        }

        ...
    ```

7. Then gracefully shutdown the health check library:

    ```
        ...

        hc.Stop()

        ...
    }
    ```

8. Set the `BuildTime`, `GitCommit` and `Version` during compile:

    Command line:

    ```
    BUILD_TIME="$(date +%s)" GIT_COMMIT="$(git rev-parse HEAD)" VERSION="$(git tag --points-at HEAD | grep ^v | head -n 1)" go build -ldflags="-X 'main.BuildTime=$BUILD_TIME' -X 'main.GitCommit=$GIT_COMMIT' -X 'main.Version=$VERSION'"
    ```

    Makefile:

    ```
    ...
    APP_NAME = app-name

    BIN_DIR    ?=.
    VERSION    ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)

    BUILD_ARCH  = $(BUILD)/$(GOOS)-$(GOARCH)
    BUILD_TIME  = $(shell date +%s)
    GIT_COMMIT  = $(shell git rev-parse HEAD)

    export GOOS   ?= $(shell go env GOOS)
    export GOARCH ?= $(shell go env GOARCH)

    ...

    build:
            @mkdir -p $(BUILD_ARCH)/$(BIN_DIR)
            go build -o $(BUILD_ARCH)/$(BIN_DIR)/$(MAIN) -ldflags="-X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)' -X 'main.Version=$(VERSION)'" cmd/$(MAIN)/main.go

    debug:
            HUMAN_LOG=1 go run -race -ldflags="-X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)' -X 'main.Version=$(VERSION)'" cmd/$(MAIN)/main.go

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
            go build -o $(BUILD_ARCH)/$(BIN_DIR)/$(APP_NAME) -ldflags "-X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT -X main.Version=$VERSION" cmd/$(APP_NAME)/main.go
    debug:
            HUMAN_LOG=1 go run -race -ldflags "-X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT -X main.Version=$VERSION" cmd/$(APP_NAME)/main.go
    ...
    ```

Implementing a checker
----------------------

Each checker measures the health of something that is required for an app to function.  This could be something internal to the app (e.g. latency, error rate, saturation, etc.) or something external (e.g. the health of an upstream app, connection to a data store, etc.).  Each checker is a function that gets the current state of whatever it is responsible for checking.

Implement a checker by creating a function (with or without a receiver) that is of the `healthcheck.Checker` type.

For example:

```

const checkName = "check name"

func Check(ctx context.Context, state *CheckState) error {
	success := rand.Float32() < 0.5
	warn := rand.Float32() < 0.5

	if success {
        state.Update(checkName, health.StatusOK, "I'm OK", 200)
	} else if warn {
        state.Update(checkName, health.StatusWarning, "degraded function of ...", 0)
	} else {
        state.Update(checkName, health.StatusWarning, "failed to ...", 503)
	}
	return nil
}
```

Note that the `statusCode` argument (last argument) to `CheckState.Update()` is only used for HTTP based checks.  If you do not have a status code then pass `0` as seen in the example above (degraded state/warning block).

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2019-2020, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
