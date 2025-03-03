package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/krateoplatformops/eventsse/internal/env"
	"github.com/krateoplatformops/eventsse/internal/handlers/getter"
	"github.com/krateoplatformops/eventsse/internal/handlers/health"
	"github.com/krateoplatformops/eventsse/internal/handlers/pub"
	"github.com/krateoplatformops/eventsse/internal/handlers/sub"
	"github.com/krateoplatformops/eventsse/internal/store"
	"github.com/rs/zerolog"

	_ "github.com/krateoplatformops/eventsse/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	serviceName    = "eventsse"
	defaultLimit   = 100
	fifoMultiplier = 10
)

func main() {
	debugOn := flag.Bool("debug", env.Bool("EVENTSSE_DEBUG", true), "dump verbose output")
	dumpEnv := flag.Bool("dump-env", env.Bool("EVENTSSE_DUMP_ENV", false), "dump environment variables")
	port := flag.Int("port", env.Int("EVENTSSE_PORT", 8181), "port to listen on")
	ttlSecs := flag.Int("ttl", env.Int("EVENTSSE_TTL", 120), "stored event exipre time in seconds")
	limit := flag.Int("limit", env.Int("EVENTSSE_GET_LIMIT", defaultLimit),
		"limits the number of results to return from 'Get' request")
	endpoints := flag.String("etcd-servers", env.String("EVENTSSE_ETCD_SERVERS", "localhost:2379"), "etcd endpoints")

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Flags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *limit <= 0 {
		*limit = defaultLimit
	}

	// Initialize the logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Default level for this log is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debugOn {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log := zerolog.New(os.Stdout).With().
		Str("service", serviceName).
		Timestamp().
		Logger()

	if log.Debug().Enabled() {
		evt := log.Debug().
			Str("debug", fmt.Sprintf("%t", *debugOn)).
			Str("port", fmt.Sprintf("%d", *port)).
			Str("ttl", fmt.Sprintf("%d", *ttlSecs)).
			Str("limit", fmt.Sprintf("%d", *limit)).
			Str("etcd-endpoints", *endpoints)

		if *dumpEnv {
			evt = evt.Strs("env-vars", os.Environ())
		}

		evt.Msg("configuration and env vars")
	}

	opts := store.Options{
		Endpoints: strings.Split(*endpoints, ","),
	}
	storage, err := store.NewClient(opts)
	if err != nil {
		log.Fatal().Err(err).Msg("could not create ETCD client")
	}
	defer storage.Close()

	if *ttlSecs <= 0 {
		*ttlSecs = 180
	}
	storage.SetTTL(*ttlSecs)

	watcher, err := store.NewWatcher(opts)
	if err != nil {
		log.Fatal().Err(err).Msg("could not create ETCD watcher")
	}

	mux := http.NewServeMux()

	healthy := int32(0)

	mux.Handle("GET /health", health.Check(&healthy, serviceName))
	mux.Handle("POST /handle", sub.Handle(sub.HandleOptions{
		Store: storage,
		TTL:   time.Duration(*ttlSecs) * time.Second,
	}))
	mux.Handle("GET /notifications", pub.SSE(watcher))
	mux.Handle("GET /events", getter.Events(storage, *limit))
	mux.Handle("GET /events/{composition}", getter.Events(storage, *limit))
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 50 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), []os.Signal{
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	}...)
	defer stop()

	go func() {
		atomic.StoreInt32(&healthy, 1)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msgf("could not listen on %s", server.Addr)
		}
	}()

	// Listen for the interrupt signal.
	log.Info().Msgf("server is ready to handle requests at @ %s", server.Addr)
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Info().Msg("server is shutting down gracefully, press Ctrl+C again to force")
	atomic.StoreInt32(&healthy, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server gracefully stopped")
}
