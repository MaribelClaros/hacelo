// Package clock provides implementations of port.Clock.
package clock

import "time"

// System returns the real wall-clock time in UTC.
type System struct{}

func (System) Now() time.Time {
	return time.Now().UTC()
}
