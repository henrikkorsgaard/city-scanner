package proximity

import "time"

type Reading struct {
	ID        int
	NodeID    int
	DeviceID  string
	Signal    int
	Timestamp time.Time
}
