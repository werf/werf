package cleanup

type Func func()

var NoOp Func = func() {}
