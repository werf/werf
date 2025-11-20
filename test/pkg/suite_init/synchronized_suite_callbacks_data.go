package suite_init

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
)

type SynchronizedSuiteCallbacksData struct {
	synchronizedBeforeSuiteNode1Funcs               []func(context.Context)
	synchronizedBeforeSuiteNode1FuncWithReturnValue func() []byte
	synchronizedBeforeSuiteAllNodesFuncs            []func(context.Context, []byte)

	synchronizedAfterSuiteAllNodesFuncs []func(context.Context)
	synchronizedAfterSuiteNode1Funcs    []func(context.Context)
}

func NewSynchronizedSuiteCallbacksData() *SynchronizedSuiteCallbacksData {
	data := &SynchronizedSuiteCallbacksData{}

	SynchronizedBeforeSuite(func(ctx SpecContext) []byte {
		for _, f := range data.synchronizedBeforeSuiteNode1Funcs {
			f(ctx)
		}
		if data.synchronizedBeforeSuiteNode1FuncWithReturnValue != nil {
			return data.synchronizedBeforeSuiteNode1FuncWithReturnValue()
		} else {
			return nil
		}
	}, func(ctx SpecContext, d []byte) {
		for _, f := range data.synchronizedBeforeSuiteAllNodesFuncs {
			f(ctx, d)
		}
	})

	SynchronizedAfterSuite(func(ctx SpecContext) {
		for _, f := range data.synchronizedAfterSuiteAllNodesFuncs {
			f(ctx)
		}
	}, func(ctx SpecContext) {
		for _, f := range data.synchronizedAfterSuiteNode1Funcs {
			f(ctx)
		}
	})

	return data
}

func (data *SynchronizedSuiteCallbacksData) AppendSynchronizedBeforeSuiteNode1Func(f func(context.Context)) bool {
	data.synchronizedBeforeSuiteNode1Funcs = append(data.synchronizedBeforeSuiteNode1Funcs, f)
	return true
}

func (data *SynchronizedSuiteCallbacksData) AppendSynchronizedBeforeSuiteAllNodesFunc(f func(context.Context, []byte)) bool {
	data.synchronizedBeforeSuiteAllNodesFuncs = append(data.synchronizedBeforeSuiteAllNodesFuncs, f)
	return true
}

func (data *SynchronizedSuiteCallbacksData) SetSynchronizedBeforeSuiteNode1FuncWithReturnValue(f func() []byte) bool {
	if data.synchronizedBeforeSuiteNode1FuncWithReturnValue != nil {
		panic("You may only call SetSynchronizedBeforeSuiteNode1FuncWithReturnValue once!")
	}
	data.synchronizedBeforeSuiteNode1FuncWithReturnValue = f
	return true
}

func (data *SynchronizedSuiteCallbacksData) AppendSynchronizedAfterSuiteAllNodesFunc(f func(context.Context)) bool {
	data.synchronizedAfterSuiteAllNodesFuncs = append(data.synchronizedAfterSuiteAllNodesFuncs, f)
	return true
}

func (data *SynchronizedSuiteCallbacksData) AppendSynchronizedAfterSuiteNode1Func(f func(context.Context)) bool {
	data.synchronizedAfterSuiteNode1Funcs = append(data.synchronizedAfterSuiteNode1Funcs, f)
	return true
}
