package utils

// support for uint8

type Equatable interface {
	Equals(w Equatable) bool
}

type WrappedUint8 struct {
	val uint8
}

func WrapUint8(v uint8) *WrappedUint8 {
	return &WrappedUint8{v}
}

func (w *WrappedUint8) Equals(other Equatable) bool {
	switch other.(type) {
	case *WrappedUint8:
		return w.val == other.(*WrappedUint8).val
	default:
		return false
	}
}

func (w *WrappedUint8) Unwrap() uint8 {
	return w.val
}

// support for bool

type WrappedBool struct {
	val bool
}

func WrapBool(v bool) *WrappedBool {
	return &WrappedBool{v}
}

func (w *WrappedBool) Equals(other Equatable) bool {
	switch other.(type) {
	case *WrappedBool:
		return w.val == other.(*WrappedBool).val
	default:
		return false
	}
}

func (w *WrappedBool) Unwrap() bool {
	return w.val
}
