package suite_init

import (
	. "github.com/onsi/ginkgo/v2"
)

type SynchronizedSuiteCallbacksData struct {
	synchronizedBeforeSuiteNode1Funcs               []func()
	synchronizedBeforeSuiteNode1FuncWithReturnValue func() []byte
	synchronizedBeforeSuiteAllNodesFuncs            []func([]byte)

	synchronizedAfterSuiteAllNodesFuncs []func()
	synchronizedAfterSuiteNode1Funcs    []func()
}

func NewSynchronizedSuiteCallbacksData() *SynchronizedSuiteCallbacksData {
	data := &SynchronizedSuiteCallbacksData{}

	SynchronizedBeforeSuite(func() []byte {
		for _, f := range data.synchronizedBeforeSuiteNode1Funcs {
			f()
		}
		if data.synchronizedBeforeSuiteNode1FuncWithReturnValue != nil {
			return data.synchronizedBeforeSuiteNode1FuncWithReturnValue()
		} else {
			return nil
		}
	}, func(d []byte) {
		for _, f := range data.synchronizedBeforeSuiteAllNodesFuncs {
			f(d)
		}
	})

	SynchronizedAfterSuite(func() {
		for _, f := range data.synchronizedAfterSuiteAllNodesFuncs {
			f()
		}
	}, func() {
		for _, f := range data.synchronizedAfterSuiteNode1Funcs {
			f()
		}
	})

	return data
}

func (data *SynchronizedSuiteCallbacksData) AppendSynchronizedBeforeSuiteNode1Func(f func()) bool {
	data.synchronizedBeforeSuiteNode1Funcs = append(data.synchronizedBeforeSuiteNode1Funcs, f)
	return true
}

func (data *SynchronizedSuiteCallbacksData) AppendSynchronizedBeforeSuiteAllNodesFunc(f func([]byte)) bool {
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

func (data *SynchronizedSuiteCallbacksData) AppendSynchronizedAfterSuiteAllNodesFunc(f func()) bool {
	data.synchronizedAfterSuiteAllNodesFuncs = append(data.synchronizedAfterSuiteAllNodesFuncs, f)
	return true
}

func (data *SynchronizedSuiteCallbacksData) AppendSynchronizedAfterSuiteNode1Func(f func()) bool {
	data.synchronizedAfterSuiteNode1Funcs = append(data.synchronizedAfterSuiteNode1Funcs, f)
	return true
}
