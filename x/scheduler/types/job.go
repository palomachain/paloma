package types

import (
	"strings"
)

const allowedJobIDCharacters = "abcdefghijklmnopqrstuvwxyz0123456789-_."

const JobIDMaxLen = 32

const JobAddressLength = 32

func (j *Job) ValidateBasic() error {
	if len(j.ID) == 0 {
		return ErrInvalid.Wrap("job id can't be empty")
	}

	if j.ID != strings.ToLower(j.ID) {
		return ErrInvalid.Wrap("job ID must be all in lowercase")
	}
	if len(j.ID) > JobIDMaxLen {
		return ErrInvalid.Wrap("job ID can't be greater than 32 characters")
	}

	if strings.Index(j.ID, "paloma") >= 0 {
		return ErrInvalid.Wrap("job ID can't contain word paloma")
	}

	if strings.Index(j.ID, "pigeon") >= 0 {
		return ErrInvalid.Wrap("job ID can't contain word pigeon")
	}

	for _, l := range j.ID {
		found := false
		for _, w := range allowedJobIDCharacters {
			if w == l {
				found = true
				break
			}
		}
		if !found {
			return ErrInvalid.Wrapf("job ID contains an invalid character: %s. allowed characters: %s", string(l), allowedJobIDCharacters)
		}
	}

	if len(j.Definition) == 0 {
		return ErrInvalid.Wrap("job definition can't be empty")
	}

	if len(j.Routing.ChainType) == 0 {
		return ErrInvalid.Wrap("job's routing information chain type can't be empty")
	}

	if len(j.Routing.ChainReferenceID) == 0 {
		return ErrInvalid.Wrap("job's routing information chain reference ID can't be empty")
	}

	if j.EnforceMEVRelay && !j.isTargetedAtChainWithMEVRelayingSupport() {
		return ErrInvalid.Wrapf("MEV relaying not supported on target chain '%s'", j.Routing.ChainReferenceID)
	}

	if err := j.Permissions.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

func (j *Job) isTargetedAtChainWithMEVRelayingSupport() bool {
	switch j.Routing.ChainReferenceID {
	case "eth-main", "bnb-main", "matic-main":
		return true
	}

	return false
}

func (p *Permissions) ValidateBasic() error {
	verify := func(runnerPerms []*Runner) error {
		for _, l := range runnerPerms {
			switch {
			case len(l.Address) > 0:
				if len(l.ChainType) == 0 || len(l.ChainReferenceID) == 0 {
					return ErrInvalid.Wrap("can't have permission for an address and chain type or chain reference ID be set to nothing")
				}
			case len(l.ChainReferenceID) > 0:
				if len(l.ChainType) == 0 {
					return ErrInvalid.Wrap("can't have permission for a chain reference ID without a chain type")
				}
			}
		}
		return nil
	}
	var err error
	err = verify(p.Blacklist)
	if err != nil {
		return err
	}
	err = verify(p.Whitelist)
	if err != nil {
		return err
	}
	return nil
}
