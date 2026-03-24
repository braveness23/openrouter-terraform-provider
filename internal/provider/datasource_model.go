package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ModelDataSource{}

func NewModelDataSource() datasource.DataSource {
	return &ModelDataSource{}
}

type ModelDataSource struct {
	client *OpenRouterClient
}

type ModelDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	ContextLength       types.Int64  `tfsdk:"context_length"`
	PromptPrice         types.String `tfsdk:"prompt_price"`
	CompletionPrice     types.String `tfsdk:"completion_price"`
	RequestPrice        types.String `tfsdk:"request_price"`
	ImagePrice          types.String `tfsdk:"image_price"`
	Tokenizer           types.String `tfsdk:"tokenizer"`
	InstructType        types.String `tfsdk:"instruct_type"`
	IsModerated         types.Bool   `tfsdk:"is_moderated"`
	SupportedParameters types.List   `tfsdk:"supported_parameters"`
}

func (d *ModelDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model"
}

func (d *ModelDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a specific OpenRouter model by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The model ID to look up (e.g. `openai/gpt-4o`, `anthropic/claude-opus-4`).",
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
				Description: "Price per prompt token in USD.",
			},
			"completion_price": schema.StringAttribute{
				Computed:    true,
				Description: "Price per completion token in USD.",
			},
			"request_price": schema.StringAttribute{
				Computed:    true,
				Description: "Price per request in USD.",
			},
			"image_price": schema.StringAttribute{
				Computed:    true,
				Description: "Price per image in USD.",
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
	}
}

func (d *ModelDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ModelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ModelDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result modelsListResponse
	if err := d.client.inferenceGet(ctx, "/models", &result); err != nil {
		resp.Diagnostics.AddError("Error Reading Models", err.Error())
		return
	}

	var found *modelAPIResponse
	for i := range result.Data {
		if result.Data[i].ID == config.ID.ValueString() {
			found = &result.Data[i]
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError("Model Not Found",
			fmt.Sprintf("No model with ID %q was found in the OpenRouter models list.", config.ID.ValueString()))
		return
	}

	paramList, diags := stringsToList(ctx, found.SupportedParameters)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := ModelDataSourceModel{
		ID:                  types.StringValue(found.ID),
		Name:                types.StringValue(found.Name),
		Description:         types.StringValue(found.Description),
		ContextLength:       types.Int64Value(found.ContextLength),
		PromptPrice:         types.StringValue(found.Pricing.Prompt),
		CompletionPrice:     types.StringValue(found.Pricing.Completion),
		RequestPrice:        types.StringValue(found.Pricing.Request),
		ImagePrice:          types.StringValue(found.Pricing.Image),
		Tokenizer:           types.StringValue(found.Architecture.Tokenizer),
		InstructType:        types.StringValue(found.Architecture.InstructType),
		IsModerated:         types.BoolValue(found.TopProvider.IsModerated),
		SupportedParameters: paramList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

