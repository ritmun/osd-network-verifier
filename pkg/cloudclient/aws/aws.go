package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ocmlog "github.com/openshift-online/ocm-sdk-go/logging"
	"github.com/openshift/osd-network-verifier/pkg/cloudclient"
	"github.com/openshift/osd-network-verifier/pkg/output"
)

// ClientIdentifier is what kind of cloud this implement supports
const ClientIdentifier = "AWS"

// Client represents an AWS Client
type Client struct {
	ec2Client   EC2Client
	clientInput *ClientInput
	output      output.Output
	// the following are extracted from clientInput here to mitigate
	// "cannot create context from nil parent" error
	logger ocmlog.Logger
	ctx    context.Context
}

// This struct provides input for the AWS cloudclient
// aws/factory combines the cloudClient interface's ClientConfig and Test Specific Config structs into an AWS ClientInput
// Cloudclient Config structs can't be imported here due to cyclic import issues
type ClientInput struct {
	Ctx          context.Context
	Logger       ocmlog.Logger
	ClientConfig *cloudclient.ClientConfig
	ExecConfig   *cloudclient.ExecConfig
}

// Extend EC2Client so that we can mock them all for testing
// to re-generate mockfile once another interface is added for testing:
// mockgen -source=pkg/cloudclient/aws/aws.go -package mocks -destination=pkg/cloudclient/mocks/mock_aws.go
type EC2Client interface {
	RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
	DescribeInstanceStatus(ctx context.Context, input *ec2.DescribeInstanceStatusInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstanceStatusOutput, error)
	DescribeInstanceTypes(ctx context.Context, input *ec2.DescribeInstanceTypesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstanceTypesOutput, error)
	GetConsoleOutput(ctx context.Context, input *ec2.GetConsoleOutputInput, optFns ...func(*ec2.Options)) (*ec2.GetConsoleOutputOutput, error)
	TerminateInstances(ctx context.Context, input *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
	DescribeVpcAttribute(ctx context.Context, input *ec2.DescribeVpcAttributeInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcAttributeOutput, error)
}

func (c *Client) ByoVPCValidator(params cloudclient.ValidateByoVpc) error {
	c.logger.Info(context.TODO(), "interface executed: %s")
	return nil
}

func (c *Client) ValidateEgress(params cloudclient.ValidateEgress) *output.Output {
	return c.validateEgress(params)
}

func (c *Client) VerifyDns(params cloudclient.ValidateDns) *output.Output {
	return c.verifyDns(params)
}

func GetEc2ClientFromInput(input *ClientInput) (*ec2.Client, error) {
	ec2Client, err := getEc2ClientFromInput(*input)
	if err != nil {
		return nil, fmt.Errorf("unable to create EC2 Client: %w", err)
	}
	return ec2Client, nil
}
