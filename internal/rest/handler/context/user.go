package context

type CtxKey string

const UserCtxKey = CtxKey("user_ctx")

type UserCtx struct {
	ID string
}
