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
	APIKey  types.String `tfsdk:"api_key"`
	BaseURL types.String `tfsdk:"base_url"`
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
		Description: "Interact with OpenRouter AI API.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "OpenRouter API key. May also be set via OPENROUTER_API_KEY environment variable.",
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
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	baseURL := "https://openrouter.ai/api/v1"
	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	client := &OpenRouterClient{
		APIKey:  apiKey,
		BaseURL: baseURL,
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *OpenRouterProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *OpenRouterProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *OpenRouterProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{}
}
