package blogweb

import (
	"context"
	"net/http"

	"github.com/blog/healthpb"
	"github.com/sirupsen/logrus"

	"github.com/go-chi/chi"
)

// Server represents the RESTful API server
type Server struct {
	logger       *logrus.Entry
	healthClient healthpb.HealthClient
}

// UseLogger applies the logger attribute
func UseLogger(logger *logrus.Entry) func(*Server) error {
	return func(s *Server) error {
		s.logger = logger
		return nil
	}
}

// UseHealthClient applies the grpc client attribute
func UseHealthClient(c healthpb.HealthClient) func(*Server) error {
	return func(s *Server) error {
		s.healthClient = c
		return nil
	}
}

// NewServer creates a REST server instance
func NewServer(opts ...func(*Server) error) (*Server, error) {
	s := &Server{}

	for _, f := range opts {
		if err := f(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// Route retruns the routing rules
func (s *Server) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/healthcheck", s.healthCheckHandler)
	return r
}

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	healthRes, err := s.healthClient.Check(context.Background(), &healthpb.HealthCheckRequest{})
	if err != nil {
		s.logger.WithError(err).Fatal("Error while calling healthcheck")
	}

	result := "UNKNOWN"
	if healthRes.GetStatus() == healthpb.HealthCheckResponse_SERVING {
		result = "SERVING"
	}

	w.Write([]byte(result))
}
