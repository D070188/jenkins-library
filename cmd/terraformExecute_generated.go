// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type terraformExecuteOptions struct {
	Command          string   `json:"command,omitempty"`
	TerraformSecrets string   `json:"terraformSecrets,omitempty"`
	GlobalOptions    []string `json:"globalOptions,omitempty"`
	AdditionalArgs   []string `json:"additionalArgs,omitempty"`
}

// TerraformExecuteCommand Executes Terraform
func TerraformExecuteCommand() *cobra.Command {
	const STEP_NAME = "terraformExecute"

	metadata := terraformExecuteMetadata()
	var stepConfig terraformExecuteOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	splunkClient := &splunk.Splunk{}
	telemetryClient := &telemetry.Telemetry{}
	provider, err := orchestrator.NewOrchestratorSpecificConfigProvider()
	if err != nil {
		log.Entry().Error(err)
		provider = &orchestrator.UnknownOrchestratorConfigProvider{}
	}

	var createTerraformExecuteCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Executes Terraform",
		Long:  `This step executes the terraform binary with the given command, and is able to fetch additional variables from vault.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			validation, err := validation.New(validation.WithJSONNamesForStructFields(), validation.WithPredefinedErrorMessages())
			if err != nil {
				return err
			}
			if err = validation.ValidateStruct(stepConfig); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			customTelemetryData := telemetry.CustomData{}
			customTelemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				customTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				customTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				customTelemetryData.Custom1Label = "PiperCommitHash"
				customTelemetryData.Custom1 = GitCommit
				customTelemetryData.Custom2Label = "PiperTag"
				customTelemetryData.Custom2 = GitTag
				customTelemetryData.Custom3Label = "Stage"
				customTelemetryData.Custom3 = provider.GetStageName()
				customTelemetryData.Custom4Label = "Orchestrator"
				customTelemetryData.Custom4 = provider.OrchestratorType()
				telemetryClient.SetData(&customTelemetryData)
				telemetryClient.Send()
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunkClient.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			terraformExecute(stepConfig, &customTelemetryData)
			customTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addTerraformExecuteFlags(createTerraformExecuteCmd, &stepConfig)
	return createTerraformExecuteCmd
}

func addTerraformExecuteFlags(cmd *cobra.Command, stepConfig *terraformExecuteOptions) {
	cmd.Flags().StringVar(&stepConfig.Command, "command", `plan`, "")
	cmd.Flags().StringVar(&stepConfig.TerraformSecrets, "terraformSecrets", os.Getenv("PIPER_terraformSecrets"), "")
	cmd.Flags().StringSliceVar(&stepConfig.GlobalOptions, "globalOptions", []string{}, "")
	cmd.Flags().StringSliceVar(&stepConfig.AdditionalArgs, "additionalArgs", []string{}, "")

}

// retrieve step metadata
func terraformExecuteMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "terraformExecute",
			Aliases:     []config.Alias{},
			Description: "Executes Terraform",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "command",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `plan`,
					},
					{
						Name: "terraformSecrets",
						ResourceRef: []config.ResourceReference{
							{
								Name:    "terraformExecuteFileVaultSecret",
								Type:    "vaultSecretFile",
								Default: "terraformExecute",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_terraformSecrets"),
					},
					{
						Name:        "globalOptions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "additionalArgs",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
				},
			},
			Containers: []config.Container{
				{Name: "terraform", Image: "hashicorp/terraform:0.14.7", EnvVars: []config.EnvVar{{Name: "TF_IN_AUTOMATION", Value: "piper"}}, Options: []config.Option{{Name: "--entrypoint", Value: ""}}},
			},
		},
	}
	return theMetaData
}
