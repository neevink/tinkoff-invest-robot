package sdk

import (
	"context"
	"crypto/tls"
	"time"

	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	api "tinkoff-invest-bot/investapi"
)

type sdk struct {
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
}

func New(address string, token string) (*sdk, error) {
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

	return &sdk{
		ctx:  ctx,
		conn: conn,

		instruments:      api.NewInstrumentsServiceClient(conn),
		marketData:       api.NewMarketDataServiceClient(conn),
		marketDataStream: api.NewMarketDataStreamServiceClient(conn),
		operations:       api.NewOperationsServiceClient(conn),
		orders:           api.NewOrdersServiceClient(conn),
		sandbox:          api.NewSandboxServiceClient(conn),
		stopOrders:       api.NewStopOrdersServiceClient(conn),
		users:            api.NewUsersServiceClient(conn),
	}, nil
}

func (s *sdk) GetTradingSchedules() ([]*api.TradingSchedule, error) {
	return nil, xerrors.Errorf("not implemented error")

	//request := &api.TradingSchedulesRequest{
	//	// todo add paramethers here
	//}
	//r, err := s.instruments.TradingSchedules(s.ctx, request)
	//if err != nil {
	//	return nil, err
	//}
	//return r.GetExchanges(), nil
}

func (s *sdk) GetShares() ([]*api.Share, error) {
	request := &api.InstrumentsRequest{
		InstrumentStatus: api.InstrumentStatus_INSTRUMENT_STATUS_BASE, // only base is accessible for trading via api
	}
	r, err := s.instruments.Shares(s.ctx, request)
	if err != nil {
		return nil, err
	}
	return r.GetInstruments(), nil
}

func (s *sdk) GetLastPrices(figi []string) ([]*api.LastPrice, error) {
	// figi it's id of share, looks like "BBG002293PJ4"
	r, err := s.marketData.GetLastPrices(s.ctx, &api.GetLastPricesRequest{Figi: figi})
	if err != nil {
		return nil, err
	}
	return r.GetLastPrices(), nil
}

func (s *sdk) GetLastPricesAll() ([]*api.LastPrice, error) {
	r, err := s.marketData.GetLastPrices(s.ctx, &api.GetLastPricesRequest{})
	if err != nil {
		return nil, err
	}
	return r.GetLastPrices(), nil
}

func (s *sdk) GetCandles(figi string, from time.Time, to time.Time, interval api.CandleInterval) ([]*api.HistoricCandle, error) {
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

func (s *sdk) GetOrderBook(figi string, depth int32) (*api.GetOrderBookResponse, error) {
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

func (s *sdk) GetAccounts() ([]*api.Account, error) {
	resp, err := s.users.GetAccounts(s.ctx, &api.GetAccountsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Accounts, nil
}

func (s *sdk) GetMarginAttributes(accountId string) (*api.GetMarginAttributesResponse, error) {
	return s.users.GetMarginAttributes(s.ctx, &api.GetMarginAttributesRequest{
		AccountId: accountId,
	})
}

func (s *sdk) GetUserInfo() (*api.GetInfoResponse, error) {
	return s.users.GetInfo(s.ctx, &api.GetInfoRequest{})
}

func (s *sdk) GetOperations(accountId string, from time.Time, to time.Time, figi string) ([]*api.Operation, error) {
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

func (s *sdk) GetPortfolio(accountId string) (*api.PortfolioResponse, error) {
	return s.operations.GetPortfolio(s.ctx, &api.PortfolioRequest{
		AccountId: accountId,
	})
}
