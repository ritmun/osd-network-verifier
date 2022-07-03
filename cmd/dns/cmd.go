package dns

import (
	"context"
	"fmt"
	"os"

	ocmlog "github.com/openshift-online/ocm-sdk-go/logging"
	"github.com/openshift/osd-network-verifier/pkg/arguments"
	"github.com/openshift/osd-network-verifier/pkg/cloudclient"
	"github.com/spf13/cobra"
)

func NewCmdValidateDns() *cobra.Command {
	config := arguments.Spec{}

	validateDnsCmd := &cobra.Command{
		Use: "dns",
		Run: func(cmd *cobra.Command, args []string) {
			// ctx
			ctx := context.TODO()

			// Create logger
			builder := ocmlog.NewStdLoggerBuilder()
			builder.Debug(config.TestSpec.Debug)
			logger, err := builder.Build()
			if err != nil {
				fmt.Printf("Unable to build logger: %s\n", err.Error())
				os.Exit(1)
			}
			client, err := cloudclient.NewClient(ctx, logger, config)
			if err != nil {
				logger.Error(ctx, "Error creating %s cloud client: %s", config.CloudProvider, err.Error())
				os.Exit(1)
			}
			out := client.VerifyDns(ctx, config.ExistingVpc.VpcID)
			out.Summary()
			if !out.IsSuccessful() {
				logger.Error(ctx, "Failure!")
				os.Exit(1)
			}

			logger.Info(ctx, "Success")
		},
	}

	validateDnsCmd.Flags().StringVar(&config.ExistingVpc.VpcID, "vpc-id", "", "ID of the VPC under test")
	validateDnsCmd.Flags().StringVar(&config.Region, "region", cloudclient.RegionDefault, fmt.Sprintf("Region to validate. Defaults to exported var %[1]v or '%[2]v' if not %[1]v set", cloudclient.RegionEnvVarStr, cloudclient.RegionDefault))
	validateDnsCmd.Flags().BoolVar(&config.TestSpec.Debug, "debug", false, "If true, enable additional debug-level logging")

	if err := validateDnsCmd.MarkFlagRequired("vpc-id"); err != nil {
		validateDnsCmd.PrintErr(err)
		os.Exit(1)
	}

	return validateDnsCmd

}
