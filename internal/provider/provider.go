package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &OpenRouterProvider{}
var _ provider.ProviderWithFunctions = &OpenRouterProvider{}

type OpenRouterProvider struct {
	version string
}

type OpenRouterProviderModel struct {
	APIKey           types.String `tfsdk:"api_key"`
	ManagementAPIKey types.String `tfsdk:"management_api_key"`
	BaseURL          types.String `tfsdk:"base_url"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OpenRouterProvider{version: version}
	}
}

func (p *OpenRouterProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "openrouter"
	resp.Version = p.version
}

func (p *OpenRouterProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with the OpenRouter AI API to manage API keys, guardrails, and query model information.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Standard OpenRouter API key for inference endpoints (/models, etc.). May also be set via the OPENROUTER_API_KEY environment variable.",
			},
			"management_api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "OpenRouter management API key, required for managing API keys, guardrails, and credits. May also be set via the OPENROUTER_MANAGEMENT_API_KEY environment variable.",
			},
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "OpenRouter API base URL. Defaults to https://openrouter.ai/api/v1.",
			},
		},
	}
}

func (p *OpenRouterProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config OpenRouterProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if !config.APIKey.IsNull() && !config.APIKey.IsUnknown() {
		apiKey = config.APIKey.ValueString()
	}

	mgmtKey := os.Getenv("OPENROUTER_MANAGEMENT_API_KEY")
	if !config.ManagementAPIKey.IsNull() && !config.ManagementAPIKey.IsUnknown() {
		mgmtKey = config.ManagementAPIKey.ValueString()
	}

	baseURL := "https://openrouter.ai/api/v1"
	if !config.BaseURL.IsNull() && !config.BaseURL.IsUnknown() {
		baseURL = config.BaseURL.ValueString()
	}

	client := &OpenRouterClient{
		APIKey:           apiKey,
		ManagementAPIKey: mgmtKey,
		BaseURL:          baseURL,
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *OpenRouterProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAPIKeyResource,
		NewGuardrailResource,
	}
}

func (p *OpenRouterProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCreditsDataSource,
		NewModelsDataSource,
		NewModelDataSource,
		NewAPIKeysDataSource,
		NewAPIKeyDataSource,
	}
}

func (p *OpenRouterProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{}
}
