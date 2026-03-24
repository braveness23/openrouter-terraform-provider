package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ActivityDataSource{}

func NewActivityDataSource() datasource.DataSource {
	return &ActivityDataSource{}
}

type ActivityDataSource struct {
	client *OpenRouterClient
}

type ActivityDataSourceModel struct {
	Date     types.String `tfsdk:"date"`
	Activity types.List   `tfsdk:"activity"`
}

type activityRowAPIResponse struct {
	Date                  string   `json:"date"`
	Model                 string   `json:"model"`
	ModelPermaslug        string   `json:"model_permaslug"`
	EndpointID            string   `json:"endpoint_id"`
	ProviderName          string   `json:"provider_name"`
	Usage                 float64  `json:"usage"`
	ByokUsageInference    float64  `json:"byok_usage_inference"`
	Requests              int64    `json:"requests"`
	PromptTokens          int64    `json:"prompt_tokens"`
	CompletionTokens      int64    `json:"completion_tokens"`
	ReasoningTokens       *int64   `json:"reasoning_tokens"`
}

type activityAPIResponse struct {
	Data []activityRowAPIResponse `json:"data"`
}

var activityRowAttrTypes = map[string]attr.Type{
	"date":                   types.StringType,
	"model":                  types.StringType,
	"model_permaslug":        types.StringType,
	"endpoint_id":            types.StringType,
	"provider_name":          types.StringType,
	"usage":                  types.Float64Type,
	"byok_usage_inference":   types.Float64Type,
	"requests":               types.Int64Type,
	"prompt_tokens":          types.Int64Type,
	"completion_tokens":      types.Int64Type,
	"reasoning_tokens":       types.Int64Type,
}

func (d *ActivityDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_activity"
}

func (d *ActivityDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves API usage activity. Returns the last 30 days of activity, or a single day when `date` is specified. Requires a management API key.",
		Attributes: map[string]schema.Attribute{
			"date": schema.StringAttribute{
				Optional:    true,
				Description: "Filter to a single day in YYYY-MM-DD format. Must be within the last 30 days. If omitted, returns all available activity.",
			},
			"activity": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Activity rows grouped by model, provider, and endpoint.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"date":             schema.StringAttribute{Computed: true, Description: "Date of activity (YYYY-MM-DD)."},
						"model":            schema.StringAttribute{Computed: true, Description: "Model ID used."},
						"model_permaslug":  schema.StringAttribute{Computed: true, Description: "Versioned model identifier."},
						"endpoint_id":      schema.StringAttribute{Computed: true, Description: "Endpoint identifier."},
						"provider_name":    schema.StringAttribute{Computed: true, Description: "Provider name."},
						"usage":            schema.Float64Attribute{Computed: true, Description: "OpenRouter credits consumed (USD)."},
						"byok_usage_inference": schema.Float64Attribute{Computed: true, Description: "External provider credits consumed (USD)."},
						"requests":         schema.Int64Attribute{Computed: true, Description: "Number of requests."},
						"prompt_tokens":    schema.Int64Attribute{Computed: true, Description: "Total prompt tokens."},
						"completion_tokens": schema.Int64Attribute{Computed: true, Description: "Total completion tokens."},
						"reasoning_tokens": schema.Int64Attribute{Computed: true, Description: "Total reasoning tokens (where applicable)."},
					},
				},
			},
		},
	}
}

func (d *ActivityDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ActivityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client.ManagementAPIKey == "" {
		resp.Diagnostics.AddError("Management API Key Required",
			"openrouter_activity requires a management_api_key in the provider configuration or OPENROUTER_MANAGEMENT_API_KEY environment variable.")
		return
	}

	var config ActivityDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := "/activity"
	if !config.Date.IsNull() && !config.Date.IsUnknown() {
		path += "?date=" + config.Date.ValueString()
	}

	var result activityAPIResponse
	if err := d.client.mgmtGet(ctx, path, &result); err != nil {
		resp.Diagnostics.AddError("Error Reading Activity", err.Error())
		return
	}

	rows := make([]attr.Value, 0, len(result.Data))
	for _, row := range result.Data {
		reasoningTokens := types.Int64Null()
		if row.ReasoningTokens != nil {
			reasoningTokens = types.Int64Value(*row.ReasoningTokens)
		}

		obj, diags := types.ObjectValue(activityRowAttrTypes, map[string]attr.Value{
			"date":                   types.StringValue(row.Date),
			"model":                  types.StringValue(row.Model),
			"model_permaslug":        types.StringValue(row.ModelPermaslug),
			"endpoint_id":            types.StringValue(row.EndpointID),
			"provider_name":          types.StringValue(row.ProviderName),
			"usage":                  types.Float64Value(row.Usage),
			"byok_usage_inference":   types.Float64Value(row.ByokUsageInference),
			"requests":               types.Int64Value(row.Requests),
			"prompt_tokens":          types.Int64Value(row.PromptTokens),
			"completion_tokens":      types.Int64Value(row.CompletionTokens),
			"reasoning_tokens":       reasoningTokens,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		rows = append(rows, obj)
	}

	activityList, diags := types.ListValue(types.ObjectType{AttrTypes: activityRowAttrTypes}, rows)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := ActivityDataSourceModel{
		Date:     config.Date,
		Activity: activityList,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
