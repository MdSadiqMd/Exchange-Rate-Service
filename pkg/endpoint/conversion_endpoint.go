package endpoint

import (
	"context"

	"github.com/MdSadiqMd/Exchange-Rate-Service/internal/domain"
	service "github.com/MdSadiqMd/Exchange-Rate-Service/internal/services"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/utils"
	"github.com/go-kit/kit/endpoint"
)

type ConversionEndpoints struct {
	Convert       endpoint.Endpoint
	GetRate       endpoint.Endpoint
	Health        endpoint.Endpoint
	PrecisionInfo endpoint.Endpoint
}

func MakeConversionEndpoints(svc service.ConversionService) ConversionEndpoints {
	return ConversionEndpoints{
		Convert:       makeConvertEndpoint(svc),
		GetRate:       makeGetRateEndpoint(svc),
		Health:        makeHealthEndpoint(),
		PrecisionInfo: makePrecisionInfoEndpoint(),
	}
}

func makeConvertEndpoint(svc service.ConversionService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(domain.ConversionRequest)
		return svc.ConvertCurrency(ctx, &req)
	}
}

func makeGetRateEndpoint(svc service.ConversionService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(utils.GetRateRequest)
		rate, err := svc.GetPrecisionRate(ctx, req.From, req.To)
		if err != nil {
			return nil, err
		}
		return struct {
			Rate domain.Money `json:"rate"`
		}{Rate: rate}, nil
	}
}

func makeHealthEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		return struct{ Status string }{"ok"}, nil
	}
}

func makePrecisionInfoEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		return struct {
			Scale   int  `json:"scale"`
			Enabled bool `json:"enabled"`
		}{domain.DefaultScale, true}, nil
	}
}
