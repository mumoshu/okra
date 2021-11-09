package cell

import (
	okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"
)

type Provider struct {
}

type CreateInput struct {
	ListenerARN string
}

func (p *Provider) CreateConfigFromAWS(input CreateInput) error {
	return nil
}

type SyncInput struct {
	Spec okrav1alpha1.CellSpec
}

func Sync(config SyncInput) error {
	return nil
}
