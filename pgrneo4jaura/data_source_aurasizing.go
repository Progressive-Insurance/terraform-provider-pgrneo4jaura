package pgrneo4jaura

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource = &auraSizingEstimateDataSource{}
)

func NewAuraSizingEstimateDataSource() datasource.DataSource {
	return &auraSizingEstimateDataSource{}
}

type auraSizingEstimateDataSource struct {
	access_token string
}

type auraSizingEstimateDataSourceModel struct {
	NodeCount           types.Int64  `tfsdk:"node_count"`
	RelationshipCount   types.Int64  `tfsdk:"relationship_count"`
	InstanceType        types.String `tfsdk:"instance_type"`
	AlgorithmCategories types.Set    `tfsdk:"algorithm_categories"`
	DidExceedMaximum    types.Bool   `tfsdk:"did_exceed_maximum"`
	RecommendedSize     types.String `tfsdk:"recommended_size"`
	MinRequiredMemory   types.String `tfsdk:"min_required_memory"`
}

func (r *auraSizingEstimateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aurasizing"
}

func (r *auraSizingEstimateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data lookup for Neo4j Aura instance sizing estimator",
		Attributes: map[string]schema.Attribute{
			"node_count": schema.Int64Attribute{
				Description: "Estimated node count.",
				Required:    true,
			},
			"relationship_count": schema.Int64Attribute{
				Description: "Estimated relationship count.",
				Required:    true,
			},
			"instance_type": schema.StringAttribute{
				Description: "Type of Neo4j Aura instance.",
				Required:    true,
			},
			"algorithm_categories": schema.SetAttribute{
				Description: "List of algorithm categories.",
				Required:    true,
				ElementType: types.StringType,
			},
			//computed
			"did_exceed_maximum": schema.BoolAttribute{
				Description: "Indicates if the instance size exceeds the maximum allowed.",
				Computed:    true,
			},
			"recommended_size": schema.StringAttribute{
				Description: "The recommended instance size.",
				Computed:    true,
			},
			"min_required_memory": schema.StringAttribute{
				Description: "The minimum required memory for the instance.",
				Computed:    true,
			},
		},
	}
}

func (r *auraSizingEstimateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.access_token = req.ProviderData.(providerData).access_token
}

func (r *auraSizingEstimateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state auraSizingEstimateDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodeCount := state.NodeCount.ValueInt64()
	relationshipCount := state.RelationshipCount.ValueInt64()
	instanceType := state.InstanceType.ValueString()

	state.NodeCount = types.Int64Value(nodeCount)
	state.RelationshipCount = types.Int64Value(relationshipCount)
	state.InstanceType = types.StringValue(instanceType)

	algorithmCategories := make([]types.String, 0, len(state.AlgorithmCategories.Elements()))
	diags = state.AlgorithmCategories.ElementsAs(ctx, &algorithmCategories, false)
	tflog.Info(ctx, strconv.Itoa(len(algorithmCategories)))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.AlgorithmCategories, diags = types.SetValueFrom(ctx, types.StringType, algorithmCategories)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "getting neo4j aura instance sizing estimate")
	var stringCategories []string
	for _, alg := range algorithmCategories {
		if !alg.IsNull() { // Check if the value is not null
			stringCategories = append(stringCategories, alg.ValueString())
		}
	}

	estimate, err := neo4jSizingEstimate(ctx, r.access_token, nodeCount, relationshipCount, instanceType, stringCategories)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Neo4j Aura instance sizing estimate.",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("estimate: %v", estimate))
	state.DidExceedMaximum = types.BoolValue(estimate["data"].(map[string]interface{})["did_exceed_maximum"].(bool))
	state.RecommendedSize = types.StringValue(estimate["data"].(map[string]interface{})["recommended_size"].(string))
	state.MinRequiredMemory = types.StringValue(estimate["data"].(map[string]interface{})["min_required_memory"].(string))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
