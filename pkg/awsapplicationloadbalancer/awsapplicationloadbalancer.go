package awsapplicationloadbalancer

import (
	"github.com/aws/aws-sdk-go/aws/session"
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
	Spec okrav1alpha1.AWSApplicationLoadBalancerConfigSpec

	Region  string
	Profile string
	Address string
	Session *session.Session
}
