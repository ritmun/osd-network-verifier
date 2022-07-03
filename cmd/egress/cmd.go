//This design is based on ocm cli commands
//https://github.com/openshift-online/ocm-cli/blob/main/cmd/ocm/list/cluster/cmd.go
package egress

import (
	"context"
	"fmt"
	"os"

	ocmlog "github.com/openshift-online/ocm-sdk-go/logging"
	"github.com/openshift/osd-network-verifier/pkg/arguments"
	"github.com/openshift/osd-network-verifier/pkg/cloudclient"
	"github.com/spf13/cobra"
)

var args = arguments.Spec{}

var ValidateEgressCmd = &cobra.Command{
	Use:     "egress",
	Aliases: []string{"egress"},
	Short:   "Verify essential openshift domains are reachable from given subnet ID.",
	Long:    `Verify essential openshift domains are reachable from given subnet ID.`,
	Example: `For AWS, ensure your credential environment vars 
AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY (also AWS_SESSION_TOKEN for STS credentials) 
are set correctly before execution.

# Verify that essential openshift domains are reachable from a given SUBNET_ID
./osd-network-verifier egress --subnet-id $(SUBNET_ID) --profile $(AWS_PROFILE)`,
	RunE: run,
}

func init() {
	fs := ValidateEgressCmd.Flags()
	fs.StringVar(
		&args.ExistingVpc.VpcSubnetID,
		"subnet-id",
		"",
		"source subnet ID",
	)
	fs.StringVar(
		&args.CloudImageID,
		"image-id",
		"",
		"(optional) cloud image for the compute instance",
	)
	fs.StringVar(
		&args.InstanceType,
		"instance-type",
		"t3.micro",
		"(optional) compute instance type",
	)
	fs.StringVar(
		&args.Region,
		"region",
		args.Region,
		fmt.Sprintf("(optional) compute instance region. If absent, environment var %[1]v will be used, if set", cloudclient.RegionEnvVarStr, cloudclient.RegionDefault),
	)
	fs.StringToStringVar(
		&args.CloudTags,
		"cloud-tags",
		cloudclient.DefaultTags,
		"(optional) comma-seperated list of tags to assign to cloud resources e.g. --cloud-tags key1=value1,key2=value2",
	)
	fs.BoolVar(
		&args.TestSpec.Debug,
		"debug",
		false,
		"(optional) if true, enable additional debug-level logging",
	)
	fs.DurationVar(
		&args.TestSpec.Timeout,
		"timeout",
		cloudclient.DefaultTime,
		"(optional) timeout for individual egress verification requests",
	)
	fs.StringVar(
		&args.KmsKeyID,
		"kms-key-id",
		"",
		"(optional) ID of KMS key used to encrypt root volumes of compute instances. Defaults to cloud account default key",
	)
	fs.StringVar(
		&args.AwsProfile,
		"profile",
		"",
		"(optional) AWS profile. If present, any credentials passed with CLI will be ignored.",
	)

	if err := ValidateEgressCmd.MarkFlagRequired("subnet-id"); err != nil {
		ValidateEgressCmd.PrintErr(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, argv []string) error {
	// ctx
	ctx := context.TODO()

	// Create logger
	builder := ocmlog.NewStdLoggerBuilder()
	builder.Debug(args.TestSpec.Debug)
	logger, err := builder.Build()
	if err != nil {
		fmt.Printf("Unable to build logger: %s\n", err.Error())
		os.Exit(1)
	}
	client, err := cloudclient.NewClient(ctx, logger, args)
	if err != nil {
		logger.Error(ctx, "Error creating %s cloud client: %s", args.CloudProvider, err.Error())
		os.Exit(1)
	}
	out := client.ValidateEgress(ctx)
	out.Summary()
	if !out.IsSuccessful() {
		logger.Error(ctx, "Failure!")
		os.Exit(1)
	}

	logger.Info(ctx, "Success")
	return nil
}
