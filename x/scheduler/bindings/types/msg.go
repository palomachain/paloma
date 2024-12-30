package types

type SchedulerMsg struct {
	Message *SchedulerMsgType `json:"scheduler_msg_type,omitempty"`
}

type SchedulerMsgType struct {
	// Contracts can create new jobs. Any number of jobs
	// may be created, so lang as job IDs stay unique.
	CreateJob *CreateJob `json:"create_job,omitempty"`
}

// CreateJob is a message to create a new job.
// JobId is a unique identifier for the job.
// ChainType is the type of chain the job is for (e.g. "evm").
// ChainReferenceId is the reference for the chain (e.g. "eth-main").
// Definition containts the ABI of the target contract.
// Payload is the data to be sent to the contract.
// PayloadModifiable indicates whether the payload can be modified.
// IsMEV indicates whether the job should be routed via an MEV pool.
type CreateJob struct {
	Job *Job `json:"job,omitempty"`
}
