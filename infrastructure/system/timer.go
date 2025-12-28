package system

import (
	"github.com/t-kuni/cqrs-example/domain/infrastructure/system"
	"time"
)

type Timer struct {
}

func NewTimer() system.ITimer {
	return &Timer{}
}

func (u Timer) Now() time.Time {
	return time.Now()
}
