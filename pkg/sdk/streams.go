package sdk

import (
	"fmt"
	"golang.org/x/xerrors"
	"io"
	"log"
	"time"

	api "tinkoff-invest-bot/investapi"
)

func (s *SDK) GetMarketDataStream(figi string) error {
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
				// read done.
				close(wait)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			payload := in.GetPayload()
			fmt.Printf("Payload: %v\n", payload)
			//fmt.Printf("%T\n", payload)
			switch payload.(type) {
			case *api.MarketDataResponse_Ping:
				a := payload.(*api.MarketDataResponse_Ping)
				fmt.Printf("statis is %s\n", a.Ping)
			case *api.MarketDataResponse_LastPrice:
				a := payload.(*api.MarketDataResponse_LastPrice)
				PrintQuotation(a.LastPrice.Price)
			default:
				fmt.Printf("can't cast payload %v with type %T", payload, payload)
			}
			//
			//lastPrice := in.GetLastPrice()
			//fmt.Printf("Received msg: %v\n", lastPrice)
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
