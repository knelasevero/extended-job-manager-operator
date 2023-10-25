package operator

import (
	"context"

	"github.com/openshift/library-go/pkg/controller/controllercmd"
)

const (
	workQueueKey          = "key"
	workQueueCMChangedKey = "CMkey"
)

type queueItem struct {
	kind string
	name string
}

func RunOperator(ctx context.Context, cc *controllercmd.ControllerContext) error {
	return nil
}
