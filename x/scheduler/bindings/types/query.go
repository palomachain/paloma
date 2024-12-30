package types

type SchedulerQuery struct {
	Query *SchedulerQueryType `json:"query,omitempty"`
}

type SchedulerQueryType struct {
	JobById *JobByIdRequest `json:"full_denom,omitempty"`
}

type JobByIdRequest struct {
	JobId string `json:"job_id"`
}

type JobByIdResponse struct {
	Job *Job `json:"job"`
}
