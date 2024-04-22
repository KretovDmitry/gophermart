package order

import "time"

type Status string

const (
	INVALID    Status = "INVALID"
	PROCESSED  Status = "PROCESSED"
	NEW        Status = "NEW"
	PROCESSING Status = "PROCESSING"
)

type Order struct {
	Number     string
	UploadetAt time.Time
	Status     Status
	ID         int
	UserID     int
}
