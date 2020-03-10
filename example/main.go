package main

// usage: to run this, open terminal in example folder where this file resides and use the following :
// HUMAN_LOG=1 go run -race -ldflags="-X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)' -X 'main.Version=$(VERSION)'" main.go
//
// for fuller description, see the packages README.md
//

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	// 1. Import the library and your HTTP server dependencies:
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-hierarchy-api/config"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

// 2. Create a new version object defining the version of your app:

var buildTime, gitCommit, version string

func main() {

	ctx, cancelHealthChecks := context.WithCancel(context.Background())

	// Likely these values would come from your app's config
	criticalTimeout := time.Minute / 4
	interval := 5 * time.Second

	buildTime = "5000"
	gitCommit = "1234567"
	version = "1.1"

	versionInfo, err := healthcheck.NewVersionInfo(
		buildTime,
		gitCommit,
		version,
	)
	if err != nil {
		log.ErrorC("failed to init health check info", err, nil)
	}

	config, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	config.ShutdownTimeout /= 5 // shorten for demo purposes

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// 2. Initialise any clients that have `Checker` type functions you wish to use
	// ??? need example here ...

	// 3. Instantiate the health check library:
	hc := healthcheck.New(versionInfo, criticalTimeout, interval)

	// 4. Register your `Checker` functions providing a short human readable name for each (it is best to try to keep the name consistent between apps where possible):

	if err = hc.AddCheck("check 1", dummyCheck1); err != nil {
		log.ErrorC("failed to add function to Check1", err, nil)
	}
	if err = hc.AddCheck("check 2", dummyCheck2); err != nil {
		log.ErrorC("failed to add function to Check2", err, nil)
	}

	// 5. Register the health handler:

	router := mux.NewRouter()
	router.HandleFunc("/health", hc.Handler)

	// 6. Start the health check library:

	hc.Start(ctx)

	// 7. Start the HTTP server:

	srv := server.New(":8080", router)
	srv.HandleOSSignals = false

	httpServerDoneChan := make(chan error)
	go func() {
		log.Debug("starting http server", log.Data{"bind_addr": config.BindAddr})
		if err := srv.ListenAndServe(); err != nil {
			log.ErrorC("server start problem", err, nil)
		}
		close(httpServerDoneChan)
	}()

	// wait (indefinitely) for an exit event (either an OS signal or the httpServerDoneChan)
	// set `err` and logData
	wantHTTPShutdown := true
	logData := log.Data{}
	select {
	case sig := <-signals:
		err = errors.New("aborting after signal")
		logData["signal"] = sig.String()
	case err = <-httpServerDoneChan:
		wantHTTPShutdown = false
	}

	// gracefully shutdown the application, closing any open resources
	logData["timeout"] = config.ShutdownTimeout
	log.ErrorC("Start shutdown", err, logData)
	shutdownContext, shutdownContextCancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)

	// 8. Then gracefully shutdown the health check library:

	//hc.Stop()
	log.Debug("Canceling healthchecks", nil)
	cancelHealthChecks()
	log.Debug("Canceled healthchecks", nil)

	go func() {
		if wantHTTPShutdown {
			if err := srv.Shutdown(shutdownContext); err != nil {
				log.ErrorC("error closing http server", err, nil)
			} else {
				log.Trace("http server shutdown", nil)
			}
		}

		shutdownContextCancel()
	}()
	log.Debug("awaiting context's Done", nil) // this message can appear before, between or maybe after the debug output
	// of "ticker.go : ctx.Done() and ticker stopped" in ticker.go

	// wait for timeout or success (cancel)
	<-shutdownContext.Done()

	<-ctx.Done()

	log.Debug("got context's Done", nil)

	fmt.Printf("No more health checks after this ... waiting %v\n", criticalTimeout+interval+interval)
	time.Sleep(criticalTimeout + interval + interval)

	log.Info("Shutdown done", log.Data{"context": shutdownContext.Err()})
	os.Exit(1)
}

func dummyCheck1(ctx context.Context, state *healthcheck.CheckState) error {
	success := rand.Float32() < 0.5
	warn := rand.Float32() < 0.5

	if success {
		state.Update(healthcheck.StatusOK, "1: I'm OK", 200)
		log.Debug("1: I'm OK", nil)
	} else if warn {
		state.Update(healthcheck.StatusWarning, "1: degraded function of ...", 0)
		log.Debug("1: degraded function of ...", nil)
	} else {
		state.Update(healthcheck.StatusWarning, "1: failed to ...", 503)
		log.Debug("1: failed to ...", nil)
	}
	return nil
}

func dummyCheck2(ctx context.Context, state *healthcheck.CheckState) error {
	success := rand.Float32() < 0.5
	warn := rand.Float32() < 0.5

	if success {
		state.Update(healthcheck.StatusOK, "2: I'm OK", 200)
		log.Debug("2: I'm OK", nil)
	} else if warn {
		state.Update(healthcheck.StatusWarning, "2: degraded function of ...", 0)
		log.Debug("2: degraded function of ...", nil)
	} else {
		state.Update(healthcheck.StatusWarning, "2: failed to ...", 503)
		log.Debug("2: failed to ...", nil)
	}
	return nil
}
