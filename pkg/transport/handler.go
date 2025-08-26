package transport

import (
	nethttp "net/http"

	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/endpoint"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
)

func MakeHTTPHandler(conversionEndpoints endpoint.ConversionEndpoints, logger log.Logger) nethttp.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	options := []http.ServerOption{
		http.ServerErrorEncoder(utils.EncodeError),
	}

	r.Route("/api/v1", func(r chi.Router) {
		r.Method("GET", "/convert", http.NewServer(
			conversionEndpoints.Convert,
			utils.DecodeConvertRequest,
			utils.EncodeResponse,
			options...,
		))
	})

	return r
}
