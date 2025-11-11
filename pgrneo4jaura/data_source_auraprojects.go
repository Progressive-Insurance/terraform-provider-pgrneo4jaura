package pgrneo4jaura

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource = &auraProjectsDataSource{}
)

func NewAuraProjectsDataSource() datasource.DataSource {
	return &auraProjectsDataSource{}
}

type auraProjectsDataSource struct {
	access_token string
}

type auraProjectsDataSourceModel struct {
	TenantID               types.String `tfsdk:"tenant_id"`
	Name                   types.String `tfsdk:"name"`
	InstanceConfigurations types.Set    `tfsdk:"instance_configurations"`
}

func (r *auraProjectsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auraprojects"
}

func (r *auraProjectsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data lookup for Neo4j Aura project configurations",
		Attributes: map[string]schema.Attribute{
			"tenant_id": schema.StringAttribute{
				Description: "Neo4j Aura tenant identifier.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(36),
					stringvalidator.LengthAtMost(36),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
						"must be a valid tenant id",
					),
				},
			},
			//computed
			"name": schema.StringAttribute{
				Description: "The tenant name.",
				Computed:    true,
			},
			"instance_configurations": schema.SetAttribute{
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"cloud_provider": types.StringType,
						"memory":         types.StringType,
						"region":         types.StringType,
						"region_name":    types.StringType,
						"storage":        types.StringType,
						"type":           types.StringType,
						"version":        types.StringType,
					},
				},
				Computed: true,
			},
		},
	}
}

func (r *auraProjectsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.access_token = req.ProviderData.(providerData).access_token
}

func (r *auraProjectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state auraProjectsDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantId := state.TenantID.ValueString()

	// Fetch project configurations
	projectConfigurations, err := neo4jGetProjectConfigurations(r.access_token, tenantId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Neo4j Aura project configurations.",
			err.Error(),
		)
		return
	}

	// Extract instance configurations from the response
	instanceConfigurationsRaw, ok := projectConfigurations["data"].(map[string]interface{})["instance_configurations"].([]interface{})
	if !ok {
		resp.Diagnostics.AddError(
			"Invalid Response",
			"'instance_configurations' is not in the expected format.",
		)
		return
	}

	// Convert instance configurations to types.Set
	var instanceConfigurations []attr.Value
	for _, rawConfig := range instanceConfigurationsRaw {
		configMap, ok := rawConfig.(map[string]interface{})
		if !ok {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"One of the instance configurations is not a valid map.",
			)
			return
		}

		// Convert each map[string]interface{} to types.Object
		instanceConfig, err := types.ObjectValue(map[string]attr.Type{
			"cloud_provider": types.StringType,
			"memory":         types.StringType,
			"region":         types.StringType,
			"region_name":    types.StringType,
			"storage":        types.StringType,
			"type":           types.StringType,
			"version":        types.StringType,
		}, map[string]attr.Value{
			"cloud_provider": types.StringValue(configMap["cloud_provider"].(string)),
			"memory":         types.StringValue(configMap["memory"].(string)),
			"region":         types.StringValue(configMap["region"].(string)),
			"region_name":    types.StringValue(configMap["region_name"].(string)),
			"storage":        types.StringValue(configMap["storage"].(string)),
			"type":           types.StringValue(configMap["type"].(string)),
			"version":        types.StringValue(configMap["version"].(string)),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Object Conversion Error",
				"Failed to convert instance configuration to ObjectValue",
			)
			return
		}

		instanceConfigurations = append(instanceConfigurations, instanceConfig)
	}

	// Convert to types.Set
	instanceConfigurationsSet, diags := types.SetValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"cloud_provider": types.StringType,
			"memory":         types.StringType,
			"region":         types.StringType,
			"region_name":    types.StringType,
			"storage":        types.StringType,
			"type":           types.StringType,
			"version":        types.StringType,
		},
	}, instanceConfigurations)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Assign to state
	state.InstanceConfigurations = instanceConfigurationsSet

	// Set other state attributes
	state.TenantID = types.StringValue(tenantId)
	state.Name = types.StringValue(fmt.Sprintf("%v", projectConfigurations["data"].(map[string]interface{})["name"]))

	// Save state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
