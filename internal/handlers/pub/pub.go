package pub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/krateoplatformops/eventsse/internal/labels"
	"github.com/krateoplatformops/eventsse/internal/store"
	"github.com/rs/zerolog"
	clientv3 "go.etcd.io/etcd/client/v3"
	corev1 "k8s.io/api/core/v1"
)

func SSE(cli clientv3.Watcher) http.Handler {
	return &handler{
		cli: cli,
	}
}

var _ http.Handler = (*handler)(nil)

type handler struct {
	cli clientv3.Watcher
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

	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	fmt.Fprintln(wri, "event: connection-established")
	fmt.Fprintln(wri, "id: 88888888")
	fmt.Fprintf(wri, "data: %s\n\n", `{"info": "Ready to watch events"}`)
	f.Flush()

	watchChan := r.cli.Watch(ctx, store.RootKey, clientv3.WithPrefix())
	for watchResp := range watchChan {
		for _, ev := range watchResp.Events {
			key := string(ev.Kv.Key)
			val := ev.Kv.Value
			if len(val) == 0 {
				continue
			}

			var obj corev1.Event
			if err := json.Unmarshal(val, &obj); err != nil {
				log.Warn().Str("key", key).Msgf("Decoding JSON event: %s", err.Error())
				continue
			}

			cid := labels.CompositionID(&obj)
			belongsToComposition := len(cid) > 0

			zle := log.Debug().
				Str("id", key).
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
			fmt.Fprintf(wri, "id: %s\n", key)
			fmt.Fprintf(wri, "data: %s\n\n", string(val))

			if belongsToComposition {
				fmt.Fprintf(wri, "event: %s\n", cid)
				fmt.Fprintf(wri, "id: %s\n", key)
				fmt.Fprintf(wri, "data: %s\n\n", string(val))
			}

			f.Flush()

			log.Info().Str("key", key).Msg("SSE Done")
		}
	}
}
