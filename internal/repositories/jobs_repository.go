package repositories

func GetJobs() *[]Job {
	return &[]Job{
		{ID: 1, Name: "job-001", Ttl: 5, Status: GetReableJobStatus(PENDING)},
		{ID: 2, Name: "job-002", Ttl: 3, Status: GetReableJobStatus(RUNNING)},
		{ID: 3, Name: "job-003", Ttl: 3, Status: GetReableJobStatus(SUBMITTED)},
		{ID: 4, Name: "job-004", Ttl: 7, Status: GetReableJobStatus(COMPLETED)},
		{ID: 5, Name: "job-005", Ttl: 2, Status: GetReableJobStatus(7)},
	}
}
