package types

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
