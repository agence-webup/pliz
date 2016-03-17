package domain

type ExecutionContext struct {
	Env string
}

func (ctx ExecutionContext) IsProd() bool {
	return ctx.Env == "prod"
}
