// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type mtaBuildOptions struct {
	MtarName                        string   `json:"mtarName,omitempty"`
	MtarGroup                       string   `json:"mtarGroup,omitempty"`
	Version                         string   `json:"version,omitempty"`
	Extensions                      string   `json:"extensions,omitempty"`
	Platform                        string   `json:"platform,omitempty" validate:"oneof=CF NEO XSA"`
	ApplicationName                 string   `json:"applicationName,omitempty"`
	Source                          string   `json:"source,omitempty"`
	Target                          string   `json:"target,omitempty"`
	DefaultNpmRegistry              string   `json:"defaultNpmRegistry,omitempty"`
	ProjectSettingsFile             string   `json:"projectSettingsFile,omitempty"`
	GlobalSettingsFile              string   `json:"globalSettingsFile,omitempty"`
	M2Path                          string   `json:"m2Path,omitempty"`
	InstallArtifacts                bool     `json:"installArtifacts,omitempty"`
	MtaDeploymentRepositoryPassword string   `json:"mtaDeploymentRepositoryPassword,omitempty"`
	MtaDeploymentRepositoryUser     string   `json:"mtaDeploymentRepositoryUser,omitempty"`
	MtaDeploymentRepositoryURL      string   `json:"mtaDeploymentRepositoryUrl,omitempty"`
	Publish                         bool     `json:"publish,omitempty"`
	Profiles                        []string `json:"profiles,omitempty"`
}

type mtaBuildCommonPipelineEnvironment struct {
	mtarFilePath string
	custom       struct {
		mtarPublishedURL   string
		publish            bool
		globalSettingsFile string
		defaultNpmRegistry string
		profiles           string
		dockerImage        string
	}
}

func (p *mtaBuildCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    interface{}
	}{
		{category: "", name: "mtarFilePath", value: p.mtarFilePath},
		{category: "custom", name: "mtarPublishedUrl", value: p.custom.mtarPublishedURL},
		{category: "custom", name: "publish", value: p.custom.publish},
		{category: "custom", name: "globalSettingsFile", value: p.custom.globalSettingsFile},
		{category: "custom", name: "defaultNpmRegistry", value: p.custom.defaultNpmRegistry},
		{category: "custom", name: "profiles", value: p.custom.profiles},
		{category: "custom", name: "dockerImage", value: p.custom.dockerImage},
	}

	errCount := 0
	for _, param := range content {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(param.category, param.name), param.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting piper environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Fatal("failed to persist Piper environment")
	}
}

// MtaBuildCommand Performs an mta build
func MtaBuildCommand() *cobra.Command {
	const STEP_NAME = "mtaBuild"

	metadata := mtaBuildMetadata()
	var stepConfig mtaBuildOptions
	var startTime time.Time
	var commonPipelineEnvironment mtaBuildCommonPipelineEnvironment
	var logCollector *log.CollectorHook

	var createMtaBuildCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Performs an mta build",
		Long:  `Executes the SAP Multitarget Application Archive Builder to create an mtar archive of the application.`,
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
			log.RegisterSecret(stepConfig.MtaDeploymentRepositoryPassword)

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
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				commonPipelineEnvironment.persist(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetryData.ErrorCategory = log.GetErrorCategory().String()
				telemetry.Send(&telemetryData)
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunk.Send(&telemetryData, logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunk.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			mtaBuild(stepConfig, &telemetryData, &commonPipelineEnvironment)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addMtaBuildFlags(createMtaBuildCmd, &stepConfig)
	return createMtaBuildCmd
}

func addMtaBuildFlags(cmd *cobra.Command, stepConfig *mtaBuildOptions) {
	cmd.Flags().StringVar(&stepConfig.MtarName, "mtarName", os.Getenv("PIPER_mtarName"), "The name of the generated mtar file including its extension.")
	cmd.Flags().StringVar(&stepConfig.MtarGroup, "mtarGroup", os.Getenv("PIPER_mtarGroup"), "The group to which the mtar artifact will be uploaded. Required when publish is True.")
	cmd.Flags().StringVar(&stepConfig.Version, "version", os.Getenv("PIPER_version"), "Version of the mtar artifact")
	cmd.Flags().StringVar(&stepConfig.Extensions, "extensions", os.Getenv("PIPER_extensions"), "The path to the extension descriptor file.")
	cmd.Flags().StringVar(&stepConfig.Platform, "platform", `CF`, "The target platform to which the mtar can be deployed.")
	cmd.Flags().StringVar(&stepConfig.ApplicationName, "applicationName", os.Getenv("PIPER_applicationName"), "The name of the application which is being built. If the parameter has been provided and no `mta.yaml` exists, the `mta.yaml` will be automatically generated using this parameter and the information (`name` and `version`) from 'package.json` before the actual build starts.")
	cmd.Flags().StringVar(&stepConfig.Source, "source", `./`, "The path to the MTA project.")
	cmd.Flags().StringVar(&stepConfig.Target, "target", `./`, "The folder for the generated `MTAR` file. If the parameter has been provided, the `MTAR` file is saved in the root of the folder provided by the argument.")
	cmd.Flags().StringVar(&stepConfig.DefaultNpmRegistry, "defaultNpmRegistry", os.Getenv("PIPER_defaultNpmRegistry"), "Url to the npm registry that should be used for installing npm dependencies.")
	cmd.Flags().StringVar(&stepConfig.ProjectSettingsFile, "projectSettingsFile", os.Getenv("PIPER_projectSettingsFile"), "Path or url to the mvn settings file that should be used as project settings file.")
	cmd.Flags().StringVar(&stepConfig.GlobalSettingsFile, "globalSettingsFile", os.Getenv("PIPER_globalSettingsFile"), "Path or url to the mvn settings file that should be used as global settings file")
	cmd.Flags().StringVar(&stepConfig.M2Path, "m2Path", os.Getenv("PIPER_m2Path"), "Path to the location of the local repository that should be used.")
	cmd.Flags().BoolVar(&stepConfig.InstallArtifacts, "installArtifacts", false, "If enabled, for npm packages this step will install all dependencies including dev dependencies. For maven it will install all artifacts to the local maven repository. Note: This happens _after_ mta build was done. The default mta build tool does not install dev-dependencies as part of the process. If you require dev-dependencies for building the mta, you will need to use a [custom builder](https://sap.github.io/cloud-mta-build-tool/configuration/#configuring-the-custom-builder)")
	cmd.Flags().StringVar(&stepConfig.MtaDeploymentRepositoryPassword, "mtaDeploymentRepositoryPassword", os.Getenv("PIPER_mtaDeploymentRepositoryPassword"), "Password for the alternative deployment repository to which mtar artifacts will be publised")
	cmd.Flags().StringVar(&stepConfig.MtaDeploymentRepositoryUser, "mtaDeploymentRepositoryUser", os.Getenv("PIPER_mtaDeploymentRepositoryUser"), "User for the alternative deployment repository to which which mtar artifacts will be publised")
	cmd.Flags().StringVar(&stepConfig.MtaDeploymentRepositoryURL, "mtaDeploymentRepositoryUrl", os.Getenv("PIPER_mtaDeploymentRepositoryUrl"), "Url for the alternative deployment repository to which mtar artifacts will be publised")
	cmd.Flags().BoolVar(&stepConfig.Publish, "publish", false, "pushed mtar artifact to altDeploymentRepositoryUrl/altDeploymentRepositoryID when set to true")
	cmd.Flags().StringSliceVar(&stepConfig.Profiles, "profiles", []string{}, "Defines list of maven build profiles to be used. profiles will overwrite existing values in the global settings xml at $M2_HOME/conf/settings.xml")

}

// retrieve step metadata
func mtaBuildMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "mtaBuild",
			Aliases:     []config.Alias{},
			Description: "Performs an mta build",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "mtarName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_mtarName"),
					},
					{
						Name:        "mtarGroup",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_mtarGroup"),
					},
					{
						Name: "version",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "artifactVersion",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "artifactVersion"}},
						Default:   os.Getenv("PIPER_version"),
					},
					{
						Name:        "extensions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "extension"}},
						Default:     os.Getenv("PIPER_extensions"),
					},
					{
						Name:        "platform",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `CF`,
					},
					{
						Name:        "applicationName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_applicationName"),
					},
					{
						Name:        "source",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `./`,
					},
					{
						Name:        "target",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `./`,
					},
					{
						Name:        "defaultNpmRegistry",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "npm/defaultNpmRegistry"}},
						Default:     os.Getenv("PIPER_defaultNpmRegistry"),
					},
					{
						Name:        "projectSettingsFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/projectSettingsFile"}},
						Default:     os.Getenv("PIPER_projectSettingsFile"),
					},
					{
						Name:        "globalSettingsFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/globalSettingsFile"}},
						Default:     os.Getenv("PIPER_globalSettingsFile"),
					},
					{
						Name:        "m2Path",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/m2Path"}},
						Default:     os.Getenv("PIPER_m2Path"),
					},
					{
						Name:        "installArtifacts",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name: "mtaDeploymentRepositoryPassword",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryPassword",
							},

							{
								Name: "mtaDeploymentRepositoryPasswordId",
								Type: "secret",
							},

							{
								Name:    "mtaDeploymentRepositoryPasswordFileVaultSecretName",
								Type:    "vaultSecretFile",
								Default: "mta-deployment-repository-passowrd",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_mtaDeploymentRepositoryPassword"),
					},
					{
						Name: "mtaDeploymentRepositoryUser",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryUsername",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_mtaDeploymentRepositoryUser"),
					},
					{
						Name: "mtaDeploymentRepositoryUrl",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryUrl",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_mtaDeploymentRepositoryUrl"),
					},
					{
						Name:        "publish",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "mta/publish"}},
						Default:     false,
					},
					{
						Name:        "profiles",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "GENERAL", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
				},
			},
			Containers: []config.Container{
				{Image: "devxci/mbtci:1.1.1"},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "commonPipelineEnvironment",
						Type: "piperEnvironment",
						Parameters: []map[string]interface{}{
							{"Name": "mtarFilePath"},
							{"Name": "custom/mtarPublishedUrl"},
							{"Name": "custom/publish"},
							{"Name": "custom/globalSettingsFile"},
							{"Name": "custom/defaultNpmRegistry"},
							{"Name": "custom/profiles"},
							{"Name": "custom/dockerImage"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
