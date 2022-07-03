package cloudclient

import (
	"context"
	"fmt"
	"os"
	"time"

	ocmlog "github.com/openshift-online/ocm-sdk-go/logging"
	"github.com/openshift/osd-network-verifier/pkg/arguments"
	awsCloudClient "github.com/openshift/osd-network-verifier/pkg/cloudclient/aws"
	"github.com/openshift/osd-network-verifier/pkg/output"
)

const GCP = "GCP"
const AWS = "AWS"

var (
	DefaultTags     = map[string]string{"osd-network-verifier": "owned", "red-hat-managed": "true", "Name": "osd-network-verifier"}
	RegionEnvVarStr = "AWS_REGION"
	RegionDefault   = "us-east-1"
	DefaultTime     = 2 * time.Second
)

// todo implement similar getter for AWS secrets and profile
func getDefaultRegion(spec arguments.Spec) string {
	if spec.Region != "" {
		return spec.Region
	}
	val, present := os.LookupEnv(RegionEnvVarStr)
	if present {
		return val
	} else {
		return RegionDefault
	}
}

// CloudClient defines the interface for a cloud agnostic implementation
// For mocking: mockgen -source=pkg/cloudclient/cloudclient.go -package mocks -destination=pkg/cloudclient/mocks/mock_cloudclient.go
type CloudClient interface {

	// ByoVPCValidator validates the configuration given by the customer
	ByoVPCValidator(ctx context.Context) error

	// ValidateEgress validates that all required targets are reachable from the vpcsubnet
	// target URLs: https://docs.openshift.com/rosa/rosa_getting_started/rosa-aws-prereqs.html#osd-aws-privatelink-firewall-prerequisites
	// Expected return value is *output.Output that's storing failures, exceptions and errors
	ValidateEgress(ctx context.Context) *output.Output

	// VerifyDns verifies that a given VPC meets the DNS requirements specified in:
	// https://docs.openshift.com/container-platform/4.10/installing/installing_aws/installing-aws-vpc.html
	// Expected return value is *output.Output that's storing failures, exceptions and errors
	VerifyDns(ctx context.Context, vpcID string) *output.Output
}

func getCloudClientType(spec arguments.Spec) string {
	if spec.AwsProfile != "" || os.Getenv("AWS_ACCESS_KEY_ID") != "" || spec.CloudProvider == "aws" {
		return AWS
	}
	if spec.CloudProvider == GCP {
		return GCP
	}
	return "unknown"
}

func NewClient(ctx context.Context, logger ocmlog.Logger,
	spec arguments.Spec) (CloudClient, error) {

	logger.Info(ctx, "Using region: %s", getDefaultRegion(spec))

	switch getCloudClientType(spec) {
	case AWS:
		if spec.AwsProfile != "" {
			logger.Info(ctx, "Using AWS profile: %s.", spec.AwsProfile)
		} else {
			logger.Info(ctx, "Using AWS secret key")
		}
		clientInput := &awsCloudClient.ClientInput{
			Ctx:             ctx,
			Logger:          logger,
			AccessKeyId:     os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
			SessionToken:    os.Getenv("AWS_SESSION_TOKEN"),
			VpcSubnetID:     spec.ExistingVpc.VpcSubnetID,
			VpcID:           spec.ExistingVpc.VpcID,
			CloudImageID:    spec.CloudImageID,
			Timeout:         spec.TestSpec.Timeout,
			KmsKeyID:        spec.KmsKeyID,
			Region:          getDefaultRegion(spec),
			InstanceType:    spec.InstanceType,
			Tags:            spec.CloudTags,
			Profile:         spec.AwsProfile,
		}
		return awsCloudClient.NewClient(clientInput)
	default:
		return nil, fmt.Errorf("No AWS credentials found. Non-AWS cloud clients are currently not supported.")
	}

}
