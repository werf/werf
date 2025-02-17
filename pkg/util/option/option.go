package option

// ValueOrDefault return value when it is not equal zero value or returns default value.
func ValueOrDefault[T comparable](value, defaultValue T) T {
	var zeroValue T

	if value != zeroValue {
		return value
	}

	return defaultValue
}

// PtrValueOrDefault returns a value behind the pointer when the pointer is nil or the default value.
//
// Borrowed from https://github.com/cidverse/go-ptr/blob/main/ptr-generic.go#L19
func PtrValueOrDefault[T any](p *T, defaultValue T) T {
	if p != nil {
		return *p
	}

	return defaultValue
}
