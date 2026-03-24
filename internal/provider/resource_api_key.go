package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &APIKeyResource{}
var _ resource.ResourceWithImportState = &APIKeyResource{}

func NewAPIKeyResource() resource.Resource {
	return &APIKeyResource{}
}

type APIKeyResource struct {
	client *OpenRouterClient
}

// APIKeyResourceModel maps the resource schema to Go types.
type APIKeyResourceModel struct {
	ID                 types.String  `tfsdk:"id"`
	Hash               types.String  `tfsdk:"hash"`
	Name               types.String  `tfsdk:"name"`
	Label              types.String  `tfsdk:"label"`
	Key                types.String  `tfsdk:"key"`
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

// apiKeyAPIResponse mirrors the OpenRouter API response for a key.
type apiKeyAPIResponse struct {
	Hash               string   `json:"hash"`
	Name               string   `json:"name"`
	Label              string   `json:"label"`
	Disabled           bool     `json:"disabled"`
	Limit              *float64 `json:"limit"`
	LimitReset         *string  `json:"limit_reset"`
	IncludeByokInLimit bool     `json:"include_byok_in_limit"`
	ExpiresAt          *string  `json:"expires_at"`
	LimitRemaining     *float64 `json:"limit_remaining"`
	Usage              float64  `json:"usage"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          *string  `json:"updated_at"`
}

// apiKeyCreateResponse is the POST /keys response, which includes the raw key value.
type apiKeyCreateResponse struct {
	Key  string            `json:"key"`
	Data apiKeyAPIResponse `json:"data"`
}

func (r *APIKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *APIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an OpenRouter API key. Requires a management API key in the provider configuration.\n\n" +
			"~> **Note:** The raw API key value (`key` attribute) is only available immediately after creation. " +
			"It cannot be retrieved later. If you lose the Terraform state, the key value is unrecoverable — " +
			"you will need to create a new key. When importing an existing key, the `key` attribute will be null.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique hash identifier of the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hash": schema.StringAttribute{
				Computed:    true,
				Description: "The unique hash identifier of the API key (same as id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Display name for the API key.",
			},
			"label": schema.StringAttribute{
				Computed:    true,
				Description: "Human-readable label for the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The raw API key value (e.g. sk-or-v1-...). Only populated after creation. Cannot be recovered after the initial apply.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the API key is disabled.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"limit": schema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Credit limit in USD. Null means unlimited.",
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"limit_reset": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "How often the credit limit resets. Valid values: `daily`, `weekly`, `monthly`. Null means never resets.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"include_byok_in_limit": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to count bring-your-own-key (external provider) usage against the limit.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"expires_at": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "ISO 8601 expiration timestamp for the key. Null means no expiration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"limit_remaining": schema.Float64Attribute{
				Computed:    true,
				Description: "Remaining credit balance in USD. Updated on each read.",
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"usage": schema.Float64Attribute{
				Computed:    true,
				Description: "Total usage in USD. Updated on each read.",
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp when the key was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp when the key was last updated.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *APIKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*OpenRouterClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *OpenRouterClient, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *APIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client.ManagementAPIKey == "" {
		resp.Diagnostics.AddError("Management API Key Required",
			"openrouter_api_key requires a management_api_key in the provider configuration or OPENROUTER_MANAGEMENT_API_KEY environment variable.")
		return
	}

	var plan APIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]any{
		"name": plan.Name.ValueString(),
	}
	if !plan.Limit.IsNull() && !plan.Limit.IsUnknown() {
		body["limit"] = plan.Limit.ValueFloat64()
	}
	if !plan.LimitReset.IsNull() && !plan.LimitReset.IsUnknown() {
		body["limit_reset"] = plan.LimitReset.ValueString()
	}
	if !plan.IncludeByokInLimit.IsNull() && !plan.IncludeByokInLimit.IsUnknown() {
		body["include_byok_in_limit"] = plan.IncludeByokInLimit.ValueBool()
	}
	if !plan.ExpiresAt.IsNull() && !plan.ExpiresAt.IsUnknown() {
		body["expires_at"] = plan.ExpiresAt.ValueString()
	}

	var createResp apiKeyCreateResponse
	if err := r.client.mgmtPost(ctx, "/keys", body, &createResp); err != nil {
		resp.Diagnostics.AddError("Error Creating API Key", err.Error())
		return
	}

	state := apiKeyResponseToModel(createResp.Data)
	// The raw key value is only available at creation time.
	state.Key = types.StringValue(createResp.Key)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *APIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var currentState APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &currentState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var env envelope[apiKeyAPIResponse]
	err := r.client.mgmtGet(ctx, "/keys/"+currentState.ID.ValueString(), &env)
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading API Key", err.Error())
		return
	}

	newState := apiKeyResponseToModel(env.Data)
	// Preserve the key — it is never returned by subsequent API calls.
	newState.Key = currentState.Key

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *APIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state APIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build patch body with only fields that have changed.
	patch := map[string]any{}

	if !plan.Name.Equal(state.Name) {
		patch["name"] = plan.Name.ValueString()
	}
	if !plan.Disabled.Equal(state.Disabled) {
		patch["disabled"] = plan.Disabled.ValueBool()
	}
	if !plan.IncludeByokInLimit.Equal(state.IncludeByokInLimit) {
		patch["include_byok_in_limit"] = plan.IncludeByokInLimit.ValueBool()
	}
	if !plan.Limit.Equal(state.Limit) {
		if plan.Limit.IsNull() {
			patch["limit"] = nil
		} else {
			patch["limit"] = plan.Limit.ValueFloat64()
		}
	}
	if !plan.LimitReset.Equal(state.LimitReset) {
		if plan.LimitReset.IsNull() {
			patch["limit_reset"] = nil
		} else {
			patch["limit_reset"] = plan.LimitReset.ValueString()
		}
	}

	var env envelope[apiKeyAPIResponse]
	if err := r.client.mgmtPatch(ctx, "/keys/"+state.ID.ValueString(), patch, &env); err != nil {
		resp.Diagnostics.AddError("Error Updating API Key", err.Error())
		return
	}

	newState := apiKeyResponseToModel(env.Data)
	// Preserve the key value across updates.
	newState.Key = state.Key

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *APIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.mgmtDelete(ctx, "/keys/"+state.ID.ValueString()); err != nil {
		if IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error Deleting API Key", err.Error())
	}
}

func (r *APIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by hash. The raw key value will be null after import — it cannot be recovered.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// apiKeyResponseToModel converts an API response struct to the Terraform state model.
func apiKeyResponseToModel(data apiKeyAPIResponse) APIKeyResourceModel {
	m := APIKeyResourceModel{
		ID:                 types.StringValue(data.Hash),
		Hash:               types.StringValue(data.Hash),
		Name:               types.StringValue(data.Name),
		Label:              types.StringValue(data.Label),
		Disabled:           types.BoolValue(data.Disabled),
		IncludeByokInLimit: types.BoolValue(data.IncludeByokInLimit),
		Usage:              types.Float64Value(data.Usage),
	}

	if data.Limit != nil {
		m.Limit = types.Float64Value(*data.Limit)
	} else {
		m.Limit = types.Float64Null()
	}

	if data.LimitReset != nil {
		m.LimitReset = types.StringValue(*data.LimitReset)
	} else {
		m.LimitReset = types.StringNull()
	}

	if data.LimitRemaining != nil {
		m.LimitRemaining = types.Float64Value(*data.LimitRemaining)
	} else {
		m.LimitRemaining = types.Float64Null()
	}

	if data.ExpiresAt != nil {
		m.ExpiresAt = types.StringValue(*data.ExpiresAt)
	} else {
		m.ExpiresAt = types.StringNull()
	}

	m.CreatedAt = types.StringValue(data.CreatedAt)

	if data.UpdatedAt != nil {
		m.UpdatedAt = types.StringValue(*data.UpdatedAt)
	} else {
		m.UpdatedAt = types.StringNull()
	}

	return m
}
