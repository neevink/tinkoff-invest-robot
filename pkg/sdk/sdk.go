package sdk

import (
	"context"
	"crypto/tls"
	"io"
	"log"

	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	api "tinkoff-invest-bot/investapi"
)

type SDK struct {
	ctx  context.Context
	conn *grpc.ClientConn

	instruments      api.InstrumentsServiceClient
	marketData       api.MarketDataServiceClient
	marketDataStream api.MarketDataStreamServiceClient
	operations       api.OperationsServiceClient
	orders           api.OrdersServiceClient
	sandbox          api.SandboxServiceClient
	stopOrders       api.StopOrdersServiceClient
	users            api.UsersServiceClient

	marketDataStreamClient api.MarketDataStreamService_MarketDataStreamClient

	marketDataConsumers map[string][]*TickerPriceConsumerInterface
}

func New(address string, token string, appName string, ctx context.Context) (*SDK, error) {
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(
			credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}),
		),
	)
	if err != nil {
		return nil, xerrors.Errorf("can't connect to gRPC server: %v", err)
	}

	marketDataStream := api.NewMarketDataStreamServiceClient(conn)
	stream, err := marketDataStream.MarketDataStream(prepareOutgoingContext(ctx, token, appName))
	if err != nil {
		return nil, xerrors.Errorf("can't careate market date stream: %v", err)
	}

	return &SDK{
		ctx:  prepareOutgoingContext(ctx, token, appName),
		conn: conn,

		instruments:      api.NewInstrumentsServiceClient(conn),
		marketData:       api.NewMarketDataServiceClient(conn),
		marketDataStream: marketDataStream,
		operations:       api.NewOperationsServiceClient(conn),
		orders:           api.NewOrdersServiceClient(conn),
		sandbox:          api.NewSandboxServiceClient(conn),
		stopOrders:       api.NewStopOrdersServiceClient(conn),
		users:            api.NewUsersServiceClient(conn),

		marketDataStreamClient: stream,

		marketDataConsumers: make(map[string][]*TickerPriceConsumerInterface, 0),
	}, nil
}

func (s *SDK) Run() {
	go func() {
		for {
			in, err := s.marketDataStreamClient.Recv()
			if err == io.EOF {
				log.Fatalf("Date stream closed")
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a message : %v", err)
			}
			for _, consumers := range s.marketDataConsumers {
				for _, c := range consumers {
					(*c).Consume(in)
				}
			}
		}
	}()
}

func (s *SDK) Shutdown() error {
	return nil
}

func prepareOutgoingContext(ctx context.Context, token string, appName string) context.Context {
	md := metadata.New(map[string]string{
		"Authorization": "Bearer " + token,
		"x-app-name":    appName,
	})
	return metadata.NewOutgoingContext(ctx, md)
}

func extractRequestError(md ...*metadata.MD) error {
	for _, m := range md {
		if errMessages, ok := (*m)["message"]; ok {
			return xerrors.Errorf("request error: %s", errMessages[0])
		}
	}
	return nil
}

func extractTrackingId(md ...*metadata.MD) string {
	for _, m := range md {
		if trackingIds, ok := (*m)["x-tracking-id"]; ok {
			return trackingIds[0]
		}
	}
	return ""
}
