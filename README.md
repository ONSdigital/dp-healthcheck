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

2. Create a new version object defining the version of your app:

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

        versionInfo := health.NewVersionInfo(
            BuildTime,
            GitCommit,
            Version,
        )

        ...
    ```

3. Initialise any clients that have `Checker` type functions you wish to use:


    ```
        ...

        // example
        cliWithChecker := s3client.NewClient("eu-west-1", "myBucket", true)

        ...
    ```

4. Instantiate the health check library:

    ```
        ...

        hc, err := health.New(versionInfo criticalTimeout, interval)
        if err != nil {
            ...
        }

        ...
    ```

5. Register your `Checker` functions providing a short human-readable name for each (it is best to try to keep the name consistent between apps where possible):

    ```
        ...

        if err = hc.AddCheck("check 1", CheckFunc1); err != nil {
            ...
        }
        if err = hc.AddCheck("mongoDB", &mongoClient.Checker); err != nil {
            ...
        }

        ...
    ```

6. Register the health handler, unless you are using the whitelist middleware described in step 8.2:

    ```
        ...

        r := mux.NewRouter()
        r.HandleFunc("/health", hc.Handler)

        ...
    ```

7. Start the health check background routine:

    ```
        ...

        hc.Start(ctx)

        ...
    ```

8. Create and start the HTTP server.

    8.1. Without whitelist (usual case):
    ```
        ...

        s := server.New(":8080", r)
                if err := s.ListenAndServe(); err != nil {
            ...
        }

        ...
    ```

    8.2. With whitelist (before applying any middleware that affects all endpoints, like auth protection):
    ```
        ...

	    middlewareChain := alice.New(middleware.Whitelist(middleware.HealthcheckFilter(hc.Handler)))

        ...

	    alice := middlewareChain.Then(r)
	    s := server.New(":8080", alice)
        if err := s.ListenAndServe(); err != nil {
            ...
        }

        ...
    ```

9. Then gracefully shutdown the health check library:

    ```
        ...

        hc.Stop()

        ...
    }
    ```

10. Set the `BuildTime`, `GitCommit` and `Version` during compile:

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

Implementing a checker
----------------------

Each checker measures the health of something that is required for an app to function.  This could be something internal to the app (e.g. latency, error rate, saturation, etc.) or something external (e.g. the health of an upstream app, connection to a data store, etc.).  Each checker is a function that gets the current state of whatever it is responsible for checking.

Implement a checker by creating a function (with or without a receiver) that is of the `healthcheck.Checker` type.

For example:

```
func Check(ctx context.Context, state *CheckState) error {
	success := rand.Float32() < 0.5
	warn := rand.Float32() < 0.5

	if success {
        state.Update(health.StatusOK, "I'm OK", 200)
	} else if warn {
        state.Update(health.StatusWarning, "degraded function of ...", 0)
	} else {
        state.Update(health.StatusWarning, "failed to ...", 503)
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
