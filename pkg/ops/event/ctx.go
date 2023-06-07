package event

import (
	"context"
)

// CtxKey 声明一个唯一key
type CtxKey string

// OpCtxKey 上下文唯一key
const OpCtxKey CtxKey = "OpCtxKey"

// AddIDToContext 添加到上下文
func AddIDToContext(ctx context.Context, s string) context.Context {
	return context.WithValue(ctx, OpCtxKey, s)
}

// GetIDFromCtx 从上下文取
func GetIDFromCtx(ctx context.Context) string {
	i := ctx.Value(OpCtxKey)
	if s, ok := i.(string); ok {
		return s
	}
	return ""
}
