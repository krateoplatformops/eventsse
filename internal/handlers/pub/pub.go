package pub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/krateoplatformops/eventsse/internal/labels"
	"github.com/krateoplatformops/eventsse/internal/store"
	"github.com/rs/zerolog"
)

func SSEx(s store.Store, limit int) http.Handler {
	return &handler{
		store: s,
		limit: limit,
	}
}

var _ http.Handler = (*handler)(nil)

type handler struct {
	store store.Store
	limit int
}

// @title EventSSE API
// @version 1.0
// @description This the Krateo EventSSE server.
// @BasePath /

// Health godoc
// @Summary SSE Endpoint
// @Description Get available events notifications
// @ID notifications
// @Produce  json
// @Success 200 {array} types.Event
// @Router /pub [get]
func (r *handler) ServeHTTP(wri http.ResponseWriter, req *http.Request) {
	log := zerolog.New(os.Stdout).With().
		Str("service", "eventsse").
		Timestamp().
		Logger()

	f, ok := wri.(http.Flusher)
	if !ok {
		msg := "http.ResponseWriter does not implement http.Flusher"
		log.Error().Msg(msg)
		http.Error(wri, msg, http.StatusInternalServerError)
		return
	}

	wri.Header().Set("Access-Control-Allow-Origin", "*")
	wri.Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS")
	wri.Header().Set("Access-Control-Expose-Headers", "Authorization,Content-Type")
	wri.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type")
	wri.Header().Set("Access-Control-Allow-Credentials", "true")

	wri.Header().Set("X-Accel-Buffering", "no")
	wri.Header().Set("Content-Type", "text/event-stream")
	wri.Header().Set("Cache-Control", "no-cache")
	wri.Header().Set("Connection", "keep-alive")

	ctx := req.Context()

	select {
	case <-ctx.Done():
		f.Flush()
		return
	default:
		keys, err := r.store.Keys(r.limit)
		if err != nil {
			log.Error().Msg("Retrieving store keys")
		} else {
			for _, k := range keys {
				all, ok, err := r.store.Get(k, store.GetOptions{Limit: 1})
				if err != nil {
					log.Err(err).Str("key", k).Msg("Unable to get Event from store")
					continue
				}
				if !ok || len(all) == 0 {
					log.Warn().Str("key", k).Msg("Event not found in store, maybe expired?")
					continue
				}

				obj := all[0]
				dat, err := json.Marshal(&obj)
				if err != nil {
					log.Err(err).Str("key", k).Msg("Encoding Event as JSON string")
					continue
				}

				cid := labels.CompositionID(&obj)
				belongsToComposition := len(cid) > 0

				zle := log.Debug().
					Str("id", k).
					Str("reason", obj.Reason).
					Str("message", obj.Message).
					Str("involvedObject.Name", obj.InvolvedObject.Name).
					Str("involvedObject.Namespace", obj.InvolvedObject.Namespace)

				if belongsToComposition {
					zle.Str("event", cid)
				} else {
					zle.Str("event", "krateo")
				}

				zle.Msg("Sending SSE")
				zle = nil

				fmt.Fprintln(wri, "event: krateo")
				fmt.Fprintf(wri, "id: %s\n", k)
				fmt.Fprintf(wri, "data: %s\n\n", string(dat))

				if belongsToComposition {
					fmt.Fprintf(wri, "event: %s\n", cid)
					fmt.Fprintf(wri, "id: %s\n", k)
					fmt.Fprintf(wri, "data: %s\n\n", string(dat))
				}

				f.Flush()

				r.store.Delete(k)
				log.Info().Str("key", k).Msg("SSE Done")
			}
		}
	}
}
