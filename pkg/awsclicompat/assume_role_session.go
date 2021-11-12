package awsclicompat

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

type SessionConfig struct {
	Region     string
	Profile    string
	AssumeRole *AssumeRoleConfig
}

func AWSCredsFromConfig(conf *SessionConfig) (*session.Session, *sts.Credentials) {
	return AWSCredsFromValues(conf.Region, conf.Profile, conf.AssumeRole)
}

func AWSCredsFromValues(region, profile string, assumeRole *AssumeRoleConfig) (*session.Session, *sts.Credentials) {
	sess := NewSession(region, profile)

	if assumeRole == nil {
		return sess, nil
	}

	assumed, creds, err := AssumeRole(sess, *assumeRole)
	if err != nil {
		panic(err)
	}

	return assumed, creds
}
