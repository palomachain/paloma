package types

import "fmt"

// JobId is a unique identifier for the job.
// ChainType is the type of chain the job is for (e.g. "evm").
// ChainReferenceId is the reference for the chain (e.g. "eth-main").
// Definition containts the ABI of the target contract.
// Payload is the data to be sent to the contract.
// PayloadModifiable indicates whether the payload can be modified.
// IsMEV indicates whether the job should be routed via an MEV pool.
type Job struct {
	JobId             string `json:"job_id"`
	ChainType         string `json:"chain_type"`
	ChainReferenceId  string `json:"chain_reference_id"`
	Definition        string `json:"definition"`
	Payload           string `json:"payload"`
	PayloadModifiable bool   `json:"payload_modifiable"`
	IsMEV             bool   `json:"is_mev"`
}

func (j Job) ValidateBasic() error {
	if j.JobId == "" {
		return fmt.Errorf("invalid job id")
	}
	if j.ChainType == "" {
		return fmt.Errorf("invalid chain type")
	}
	if j.ChainReferenceId == "" {
		return fmt.Errorf("invalid chain reference id")
	}
	if j.Definition == "" {
		return fmt.Errorf("invalid definition")
	}
	if j.Payload == "" {
		return fmt.Errorf("invalid payload")
	}
	return nil
}
