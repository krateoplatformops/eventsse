package getter

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/krateoplatformops/eventsse/internal/store"
	"github.com/rs/zerolog"
)

const (
	defaultLimit = 500
)

func Events(storage *store.Client, limit int) http.Handler {
	h := &handler{
		storage:  storage,
		maxLimit: limit,
	}

	if h.maxLimit < 0 || h.maxLimit > defaultLimit {
		h.maxLimit = defaultLimit
	}

	return h
}

var _ http.Handler = (*handler)(nil)

type handler struct {
	storage  *store.Client
	maxLimit int
}

func (r *handler) ServeHTTP(wri http.ResponseWriter, req *http.Request) {
	log := zerolog.New(os.Stdout).With().
		Str("service", "eventsse").
		Timestamp().
		Logger()

	key := r.storage.PrepareKey("", req.PathValue("composition"))

	limit := r.maxLimit
	if v := req.URL.Query().Get("limit"); len(v) > 0 {
		x, err := strconv.Atoi(v)
		if err == nil {
			limit = x
		}
	}

	max := min(r.maxLimit, defaultLimit)
	if limit < 0 || limit > max {
		limit = max
	}

	log.Info().
		Int("limit", limit).
		Str("key", key).Msg("request received")

	all, ok, err := r.storage.Get(key, store.GetOptions{
		Limit: limit,
	})
	if err != nil {
		log.Error().Msg(err.Error())
		http.Error(wri, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		log.Info().
			Int("limit", limit).
			Str("key", key).Msg("no event found")
		wri.WriteHeader(http.StatusNoContent)
		return
	}

	log.Info().
		Int("limit", limit).
		Str("key", key).Msgf("[%d] events found", len(all))

	wri.Header().Set("Content-Type", "application/json")
	wri.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(wri).Encode(all); err != nil {
		log.Error().Msg(err.Error())
		http.Error(wri, err.Error(), http.StatusInternalServerError)
		return
	}
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
