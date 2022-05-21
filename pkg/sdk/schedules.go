package sdk

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	api "tinkoff-invest-bot/investapi"
)

// CanTradeNow Получает расписание торгов на указанную дату
func (s *SDK) CanTradeNow(exchange string) (bool, string, error) {
	var header, trailer metadata.MD

	now := timestamppb.New(time.Now().UTC())
	resp, err := s.instruments.TradingSchedules(
		s.ctx, &api.TradingSchedulesRequest{
			From:     now,
			To:       now,
			Exchange: exchange,
		},
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)
	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return false, trackingId, extractedError
		}
		return false, trackingId, err
	}

	if len((*resp).GetExchanges()) > 0 && len((*resp).GetExchanges()[0].GetDays()) > 0 {
		schedule := (*resp).GetExchanges()[0].GetDays()[0]
		if !schedule.GetIsTradingDay() {
			return false, trackingId, nil
		}

		startTime := schedule.GetStartTime().AsTime()
		endTime := schedule.GetEndTime().AsTime()
		nowTime := time.Now().UTC()

		if startTime.Sub(nowTime) < 0 && nowTime.Sub(endTime) < 0 {
			return true, trackingId, err
		}
		return false, trackingId, err
	}
	return false, trackingId, nil
}
