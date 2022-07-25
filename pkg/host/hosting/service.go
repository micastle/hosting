package hosting

import "context"

type Service interface {
	Run()
	Stop(ctx context.Context) error
}
