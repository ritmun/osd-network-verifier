package byovpc

import (
	"context"
	"fmt"
	"os"

	ocmlog "github.com/openshift-online/ocm-sdk-go/logging"
	"github.com/openshift/osd-network-verifier/pkg/cloudclient"
	"github.com/spf13/cobra"
)

var debug bool

func NewCmdByovpc() *cobra.Command {
	byovpcCmd := &cobra.Command{
		Use:   "byovpc",
		Short: "Verify subnet configuration of a specific VPC",
		Run: func(cmd *cobra.Command, args []string) {
			// Create logger
			builder := ocmlog.NewStdLoggerBuilder()
			builder.Debug(debug)
			logger, err := builder.Build()
			if err != nil {
				fmt.Printf("Unable to build logger: %s\n", err.Error())
				os.Exit(1)
			}

			ctx := context.TODO()

			// TODO when this command is actually used, most if not all of the following should be command line options
			region := os.Getenv("AWS_REGION")
			instanceType := "t3.micro"
			tags := map[string]string{}

			var cli cloudclient.CloudClient
			cli, err = cloudclient.NewClient(ctx, logger, region, instanceType, tags, "aws", "")
			if err != nil {
				logger.Error(ctx, err.Error())
				os.Exit(1)
			}

			err = cli.ByoVPCValidator(ctx)
			if err != nil {
				logger.Error(ctx, err.Error())
				os.Exit(1)
			}

			logger.Info(ctx, "Success")
		},
	}

	byovpcCmd.Flags().BoolVar(&debug, "debug", false, "If true, enable additional debug-level logging")

	return byovpcCmd
}
