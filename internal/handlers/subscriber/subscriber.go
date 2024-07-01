package subscriber

import (
	"net/http"
	"time"

	"github.com/krateoplatformops/eventsse/internal/handlers"
	"github.com/krateoplatformops/eventsse/internal/httputil/decode"
	"github.com/krateoplatformops/eventsse/internal/queue"
	"github.com/rs/zerolog"
)

func Handle(broker queue.Broker, verbose bool) handlers.Handler {
	return &handler{
		verbose: verbose,
		broker:  broker,
	}
}

var _ handlers.Handler = (*handler)(nil)

type handler struct {
	verbose bool
	broker  queue.Broker
}

func (r *handler) Name() string {
	return "subscriber"
}

func (r *handler) Pattern() string {
	return "/handle"
}

func (r *handler) Methods() []string {
	return []string{http.MethodPost}
}

func (r *handler) Handler() http.HandlerFunc {
	return func(wri http.ResponseWriter, req *http.Request) {
		log := zerolog.Ctx(req.Context()).With().Logger()
		if r.verbose {
			log = log.Level(zerolog.DebugLevel)
		}

		var nfo EventInfo
		err := decode.JSONBody(wri, req, &nfo)
		if err != nil && !decode.IsEmptyBodyError(err) {
			log.Error().Msg(err.Error())
			http.Error(wri, err.Error(), http.StatusBadRequest)
			return
		}

		if r.verbose {
			log.Debug().Interface("event", nfo).Msg("Event received")
		}

		q, err := r.broker.Queue("events")
		if err != nil && !decode.IsEmptyBodyError(err) {
			log.Error().Msg(err.Error())
			http.Error(wri, err.Error(), http.StatusInternalServerError)
			return
		}

		j, err := queue.NewJob(nfo.Metadata.Name)
		if err != nil && !decode.IsEmptyBodyError(err) {
			log.Error().Msg(err.Error())
			http.Error(wri, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := j.Encode(nfo); err != nil {
			log.Error().Msg(err.Error())
			http.Error(wri, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := q.Publish(j); err != nil {
			log.Error().Msg(err.Error())
			http.Error(wri, err.Error(), http.StatusInternalServerError)
			return
		}

		wri.WriteHeader(http.StatusOK)
	}
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
