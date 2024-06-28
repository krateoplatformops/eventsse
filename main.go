package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/krateoplatformops/eventsse/internal/env"
	"github.com/krateoplatformops/eventsse/internal/handlers"
	"github.com/krateoplatformops/eventsse/internal/handlers/health"
	"github.com/krateoplatformops/eventsse/internal/handlers/subscriber"
	"github.com/krateoplatformops/eventsse/internal/middlewares/cors"
	"github.com/krateoplatformops/eventsse/internal/queue"
	"github.com/rs/zerolog"
)

const (
	serviceName = "eventsse"
)

func main() {
	debugOn := flag.Bool("debug", env.Bool("EVENTSSE_DEBUG", false), "dump verbose output")
	dumpEnv := flag.Bool("dump-env", env.Bool("EVENTSSE_DUMP_ENV", false), "dump environment variables")
	corsOn := flag.Bool("cors", env.Bool("EVENTSSE_CORS", true), "enable or disable CORS")
	port := flag.Int("port", env.Int("EVENTSSE_PORT", 8181), "port to listen on")
	queueMaxCapacity := flag.Int("queue-max-capacity",
		env.Int("EVENTSSE_QUEUE_MAX_CAPACITY", 10), "notification queue buffer size")
	queueWorkerThreads := flag.Int("queue-worker-threads",
		env.Int("EVENTSSE_QUEUE_WORKER_THREADS", 50), "number of worker threads in the notification queue")

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Flags:")
		flag.PrintDefaults()
	}

	flag.Parse()

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
			Str("cors", fmt.Sprintf("%t", *corsOn)).
			Str("port", fmt.Sprintf("%d", *port)).
			Str("queueMaxCapacity", fmt.Sprintf("%d", *queueMaxCapacity)).
			Str("queueWorkerThreads", fmt.Sprintf("%d", *queueWorkerThreads))

		if *dumpEnv {
			evt = evt.Strs("env-vars", os.Environ())
		}

		evt.Msg("configuration and env vars")
	}

	broker, err := queue.NewBroker("memory://")
	if err != nil {
		log.Fatal().Err(err).Msgf("could not create memory broker")
	}
	defer func() {
		if err := broker.Close(); err != nil {
			log.Info().Err(err).Msgf("could not close memory broker")
		}
	}()

	healthy := int32(0)

	all := []handlers.Handler{}
	all = append(all, health.Check(&healthy, serviceName))
	all = append(all, subscriber.Handle(broker, *debugOn))

	handler := handlers.Serve(all)
	if *corsOn {
		c := cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "OPTIONS"},
			AllowedHeaders: []string{
				"Accept",
				"Authorization",
				"Content-Type",
				"X-Auth-Code",
				"X-Krateo-User",
				"X-Krateo-Groups",
			},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300,
		})

		handler = c.Handler(handler)
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      handler,
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
