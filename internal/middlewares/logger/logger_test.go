package logger

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func dummyHandler(w http.ResponseWriter, r *http.Request) {
	// Recupera il logger dal contesto
	logger := zerolog.Ctx(r.Context())
	if logger == nil {
		http.Error(w, "Logger not found in context", http.StatusInternalServerError)
		return
	}

	// Logga un messaggio per verificare che il logger funzioni
	logger.Info().Msg("Logging from DummyHandler")
	w.WriteHeader(http.StatusOK)
}

func TestLoggerMiddleware(t *testing.T) {
	// Crea un logger zerolog per test con output su /dev/null
	logger := zerolog.New(zerolog.ConsoleWriter{Out: io.Discard}).With().Timestamp().Logger()

	// Wrappa il dummy handler con il middleware Logger
	handler := Logger(logger)(http.HandlerFunc(dummyHandler))

	// Crea una richiesta HTTP di esempio
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	// Registra la risposta
	rr := httptest.NewRecorder()

	// Serve la richiesta utilizzando il middleware e l'handler
	handler.ServeHTTP(rr, req)

	// Verifica che lo stato della risposta sia 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Puoi aggiungere ulteriori verifiche se hai un mock o un sistema per catturare i log
}
