package server

import (
	"net/http"
	"os"

	"github.com/cloudfoundry-incubator/receptor/api"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/rata"
)

type Server struct {
	Address string
	Logger  lager.Logger
}

func (s *Server) Run(sigChan <-chan os.Signal, readyChan chan<- struct{}) error {
	handlers := s.NewHandlers(s.Logger)
	for key, handler := range handlers {
		handlers[key] = LogWrap(handler, s.Logger)
	}

	router, err := rata.NewRouter(api.Routes, handlers)
	if err != nil {
		return err
	}

	server := ifrit.Invoke(http_server.New(s.Address, router))

	close(readyChan)

	for {
		select {
		case sig := <-sigChan:
			server.Signal(sig)
			s.Logger.Info("server.signaled-to-stop")
		case err := <-server.Wait():
			if err != nil {
				s.Logger.Error("server-failed", err)
			}

			s.Logger.Info("server.stopped")
			return err
		}
	}
}

func (s *Server) NewHandlers(logger lager.Logger) rata.Handlers {
	return rata.Handlers{
		api.CreateTask: NewCreateTaskHandler(logger),
	}
}

func LogWrap(handler http.Handler, logger lager.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestLog := logger.Session("request", lager.Data{
			"method":  r.Method,
			"request": r.URL.String(),
		})

		requestLog.Info("serving")

		handler.ServeHTTP(w, r)

		requestLog.Info("done")
	}
}

type CreateTaskHandler struct {
	logger lager.Logger
}

func NewCreateTaskHandler(logger lager.Logger) *CreateTaskHandler {
	return &CreateTaskHandler{
		logger: logger,
	}
}

func (h *CreateTaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := h.logger.Session("create-task-handler")
	w.Write(201)
	log.Info("created")
}
