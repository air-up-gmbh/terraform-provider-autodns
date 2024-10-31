package provider

import (
	"context"
	"os"
	"terraform-provider-autodns/internal/api"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const DEFAULT_API_ENDPOINT = "api.autodns.com/v1"
const DEFAULT_API_CONTEXT = "4"

// Ensure AutoDNSProvider satisfies the interfaces we need.
var (
	_ provider.Provider = &AutoDNSProvider{}
)

// AutoDNSProvider defines the provider implementation.
type AutoDNSProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// AutoDNSProviderModel describes the provider data model.
type AutoDNSProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Context  types.String `tfsdk:"context"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (p *AutoDNSProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "autodns"
	resp.Version = p.version
}

func (p *AutoDNSProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with AutoDNS API.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "AutoDNS api endpoint. May also be provided via AUTODNS_ENDPOINT environment variable.",
				Optional:            true,
			},
			"context": schema.StringAttribute{
				MarkdownDescription: "Context '1' refers to the demo system, context '4' or the PersonalAutoDNS context number refer to the live system." +
					"May also be provided via AUTODNS_CONTEXT environment variable.",
				Optional: true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "AutoDNS username. May also be provided via AUTODNS_USERNAME environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "AutoDNS password. May also be provided via AUTODNS_PASSWORD environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *AutoDNSProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config AutoDNSProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	tflog.Info(ctx, "configuring AutoDNS client")

	// Get values from environment variables
	endpoint := os.Getenv("AUTODNS_ENDPOINT")
	context := os.Getenv("AUTODNS_CONTEXT")
	username := os.Getenv("AUTODNS_USERNAME")
	password := os.Getenv("AUTODNS_PASSWORD")

	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}

	if !config.Context.IsNull() {
		context = config.Context.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	tflog.Debug(ctx, "creating AutoDNS client")

	if endpoint == "" {
		endpoint = DEFAULT_API_ENDPOINT
	}

	if context == "" {
		context = DEFAULT_API_CONTEXT
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown AutoDNS API Username",
			"The provider cannot create the AutoDNS API client as there is an unknown configuration value for the AutoDNS API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the AUTODNS_USERNAME environment variable.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown AutoDNS API Password",
			"The provider cannot create the AutoDNS API client as there is an unknown configuration value for the AutoDNS API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the AUTODNS_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create our API client
	client := api.NewClient(endpoint, context, username, password)

	resp.DataSourceData = client
	resp.ResourceData = client

	ctx = tflog.SetField(ctx, "autodns_endpoint", endpoint)
	ctx = tflog.SetField(ctx, "autodns_context", context)
	ctx = tflog.SetField(ctx, "autodns_username", username)
	tflog.Info(ctx, "configured AutoDNS client successfully")
}

// Resources registers our resources with the provider.
func (p *AutoDNSProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewRecordResource,
	}
}

// DataSources registers our datasources with the provider.
func (p *AutoDNSProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewZoneDataSource,
	}
}

// New returns a new instance of the provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AutoDNSProvider{
			version: version,
		}
	}
}
