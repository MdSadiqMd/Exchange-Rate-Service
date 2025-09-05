package transport

import (
	nethttp "net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"

	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/endpoint"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/utils"
)

func MakeHTTPHandler(e endpoint.ConversionEndpoints, logger log.Logger) nethttp.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(utils.EncodeError),
	}

	r.Route("/api/v2", func(r chi.Router) {
		r.Method(
			"GET",
			"/health",
			kithttp.NewServer(
				e.Health,
				utils.DecodeEmptyRequest,
				utils.EncodeResponse,
				opts...,
			),
		)
		r.Method(
			"GET",
			"/precision",
			kithttp.NewServer(
				e.PrecisionInfo,
				utils.DecodeEmptyRequest,
				utils.EncodeResponse,
				opts...,
			),
		)
		r.Method(
			"GET",
			"/rates/{from}/{to}",
			kithttp.NewServer(
				e.GetRate,
				utils.DecodeGetRateRequest,
				utils.EncodeResponse,
				opts...,
			),
		)
		r.Method(
			"POST",
			"/convert",
			kithttp.NewServer(
				e.Convert,
				utils.DecodeConvertRequest,
				utils.EncodeResponse,
				opts...,
			),
		)
	})

	return r
}
