package filter

var (
	DanglingTrue  = NewFilter("dangling", "true")
	DanglingFalse = NewFilter("dangling", "false")
)

const LabelPrefix = "label"
