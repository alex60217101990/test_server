package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/alex60217101990/test_server/internal/ticker_service"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jpillora/ipfilter"
)

type Server struct {
	ticker ticker_service.TickerService
	logger *log.Logger
	ctx    context.Context
	server *http.Server
}

func NewServer(addr string, options ...func(*Server) error) *Server {
	s := &Server{}
	for _, op := range options {
		err := op(s)
		if err != nil {
			if s.logger != nil {
				s.logger.Fatalf(
					"package: 'server', type: 'Server', method: 'NewServer', fatal: %v, stack: %s",
					err, string(debug.Stack()),
				)
			} else {
				log.Fatalf(
					"package: 'server', type: 'Server', method: 'NewServer', fatal: %v, stack: %s",
					err, string(debug.Stack()),
				)
			}
		}
	}
	s.ticker = ticker_service.NewTicker(s.logger)
	if s.server == nil {
		s.server = &http.Server{
			Addr:           addr,
			Handler:        s.GetRouter(),
			MaxHeaderBytes: 1 << 20,
			ReadTimeout:    time.Second * 7,
			WriteTimeout:   time.Second * 7,
			ErrorLog:       s.logger,
		}
	}
	return s
}

func SetLogger(logger *log.Logger) func(*Server) error {
	return func(serv *Server) error {
		serv.logger = logger
		return nil
	}
}

func SetContext(ctx context.Context) func(*Server) error {
	return func(serv *Server) error {
		serv.ctx = ctx
		return nil
	}
}

func SetHTTPServer(httpServer *http.Server) func(*Server) error {
	return func(serv *Server) error {
		serv.server = httpServer
		return nil
	}
}

func (s *Server) GetRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	// [routes]
	v1 := router.PathPrefix("/api").Subrouter()
	v1.Handle("/hash", handlers.LoggingHandler(
		os.Stdout,
		ipfilter.Wrap(
			http.HandlerFunc(s.getLastHash),
			ipfilter.Options{
				//block requests from China by IP (for secure from China DDOS)
				BlockedCountries: []string{"CN"},
			},
		),
	))
	return router
}

func (s *Server) Run() {
	s.ticker.Loop(s.ctx, 10)
	if s.logger != nil {
		s.logger.Fatalf(
			"package: 'server', type: 'Server', method: 'Run', fatal: %v, stack: %s",
			s.server.ListenAndServe(),
			string(debug.Stack()),
		)
	} else {
		log.Fatalf(
			"package: 'server', type: 'Server', method: 'Run', fatal: %v, stack: %s",
			s.server.ListenAndServe(),
			string(debug.Stack()),
		)
	}
}

func (s *Server) Close() {
	s.server.Shutdown(s.ctx)
}

func (s *Server) getLastHash(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			if s.logger != nil {
				s.logger.Printf("package: 'server', type: 'Server', method: 'getLastHash', fatal: %v, stack: %s", r, string(debug.Stack()))
			} else {
				log.Printf("package: 'server', type: 'Server', method: 'getLastHash', fatal: %v, stack: %s", r, string(debug.Stack()))
			}
			s.ErrorHandler(w, r.(error))
		}
	}()
	values, err := s.ticker.GetLatestValues()
	if err != nil {
		s.ErrorHandler(w, err)
		if s.logger != nil {
			s.logger.Printf(
				"package: 'server', type: 'Server', method: 'getLastHash', error: %v, stack: %s",
				err, string(debug.Stack()),
			)
		} else {
			log.Printf(
				"package: 'server', type: 'Server', method: 'getLastHash', error: %v, stack: %s",
				err, string(debug.Stack()),
			)
		}
		return
	}
	if values != nil {
		s.GetValuesHandler(w, values)
		return
	}
	err = fmt.Errorf("cache latest rand values is empty")
	s.ErrorHandler(w, err)
	if s.logger != nil {
		s.logger.Printf(
			"package: 'server', type: 'Server', method: 'getLastHash', error: %v, stack: %s",
			err, string(debug.Stack()),
		)
	} else {
		log.Printf(
			"package: 'server', type: 'Server', method: 'getLastHash', error: %v, stack: %s",
			err, string(debug.Stack()),
		)
	}
}

func (s *Server) ErrorHandler(w http.ResponseWriter, err error) {
	httpError := make(map[string]interface{})
	httpError["status"] = http.StatusInternalServerError
	httpError["message"] = err.Error()
	js, err := json.Marshal(httpError)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (s *Server) GetValuesHandler(w http.ResponseWriter, values []string) {
	httpResponse := make(map[string]interface{})
	httpResponse["status"] = http.StatusOK
	httpResponse["values"] = values
	js, err := json.Marshal(httpResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
