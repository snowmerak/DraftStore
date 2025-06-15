package webapi

import (
	"github.com/go-chi/chi/v5"
	"github.com/snowmerak/DraftStore/lib/controller/webapi/handler"
	"github.com/snowmerak/DraftStore/lib/service/draft"
	"github.com/snowmerak/DraftStore/lib/util/logger"
)

type Server struct {
	router       chi.Router
	address      string
	draftHandler *handler.DraftHandler
}

type ServerOptions struct {
	Router       chi.Router
	Address      string
	DraftService *draft.Service
}

func NewServer(option ServerOptions) *Server {
	log := logger.GetServiceLogger("webapi-controller")

	draftHandler := handler.NewDraftHandler(option.DraftService)

	// Register routes
	draftHandler.RegisterRoutes(option.Router)

	server := &Server{
		router:       option.Router,
		address:      option.Address,
		draftHandler: draftHandler,
	}

	log.Info().
		Str("address", server.address).
		Msg("WebAPI server controller initialized")

	return server
}

// GetRouter returns the configured router
func (s *Server) GetRouter() chi.Router {
	return s.router
}

// GetAddress returns the server address
func (s *Server) GetAddress() string {
	return s.address
}
