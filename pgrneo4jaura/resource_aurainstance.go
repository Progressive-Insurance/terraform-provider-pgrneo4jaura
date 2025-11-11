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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &neo4jAuraResource{}
	_ resource.ResourceWithConfigure   = &neo4jAuraResource{}
	_ resource.ResourceWithImportState = &neo4jAuraResource{}
)

func NewAuraInstanceResource() resource.Resource {
	return &neo4jAuraResource{}
}

type neo4jAuraResource struct {
	access_token string
}

type neo4jAuraResourceModel struct {
	ID              types.String `tfsdk:"id"`
	ConnectionURL   types.String `tfsdk:"connection_url"`
	Version         types.String `tfsdk:"version"`
	Region          types.String `tfsdk:"region"`
	Memory          types.String `tfsdk:"memory"`
	InstanceType    types.String `tfsdk:"type"`
	TenantID        types.String `tfsdk:"tenant_id"`
	CloudProvider   types.String `tfsdk:"cloud_provider"`
	Name            types.String `tfsdk:"name"`
	Storage         types.String `tfsdk:"storage"`
	Paused          types.Bool   `tfsdk:"paused"`
	NeoUser         types.Bool   `tfsdk:"n4jusr"`
	NeoPwd          types.String `tfsdk:"n4jpwd"`
	CMK             types.String `tfsdk:"customer_managed_key_id"`
	VectorOptimized types.Bool   `tfsdk:"vector_optimized"`
	GDSPlugin       types.Bool   `tfsdk:"graph_analytics_plugin"`
	MetricsURL      types.String `tfsdk:"metrics_integration_url"`
	Secondaries     types.Int64  `tfsdk:"secondary_count"`
}

// tenant_id,storage,cloud_provider,type,version,name,region,memory,
func (r *neo4jAuraResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aurainstance"
}

func (r *neo4jAuraResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Neo4j Aura Instance",
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
				Description: "Neo4j Aura instance cloud provider.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"gcp", "aws", "azure"}...),
				},
			},
			"type": schema.StringAttribute{
				Description: "Neo4j Aura instance type.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"enterprise-db", "enterprise-ds", "professional-db", "professional-ds", "free-db"}...),
				},
			},
			"version": schema.StringAttribute{
				Description: "Neo4j Aura version.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"4", "5"}...),
				},
			},
			"name": schema.StringAttribute{
				Description: "Neo4j Aura instance name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": schema.StringAttribute{
				Description: "Neo4j Aura instance region.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"us-east-1", "us-west-2", "eu-west-1", "sa-east-1", "ap-southeast-1", "australia-southeast1", "us-east1", "us-central1", "us-west1", "europe-west2", "europe-west1", "europe-west3", "asia-east1", "asia-east2", "eastus", "westus3", "francecentral", "brazilsouth", "koreacentral"}...),
				},
			},
			"memory": schema.StringAttribute{
				Description: "Neo4j Aura instance memory size.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^(\d+GB)$`),
						"must be a valid memory setting",
					),
				},
			},
			"paused": schema.BoolAttribute{
				Description: "Neo4j instances running state.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"secondary_count": schema.Int64Attribute{
				Description: "Number of secondary Neo4j Aura instances.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"customer_managed_key_id": schema.StringAttribute{
				Description: "Neo4j Aura Customer Managed Key (CMK).",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vector_optimized": schema.BoolAttribute{
				Description: "An optional vector optimization configuration to be set during instance creation.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"graph_analytics_plugin": schema.BoolAttribute{
				Description: "An optional graph analytics plugin configuration to be set during instance creation.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"n4jusr": schema.BoolAttribute{
				Description: "Controls retrieval of default neo4j user password upon creation.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			//computed, no default (retrieved after create)
			//check auraprojects list for available memory/storage pairs. the resource takes memory value and computes storage value
			"connection_url": schema.StringAttribute{
				Description: "Neo4j Aura connection url.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"storage": schema.StringAttribute{
				Description: "Neo4j Aura instance storage. The amount of storage depends on the amount of memory allocated for your instance.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"n4jpwd": schema.StringAttribute{
				Description: "Default neo4j user password.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"metrics_integration_url": schema.StringAttribute{
				Description: "Neo4j Aura instance metrics url.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *neo4jAuraResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.access_token = req.ProviderData.(providerData).access_token
}

func (r *neo4jAuraResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan neo4jAuraResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	version := plan.Version.ValueString()
	region := plan.Region.ValueString()
	memory := plan.Memory.ValueString()
	instanceType := plan.InstanceType.ValueString()
	tenantID := plan.TenantID.ValueString()
	cloudProvider := plan.CloudProvider.ValueString()
	name := plan.Name.ValueString()
	paused := plan.Paused.ValueBool()
	n4jusr := plan.NeoUser.ValueBool()
	cmk := plan.CMK.ValueString()
	vectorOptimized := plan.VectorOptimized.ValueBool()
	gdsPluginIncluded := plan.GDSPlugin.ValueBool()
	secondaryCount := plan.Secondaries.ValueInt64()

	tflog.Info(ctx, fmt.Sprintf("creating neo4j %s instance", instanceType))
	instance, err := neo4jCreateInstance(ctx, r.access_token, version, region, memory, name, instanceType, tenantID, cloudProvider, cmk, vectorOptimized, gdsPluginIncluded)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Neo4j Aura instance",
			"Could not create Neo4j Aura instance. Received error: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "created neo4j instance")
	tflog.Debug(ctx, "instance details: %v", instance)
	instanceID := instance["data"].(map[string]interface{})["id"].(string)
	storage := instance["data"].(map[string]interface{})["storage"].(string)
	if paused {
		pauseResponse, err := neo4jPauseInstance(ctx, r.access_token, instanceID, true) //wait for pause to complete
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Pausing Neo4j Aura instance",
				"Could not pause Neo4j Aura instance. Received error: "+err.Error(),
			)
			return
		}
		tflog.Debug(ctx, fmt.Sprintf("pause respose: %v", pauseResponse))
	}
	if secondaryCount > 0 {
		updateSecondariesResponse, statusCode, err := neo4jUpdateSecondariesCount(ctx, r.access_token, instanceID, secondaryCount)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Secondaries Count",
				"Could not update secondaries count for Neo4j Aura instance. Received error: "+err.Error(),
			)
			return
		}
		tflog.Debug(ctx, fmt.Sprintf("update secondaries respose(%d): %v", statusCode, updateSecondariesResponse))
	}

	plan.ID = types.StringValue(instance["data"].(map[string]interface{})["id"].(string))
	plan.ConnectionURL = types.StringValue(instance["data"].(map[string]interface{})["connection_url"].(string))
	plan.MetricsURL = types.StringValue(instance["data"].(map[string]interface{})["metrics_integration_url"].(string))
	if n4jusr {
		plan.NeoPwd = types.StringValue(instance["data"].(map[string]interface{})["n4jpwd"].(string))
	} else {
		plan.NeoPwd = types.StringValue("N/A")
	}
	plan.Storage = types.StringValue(storage)
	plan.CMK = types.StringValue(cmk)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *neo4jAuraResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state neo4jAuraResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	instance, statusCode, err := neo4jGetInstance(r.access_token, id)
	tflog.Info(ctx, "reading neo4j instance")
	tflog.Debug(ctx, fmt.Sprintf("instance details (http: %d): %v", statusCode, instance))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Neo4j Aura instance",
			"Could not read Neo4j Aura instance. Received error: "+err.Error()+","+req.State.Raw.String(),
		)
		return
	}

	state.ID = types.StringValue(instance["data"].(map[string]interface{})["id"].(string))
	paused := instance["data"].(map[string]interface{})["status"].(string) == "paused"
	state.Secondaries = types.Int64Value(instance["data"].(map[string]interface{})["secondaries_count"].(int64))
	state.Paused = types.BoolValue(paused)
	state.Memory = types.StringValue(instance["data"].(map[string]interface{})["memory"].(string))
	state.VectorOptimized = types.BoolValue(instance["data"].(map[string]interface{})["vector_optimized"].(bool))
	state.GDSPlugin = types.BoolValue(instance["data"].(map[string]interface{})["graph_analytics_plugin"].(bool))
	state.MetricsURL = types.StringValue(instance["data"].(map[string]interface{})["metrics_integration_url"].(string))
	if !paused {
		state.ConnectionURL = types.StringValue(instance["data"].(map[string]interface{})["connection_url"].(string))
		state.Storage = types.StringValue(instance["data"].(map[string]interface{})["storage"].(string))
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func doSerializedUpdates(ctx context.Context, token string, instanceID string, state neo4jAuraResourceModel, plan neo4jAuraResourceModel, resp *resource.UpdateResponse) (map[string]bool, error) {
	updates := map[string]bool{
		"memory":                 false,
		"vector_optimized":       false,
		"graph_analytics_plugin": false,
		"secondaries_count":      false,
	}

	//decrease secondary instances to do modifications to less instances
	if state.Secondaries.ValueInt64() > plan.Secondaries.ValueInt64() {
		tflog.Info(ctx, "decreasing neo4j instance secondaries_count")
		updateResponse, statusCode, err := neo4jUpdateSecondariesCount(ctx, token, instanceID, plan.Secondaries.ValueInt64())
		tflog.Debug(ctx, fmt.Sprintf("Update secondaries_count response (%d): %v", statusCode, updateResponse))
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Neo4j Aura instance secondaries_count",
				"Could not update Neo4j Aura instance secondaries_count. Received error: "+err.Error(),
			)
			return nil, err
		}
		updates["secondaries_count"] = true
	}

	if state.Memory != plan.Memory {
		tflog.Info(ctx, "updating neo4j instance memory")
		updateResponse, statusCode, err := neo4jUpdateMemory(ctx, token, instanceID, plan.Memory.ValueString())
		tflog.Debug(ctx, fmt.Sprintf("update response(%d): %v", statusCode, updateResponse))
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Neo4j Aura instance",
				"Could not update Neo4j Aura instance memory. Received error: "+err.Error(),
			)
			return nil, err
		}
		updates["memory"] = true
	}

	if state.VectorOptimized != plan.VectorOptimized {
		tflog.Info(ctx, "updating neo4j instance vector optimzation")
		updateResponse, statusCode, err := neo4jUpdateVectorOptimization(ctx, token, instanceID, plan.VectorOptimized.ValueBool())
		tflog.Debug(ctx, fmt.Sprintf("Update vector optimzation response (%d): %v", statusCode, updateResponse))
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Neo4j Aura instance vector optimization",
				"Could not update Neo4j Aura instance vector optimization. Received error: "+err.Error(),
			)
			return nil, err
		}
		updates["vector_optimized"] = true
	}

	if state.GDSPlugin != plan.GDSPlugin {
		tflog.Info(ctx, "updating neo4j instance graph_analytics_plugin")
		updateResponse, statusCode, err := neo4jUpdateIncludeGraphPlugin(ctx, token, instanceID, plan.GDSPlugin.ValueBool())
		tflog.Debug(ctx, fmt.Sprintf("Update graph_analytics_plugin response (%d): %v", statusCode, updateResponse))
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Neo4j Aura instance graph_analytics_plugin",
				"Could not update Neo4j Aura instance graph_analytics_plugin. Received error: "+err.Error(),
			)
			return nil, err
		}
		updates["graph_analytics_plugin"] = true
	}

	//increase secondary instances after modifications
	if state.Secondaries.ValueInt64() < plan.Secondaries.ValueInt64() {
		tflog.Info(ctx, "increasing neo4j instance secondaries_count")
		updateResponse, statusCode, err := neo4jUpdateSecondariesCount(ctx, token, instanceID, plan.Secondaries.ValueInt64())
		tflog.Debug(ctx, fmt.Sprintf("Update secondaries_count response (%d): %v", statusCode, updateResponse))
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Neo4j Aura instance secondaries_count",
				"Could not update Neo4j Aura instance secondaries_count. Received error: "+err.Error(),
			)
			return nil, err
		}
		updates["secondaries_count"] = true
	}

	return updates, nil
}

func (r *neo4jAuraResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state neo4jAuraResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan neo4jAuraResourceModel
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	instanceID := state.ID.ValueString()
	name := plan.Name.ValueString()
	tflog.Info(ctx, fmt.Sprintf("updating neo4j instance %s with id %s", name, instanceID))

	// renaming can be performed paused/unpaused
	if state.Name != plan.Name {
		tflog.Info(ctx, "renaming neo4j instance")
		renameResponse, err := neo4jRenameInstance(r.access_token, instanceID, name)
		tflog.Debug(ctx, fmt.Sprintf("Rename response: %v", renameResponse))
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Renaming Neo4j Aura instance",
				"Could not rename Neo4j Aura instance. Received error: "+err.Error(),
			)
			return
		}
	}

	// adjust before pause
	if plan.Paused.ValueBool() { //instance will be paused
		// updateResponse, err := doCombinedUpdates(ctx, r.access_token, instanceID, state, plan, resp)
		updates, err := doSerializedUpdates(ctx, r.access_token, instanceID, state, plan, resp)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Neo4j Aura instance",
				"Could not perform updates to Neo4j Aura instance. Received error: "+err.Error(),
			)
			return
		}
		if updates["memory"] || updates["vector_optimized"] || updates["graph_analytics_plugin"] || updates["secondaries_count"] { //if updateResponse != nil { when combinedupates
			tflog.Info(ctx, fmt.Sprintf("update objects: %v", updates))
		} else {
			tflog.Info(ctx, "no updates to make before pause")
		}
	}

	// pause / unpause
	if state.Paused != plan.Paused {
		if plan.Paused.ValueBool() {
			tflog.Info(ctx, "pausing neo4j instance")
			pauseResponse, err := neo4jPauseInstance(ctx, r.access_token, instanceID, true)
			tflog.Debug(ctx, fmt.Sprintf("Pause response: %v", pauseResponse))
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Pausing Neo4j Aura instance",
					"Could not pause Neo4j Aura instance. Received error: "+err.Error(),
				)
				return
			}
		} else {
			tflog.Info(ctx, "resuming neo4j instance")
			resumeResponse, err := neo4jResumeInstance(ctx, r.access_token, instanceID, true)
			tflog.Debug(ctx, fmt.Sprintf("Resume response: %v", resumeResponse))
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Resuming Neo4j Aura instance",
					"Could not resume Neo4j Aura instance. Received error: "+err.Error(),
				)
				return
			}
		}
	}

	// adjust after unpause/resume
	if !plan.Paused.ValueBool() { //instance was unpaused/resumed
		// updateResponse, err := doCombinedUpdates(ctx, r.access_token, instanceID, state, plan, resp)
		updates, err := doSerializedUpdates(ctx, r.access_token, instanceID, state, plan, resp)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Neo4j Aura instance",
				"Could not perform updates to Neo4j Aura instance. Received error: "+err.Error(),
			)
			return
		}
		if updates["memory"] || updates["vector_optimized"] || updates["graph_analytics_plugin"] || updates["secondaries_count"] { //if updateResponse != nil { when combinedupates
			tflog.Info(ctx, fmt.Sprintf("update objects: %v", updates))
		} else {
			tflog.Info(ctx, "no updates to make before pause")
		}
	}

	state.Name = plan.Name
	state.Paused = plan.Paused
	state.Memory = plan.Memory
	state.VectorOptimized = plan.VectorOptimized
	state.GDSPlugin = plan.GDSPlugin
	state.Secondaries = plan.Secondaries

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *neo4jAuraResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state neo4jAuraResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	tflog.Info(ctx, fmt.Sprintf("deleting neo4j instance with id %s", id))
	_, err := neo4jDeleteInstance(ctx, r.access_token, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Neo4j Aura instance",
			"Could not delete Neo4j Aura instance. Received error: "+err.Error(),
		)
		return
	}
}

// NOTE about "version"
//   - /instances/{instanceId} response does not include instance version.
//   - therefore, terraform import must additionaly be passed the version to sync state
//   - the endpoint could also be updated to reflect the instances version so its 1 less input for the import
//     ie. terraform import resource_type.resource_name <aura_instance_id> (* DOESNT WORK WITHOUT VERSION)
//
// terraform import pgrneo4jaura_aurainstance.myinstance <INSTANCE ID>,<INSTANCE VERSION>,<INCL N4J USR>,(,<N4J USR PWD>)(,<INSTANCE MEMORY>)
func (r *neo4jAuraResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importParts := strings.Split(req.ID, ",")
	if len(importParts) < 3 || len(importParts) > 5 {
		resp.Diagnostics.AddError(
			"Error Importing Neo4j Aura instance",
			"Could not import Neo4j Aura instance.\nPlease ensure you run \"terraform import resource_type.resource_name <aura_instance_id>,<instance_version>,<include_neo4j_user>,(,<neo4j_user_pwd>)(,<instance_memory>)",
		)
		return
	}

	id := importParts[0]
	version := importParts[1]
	n4jUserIncl := importParts[2]

	tflog.Info(ctx, "importing neo4j instance")
	instance, statusCode, err := neo4jGetInstance(r.access_token, id)
	tflog.Debug(ctx, fmt.Sprintf("instance details (http: %d): %v", statusCode, instance))

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Neo4j Aura instance",
			"Could not import Neo4j Aura instance "+id+". Received error: "+err.Error(),
		)
		return
	}

	n4jusr := types.BoolValue(n4jUserIncl == "true")
	n4jpwd := "N/A"
	if n4jUserIncl == "true" {
		n4jpwd = importParts[3]
	}

	connection_url := "" //null connection_url in paused instances
	paused := instance["data"].(map[string]interface{})["status"].(string) == "paused"
	memory, storage := "paused", "paused"
	if !paused {
		connection_url = instance["data"].(map[string]interface{})["connection_url"].(string) //null in paused instances
		memory = instance["data"].(map[string]interface{})["memory"].(string)                 //missing in paused instances
		storage = instance["data"].(map[string]interface{})["storage"].(string)               //missing in paused instances
	} else {
		if len(importParts) != 5 {
			resp.Diagnostics.AddError(
				"Error Importing Neo4j Aura instance",
				"Could not import Neo4j Aura instance.\nPlease ensure you run:\n  terraform import resource_type.resource_name <aura_instance_id>,<instance_version>,<include_neo4j_user>,(,<neo4j_user_pwd>)(,<instance_memory>)\nIf instance is paused for import you must also specify the memory in GB.",
			)
			return
		}
		memory = importParts[4]
	}
	cloud_provider := instance["data"].(map[string]interface{})["cloud_provider"].(string)
	name := instance["data"].(map[string]interface{})["name"].(string)
	region := instance["data"].(map[string]interface{})["region"].(string)
	tenant_id := instance["data"].(map[string]interface{})["tenant_id"].(string)
	instanceType := instance["data"].(map[string]interface{})["type"].(string)

	secondaries := types.Int64Value(instance["data"].(map[string]interface{})["secondaries_count"].(int64))
	vectorOptimized := types.BoolValue(instance["data"].(map[string]interface{})["vector_optimized"].(bool))
	gdsPlugin := types.BoolValue(instance["data"].(map[string]interface{})["graph_analytics_plugin"].(bool))
	metricsUrl := types.StringValue(instance["data"].(map[string]interface{})["metrics_integration_url"].(string))
	cmk := types.StringValue(instance["data"].(map[string]interface{})["customer_managed_key_id"].(string))

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("connection_url"), connection_url)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cloud_provider"), cloud_provider)...)

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("customer_managed_key_id"), cmk)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vector_optimized"), vectorOptimized)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("secondary_count"), secondaries)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("graph_analytics_plugin"), gdsPlugin)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("metrics_integration_url"), metricsUrl)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("n4jusr"), n4jusr)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("n4jpwd"), n4jpwd)...)

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("memory"), memory)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), region)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("storage"), storage)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), tenant_id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), instanceType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("version"), version)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("paused"), paused)...)
}
