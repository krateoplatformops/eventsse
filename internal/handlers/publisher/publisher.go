package publisher

import (
	"fmt"
	"net/http"

	"github.com/krateoplatformops/eventsse/internal/handlers"
	"github.com/krateoplatformops/eventsse/internal/httputil/decode"
	"github.com/krateoplatformops/eventsse/internal/queue"
	"github.com/rs/zerolog"
)

func Handle(broker queue.Broker) handlers.Handler {
	return &handler{
		verbose: true,
		broker:  broker,
	}
}

var _ handlers.Handler = (*handler)(nil)

type handler struct {
	verbose bool
	broker  queue.Broker
}

func (r *handler) Name() string {
	return "publisher"
}

func (r *handler) Pattern() string {
	return "/events"
}

func (r *handler) Methods() []string {
	return []string{http.MethodGet}
}

func (r *handler) Handler() http.HandlerFunc {
	return func(wri http.ResponseWriter, req *http.Request) {
		log := zerolog.Ctx(req.Context()).With().Logger()
		if r.verbose {
			log = log.Level(zerolog.DebugLevel)
		}

		q, err := r.broker.Queue("events")
		if err != nil && !decode.IsEmptyBodyError(err) {
			log.Error().Msg(err.Error())
			http.Error(wri, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set CORS headers to allow all origins. You may want to restrict this to specific origins in a production environment.
		wri.Header().Set("Access-Control-Allow-Origin", "*")
		wri.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		wri.Header().Set("Content-Type", "text/event-stream")
		wri.Header().Set("Cache-Control", "no-cache")
		wri.Header().Set("Connection", "keep-alive")

		ctx := req.Context()

		select {
		case <-ctx.Done():
			//someCleanup(ctx.Err())
			return
		default:
			for {
				iter, _ := q.Consume(1)
				job, err := iter.Next()
				if err != nil {
					log.Error().Msg(err.Error())
					break
				}

				if r.verbose {
					log.Debug().Str("id", job.ID).Msg("Sending SSE")
				}

				fmt.Fprintln(wri, "event: krateo")
				fmt.Fprintf(wri, "id: %s\n", job.ID)
				fmt.Fprintf(wri, "data: %s\n\n", string(job.Raw))

				wri.(http.Flusher).Flush()
			}

		}
	}
}
