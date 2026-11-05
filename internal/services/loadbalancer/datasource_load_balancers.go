package loadbalancer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/loadbalancer/datasource_load_balancer"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &loadbalancersDataSource{}

// loadbalancersDataSource is the Terraform data source implementation for numspot_load_balancers.
type loadbalancersDataSource struct {
	services.BaseDataSource
}

// NewLoadBalancersDataSource returns a new loadbalancersDataSource instance.
func NewLoadBalancersDataSource() datasource.DataSource {
	return &loadbalancersDataSource{}
}

// Metadata sets the data source type name.
func (d *loadbalancersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancers"
}

// Schema defines the schema for the data source.
func (d *loadbalancersDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_load_balancer.LoadBalancerDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *loadbalancersDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_load_balancer.LoadBalancerModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	params := deserializeReadLoadBalancers(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	result, err := core.ReadLoadBalancers(ctx, d.Provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read load_balancers", err.Error())
		return
	}

	objectItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *result, mappingItemsValue)
	if serializeDiags.HasError() {
		response.Diagnostics.Append(serializeDiags...)
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, objectItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func mappingItemsValue(ctx context.Context, item *api.LoadBalancer) (datasource_load_balancer.ItemsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	tagsList := types.ListNull(datasource_load_balancer.ItemsValue{}.Type(ctx))

	if item.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_load_balancer.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_load_balancer.ItemsValue{}, serializeDiags
		}
	}

	value, valueDiags := datasource_load_balancer.NewItemsValue(datasource_load_balancer.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"application_sticky_cookie_policies": mappingApplicationStickyCookiePolicies(ctx, item, &diags),
		"availability_zone_names":            azNamesToTfList(ctx, item.AvailabilityZoneNames, &diags),
		"backend_ips":                        utils.FromStringListPointerToTfStringList(ctx, item.BackendIps, &diags),
		"backend_vm_ids":                     utils.FromStringListPointerToTfStringList(ctx, item.BackendVmIds, &diags),
		"dns_name":                           types.StringValue(utils.ConvertStringPtrToString(item.DnsName)),
		"health_check":                       mappingHealthCheck(ctx, item.HealthCheck, &diags),
		"listeners":                          mappingListeners(ctx, item, &diags),
		"name":                               types.StringValue(utils.ConvertStringPtrToString(item.Name)),
		"public_ip":                          types.StringValue(utils.ConvertStringPtrToString(item.PublicIp)),
		"secured_cookies":                    types.BoolValue(*item.SecuredCookies),
		"security_groups":                    utils.FromStringListPointerToTfStringList(ctx, item.SecurityGroups, &diags),
		"sticky_cookie_policies":             mappingStickyCookiePolicies(ctx, item, &diags),
		"subnets":                            utils.FromStringListPointerToTfStringList(ctx, item.Subnets, &diags),
		"type":                               types.StringValue(utils.ConvertStringPtrToString(item.Type)),
		"vpc_id":                             types.StringValue(utils.ConvertStringPtrToString(item.VpcId)),
		"tags":                               tagsList,
	})
	diags.Append(valueDiags...)
	return value, diags
}

func mappingTags(ctx context.Context, tag *api.ResourceTag) (datasource_load_balancer.TagsValue, diag.Diagnostics) {
	return datasource_load_balancer.NewTagsValue(datasource_load_balancer.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}

func mappingApplicationStickyCookiePolicie(
	ctx context.Context,
	item *api.ApplicationStickyCookiePolicy,
) (datasource_load_balancer.ApplicationStickyCookiePoliciesValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, d := datasource_load_balancer.NewApplicationStickyCookiePoliciesValue(
		datasource_load_balancer.ApplicationStickyCookiePoliciesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"cookie_name": types.StringPointerValue(item.CookieName),
			"policy_name": types.StringPointerValue(item.PolicyName),
		},
	)
	diags.Append(d...)
	return value, diags
}

func mappingApplicationStickyCookiePolicies(ctx context.Context, item *api.LoadBalancer, diags *diag.Diagnostics) types.List {
	if item.ApplicationStickyCookiePolicies == nil {
		return types.ListNull(datasource_load_balancer.ApplicationStickyCookiePoliciesValue{}.Type(ctx))
	}
	items, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.ApplicationStickyCookiePolicies, mappingApplicationStickyCookiePolicie)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
		return types.ListNull(datasource_load_balancer.ApplicationStickyCookiePoliciesValue{}.Type(ctx))
	}
	listValueItems := utils.CreateListValueItems(ctx, items, &serializeDiags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	return listValueItems
}

func mappingHealthCheck(ctx context.Context, item *api.HealthCheck, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_load_balancer.HealthCheckValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_load_balancer.NewHealthCheckValue(datasource_load_balancer.HealthCheckValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"check_interval":      utils.FromIntToTfInt64(item.CheckInterval),
		"healthy_threshold":   utils.FromIntToTfInt64(item.HealthyThreshold),
		"path":                types.StringPointerValue(item.Path),
		"port":                utils.FromIntToTfInt64(item.Port),
		"protocol":            types.StringValue(item.Protocol),
		"timeout":             utils.FromIntToTfInt64(item.Timeout),
		"unhealthy_threshold": utils.FromIntToTfInt64(item.UnhealthyThreshold),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func mappingListener(ctx context.Context, item *api.Listener) (datasource_load_balancer.ListenersValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, d := datasource_load_balancer.NewListenersValue(datasource_load_balancer.ListenersValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"backend_port":           utils.FromIntPtrToTfInt64(item.BackendPort),
		"backend_protocol":       types.StringPointerValue(item.BackendProtocol),
		"load_balancer_port":     utils.FromIntPtrToTfInt64(item.LoadBalancerPort),
		"load_balancer_protocol": types.StringPointerValue(item.LoadBalancerProtocol),
		"policy_names":           utils.FromStringListPointerToTfStringList(ctx, item.PolicyNames, &diags),
		"server_certificate_id":  types.StringPointerValue(item.ServerCertificateId),
	})
	diags.Append(d...)
	return value, diags
}

func mappingListeners(ctx context.Context, item *api.LoadBalancer, diags *diag.Diagnostics) types.List {
	if item.Listeners == nil {
		return types.ListNull(datasource_load_balancer.ListenersValue{}.Type(ctx))
	}
	items, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.Listeners, mappingListener)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
		return types.ListNull(datasource_load_balancer.ListenersValue{}.Type(ctx))
	}
	listValueItems := utils.CreateListValueItems(ctx, items, &serializeDiags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	return listValueItems
}

func mappingStickyCookiePolicie(
	ctx context.Context,
	item *api.LoadBalancerStickyCookiePolicy,
) (datasource_load_balancer.StickyCookiePoliciesValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, d := datasource_load_balancer.NewStickyCookiePoliciesValue(
		datasource_load_balancer.StickyCookiePoliciesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"cookie_expiration_period": utils.FromIntPtrToTfInt64(item.CookieExpirationPeriod),
			"policy_name":              types.StringPointerValue(item.PolicyName),
		},
	)
	diags.Append(d...)
	return value, diags
}

func mappingStickyCookiePolicies(ctx context.Context, item *api.LoadBalancer, diags *diag.Diagnostics) types.List {
	if item.StickyCookiePolicies == nil {
		return types.ListNull(datasource_load_balancer.StickyCookiePoliciesValue{}.Type(ctx))
	}
	items, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.StickyCookiePolicies, mappingStickyCookiePolicie)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
		return types.ListNull(datasource_load_balancer.StickyCookiePoliciesValue{}.Type(ctx))
	}
	listValueItems := utils.CreateListValueItems(ctx, items, &serializeDiags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	return listValueItems
}

func deserializeReadLoadBalancers(
	ctx context.Context,
	tf *datasource_load_balancer.LoadBalancerModel,
	diags *diag.Diagnostics,
) *api.ReadLoadBalancersParams {
	return &api.ReadLoadBalancersParams{
		LoadBalancerNames: utils.ConvertTfListToArrayOfString(ctx, tf.LoadBalancerNames, diags),
		Page: func() *api.PageParameter {
			if utils.IsTfValueNull(tf.Page) {
				return nil
			}
			return &api.PageParameter{NextToken: tf.Page.NextToken.ValueStringPointer(), Size: utils.FromTfInt64ToInt32Ptr(tf.Page.Size)}
		}(),
	}
}
