package engine

import (
	"fmt"

	robotProto "tinkoff-invest-bot/robot/proto"
)

type InvestRobot struct {
	config *robotProto.RobotConfig
}

func New(config *robotProto.RobotConfig) *InvestRobot {
	return &InvestRobot{
		config: config,
	}
}

func (r *InvestRobot) Run() error {
	fmt.Println("InvestRobot successfully start")
	return nil
}
