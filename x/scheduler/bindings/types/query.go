package types

type Query struct {
	JobById *JobByIdRequest `json:"job_id,omitempty"`
}

type JobByIdRequest struct {
	JobId string `json:"job_id"`
}

type JobByIdResponse struct {
	Job *Job `json:"job"`
}
