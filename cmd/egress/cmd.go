package egress

import (
	"context"
	"fmt"
	"os"
	"time"

	ocmlog "github.com/openshift-online/ocm-sdk-go/logging"
	"github.com/openshift/osd-network-verifier/pkg/cloudclient"
	"github.com/spf13/cobra"
)

var (
	defaultTags     = map[string]string{"osd-network-verifier": "owned", "red-hat-managed": "true", "Name": "osd-network-verifier"}
	regionEnvVarStr = "AWS_REGION"
	regionDefault   = "us-east-2"
)

//Test specific commandline args
type egressConfig struct {
	vpcSubnetID  string
	cloudImageID string
	instanceType string
}

func getDefaultRegion() string {
	val, present := os.LookupEnv(regionEnvVarStr)
	if present {
		return val
	} else {
		return regionDefault
	}
}
func NewCmdValidateEgress() *cobra.Command {
	config := egressConfig{}
	cmdOptions := cloudclient.CmdOptions{}

	validateEgressCmd := &cobra.Command{
		Use:        "egress",
		Aliases:    nil,
		SuggestFor: nil,
		Short:      "Verify essential openshift domains are reachable from given subnet ID.",
		Long:       `Verify essential openshift domains are reachable from given subnet ID.`,
		Example: `For AWS, ensure your credential environment vars 
AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY (also AWS_SESSION_TOKEN for STS credentials) 
are set correctly before execution.

# Verify that essential openshift domains are reachable from a given SUBNET_ID
./osd-network-verifier egress --subnet-id $(SUBNET_ID) --profile $(AWS_PROFILE)`,

		Run: func(cmd *cobra.Command, args []string) {
			// ctx
			ctx := context.TODO()

			// Create logger
			builder := ocmlog.NewStdLoggerBuilder()
			builder.Debug(cmdOptions.Debug)
			logger, err := builder.Build()
			if err != nil {
				fmt.Printf("Unable to build logger: %s\n", err.Error())
				os.Exit(1)
			}
			var cli cloudclient.CloudClient
			logger.Info(ctx, "Using region: %s", cmdOptions.Region)
			if cmdOptions.AwsProfile != "" || os.Getenv("AWS_ACCESS_KEY_ID") != "" {
				// For AWS type
				if cmdOptions.AwsProfile != "" {
					logger.Info(ctx, "Using AWS profile: %s.", cmdOptions.AwsProfile)
				} else {
					logger.Info(ctx, "Using provided AWS credentials")
				}
				//cli, err = cloudclient.NewClient(ctx, logger, cmdOptions.region, config.instanceType, cmdOptions.cloudTags, "aws", config.awsProfile)
				cli, err = cloudclient.NewClient(ctx, logger, config.instanceType, "aws", cmdOptions)
			} else {
				//	todo after GCP is implemented, check GCP type using creds
				logger.Error(ctx, "No AWS credentials found.")
			}
			if err != nil {
				logger.Error(ctx, "Error creating cloud client: %s", err.Error())
				os.Exit(1)
			}
			out := cli.ValidateEgress(ctx, config.vpcSubnetID, config.cloudImageID, cmdOptions.KmsKeyID, cmdOptions.Timeout)
			out.Summary()
			if !out.IsSuccessful() {
				logger.Error(ctx, "Failure!")
				os.Exit(1)
			}

			logger.Info(ctx, "Success")
		},

		FParseErrWhitelist: cobra.FParseErrWhitelist{},
		CompletionOptions:  cobra.CompletionOptions{},
		TraverseChildren:   false,
	}

	validateEgressCmd.Flags().StringVar(&config.vpcSubnetID, "subnet-id", "", "source subnet ID")
	validateEgressCmd.Flags().StringVar(&config.cloudImageID, "image-id", "", "(optional) cloud image for the compute instance")
	validateEgressCmd.Flags().StringVar(&config.instanceType, "instance-type", "t3.micro", "(optional) compute instance type")
	validateEgressCmd.Flags().StringVar(&cmdOptions.Region, "region", getDefaultRegion(), fmt.Sprintf("(optional) compute instance region. If absent, environment var %[1]v will be used, if set", regionEnvVarStr, regionDefault))
	validateEgressCmd.Flags().StringToStringVar(&cmdOptions.CloudTags, "cloud-tags", defaultTags, "(optional) comma-seperated list of tags to assign to cloud resources e.g. --cloud-tags key1=value1,key2=value2")
	validateEgressCmd.Flags().BoolVar(&cmdOptions.Debug, "debug", false, "(optional) if true, enable additional debug-level logging")
	validateEgressCmd.Flags().DurationVar(&cmdOptions.Timeout, "timeout", 1*time.Second, "(optional) timeout for individual egress verification requests")
	validateEgressCmd.Flags().StringVar(&cmdOptions.KmsKeyID, "kms-key-id", "", "(optional) ID of KMS key used to encrypt root volumes of compute instances. Defaults to cloud account default key")
	validateEgressCmd.Flags().StringVar(&cmdOptions.AwsProfile, "profile", "", "(optional) AWS profile. If present, any credentials passed with CLI will be ignored.")

	if err := validateEgressCmd.MarkFlagRequired("subnet-id"); err != nil {
		validateEgressCmd.PrintErr(err)
		os.Exit(1)
	}

	return validateEgressCmd

}
