package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &GuardrailResource{}
var _ resource.ResourceWithImportState = &GuardrailResource{}

func NewGuardrailResource() resource.Resource {
	return &GuardrailResource{}
}

type GuardrailResource struct {
	client *OpenRouterClient
}

type GuardrailResourceModel struct {
	ID               types.String  `tfsdk:"id"`
	Name             types.String  `tfsdk:"name"`
	Description      types.String  `tfsdk:"description"`
	LimitUSD         types.Float64 `tfsdk:"limit_usd"`
	ResetInterval    types.String  `tfsdk:"reset_interval"`
	AllowedProviders types.List    `tfsdk:"allowed_providers"`
	IgnoredProviders types.List    `tfsdk:"ignored_providers"`
	AllowedModels    types.List    `tfsdk:"allowed_models"`
	EnforceZDR       types.Bool    `tfsdk:"enforce_zdr"`
	CreatedAt        types.String  `tfsdk:"created_at"`
	UpdatedAt        types.String  `tfsdk:"updated_at"`
}

type guardrailAPIResponse struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Description      *string   `json:"description"`
	LimitUSD         *float64  `json:"limit_usd"`
	ResetInterval    *string   `json:"reset_interval"`
	AllowedProviders []string  `json:"allowed_providers"`
	IgnoredProviders []string  `json:"ignored_providers"`
	AllowedModels    []string  `json:"allowed_models"`
	EnforceZDR       *bool     `json:"enforce_zdr"`
	CreatedAt        string    `json:"created_at"`
	UpdatedAt        *string   `json:"updated_at"`
}

func (r *GuardrailResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_guardrail"
}

func (r *GuardrailResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an OpenRouter guardrail. Guardrails enforce spending limits and restrict which providers or models can be used. Requires a management API key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The UUID identifier of the guardrail.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Display name for the guardrail.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional description of the guardrail.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"limit_usd": schema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Spending cap in USD. Null means unlimited.",
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"reset_interval": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "How often the spending limit resets. Valid values: `daily`, `weekly`, `monthly`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"allowed_providers": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of provider IDs to permit. If set, only these providers are allowed.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"ignored_providers": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of provider IDs to block.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"allowed_models": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of model canonical slugs to allow. If set, only these models can be used.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"enforce_zdr": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Require zero data retention (ZDR) across all providers.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp when the guardrail was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp when the guardrail was last updated.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *GuardrailResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GuardrailResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client.ManagementAPIKey == "" {
		resp.Diagnostics.AddError("Management API Key Required",
			"openrouter_guardrail requires a management_api_key in the provider configuration or OPENROUTER_MANAGEMENT_API_KEY environment variable.")
		return
	}

	var plan GuardrailResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]any{
		"name": plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}
	if !plan.LimitUSD.IsNull() && !plan.LimitUSD.IsUnknown() {
		body["limit_usd"] = plan.LimitUSD.ValueFloat64()
	}
	if !plan.ResetInterval.IsNull() && !plan.ResetInterval.IsUnknown() {
		body["reset_interval"] = plan.ResetInterval.ValueString()
	}
	if !plan.EnforceZDR.IsNull() && !plan.EnforceZDR.IsUnknown() {
		body["enforce_zdr"] = plan.EnforceZDR.ValueBool()
	}
	if !plan.AllowedProviders.IsNull() && !plan.AllowedProviders.IsUnknown() {
		body["allowed_providers"] = listToStrings(ctx, plan.AllowedProviders)
	}
	if !plan.IgnoredProviders.IsNull() && !plan.IgnoredProviders.IsUnknown() {
		body["ignored_providers"] = listToStrings(ctx, plan.IgnoredProviders)
	}
	if !plan.AllowedModels.IsNull() && !plan.AllowedModels.IsUnknown() {
		body["allowed_models"] = listToStrings(ctx, plan.AllowedModels)
	}

	var env envelope[guardrailAPIResponse]
	if err := r.client.mgmtPost(ctx, "/guardrails", body, &env); err != nil {
		resp.Diagnostics.AddError("Error Creating Guardrail", err.Error())
		return
	}

	state, diags := guardrailResponseToModel(ctx, env.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GuardrailResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var currentState GuardrailResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &currentState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var env envelope[guardrailAPIResponse]
	if err := r.client.mgmtGet(ctx, "/guardrails/"+currentState.ID.ValueString(), &env); err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Guardrail", err.Error())
		return
	}

	state, diags := guardrailResponseToModel(ctx, env.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GuardrailResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state GuardrailResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	patch := map[string]any{}

	if !plan.Name.Equal(state.Name) {
		patch["name"] = plan.Name.ValueString()
	}
	if !plan.Description.Equal(state.Description) {
		if plan.Description.IsNull() {
			patch["description"] = nil
		} else {
			patch["description"] = plan.Description.ValueString()
		}
	}
	if !plan.LimitUSD.Equal(state.LimitUSD) {
		if plan.LimitUSD.IsNull() {
			patch["limit_usd"] = nil
		} else {
			patch["limit_usd"] = plan.LimitUSD.ValueFloat64()
		}
	}
	if !plan.ResetInterval.Equal(state.ResetInterval) {
		if plan.ResetInterval.IsNull() {
			patch["reset_interval"] = nil
		} else {
			patch["reset_interval"] = plan.ResetInterval.ValueString()
		}
	}
	if !plan.EnforceZDR.Equal(state.EnforceZDR) {
		patch["enforce_zdr"] = plan.EnforceZDR.ValueBool()
	}
	if !plan.AllowedProviders.Equal(state.AllowedProviders) {
		patch["allowed_providers"] = listToStrings(ctx, plan.AllowedProviders)
	}
	if !plan.IgnoredProviders.Equal(state.IgnoredProviders) {
		patch["ignored_providers"] = listToStrings(ctx, plan.IgnoredProviders)
	}
	if !plan.AllowedModels.Equal(state.AllowedModels) {
		patch["allowed_models"] = listToStrings(ctx, plan.AllowedModels)
	}

	var env envelope[guardrailAPIResponse]
	if err := r.client.mgmtPatch(ctx, "/guardrails/"+state.ID.ValueString(), patch, &env); err != nil {
		resp.Diagnostics.AddError("Error Updating Guardrail", err.Error())
		return
	}

	newState, diags := guardrailResponseToModel(ctx, env.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *GuardrailResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GuardrailResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.mgmtDelete(ctx, "/guardrails/"+state.ID.ValueString()); err != nil {
		if IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error Deleting Guardrail", err.Error())
	}
}

func (r *GuardrailResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func guardrailResponseToModel(ctx context.Context, data guardrailAPIResponse) (GuardrailResourceModel, diag.Diagnostics) {
	m := GuardrailResourceModel{
		ID:        types.StringValue(data.ID),
		Name:      types.StringValue(data.Name),
		CreatedAt: types.StringValue(data.CreatedAt),
	}

	if data.Description != nil {
		m.Description = types.StringValue(*data.Description)
	} else {
		m.Description = types.StringNull()
	}

	if data.LimitUSD != nil {
		m.LimitUSD = types.Float64Value(*data.LimitUSD)
	} else {
		m.LimitUSD = types.Float64Null()
	}

	if data.ResetInterval != nil {
		m.ResetInterval = types.StringValue(*data.ResetInterval)
	} else {
		m.ResetInterval = types.StringNull()
	}

	if data.EnforceZDR != nil {
		m.EnforceZDR = types.BoolValue(*data.EnforceZDR)
	} else {
		m.EnforceZDR = types.BoolNull()
	}

	if data.UpdatedAt != nil {
		m.UpdatedAt = types.StringValue(*data.UpdatedAt)
	} else {
		m.UpdatedAt = types.StringNull()
	}

	var diags diag.Diagnostics

	allowedProviders, d := stringsToList(ctx, data.AllowedProviders)
	diags = append(diags, d...)
	m.AllowedProviders = allowedProviders

	ignoredProviders, d := stringsToList(ctx, data.IgnoredProviders)
	diags = append(diags, d...)
	m.IgnoredProviders = ignoredProviders

	allowedModels, d := stringsToList(ctx, data.AllowedModels)
	diags = append(diags, d...)
	m.AllowedModels = allowedModels

	return m, diags
}

// listToStrings converts a types.List of strings to a []string.
func listToStrings(ctx context.Context, list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var result []string
	_ = list.ElementsAs(ctx, &result, false)
	return result
}

// stringsToList converts a []string to a types.List of strings.
func stringsToList(ctx context.Context, strs []string) (types.List, diag.Diagnostics) {
	if strs == nil {
		return types.ListNull(types.StringType), nil
	}
	elems := make([]attr.Value, len(strs))
	for i, s := range strs {
		elems[i] = types.StringValue(s)
	}
	return types.ListValue(types.StringType, elems)
}
