package strategies

type TradingStrategy interface {
	Name() string
	Start() error
	Stop() error
}
