package sdk

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"time"

	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

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

func New(address string, token string) (*SDK, error) {
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(
			credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}),
		),
	)
	if err != nil {
		return nil, xerrors.Errorf("can't connect to gRPC server: %v", err)
	}

	md := metadata.New(map[string]string{"Authorization": "Bearer " + token})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	marketDataStream := api.NewMarketDataStreamServiceClient(conn)
	stream, err := marketDataStream.MarketDataStream(ctx)
	if err != nil {
		return nil, xerrors.Errorf("can't careate market date stream: %v", err)
	}

	return &SDK{
		ctx:  ctx,
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
				log.Fatalf("Failed to receive a note : %v", err)
			}
			payload := in.GetPayload()
			fmt.Printf("Payload: %v\n", payload)

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

func (s *SDK) GetTradingSchedules() ([]*api.TradingSchedule, error) {
	return nil, xerrors.Errorf("not implemented error")
}

func (s *SDK) GetShares() ([]*api.Share, error) {
	request := &api.InstrumentsRequest{
		InstrumentStatus: api.InstrumentStatus_INSTRUMENT_STATUS_BASE, // only base is accessible for trading via api
	}
	r, err := s.instruments.Shares(s.ctx, request)
	if err != nil {
		return nil, err
	}
	return r.GetInstruments(), nil
}

func (s *SDK) GetInstrumentByFigi(figi string) (*api.Instrument, error) {
	request := &api.InstrumentRequest{
		IdType: api.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI,
		Id:     figi,
	}
	r, err := s.instruments.GetInstrumentBy(s.ctx, request)
	if err != nil {
		return nil, err
	}
	return r.GetInstrument(), nil
}

func (s *SDK) GetLastPrices(figi []string) ([]*api.LastPrice, error) {
	// figi it's id of share, looks like "BBG002293PJ4"
	r, err := s.marketData.GetLastPrices(s.ctx, &api.GetLastPricesRequest{Figi: figi})
	if err != nil {
		return nil, err
	}
	return r.GetLastPrices(), nil
}

func (s *SDK) GetLastPrice(figi string) (*api.LastPrice, error) {
	r, err := s.marketData.GetLastPrices(s.ctx, &api.GetLastPricesRequest{Figi: []string{figi}})
	if err != nil {
		return nil, err
	}
	return r.GetLastPrices()[0], nil
}

func (s *SDK) GetLastPricesAll() ([]*api.LastPrice, error) {
	r, err := s.marketData.GetLastPrices(s.ctx, &api.GetLastPricesRequest{})
	if err != nil {
		return nil, err
	}
	return r.GetLastPrices(), nil
}

func (s *SDK) GetCandles(figi string, from time.Time, to time.Time, interval api.CandleInterval) ([]*api.HistoricCandle, error) {
	mr := &api.GetCandlesRequest{
		Figi:     figi,
		From:     timestamppb.New(from),
		To:       timestamppb.New(to),
		Interval: interval,
	}
	r, err := s.marketData.GetCandles(s.ctx, mr)
	if err != nil {
		return nil, err
	}
	return r.GetCandles(), nil
}

func (s *SDK) GetOrderBook(figi string, depth int32) (*api.GetOrderBookResponse, error) {
	var or api.GetOrderBookRequest
	or.Figi = figi
	or.Depth = depth
	if or.Depth == 0 {
		or.Depth = 10
	}

	r, err := s.marketData.GetOrderBook(s.ctx, &or)

	if err != nil {
		return nil, err
	}

	return r, nil
}

func (s *SDK) GetAccounts() ([]*api.Account, error) {
	resp, err := s.users.GetAccounts(s.ctx, &api.GetAccountsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Accounts, nil
}

func (s *SDK) GetMarginAttributes(accountId string) (*api.GetMarginAttributesResponse, error) {
	return s.users.GetMarginAttributes(s.ctx, &api.GetMarginAttributesRequest{
		AccountId: accountId,
	})
}

func (s *SDK) GetUserInfo() (*api.GetInfoResponse, error) {
	return s.users.GetInfo(s.ctx, &api.GetInfoRequest{})
}

func (s *SDK) GetOperations(accountId string, from time.Time, to time.Time, figi string) ([]*api.Operation, error) {
	var or api.OperationsRequest
	or.AccountId = accountId

	tsFrom := timestamppb.New(from)
	tsTo := timestamppb.New(to)

	or.From = tsFrom
	or.To = tsTo
	or.Figi = figi

	r, err := s.operations.GetOperations(s.ctx, &or)
	if err != nil {
		return nil, err
	}
	return r.GetOperations(), nil
}

func (s *SDK) GetPortfolio(accountId string) (*api.PortfolioResponse, error) {
	return s.operations.GetPortfolio(s.ctx, &api.PortfolioRequest{
		AccountId: accountId,
	})
}
