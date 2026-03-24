package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &APIKeysDataSource{}

func NewAPIKeysDataSource() datasource.DataSource {
	return &APIKeysDataSource{}
}

type APIKeysDataSource struct {
	client *OpenRouterClient
}

type APIKeysDataSourceModel struct {
	IncludeDisabled types.Bool `tfsdk:"include_disabled"`
	Keys            types.List `tfsdk:"keys"`
}

// apiKeyListResponse is the raw response for GET /keys
type apiKeyListResponse struct {
	Data []apiKeyAPIResponse `json:"data"`
}

// apiKeyItemAttrTypes defines the attribute types for a key item in the list.
var apiKeyItemAttrTypes = map[string]attr.Type{
	"hash":                 types.StringType,
	"name":                 types.StringType,
	"label":                types.StringType,
	"disabled":             types.BoolType,
	"limit":                types.Float64Type,
	"limit_reset":          types.StringType,
	"include_byok_in_limit": types.BoolType,
	"expires_at":           types.StringType,
	"limit_remaining":      types.Float64Type,
	"usage":                types.Float64Type,
	"created_at":           types.StringType,
	"updated_at":           types.StringType,
}

func (d *APIKeysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_keys"
}

func (d *APIKeysDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	keyAttrs := map[string]schema.Attribute{
		"hash":  schema.StringAttribute{Computed: true, Description: "Unique hash identifier of the key."},
		"name":  schema.StringAttribute{Computed: true, Description: "Display name of the key."},
		"label": schema.StringAttribute{Computed: true, Description: "Human-readable label."},
		"disabled": schema.BoolAttribute{Computed: true, Description: "Whether the key is disabled."},
		"limit":    schema.Float64Attribute{Computed: true, Description: "Credit limit in USD."},
		"limit_reset": schema.StringAttribute{Computed: true, Description: "Limit reset interval."},
		"include_byok_in_limit": schema.BoolAttribute{Computed: true, Description: "Whether BYOK usage counts against limit."},
		"expires_at":       schema.StringAttribute{Computed: true, Description: "Expiration timestamp."},
		"limit_remaining":  schema.Float64Attribute{Computed: true, Description: "Remaining credit balance in USD."},
		"usage":            schema.Float64Attribute{Computed: true, Description: "Total usage in USD."},
		"created_at":       schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
		"updated_at":       schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
	}

	resp.Schema = schema.Schema{
		Description: "Retrieves a list of all API keys. Requires a management API key. Note: the raw key values are never returned by this endpoint.",
		Attributes: map[string]schema.Attribute{
			"include_disabled": schema.BoolAttribute{
				Optional:    true,
				Description: "When true, includes disabled keys in the results.",
			},
			"keys": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of API keys.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: keyAttrs,
				},
			},
		},
	}
}

func (d *APIKeysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *APIKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client.ManagementAPIKey == "" {
		resp.Diagnostics.AddError("Management API Key Required",
			"openrouter_api_keys requires a management_api_key in the provider configuration or OPENROUTER_MANAGEMENT_API_KEY environment variable.")
		return
	}

	var config APIKeysDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result apiKeyListResponse
	if err := d.client.mgmtGet(ctx, "/keys", &result); err != nil {
		resp.Diagnostics.AddError("Error Reading API Keys", err.Error())
		return
	}

	keys := result.Data
	if !config.IncludeDisabled.IsNull() && !config.IncludeDisabled.ValueBool() {
		var active []apiKeyAPIResponse
		for _, k := range keys {
			if !k.Disabled {
				active = append(active, k)
			}
		}
		keys = active
	}

	keyObjects := make([]attr.Value, 0, len(keys))
	for _, k := range keys {
		m := apiKeyResponseToModel(k)
		obj, diags := types.ObjectValue(apiKeyItemAttrTypes, map[string]attr.Value{
			"hash":                  m.Hash,
			"name":                  m.Name,
			"label":                 m.Label,
			"disabled":              m.Disabled,
			"limit":                 m.Limit,
			"limit_reset":           m.LimitReset,
			"include_byok_in_limit": m.IncludeByokInLimit,
			"expires_at":            m.ExpiresAt,
			"limit_remaining":       m.LimitRemaining,
			"usage":                 m.Usage,
			"created_at":            m.CreatedAt,
			"updated_at":            m.UpdatedAt,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		keyObjects = append(keyObjects, obj)
	}

	keyList, diags := types.ListValue(types.ObjectType{AttrTypes: apiKeyItemAttrTypes}, keyObjects)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := APIKeysDataSourceModel{
		IncludeDisabled: config.IncludeDisabled,
		Keys:            keyList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
