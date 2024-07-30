package getter

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/krateoplatformops/eventsse/internal/store"
	"github.com/rs/zerolog"
)

const (
	defaultLimit = 50
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

// @title EventSSE API
// @version 1.0
// @description This the Krateo EventSSE server.
// @BasePath /

// Events godoc
// @Summary List all events related to a composition
// @Description list composition events
// @ID events
// @Produce  json
// @Param composition path string false "Composition Identifier"
// @Param limit query int false "Max number of events"
// @Success 200 {array} map[string]any
// @Router /events [get]
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

	sort.Slice(all, func(i, j int) bool {
		return all[i].LastTimestamp.Time.After(all[j].LastTimestamp.Time)
	})

	log.Info().
		Int("limit", limit).
		Str("key", key).Msgf("[%d] events found", len(all))

	wri.Header().Set("Access-Control-Allow-Origin", "*")
	wri.Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS")
	wri.Header().Set("Access-Control-Expose-Headers", "Authorization,Content-Type")
	wri.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type")
	wri.Header().Set("Access-Control-Allow-Credentials", "true")
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
