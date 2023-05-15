package gcpverifier

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/openshift/osd-network-verifier/pkg/output"
	"github.com/openshift/osd-network-verifier/pkg/verifier"
)

const cloudImageIDDefault = "cos-97-lts"

// validateEgress performs validation process for egress
// Basic workflow is:
// - prepare for ComputeService instance creation
// - create instance and wait till it gets ready, wait for gcpUserData script execution
// - find unreachable endpoints & parse output, then terminate instance
// - return `g.output` which stores the execution results
func (g *GcpVerifier) ValidateEgress(vei verifier.ValidateEgressInput) *output.Output {
	g.Logger.Debug(vei.Ctx, "Using configured timeout of %s for each egress request", vei.Timeout.String())
	//default gcp machine e2
	if vei.InstanceType == "" {
		vei.InstanceType = "e2-standard-2"
	}

	if err := g.validateMachineType(vei.GCP.ProjectID, vei.GCP.Zone, vei.InstanceType); err != nil {
		return g.Output.AddError(fmt.Errorf("instance type %s is invalid: %s", vei.InstanceType, err))
	}

	userDataVariables := map[string]string{
		"AWS_REGION":               "us-east-2", // Not sure if this is the correct data
		"USERDATA_BEGIN":           "USERDATA BEGIN",
		"USERDATA_END":             userdataEndVerifier,
		"VALIDATOR_START_VERIFIER": "VALIDATOR START",
		"VALIDATOR_END_VERIFIER":   "VALIDATOR END",
		"VALIDATOR_IMAGE":          networkValidatorImage,
		"TIMEOUT":                  vei.Timeout.String(),
		"HTTP_PROXY":               vei.Proxy.HttpProxy,
		"HTTPS_PROXY":              vei.Proxy.HttpsProxy,
		"CACERT":                   base64.StdEncoding.EncodeToString([]byte(vei.Proxy.Cacert)),
		"NOTLS":                    strconv.FormatBool(vei.Proxy.NoTls),
	}

	userData, err := generateUserData(userDataVariables)
	if err != nil {
		return g.Output.AddError(err)
	}

	g.Logger.Debug(vei.Ctx, "Generated userdata script:\n---\n%s\n---", userData)

	if vei.CloudImageID == "" {
		vei.CloudImageID = cloudImageIDDefault
	}

	//for random name
	rand.Seed(time.Now().UnixNano())

	//image list https://cloud.google.com/compute/docs/images/os-details#red_hat_enterprise_linux_rhel

	instance, err := g.createComputeServiceInstance(createComputeServiceInstanceInput{
		projectID:    vei.GCP.ProjectID,
		zone:         vei.GCP.Zone,
		vpcSubnetID:  fmt.Sprintf("projects/%s/regions/%s/subnetworks/%s", vei.GCP.ProjectID, vei.GCP.Region, vei.SubnetID),
		userdata:     userData,
		machineType:  vei.InstanceType,
		instanceName: fmt.Sprintf("verifier-%v", rand.Intn(10000)),
		sourceImage:  fmt.Sprintf("projects/cos-cloud/global/images/family/%s", vei.CloudImageID),
		networkName:  fmt.Sprintf("projects/%s/global/networks/%s", vei.GCP.ProjectID, vei.GCP.VpcName),
		tags:         vei.Tags,
	})
	if err != nil {
		err = g.GcpClient.TerminateComputeServiceInstance(vei.GCP.ProjectID, vei.GCP.Zone, instance.Name)
		if err != nil {
			g.Output.AddError(err)
		}
		return g.Output.AddError(err) // fatal
	}

	g.Logger.Debug(vei.Ctx, "Waiting for ComputeService instance %s to be running", instance.Name)
	if instanceReadyErr := g.waitForComputeServiceInstanceCompletion(vei.GCP.ProjectID, vei.GCP.Zone, instance.Name); instanceReadyErr != nil {
		err = g.GcpClient.TerminateComputeServiceInstance(vei.GCP.ProjectID, vei.GCP.Zone, instance.Name) // try to terminate the created instanc
		if err != nil {
			g.Output.AddError(err)
		}
		return g.Output.AddError(instanceReadyErr) // fatal
	}

	g.Logger.Info(vei.Ctx, "Gathering and parsing console log output...")

	err = g.findUnreachableEndpoints(vei.GCP.ProjectID, vei.GCP.Zone, instance.Name)
	if err != nil {
		g.Output.AddError(err)
	}

	err = g.GcpClient.TerminateComputeServiceInstance(vei.GCP.ProjectID, vei.GCP.Zone, instance.Name)
	if err != nil {
		g.Output.AddError(err)
	}

	return &g.Output
}

// TODO():
func (g *GcpVerifier) VerifyDns(vdi verifier.VerifyDnsInput) *output.Output {
	return &output.Output{}
}
