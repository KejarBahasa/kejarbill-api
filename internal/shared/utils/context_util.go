package utils

import (
	"context"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/constants"
)

func GetRequestID(ctx context.Context) string {
	v, _ := ctx.Value(constants.CtxKeyRequestID).(string)
	return v
}
