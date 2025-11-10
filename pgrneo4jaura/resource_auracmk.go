package pgrneo4jaura

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &neo4jAuraCMKResource{}
	_ resource.ResourceWithConfigure   = &neo4jAuraCMKResource{}
	_ resource.ResourceWithImportState = &neo4jAuraCMKResource{}
)

func NewAuraCMKResource() resource.Resource {
	return &neo4jAuraCMKResource{}
}

type neo4jAuraCMKResource struct {
	access_token string
}

type neo4jAuraCMKResourceModel struct {
	ID            types.String `tfsdk:"id"`
	TenantID      types.String `tfsdk:"tenant_id"`
	Region        types.String `tfsdk:"region"`
	InstanceType  types.String `tfsdk:"instance_type"`
	CloudProvider types.String `tfsdk:"cloud_provider"`
	Name          types.String `tfsdk:"name"`
	KeyID         types.String `tfsdk:"key_id"`
	Created       types.String `tfsdk:"created"`
}

// tenant_id,storage,cloud_provider,type,version,name,region,memory,
func (r *neo4jAuraCMKResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auracmk"
}

func (r *neo4jAuraCMKResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Neo4j Aura Customer Managed Key (CMK)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "identifier for resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": schema.StringAttribute{
				Description: "Neo4j Aura tenant identifier.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(36),
					stringvalidator.LengthAtMost(36),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
						"must be a valid tenant id",
					),
				},
			},
			"cloud_provider": schema.StringAttribute{
				Description: "Neo4j Aura CMK cloud provider.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"gcp", "aws", "azure"}...),
				},
			},
			"instance_type": schema.StringAttribute{
				Description: "Neo4j Aura CMK type.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"enterprise-db", "enterprise-ds", "professional-db", "professional-ds", "free-db"}...),
				},
			},
			"name": schema.StringAttribute{
				Description: "Neo4j Aura CMK name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Description: "Neo4j Aura CMK region.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"us-east-1", "us-west-2", "eu-west-1", "sa-east-1", "ap-southeast-1", "australia-southeast1", "us-east1", "us-central1", "us-west1", "europe-west2", "europe-west1", "europe-west3", "asia-east1", "asia-east2", "eastus", "westus3", "francecentral", "brazilsouth", "koreacentral"}...),
				},
			},
			"key_id": schema.StringAttribute{
				Description: "CMK key id.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created": schema.StringAttribute{
				Description: "Neo4j Aura CMK created at date/time",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *neo4jAuraCMKResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.access_token = req.ProviderData.(providerData).access_token
}

func (r *neo4jAuraCMKResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan neo4jAuraCMKResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := plan.TenantID.ValueString()
	region := plan.Region.ValueString()
	instanceType := plan.InstanceType.ValueString()
	cloudProvider := plan.CloudProvider.ValueString()
	name := plan.Name.ValueString()
	keyId := plan.KeyID.ValueString()

	tflog.Info(ctx, fmt.Sprintf("creating neo4j cmk %s", name))
	cmk, err := neo4jCreateCMK(ctx, r.access_token, tenantID, name, region, instanceType, cloudProvider, keyId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Neo4j Aura CMK",
			"Could not create Neo4j Aura CMK. Received error: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "created neo4j cmk")
	tflog.Debug(ctx, "cmk details: %v", cmk)

	plan.ID = types.StringValue(cmk["data"].(map[string]interface{})["id"].(string))
	plan.Created = types.StringValue(cmk["data"].(map[string]interface{})["created"].(string))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *neo4jAuraCMKResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state neo4jAuraCMKResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	cmk, statusCode, err := neo4jGetCMK(r.access_token, id)
	tflog.Info(ctx, fmt.Sprintf("reading neo4j cmk %s", id))
	tflog.Debug(ctx, fmt.Sprintf("cmk details (http: %d): %v", statusCode, cmk))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Neo4j Aura CMK",
			"Could not read Neo4j Aura CMK. Received error: "+err.Error()+","+req.State.Raw.String(),
		)
		return
	}

	state.ID = types.StringValue(cmk["data"].(map[string]interface{})["id"].(string))
	state.Created = types.StringValue(cmk["data"].(map[string]interface{})["created"].(string))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *neo4jAuraCMKResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state neo4jAuraCMKResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan neo4jAuraCMKResourceModel
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("neo4j cmk should require replace for any updates"))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *neo4jAuraCMKResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state neo4jAuraCMKResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	tflog.Info(ctx, fmt.Sprintf("deleting neo4j cmk with id %s", id))
	err := neo4jDeleteCMK(ctx, r.access_token, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Neo4j Aura CMK",
			"Could not delete Neo4j Aura CMK. Received error: "+err.Error(),
		)
		return
	}
	return
}

// terraform import pgrneo4jaura_auracmk.mycmk <CMK NAME>
func (r *neo4jAuraCMKResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importParts := strings.Split(req.ID, ",")
	if len(importParts) != 1 {
		resp.Diagnostics.AddError(
			"Error Importing Neo4j Aura CMK",
			"Could not import Neo4j Aura CMK.\nPlease ensure you run \"terraform import resource_type.resource_name <cmk_name>",
		)
		return
	}
	name := importParts[0]

	cmks, err := neo4jGetCMKs(r.access_token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Neo4j Aura CMK",
			"Could not retrieve list of Neo4j Aura CMKs. Received error: "+err.Error(),
		)
		return
	}

	id := ""
	if data, ok := cmks["data"].([]interface{}); ok {
		for _, item := range data {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemMap["name"] == name {
					id = itemMap["id"].(string)
				}
			} else {
				fmt.Println("Invalid item type in data")
			}
		}
	} else {
		resp.Diagnostics.AddError(
			"Error Importing Neo4j Aura CMK",
			"Could not retrieve list of Neo4j Aura CMKs. Received error: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "importing neo4j cmk")
	cmk, statusCode, err := neo4jGetCMK(r.access_token, id)
	tflog.Debug(ctx, fmt.Sprintf("cmk details (http: %d): %v", statusCode, cmk))

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Neo4j Aura CMK",
			"Could not import Neo4j Aura CMK "+id+". Received error: "+err.Error(),
		)
		return
	}

	cloud_provider := cmk["data"].(map[string]interface{})["cloud_provider"].(string)
	created := cmk["data"].(map[string]interface{})["created"].(string)
	instanceType := cmk["data"].(map[string]interface{})["instance_type"].(string)
	keyId := cmk["data"].(map[string]interface{})["key_id"].(string)
	region := cmk["data"].(map[string]interface{})["region"].(string)
	tenant_id := cmk["data"].(map[string]interface{})["tenant_id"].(string)

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("created"), created)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cloud_provider"), cloud_provider)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("instance_type"), instanceType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key_id"), keyId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), region)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), tenant_id)...)
}
