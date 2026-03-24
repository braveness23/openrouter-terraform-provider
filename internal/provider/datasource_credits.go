package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CreditsDataSource{}

func NewCreditsDataSource() datasource.DataSource {
	return &CreditsDataSource{}
}

type CreditsDataSource struct {
	client *OpenRouterClient
}

type CreditsDataSourceModel struct {
	TotalCredits types.Float64 `tfsdk:"total_credits"`
	TotalUsage   types.Float64 `tfsdk:"total_usage"`
}

type creditsAPIResponse struct {
	TotalCredits float64 `json:"total_credits"`
	TotalUsage   float64 `json:"total_usage"`
}

func (d *CreditsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credits"
}

func (d *CreditsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the current account credit balance from OpenRouter. Requires a management API key.",
		Attributes: map[string]schema.Attribute{
			"total_credits": schema.Float64Attribute{
				Computed:    true,
				Description: "Total credits purchased in USD.",
			},
			"total_usage": schema.Float64Attribute{
				Computed:    true,
				Description: "Total credits consumed in USD.",
			},
		},
	}
}

func (d *CreditsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*OpenRouterClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *OpenRouterClient, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *CreditsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client.ManagementAPIKey == "" {
		resp.Diagnostics.AddError("Management API Key Required",
			"openrouter_credits requires a management_api_key in the provider configuration or OPENROUTER_MANAGEMENT_API_KEY environment variable.")
		return
	}

	var env envelope[creditsAPIResponse]
	if err := d.client.mgmtGet(ctx, "/credits", &env); err != nil {
		resp.Diagnostics.AddError("Error Reading Credits", err.Error())
		return
	}

	state := CreditsDataSourceModel{
		TotalCredits: types.Float64Value(env.Data.TotalCredits),
		TotalUsage:   types.Float64Value(env.Data.TotalUsage),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
