package sdk

import (
	"fmt"
	"io"
	"log"
	"time"

	"golang.org/x/xerrors"

	api "tinkoff-invest-bot/investapi"
)

type TickerPriceConsumer struct {
	ticker string
}

func NewTickerPriceConsumer(ticker string) *TickerPriceConsumer {
	return &TickerPriceConsumer{
		ticker: ticker,
	}
}

func (consumer *TickerPriceConsumer) Consume(data *api.MarketDataResponse) {
	price := data.GetLastPrice()
	if price != nil {
		fmt.Printf("price figi %s is: \n", price.GetFigi())
		PrintQuotation(price.Price)
		fmt.Printf("\n")
	}
}

func (s *SDK) SubscribeMarketData(figi string, consumer *TickerPriceConsumerInterface) error {
	consumers, contains := s.marketDataConsumers[figi]
	if !contains {
		subscribeRequest := api.MarketDataRequest{
			Payload: &api.MarketDataRequest_SubscribeLastPriceRequest{
				SubscribeLastPriceRequest: &api.SubscribeLastPriceRequest{
					SubscriptionAction: api.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE,
					Instruments: []*api.LastPriceInstrument{
						{Figi: figi},
					},
				},
			},
		}
		if err := s.marketDataStreamClient.Send(&subscribeRequest); err != nil {
			return err
		}
		s.marketDataConsumers[figi] = make([]*TickerPriceConsumerInterface, 0)
	}

	s.marketDataConsumers[figi] = append(consumers, consumer)
	return nil
}

func (s *SDK) UnsubscribeMarketData(figi string, consumer *TickerPriceConsumerInterface) error {
	consumers, contains := s.marketDataConsumers[figi]
	if !contains {
		return xerrors.Errorf("no such consumer subscribed on figi %s", figi)
	}

	for i, c := range consumers {
		if c == consumer {
			s.marketDataConsumers[figi] = append(consumers[:i], consumers[i+1:]...)
			break
		}
	}

	if len(consumers) == 0 {
		unsubscribeRequest := api.MarketDataRequest{
			Payload: &api.MarketDataRequest_SubscribeLastPriceRequest{
				SubscribeLastPriceRequest: &api.SubscribeLastPriceRequest{
					SubscriptionAction: api.SubscriptionAction_SUBSCRIPTION_ACTION_UNSUBSCRIBE,
					Instruments: []*api.LastPriceInstrument{
						{Figi: figi},
					},
				},
			},
		}
		err := s.marketDataStreamClient.Send(&unsubscribeRequest)
		if err != nil {
			return err
		}
		// add unsubscribe grpc request
		delete(s.marketDataConsumers, figi)
	}
	return nil
}

func (s *SDK) startMarketDataStream(figi string) error {
	stream, err := s.marketDataStream.MarketDataStream(s.ctx)
	if err != nil {
		return err
	}

	wait := make(chan struct{})
	r := api.MarketDataRequest{
		Payload: &api.MarketDataRequest_SubscribeLastPriceRequest{
			SubscribeLastPriceRequest: &api.SubscribeLastPriceRequest{
				SubscriptionAction: api.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE,
				Instruments: []*api.LastPriceInstrument{
					{Figi: figi},
				},
			},
		},
	}
	if err := stream.Send(&r); err != nil {
		return xerrors.Errorf("Failed to send subscribe request: %v", err)
	}

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(wait)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive message : %v", err)
			}
			payload := in.GetPayload()
			fmt.Printf("Payload: %v\n", payload)
			switch payload.(type) {
			case *api.MarketDataResponse_Ping:
				a := payload.(*api.MarketDataResponse_Ping)
				fmt.Printf("statis is %s\n", a.Ping)
			case *api.MarketDataResponse_SubscribeLastPriceResponse:
				a := payload.(*api.MarketDataResponse_SubscribeLastPriceResponse)
				fmt.Printf("SubscribeLastPriceResponse is %v\n", a.SubscribeLastPriceResponse)
			case *api.MarketDataResponse_LastPrice:
				a := payload.(*api.MarketDataResponse_LastPrice)
				PrintQuotation(a.LastPrice.Price)
			default:
				fmt.Printf("can't cast payload %v with type %T", payload, payload)
			}
		}
	}()
	time.Sleep(3 * time.Minute)
	err = stream.CloseSend()
	if err != nil {
		return err
	}
	<-wait
	return nil
}
