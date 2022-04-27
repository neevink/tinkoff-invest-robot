package main

import (
	"fmt"

	"tinkoff-invest-bot/robot/pkg/engine"

	robotProto "tinkoff-invest-bot/robot/proto"
)

func main() {
	conf := &robotProto.RobotConfig{TinkoffApiEndpoint: ""}

	robotInstance := engine.New(conf)
	if err := robotInstance.Run(); err != nil {
		fmt.Println()
	}

}
