package main

import (
	"context"
	"fmt"

	"tinkoff-invest-bot/robot/pkg/engine"
	robotProto "tinkoff-invest-bot/robot/proto"
)

func main() {
	conf := &robotProto.RobotConfig{
		TinkoffApiEndpoint: "invest-public-api.tinkoff.ru:443",
		AccessToken:        "#",
	}

	ctx := context.Background()

	robotInstance := engine.New(conf)
	if err := robotInstance.Run(ctx); err != nil {
		fmt.Println(err)
	}

}
