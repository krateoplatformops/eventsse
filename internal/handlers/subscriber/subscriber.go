package subscriber

import (
	"net/http"
	"os"
	"time"

	"github.com/krateoplatformops/eventsse/internal/cache"
	"github.com/krateoplatformops/eventsse/internal/httputil/decode"
	"github.com/rs/zerolog"
)

func Handle(ttlCache *cache.TTLCache[string, EventInfo]) http.Handler {
	return &handler{
		ttlCache: ttlCache,
	}
}

var _ http.Handler = (*handler)(nil)

type handler struct {
	ttlCache *cache.TTLCache[string, EventInfo]
}

func (r *handler) ServeHTTP(wri http.ResponseWriter, req *http.Request) {
	// if req.Method != http.MethodPost {
	// 	wri.Header().Set("Allow", "POST")
	// 	http.Error(wri, "405 method not allowed", http.StatusMethodNotAllowed)
	// 	return
	// }

	log := zerolog.New(os.Stdout).With().
		Str("service", "eventsse").
		Timestamp().
		Logger()

	var nfo EventInfo
	err := decode.JSONBody(wri, req, &nfo)
	if err != nil && !decode.IsEmptyBodyError(err) {
		log.Error().Msg(err.Error())
		http.Error(wri, err.Error(), http.StatusBadRequest)
		return
	}

	id := nfo.Metadata.UID
	r.ttlCache.Set(id, nfo, time.Minute*10)
	log.Info().Str("id", id).Msg("Event received")

	wri.WriteHeader(http.StatusOK)
	wri.Header().Set("Content-Type", "text/plain")
	wri.Write([]byte(id))

}

type InvolvedObject struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	UID        string `json:"uid"`
}

type Metadata struct {
	CreationTimestamp time.Time `json:"creationTimestamp"`
	Name              string    `json:"name"`
	Namespace         string    `json:"namespace"`
	UID               string    `json:"uid"`
}

type EventInfo struct {
	Type           string         `json:"type"`
	Reason         string         `json:"reason"`
	DeploymentId   string         `json:"deploymentId"`
	Time           int64          `json:"time"`
	Message        string         `json:"message"`
	Source         string         `json:"source"`
	InvolvedObject InvolvedObject `json:"involvedObject"`
	Metadata       Metadata       `json:"metadata"`
}
