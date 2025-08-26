package endpoint

import (
	"context"

	"github.com/MdSadiqMd/Exchange-Rate-Service/internal/domain"
	service "github.com/MdSadiqMd/Exchange-Rate-Service/internal/services"
	"github.com/go-kit/kit/endpoint"
)

type ConversionEndpoints struct {
	Convert endpoint.Endpoint
}

func MakeConversionEndpoints(svc service.ConversionService) ConversionEndpoints {
	return ConversionEndpoints{
		Convert: makeConvertEndpoint(svc),
	}
}

type convertResponse struct {
	*domain.ConversionResponse
}

func (r convertResponse) Error() error { return nil }

func makeConvertEndpoint(svc service.ConversionService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(domain.ConversionRequest)
		result, err := svc.ConvertCurrency(ctx, &req)
		if err != nil {
			return nil, err
		}

		return convertResponse{ConversionResponse: result}, nil
	}
}
