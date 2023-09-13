package historyiniface

import (
	"context"

	"code.justin.tv/foundation/history.v2/historyin"
)

// API adds history events
type API interface {
	Add(context.Context, *historyin.Audit) error
}
