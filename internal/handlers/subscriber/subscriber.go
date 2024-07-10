package subscriber

import (
	"net/http"
	"os"
	"path"
	"time"

	"github.com/krateoplatformops/eventsse/internal/cache"
	"github.com/krateoplatformops/eventsse/internal/httputil/decode"
	"github.com/krateoplatformops/eventsse/internal/labels"
	"github.com/krateoplatformops/eventsse/internal/store"
	"github.com/rs/zerolog"

	corev1 "k8s.io/api/core/v1"
)

type HandleOptions struct {
	TTLCache    *cache.TTLCache[string, corev1.Event]
	StoreClient *store.Client
}

func Handle(opts HandleOptions) http.Handler {
	return &handler{
		ttlCache:    opts.TTLCache,
		storeClient: opts.StoreClient,
	}
}

var _ http.Handler = (*handler)(nil)

type handler struct {
	ttlCache    *cache.TTLCache[string, corev1.Event]
	storeClient *store.Client
}

func (r *handler) ServeHTTP(wri http.ResponseWriter, req *http.Request) {
	log := zerolog.New(os.Stdout).With().
		Str("service", "eventsse").
		Timestamp().
		Logger()

	var nfo corev1.Event
	err := decode.JSONBody(wri, req, &nfo)
	if err != nil && !decode.IsEmptyBodyError(err) {
		log.Error().Msg(err.Error())
		http.Error(wri, err.Error(), http.StatusBadRequest)
		return
	}

	key := nfo.FirstTimestamp.Format("2006010215")
	if val, ok := labels.CompositionID(&nfo); ok {
		key = path.Join(key, val)
	}
	key = path.Join(key, string(nfo.UID))

	log.Info().Str("key", key).Msg("Event received")

	if err := r.storeClient.Set(key, &nfo); err != nil {
		log.Error().Msg(err.Error())
		http.Error(wri, err.Error(), http.StatusInternalServerError)
		return
	}

	r.ttlCache.Set(key, nfo, time.Minute*2)
	log.Info().Str("key", key).Msg("Event stored")

	wri.WriteHeader(http.StatusOK)
	wri.Header().Set("Content-Type", "text/plain")
	wri.Write([]byte(key))
}
