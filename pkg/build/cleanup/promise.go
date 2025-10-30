package cleanup

import "github.com/samber/lo"

func NewPromise() *Promise {
	return &Promise{}
}

// Promise is an abstraction for defer and compose multiple cleanup functions.
type Promise struct {
	forgotten bool
	funcs     []Func
}

func (p *Promise) Give() {
	if p.forgotten {
		return
	}
	fn := p.compose()
	fn()
}

func (p *Promise) Forget() Func {
	p.forgotten = true
	return p.compose()
}

func (p *Promise) Add(fn Func) {
	p.funcs = append(p.funcs, fn)
}

func (p *Promise) compose() Func {
	return func() {
		for _, callbackFn := range p.funcs {
			lo.Try0(callbackFn)
		}
	}
}
