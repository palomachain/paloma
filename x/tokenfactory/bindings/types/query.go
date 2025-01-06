package types

type Query struct {
	FullDenom       *FullDenom       `json:"full_denom,omitempty"`
	Admin           *DenomAdmin      `json:"admin,omitempty"`
	Metadata        *GetMetadata     `json:"metadata,omitempty"`
	DenomsByCreator *DenomsByCreator `json:"denoms_by_creator,omitempty"`
	Params          *GetParams       `json:"params,omitempty"`
}

type FullDenom struct {
	CreatorAddr string `json:"creator_addr"`
	Subdenom    string `json:"subdenom"`
}

type FullDenomResponse struct {
	Denom string `json:"denom"`
}

type GetMetadata struct {
	Denom string `json:"denom"`
}

type MetadataResponse struct {
	Metadata *Metadata `json:"metadata,omitempty"`
}

type DenomAdmin struct {
	Denom string `json:"denom"`
}

type AdminResponse struct {
	Admin string `json:"admin"`
}

type DenomsByCreator struct {
	Creator string `json:"creator"`
}

type DenomsByCreatorResponse struct {
	Denoms []string `json:"denoms"`
}

type GetParams struct{}

type ParamsResponse struct {
	Params Params `json:"params"`
}
