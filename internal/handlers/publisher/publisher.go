package publisher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/krateoplatformops/eventsse/internal/cache"
	"github.com/krateoplatformops/eventsse/internal/labels"
	"github.com/rs/zerolog"
	corev1 "k8s.io/api/core/v1"
)

func SSE(ttlCache *cache.TTLCache[string, corev1.Event]) http.Handler {
	return &handler{
		ttlCache: ttlCache,
	}
}

var _ http.Handler = (*handler)(nil)

type handler struct {
	ttlCache *cache.TTLCache[string, corev1.Event]
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
// @Router /notifications [get]
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
		for _, k := range r.ttlCache.Keys() {
			obj, ok := r.ttlCache.Get(k)
			if !ok {
				log.Warn().Str("key", k).Msg("Event not found in cache, maybe expired?")
				continue
			}

			dat, err := json.Marshal(&obj)
			if err != nil {
				log.Error().Str("key", k).Msg("Encoding Event as JSON string")
				continue
			}

			log.Info().Str("key", k).Msg("Sending SSE")

			fmt.Fprintln(wri, "event: krateo")
			fmt.Fprintf(wri, "id: %s\n", k)
			fmt.Fprintf(wri, "data: %s\n\n", string(dat))

			cid := labels.CompositionID(&obj)
			if len(cid) > 0 {
				fmt.Fprintf(wri, "event: %s\n", cid)
				fmt.Fprintf(wri, "id: %s\n", k)
				fmt.Fprintf(wri, "data: %s\n\n", string(dat))
			}

			f.Flush()

			r.ttlCache.Remove(k)
			log.Info().Str("key", k).Msg("SSE Done")
		}
	}
}
