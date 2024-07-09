package getter

import (
	"encoding/json"
	"net/http"
	"os"
	"path"

	"github.com/krateoplatformops/eventsse/internal/store"
	"github.com/rs/zerolog"
)

func Events(storage *store.Client) http.Handler {
	return &handler{
		storage: storage,
	}
}

var _ http.Handler = (*handler)(nil)

type handler struct {
	storage *store.Client
}

func (r *handler) ServeHTTP(wri http.ResponseWriter, req *http.Request) {
	log := zerolog.New(os.Stdout).With().
		Str("service", "eventsse").
		Timestamp().
		Logger()

	key := req.PathValue("date")
	if val := req.PathValue("composition"); len(val) > 0 {
		key = path.Join(key, val)
	}
	if val := req.PathValue("event"); len(val) > 0 {
		key = path.Join(key, val)
	}
	log.Info().Str("key", key).Msg("request received")

	all, ok, err := r.storage.Get(key)
	if err != nil {
		log.Error().Msg(err.Error())
		http.Error(wri, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		log.Info().Str("key", key).Msg("no event found")
		wri.WriteHeader(http.StatusNoContent)
		return
	}

	log.Info().Str("key", key).Msgf("[%d] events found", len(all))

	wri.Header().Set("Content-Type", "application/json")
	wri.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(wri).Encode(all); err != nil {
		log.Error().Msg(err.Error())
		http.Error(wri, err.Error(), http.StatusInternalServerError)
		return
	}
}
