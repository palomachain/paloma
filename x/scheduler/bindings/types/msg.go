package types

import "fmt"

type Message struct {
	// Contracts can create new jobs. Any number of jobs
	// may be created, so lang as job IDs stay unique.
	CreateJob *CreateJob `json:"create_job,omitempty"`

	// Contracts can execute jobs.
	ExecuteJob *ExecuteJob `json:"execute_job,omitempty"`
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

func (c CreateJob) ValidateBasic() error {
	if c.Job == nil {
		return fmt.Errorf("job cannot be nil")
	}

	return c.Job.ValidateBasic()
}

// ExecyteJob is a message to execute a job.
// JobId is the unique identifier for the job.
// Sender is the address of the sender.
// Payload is the data to be sent to the contract.
type ExecuteJob struct {
	JobID   string `json:"job_id"`
	Sender  string `json:"sender"`
	Payload []byte `json:"payload"`
}

func (j ExecuteJob) ValidateBasic() error {
	if j.JobID == "" {
		return fmt.Errorf("job_id cannot be empty")
	}

	if j.Sender == "" {
		return fmt.Errorf("sender cannot be empty")
	}

	if len(j.Payload) < 1 {
		return fmt.Errorf("payload cannot be empty")
	}

	return nil
}
