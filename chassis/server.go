package chassis

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"

	"github.com/veganbase/backend/chassis/pubsub"
)

// Server represents the shared functionality used in all backend REST
// servers.
type Server struct {
	AppName  string
	Ctx      context.Context
	Srv      *http.Server
	PubSub   pubsub.PubSub
	muAtExit sync.Mutex
	atExit   []func()
}

// Init initialises all the common infrastructure used by REST
// servers.
func (s *Server) Init(appname string, project string, port int, credentials string, r chi.Router) {
	// Randomise ID generation.
	rand.Seed(int64(time.Now().Nanosecond()))

	s.AppName = appname
	s.Ctx = context.Background()
	s.atExit = []func(){}

	// StackDriver export.
	// TODO: THIS SEEMS TO SET UP DISTRIBUTED TRACING AND METRIC EXPORT
	// VIA StackDriver. HOW DOES LOG AGGREGATION FROM STDOUT WORK WITH
	// THIS?
	// if project != "dev" {
	// 	se, err := stackdriver.NewExporter(stackdriver.Options{
	// 		ProjectID: project,
	// 	})
	// 	if err != nil {
	// 		log.Fatal().Err(err).Msg("failed to create Stackdriver exporter")
	// 	}

	// 	// TODO: THIS IS FOR DISTRIBUTED TRACING SUPPORT, BUT IT DOESN'T
	// 	// DO ANYTHING UNLESS YOU SET UP SPANS. IS THERE A MIDDLEWARE TO
	// 	// DO THIS FOR EACH REQUEST? HOW DOES IT INTERACT WITH THE
	// 	// Request-Id CORRELATION ID SETUP? WHERE SHOULD WE SET UP MORE
	// 	// SPANS?
	// 	trace.RegisterExporter(se)

	// 	// TODO: WE SEEM TO EXPORT STATISTICS BOTH VIA StackDriver AND VIA
	// 	// Prometheus (SEE BELOW). WHAT'S THE RIGHT WAY TO DO THIS?
	// 	view.RegisterExporter(se)
	// }

	// TODO: ADD APPLICATION-SPECIFIC METRICS.
	// pe, err := prometheus.NewExporter(prometheus.Options{
	// 	Namespace: appname,
	// })
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("failed to create Prometheus exporter")
	// }
	// TODO: WHAT ARE THESE THINGS ACTUALLY DOING?
	// view.RegisterExporter(pe)
	// trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	// TODO: WE SEEM TO HAVE A REST METRICS ENDPOINT AS WELL AS
	// EXPORTING VIA StackDriver. NOT SURE THAT'S RIGHT.
	// r.Handle("/metrics", pe)

	// ADD OpenCensus DEFAULT VIEWS FOR A SERVICE. (CHECK WHAT THESE
	// ACTUALLY ARE...)
	// if err := view.Register(ochttp.DefaultServerViews...); err != nil {
	// 	log.Fatal().Msg("failed to register ochttp.DefaultServerViews")
	// }
	// h := &ochttp.Handler{Handler: r}
	s.Srv = &http.Server{
		// Handler: h,
		Handler: r,
		Addr:    fmt.Sprintf(":%d", port),
	}

	// Initialise Pub/Sub.
	if credentials != "dev" {
		var err error
		s.PubSub, err = pubsub.NewGoogleClient(s.Ctx, project, credentials)
		if err != nil {
			log.Fatal().Err(err).Msg("couldn't connect to Google Pub/Sub service")
		}
	} else {
		s.PubSub = pubsub.NewMockPubSub(map[string][][]byte{})
	}

	// Make sure pub/sub gets cleaned up on exit: this removes unique
	// subscriptions added for fan-out processing.
	s.AddAtExit(s.PubSub.Close)
}

// InitSimple initialises all the common infrastructure used by
// servers that only provide healthcheck REST endpoints.
func (s *Server) InitSimple(appname, project string, port int, credentials string) {
	s.Init(appname, project, port, credentials, s.healthRoutes())
}

// Healthcheck endpoints for simple servers.
func (s *Server) healthRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", Health)
	r.Get("/healthz", Health)
	return r
}

// Serve runs a server event loop.
func (s *Server) Serve() {
	errChan := make(chan error, 0)
	go func() {
		log.Info().
			Str("address", s.Srv.Addr).
			Msg("server started")
		err := s.Srv.ListenAndServe()
		if err != nil {
			errChan <- err
		}
	}()

	signalCh := make(chan os.Signal, 0)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	var err error

	select {
	case <-signalCh:
	case err = <-errChan:
	}

	s.shutdown()
	s.runAtShutdown()

	if err == nil {
		log.Info().Msg("server shutting down")
	} else {
		log.Fatal().Err(err).Msg("server failed")
	}
}

// AddAtExit adds an exit handler function.
func (s *Server) AddAtExit(fn func()) {
	s.muAtExit.Lock()
	defer s.muAtExit.Unlock()
	s.atExit = append(s.atExit, fn)
}

// Shut down server.
func (s *Server) shutdown() {
	ctx, cancel := context.WithTimeout(s.Ctx, 10*time.Second)
	defer cancel()
	if err := s.Srv.Shutdown(ctx); err != nil {
		log.Error().Err(err)
	}
}

// Run at-exit processing.
func (s *Server) runAtShutdown() {
	s.muAtExit.Lock()
	defer s.muAtExit.Unlock()
	for _, fn := range s.atExit {
		fn()
	}
}

// SimpleHandlerFunc is a HTTP handler function that signals internal
// errors by returning a normal Go error, and when successful returns
// a response body to be marshalled to JSON. It can be wrapped in the
// SimpleHandler middleware to produce a normal HTTP handler function.
type SimpleHandlerFunc func(w http.ResponseWriter, r *http.Request) (interface{}, error)

// SimpleHandler wraps a simpleHandler-style HTTP handler function to
// turn it into a normal HTTP handler function. Go errors from the
// inner handler are returned to the caller as "500 Internal Server
// Error" responses. Returns from successful processing by the inner
// handler and marshalled into a JSON response body.
func SimpleHandler(inner SimpleHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Run internal handler: returns a marshalable result and an
		// error, either of which may be nil.
		result, err := inner(w, r)

		// Propagate Go errors as "500 Internal Server Error" responses.
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("handling %q: %v", r.RequestURI, err)
			return
		}

		// No response body, so internal handler dealt with response
		// setup.
		if result == nil {
			return
		}

		// Marshal JSON response body.
		body, err := json.Marshal(result)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("handling %q: %v", r.RequestURI, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}
}
