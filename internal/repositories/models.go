package repositories

type Job struct {
	ID     int    `json:"jobId"`
	Name   string `json:"name"`
	Ttl    int    `json:"ttl"`
	Status string `json:""`
}

const (
	SUBMITTED = iota + 1
	PENDING
	RUNNING
	COMPLETED
)

func GetReableJobStatus(status int) string {
	switch status {
	case SUBMITTED:
		return "SUBMITTED"
	case PENDING:
		return "PENDING"
	case RUNNING:
		return "RUNNING"
	case COMPLETED:
		return "COMPLETED"
	}

	return "UNKNOWN"
}
