// argument command struct based on ocm cli
// https://github.com/openshift-online/ocm-cli/blob/main/pkg/arguments/arguments.go

package arguments

import "time"

type TestSpec struct {
	Debug   bool
	Timeout time.Duration
}

// common commandline args
type Spec struct {
	TestSpec      TestSpec
	CloudProvider string // not required. if provided, currently only supports "aws".
	CloudTags     map[string]string
	Region        string
	AwsProfile    string
	InstanceType  string
	CloudImageID  string
	KmsKeyID      string
	ExistingVpc   VPC
}

type VPC struct {
	VpcSubnetID string
	VpcID       string
}
