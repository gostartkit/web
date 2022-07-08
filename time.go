package web

import "time"

// Now return *time.Time
func Now() *time.Time {
	now := time.Now()
	return &now
}

// After return *time.Time
func After(d time.Duration) *time.Time {
	now := Now().Add(d)
	return &now
}
