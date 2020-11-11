package utils

import (
	"sync/atomic"
)

type (
	AtomicBool int32
)

func (b *AtomicBool) IsSet() bool {
	return atomic.LoadInt32((*int32)(b)) != 0
}

func (b *AtomicBool) SetTrue() {
	atomic.StoreInt32((*int32)(b), 1)
}

func (b *AtomicBool) SetTrueIfFalse() bool {
	return atomic.CompareAndSwapInt32((*int32)(b), 0, 1)
}

func (b *AtomicBool) SetFalse() {
	atomic.StoreInt32((*int32)(b), 0)
}

func (b *AtomicBool) SetFalseIfTrue() bool {
	return atomic.CompareAndSwapInt32((*int32)(b), 1, 0)
}

func Int(v int) *int {
	return &v
}

func IntValue(v *int) int {
	if v != nil {
		return *v
	}

	return 0
}

func String(v string) *string {
	return &v
}

func StringValue(v *string) string {
	if v != nil {
		return *v
	}

	return ""
}

func Bool(v bool) *bool {
	return &v
}

func BoolValue(v *bool) bool {
	if v != nil {
		return *v
	}

	return false
}

func Float64(v float64) *float64 {
	return &v
}

func Float64Value(v *float64) float64 {
	if v != nil {
		return *v
	}

	return 0
}
