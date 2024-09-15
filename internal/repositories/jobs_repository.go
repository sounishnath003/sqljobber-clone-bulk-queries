package repositories

import "github.com/sounishnath003/jobprocessor/internal/models"

func GetJobs() *[]models.Job {
	return &[]models.Job{
		{ID: 1, Name: "job-001", Ttl: 5, Status: models.StatusPending},
		{ID: 2, Name: "job-002", Ttl: 3, Status: models.StatusStarted},
		{ID: 3, Name: "job-003", Ttl: 3, Status: models.StatusSuccess},
		{ID: 4, Name: "job-004", Ttl: 7, Status: models.StatusRetry},
		{ID: 5, Name: "job-005", Ttl: 2, Status: models.StatusFailure},
	}
}
