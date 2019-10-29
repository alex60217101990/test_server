package ticker_service

import "context"

type TickerService interface {
	Loop(ctx context.Context, secondInterval int)
	GetLatestValues() ([]string, error)
}
