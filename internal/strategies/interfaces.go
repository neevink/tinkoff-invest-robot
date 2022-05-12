package strategies

type TradingStrategy interface {
	Name() string
	Start() error
	Step() error
	Stop() error
}
