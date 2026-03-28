package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
	"github.com/realnedsanders/terraform-provider-vastai/internal/services/apikey"
	"github.com/realnedsanders/terraform-provider-vastai/internal/services/envvar"
	"github.com/realnedsanders/terraform-provider-vastai/internal/services/subaccount"
)

// Ensure VastaiProvider satisfies the provider.Provider interface.
var _ provider.Provider = &VastaiProvider{}

// VastaiProvider defines the provider implementation.
type VastaiProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// VastaiProviderModel describes the provider data model.
type VastaiProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
	APIURL types.String `tfsdk:"api_url"`
}

// New returns a function that creates a new VastaiProvider instance.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &VastaiProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name and version.
func (p *VastaiProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "vastai"
	resp.Version = p.version
}

// Schema defines the provider-level schema.
func (p *VastaiProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Vast.ai GPU compute infrastructure.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "Vast.ai API key. Can also be set via VASTAI_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_url": schema.StringAttribute{
				Description: "Vast.ai API base URL. Can also be set via VASTAI_API_URL environment variable. Defaults to https://console.vast.ai.",
				Optional:    true,
			},
		},
	}
}

// Configure prepares the Vast.ai API client for data sources and resources.
func (p *VastaiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config VastaiProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for unknown values before accessing them. Unknown values can occur
	// when a provider attribute references a resource attribute that has not
	// yet been computed during the plan phase.
	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddWarning(
			"Unknown Vast.ai API Key",
			"The provider cannot create the Vast.ai API client as there is an unknown configuration value for the Vast.ai API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the VASTAI_API_KEY environment variable.",
		)
		return
	}

	if config.APIURL.IsUnknown() {
		resp.Diagnostics.AddWarning(
			"Unknown Vast.ai API URL",
			"The provider cannot create the Vast.ai API client as there is an unknown configuration value for the Vast.ai API URL. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the VASTAI_API_URL environment variable.",
		)
		return
	}

	// Default values from environment variables, with config overrides.
	apiKey := os.Getenv("VASTAI_API_KEY")
	apiURL := os.Getenv("VASTAI_API_URL")

	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if !config.APIURL.IsNull() {
		apiURL = config.APIURL.ValueString()
	}

	if apiURL == "" {
		apiURL = "https://console.vast.ai"
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing Vast.ai API Key",
			"The provider requires a Vast.ai API key to authenticate. "+
				"Set the api_key attribute in the provider configuration block or use the VASTAI_API_KEY environment variable.",
		)
		return
	}

	// Create API client and inject into resource/data source data.
	c := client.NewVastAIClient(apiKey, apiURL, p.version)
	resp.ResourceData = c
	resp.DataSourceData = c
}

// Resources defines the resources implemented in the provider.
func (p *VastaiProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		apikey.NewApiKeyResource,
		envvar.NewEnvVarResource,
		subaccount.NewSubaccountResource,
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *VastaiProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
