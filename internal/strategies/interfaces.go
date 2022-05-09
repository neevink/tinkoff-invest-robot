package strategies

type TradingStrategy interface {
	Start() error
	Step() error
	Stop() error
}
