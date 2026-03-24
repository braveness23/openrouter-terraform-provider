package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &APIKeyDataSource{}

func NewAPIKeyDataSource() datasource.DataSource {
	return &APIKeyDataSource{}
}

type APIKeyDataSource struct {
	client *OpenRouterClient
}

type APIKeyDataSourceModel struct {
	Hash               types.String  `tfsdk:"hash"`
	Name               types.String  `tfsdk:"name"`
	Label              types.String  `tfsdk:"label"`
	Disabled           types.Bool    `tfsdk:"disabled"`
	Limit              types.Float64 `tfsdk:"limit"`
	LimitReset         types.String  `tfsdk:"limit_reset"`
	IncludeByokInLimit types.Bool    `tfsdk:"include_byok_in_limit"`
	ExpiresAt          types.String  `tfsdk:"expires_at"`
	LimitRemaining     types.Float64 `tfsdk:"limit_remaining"`
	Usage              types.Float64 `tfsdk:"usage"`
	CreatedAt          types.String  `tfsdk:"created_at"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
}

func (d *APIKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (d *APIKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves metadata for a specific OpenRouter API key by hash. Requires a management API key. Note: the raw key value is never returned by this endpoint.",
		Attributes: map[string]schema.Attribute{
			"hash": schema.StringAttribute{
				Required:    true,
				Description: "The unique hash identifier of the API key to look up.",
			},
			"name":  schema.StringAttribute{Computed: true, Description: "Display name of the key."},
			"label": schema.StringAttribute{Computed: true, Description: "Human-readable label."},
			"disabled": schema.BoolAttribute{Computed: true, Description: "Whether the key is disabled."},
			"limit":    schema.Float64Attribute{Computed: true, Description: "Credit limit in USD."},
			"limit_reset": schema.StringAttribute{Computed: true, Description: "Limit reset interval."},
			"include_byok_in_limit": schema.BoolAttribute{Computed: true, Description: "Whether BYOK usage counts against limit."},
			"expires_at":      schema.StringAttribute{Computed: true, Description: "Expiration timestamp."},
			"limit_remaining": schema.Float64Attribute{Computed: true, Description: "Remaining credit balance in USD."},
			"usage":           schema.Float64Attribute{Computed: true, Description: "Total usage in USD."},
			"created_at":      schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
			"updated_at":      schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
		},
	}
}

func (d *APIKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *APIKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client.ManagementAPIKey == "" {
		resp.Diagnostics.AddError("Management API Key Required",
			"openrouter_api_key data source requires a management_api_key in the provider configuration or OPENROUTER_MANAGEMENT_API_KEY environment variable.")
		return
	}

	var config APIKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var env envelope[apiKeyAPIResponse]
	if err := d.client.mgmtGet(ctx, "/keys/"+config.Hash.ValueString(), &env); err != nil {
		resp.Diagnostics.AddError("Error Reading API Key", err.Error())
		return
	}

	m := apiKeyResponseToModel(env.Data)
	state := APIKeyDataSourceModel{
		Hash:               m.Hash,
		Name:               m.Name,
		Label:              m.Label,
		Disabled:           m.Disabled,
		Limit:              m.Limit,
		LimitReset:         m.LimitReset,
		IncludeByokInLimit: m.IncludeByokInLimit,
		ExpiresAt:          m.ExpiresAt,
		LimitRemaining:     m.LimitRemaining,
		Usage:              m.Usage,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
