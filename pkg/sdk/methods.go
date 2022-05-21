package sdk

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	api "tinkoff-invest-bot/investapi"
)

func (s *SDK) GetShares() ([]*api.Share, string, error) {
	var header, trailer metadata.MD
	r, err := s.instruments.Shares(
		s.ctx,
		&api.InstrumentsRequest{
			InstrumentStatus: api.InstrumentStatus_INSTRUMENT_STATUS_BASE, // only base is accessible for trading via api
		},
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)

	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return nil, trackingId, extractedError
		}
		return nil, trackingId, err
	}
	return r.GetInstruments(), trackingId, nil
}

func (s *SDK) GetInstrumentByFigi(figi string) (*api.Instrument, string, error) {
	var header, trailer metadata.MD
	r, err := s.instruments.GetInstrumentBy(s.ctx, &api.InstrumentRequest{
		IdType: api.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI,
		Id:     figi,
	})

	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return nil, trackingId, extractedError
		}
		return nil, trackingId, err
	}
	return r.GetInstrument(), trackingId, nil
}

func (s *SDK) GetLastPrices(figi []string) ([]*api.LastPrice, string, error) {
	// figi it's id of share, looks like "BBG002293PJ4"
	var header, trailer metadata.MD
	r, err := s.marketData.GetLastPrices(
		s.ctx,
		&api.GetLastPricesRequest{Figi: figi},
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)
	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return nil, trackingId, extractedError
		}
		return nil, trackingId, err
	}
	return r.GetLastPrices(), trackingId, nil
}

func (s *SDK) GetLastPrice(figi string) (*api.LastPrice, string, error) {
	var header, trailer metadata.MD
	r, err := s.marketData.GetLastPrices(
		s.ctx,
		&api.GetLastPricesRequest{Figi: []string{figi}},
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)
	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return nil, trackingId, extractedError
		}
		return nil, trackingId, err
	}

	return r.GetLastPrices()[0], trackingId, nil
}

func (s *SDK) GetLastPricesAll() ([]*api.LastPrice, string, error) {
	var header, trailer metadata.MD
	r, err := s.marketData.GetLastPrices(s.ctx, &api.GetLastPricesRequest{}, grpc.Header(&header), grpc.Trailer(&trailer))
	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return nil, trackingId, extractedError
		}
		return nil, trackingId, err
	}
	return r.GetLastPrices(), trackingId, nil
}

func (s *SDK) GetCandles(figi string, from time.Time, to time.Time, interval api.CandleInterval) ([]*api.HistoricCandle, string, error) {
	var header, trailer metadata.MD
	r, err := s.marketData.GetCandles(
		s.ctx,
		&api.GetCandlesRequest{
			Figi:     figi,
			From:     timestamppb.New(from),
			To:       timestamppb.New(to),
			Interval: interval,
		},
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)

	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return nil, trackingId, extractedError
		}
		return nil, trackingId, err
	}
	return r.GetCandles(), trackingId, nil
}

func (s *SDK) GetOrderBook(figi string, depth int32) (*api.GetOrderBookResponse, string, error) {
	var header, trailer metadata.MD
	r, err := s.marketData.GetOrderBook(
		s.ctx,
		&api.GetOrderBookRequest{
			Figi:  figi,
			Depth: depth,
		},
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)

	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return nil, trackingId, extractedError
		}
		return nil, trackingId, err
	}
	return r, trackingId, nil
}

func (s *SDK) GetAccounts() ([]*api.Account, string, error) {
	var header, trailer metadata.MD
	resp, err := s.users.GetAccounts(s.ctx, &api.GetAccountsRequest{}, grpc.Header(&header), grpc.Trailer(&trailer))

	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return nil, trackingId, extractedError
		}
		return nil, trackingId, err
	}
	return resp.Accounts, trackingId, nil
}

func (s *SDK) GetMarginAttributes(accountId string) (*api.GetMarginAttributesResponse, string, error) {
	var header, trailer metadata.MD
	resp, err := s.users.GetMarginAttributes(
		s.ctx,
		&api.GetMarginAttributesRequest{
			AccountId: accountId,
		},
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)

	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return nil, trackingId, extractedError
		}
		return nil, trackingId, err
	}
	return resp, trackingId, nil
}

func (s *SDK) GetUserInfo() (*api.GetInfoResponse, string, error) {
	var header, trailer metadata.MD
	resp, err := s.users.GetInfo(
		s.ctx,
		&api.GetInfoRequest{},
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)

	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return nil, trackingId, extractedError
		}
		return nil, trackingId, err
	}
	return resp, trackingId, nil
}

func (s *SDK) GetOperations(accountId string, from time.Time, to time.Time, figi string) ([]*api.Operation, string, error) {
	var header, trailer metadata.MD

	r, err := s.operations.GetOperations(
		s.ctx,
		&api.OperationsRequest{
			AccountId: accountId,
			From:      timestamppb.New(from),
			To:        timestamppb.New(to),
			Figi:      figi,
		},
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)

	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return nil, trackingId, extractedError
		}
		return nil, trackingId, err
	}
	return r.GetOperations(), trackingId, nil
}

func (s *SDK) GetPortfolio(accountId string) (*api.PortfolioResponse, string, error) {
	var header, trailer metadata.MD

	resp, err := s.operations.GetPortfolio(
		s.ctx,
		&api.PortfolioRequest{
			AccountId: accountId,
		},
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)

	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return nil, trackingId, extractedError
		}
		return nil, trackingId, err
	}
	return resp, trackingId, nil
}

func (s *SDK) RealMarketBuy(figi string, quantity int64, accountId string, orderId string) (*api.PostOrderResponse, string, error) {
	return s.postMarketOrder(figi, quantity, api.OrderDirection_ORDER_DIRECTION_BUY, accountId, orderId)
}

func (s *SDK) RealMarketSell(figi string, quantity int64, accountId string, orderId string) (*api.PostOrderResponse, string, error) {
	return s.postMarketOrder(figi, quantity, api.OrderDirection_ORDER_DIRECTION_SELL, accountId, orderId)
}

func (s *SDK) postMarketOrder(figi string, quantity int64, direction api.OrderDirection, accountId string, orderId string) (*api.PostOrderResponse, string, error) {
	var header, trailer metadata.MD

	resp, err := s.orders.PostOrder(
		s.ctx,
		&api.PostOrderRequest{
			Figi:      figi,
			Quantity:  quantity,
			Price:     nil,
			Direction: direction,
			AccountId: accountId,
			OrderType: api.OrderType_ORDER_TYPE_MARKET,
			OrderId:   orderId,
		},
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)

	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return nil, trackingId, extractedError
		}
		return nil, trackingId, err
	}
	return resp, trackingId, nil
}
