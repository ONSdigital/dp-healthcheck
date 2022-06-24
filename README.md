# dp-healthcheck

A health check library for ONS Digital Publishing that implements the [Health Check Specification](https://github.com/ONSdigital/dp/blob/master/standards/HEALTH_CHECK_SPECIFICATION.md).

All Digital Publishing apps must implement a health check using this library.  Functions that implement the `Checker` type are registered with the library which will check internal and external measures of the app's health.  The library will then call these functions periodically to determine the overall health of the app and report this back using the included handler.

## Getting started

* [Add health check to an app](#adding-a-health-check-to-an-app)
* [Subscribing an app to health changes](#subscribing-an-app-to-health-changes)
* [Implementing a `Checker` function](#implementing-a-checker)

## Adding a health check to an app

1. Import the library and your HTTP server dependencies:

    ```go
    package main

    import (
        ...
        health "github.com/ONSdigital/dp-healthcheck/healthcheck"
        dphttp "github.com/ONSdigital/dp-net/v2/http"
        "github.com/gorilla/mux"
        ...
    )
    ```

2. Create a new version object defining the version of your app:

    ```go
    ...

    var BuildTime, GitCommit, Version string

    ...

    func main() {
        ctx := context.Context(context.Background())

        ...

        // Likely these values would come from your app's config
        criticalTimeout := 90 * time.Second // defined by env var HEALTHCHECK_CRITICAL_TIMEOUT
        interval := 30 * time.Second // defined by env var HEALTHCHECK_INTERVAL

        ...

        versionInfo := health.NewVersionInfo(
            BuildTime,
            GitCommit,
            Version,
        )

        ...
    ```

3. Initialise any clients that have `Checker` type functions you wish to use

4. Instantiate the health check library:

    ```go
        ...

        hc, err := health.New(versionInfo criticalTimeout, interval)
        if err != nil {
            ...
        }

        ...
    ```

5. Register your `Checker` functions providing a short human readable name for each (it is best to try to keep the name consistent between apps where possible):

    ```go
        ...

        if _, err = hc.AddCheck("check 1", CheckFunc1); err != nil {
            ...
        }
        if _, err = hc.AddCheck("mongoDB", &mongoClient.Checker); err != nil {
            ...
        }

        ...
    ```

6. Register the health handler:

    ```go
        ...

        r := mux.NewRouter()
        r.HandleFunc("/health", hc.Handler)

        ...
    ```

7. Start the health check library:

    ```go
        ...

        hc.Start(ctx)

        ...
    ```

8. Start the HTTP server:

    ```go
        ...

        s := dphttp.NewServer(":8080", r
        if err := s.ListenAndServe(); err != nil {
            ...
        }

        ...
    ```

9. Then gracefully shutdown the health check library:

    ```go
        ...

        hc.Stop()

        ...
    }
    ```

10. Set the `BuildTime`, `GitCommit` and `Version` during compile:

    Command line:

    ```sh
    BUILD_TIME="$(date +%s)" GIT_COMMIT="$(git rev-parse HEAD)" VERSION="$(git tag --points-at HEAD | grep ^v | head -n 1)" go build -ldflags="-X 'main.BuildTime=$BUILD_TIME' -X 'main.GitCommit=$GIT_COMMIT' -X 'main.Version=$VERSION'"
    ```

    Makefile:

    ```sh
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


## Subscribing an app to health changes

In step 5 of `Adding a health check to an app` you registered the checkers. Note that `AddCheck` returns a Check struct along with any error during AddCheck execution.

If your app needs ot perform any action on a state change, you can subscribe it to the changes on a combined health state of a set of Checks.

The subscriber must implement `Subscriber` interface, as defined here:

    ```go
    type Subscriber interface {
        OnHealthUpdate(status string)
    }
    ```

Example of a basic Subscriber implementation that logs the state change event:

    ```go
    type MySubscriber struct {
    }

    func (m *MySubscriber) OnHealthUpdate(status string) {
        log.Info(context.Background(), "health update", log.Data{"status": status})
    }
    ```

As it stands, dp-kafka v3 ConsumerGroup is the only library that implements the Subscriber interface.
It starts consuming messages on a healthy state and stops on an unhealthy state; hence we should be careful to subscribe a kafka consumer to only one healthcheck instance (multiple checks may be subscribed).

After calling AddCheck, you can register the returned checkers that you are interested in. You may call `Subscribe` multiple times to subscribe to more checks:

    ```go
    mySubscriber := &MySubscriber{}

    check1, err1 := hc.AddCheck("check 1", CheckFunc1)
    _, err2 := hc.AddCheck("check 2", CheckFunc2)
    hc.Subscribe(mySubscriber, check1, check2)

    if(someFlag){
        check3, err3 := hc.AddCheck("check 3", CheckFunc3)
        hc.Subscribe(mySubscriber, check3)
    }

    ```

Or you may register all the checks that have been added by calling `SubscribeAll`:

    ```go
    mySubscriber := &MySubscriber{}

    check1, err1 := hc.AddCheck("check 1", CheckFunc1)
    ...
    checkN, errN := hc.AddCheck("check N", CheckFuncN)

	hc.SubscribeAll(mySubscriber)
    ```

The `OnHealthUpdate` function will be invoked every time there is a change in any of the checkers, with the combined state of the checkers you are subscribed to as a parameter. Note that the combined state might not change from one call to another.

## Implementing a checker

Each checker measures the health of something that is required for an app to function.  This could be something internal to the app (e.g. latency, error rate, saturation, etc.) or something external (e.g. the health of an upstream app, connection to a data store, etc.).  Each checker is a function that gets the current state of whatever it is responsible for checking.

Implement a checker by creating a function (with or without a receiver) that is of the `healthcheck.Checker` type.

For example:

```go
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

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

Copyright © 2019-2021, Office for National Statistics [https://www.ons.gov.uk](https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
