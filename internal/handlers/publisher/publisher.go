package publisher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/krateoplatformops/eventsse/internal/cache"
	"github.com/krateoplatformops/eventsse/internal/handlers/subscriber"
	"github.com/rs/zerolog"
)

func SSE(ttlCache *cache.TTLCache[string, subscriber.EventInfo]) http.Handler {
	return &handler{
		ttlCache: ttlCache,
	}
}

var _ http.Handler = (*handler)(nil)

type handler struct {
	ttlCache *cache.TTLCache[string, subscriber.EventInfo]
}

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
		//for {
		for _, k := range r.ttlCache.Keys() {
			obj, ok := r.ttlCache.Get(k)
			if !ok {
				log.Warn().Str("id", k).Msg("Event not found in cache, maybe expired?")
				continue
			}

			dat, err := json.Marshal(&obj)
			if err != nil {
				log.Error().Str("id", k).Msg("Encoding Event as JSON string")
				continue
			}

			log.Info().Str("id", k).Msg("Sending SSE")

			fmt.Fprintln(wri, "event: krateo")
			//fmt.Fprintf(wri, "id: %s\n", k)
			fmt.Fprintf(wri, "data: %s\n\n", string(dat))
			f.Flush()

			r.ttlCache.Remove(k)
			log.Info().Str("id", k).Msg("SSE Done")
		}
		//}
	}
}
