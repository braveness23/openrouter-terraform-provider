package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ModelsDataSource{}

func NewModelsDataSource() datasource.DataSource {
	return &ModelsDataSource{}
}

type ModelsDataSource struct {
	client *OpenRouterClient
}

// modelAPIResponse mirrors the OpenRouter model object.
type modelAPIResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	ContextLength int64   `json:"context_length"`
	Created       int64   `json:"created"`
	Pricing       struct {
		Prompt     string `json:"prompt"`
		Completion string `json:"completion"`
		Request    string `json:"request"`
		Image      string `json:"image"`
	} `json:"pricing"`
	Architecture struct {
		Tokenizer    string `json:"tokenizer"`
		InstructType string `json:"instruct_type"`
	} `json:"architecture"`
	TopProvider struct {
		ContextLength       *int64 `json:"context_length"`
		MaxCompletionTokens *int64 `json:"max_completion_tokens"`
		IsModerated         bool   `json:"is_moderated"`
	} `json:"top_provider"`
	SupportedParameters []string `json:"supported_parameters"`
}

type modelsListResponse struct {
	Data []modelAPIResponse `json:"data"`
}

// ModelsDataSourceModel is the Terraform schema model for the data source.
type ModelsDataSourceModel struct {
	SupportedParameters types.List   `tfsdk:"supported_parameters"`
	Models              types.List   `tfsdk:"models"`
}

// modelAttrTypes defines the attribute types for a single model object in the list.
var modelAttrTypes = map[string]attr.Type{
	"id":                   types.StringType,
	"name":                 types.StringType,
	"description":          types.StringType,
	"context_length":       types.Int64Type,
	"prompt_price":         types.StringType,
	"completion_price":     types.StringType,
	"request_price":        types.StringType,
	"image_price":          types.StringType,
	"tokenizer":            types.StringType,
	"instruct_type":        types.StringType,
	"is_moderated":         types.BoolType,
	"supported_parameters": types.ListType{ElemType: types.StringType},
}

func (d *ModelsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_models"
}

func (d *ModelsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the list of all available models from OpenRouter.",
		Attributes: map[string]schema.Attribute{
			"supported_parameters": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Filter models to only those supporting all listed parameters (e.g. `temperature`, `tools`).",
			},
			"models": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of available models.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique model identifier (e.g. `openai/gpt-4o`).",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Human-readable model name.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Model description.",
						},
						"context_length": schema.Int64Attribute{
							Computed:    true,
							Description: "Maximum context window in tokens.",
						},
						"prompt_price": schema.StringAttribute{
							Computed:    true,
							Description: "Price per prompt token in USD (as a string to preserve precision).",
						},
						"completion_price": schema.StringAttribute{
							Computed:    true,
							Description: "Price per completion token in USD.",
						},
						"request_price": schema.StringAttribute{
							Computed:    true,
							Description: "Price per request in USD (for fixed-price models).",
						},
						"image_price": schema.StringAttribute{
							Computed:    true,
							Description: "Price per image in USD (for vision models).",
						},
						"tokenizer": schema.StringAttribute{
							Computed:    true,
							Description: "Tokenizer type used by the model.",
						},
						"instruct_type": schema.StringAttribute{
							Computed:    true,
							Description: "Instruction format expected by the model.",
						},
						"is_moderated": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the model applies content moderation.",
						},
						"supported_parameters": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Parameters supported by this model.",
						},
					},
				},
			},
		},
	}
}

func (d *ModelsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ModelsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ModelsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result modelsListResponse
	if err := d.client.inferenceGet(ctx, "/models", &result); err != nil {
		resp.Diagnostics.AddError("Error Reading Models", err.Error())
		return
	}

	// Apply client-side filtering on supported_parameters if provided.
	filtered := result.Data
	if !config.SupportedParameters.IsNull() && !config.SupportedParameters.IsUnknown() {
		var filterParams []string
		resp.Diagnostics.Append(config.SupportedParameters.ElementsAs(ctx, &filterParams, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		filtered = filterModelsByParameters(filtered, filterParams)
	}

	modelObjects := make([]attr.Value, 0, len(filtered))
	for _, m := range filtered {
		paramList, diags := stringsToList(ctx, m.SupportedParameters)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		obj, diags := types.ObjectValue(modelAttrTypes, map[string]attr.Value{
			"id":                   types.StringValue(m.ID),
			"name":                 types.StringValue(m.Name),
			"description":          types.StringValue(m.Description),
			"context_length":       types.Int64Value(m.ContextLength),
			"prompt_price":         types.StringValue(m.Pricing.Prompt),
			"completion_price":     types.StringValue(m.Pricing.Completion),
			"request_price":        types.StringValue(m.Pricing.Request),
			"image_price":          types.StringValue(m.Pricing.Image),
			"tokenizer":            types.StringValue(m.Architecture.Tokenizer),
			"instruct_type":        types.StringValue(m.Architecture.InstructType),
			"is_moderated":         types.BoolValue(m.TopProvider.IsModerated),
			"supported_parameters": paramList,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		modelObjects = append(modelObjects, obj)
	}

	modelsList, diags := types.ListValue(types.ObjectType{AttrTypes: modelAttrTypes}, modelObjects)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := ModelsDataSourceModel{
		SupportedParameters: config.SupportedParameters,
		Models:              modelsList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// filterModelsByParameters returns only models that support all of the given parameters.
func filterModelsByParameters(models []modelAPIResponse, params []string) []modelAPIResponse {
	if len(params) == 0 {
		return models
	}
	var result []modelAPIResponse
	for _, m := range models {
		supported := make(map[string]bool, len(m.SupportedParameters))
		for _, p := range m.SupportedParameters {
			supported[p] = true
		}
		match := true
		for _, p := range params {
			if !supported[p] {
				match = false
				break
			}
		}
		if match {
			result = append(result, m)
		}
	}
	return result
}
