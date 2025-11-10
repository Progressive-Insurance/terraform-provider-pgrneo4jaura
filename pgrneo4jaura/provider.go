package pgrneo4jaura

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &pgrneo4jaura_provider{}

func New() provider.Provider {
	return &pgrneo4jaura_provider{}
}

type pgrneo4jaura_provider struct{}

type pgrneo4jauraProviderModel struct {
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

type providerData struct {
	access_token string
}

func (p *pgrneo4jaura_provider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pgrneo4jaura"
}

func (p *pgrneo4jaura_provider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Progressive Neo4j Aura Provider",
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Optional:    true,
				Description: "Progressive Neo4j Aura API client id.",
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Progressive Neo4j Aura API client secret.",
			},
		},
	}
}

func (p *pgrneo4jaura_provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data providerData
	var config pgrneo4jauraProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ClientID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Unknown Neo4j Aura client id",
			"The provider cannot authenticate as there is an unkonwn configuration value for the client id. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PGRNEO4J_CLIENTID environment variable.",
		)
	}

	if config.ClientSecret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Unknown Neo4j Aura client secret",
			"The provider cannot authenticate as there is an unkonwn configuration value for the client secret. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PGRNEO4J_CLIENTSECERET environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client_id := os.Getenv("PGRNEO4J_CLIENTID")
	client_secret := os.Getenv("PGRNEO4J_CLIENTSECERET")

	if !config.ClientID.IsNull() {
		client_id = config.ClientID.ValueString()
	}

	if !config.ClientSecret.IsNull() {
		client_secret = config.ClientSecret.ValueString()
	}

	if client_id == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Missing Neo4j Aura client id.",
			"The provider cannot authenticate to Neo4j Aura without a valid client id/client secret.",
		)
	}

	if client_secret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Missing Neo4j Aura client secret.",
			"The provider cannot authenticate to Neo4j Aura without a valid client id/client secret.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	access_token, err := getNeo4jAuraAuthToken(client_id, client_secret)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to authenticate to Neo4j Aura",
			"An unexpected error occured when authenicating to Neo4j Aura. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	data.access_token = access_token
	resp.ResourceData = data
	resp.DataSourceData = data
}

func (p *pgrneo4jaura_provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAuraSizingEstimateDataSource,
		NewAuraProjectsDataSource,
	}
}

func (p *pgrneo4jaura_provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAuraInstanceResource,
		NewAuraCMKResource,
	}
}
