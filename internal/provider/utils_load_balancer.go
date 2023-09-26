package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_load_balancer"
)

func LoadBalancerFromTfToHttp(tf resource_load_balancer.LoadBalancerModel) *api.LoadBalancerSchema {
	return &api.LoadBalancerSchema{}
}

func LoadBalancerFromHttpToTf(http *api.LoadBalancerSchema) resource_load_balancer.LoadBalancerModel {
	return resource_load_balancer.LoadBalancerModel{}
}

func LoadBalancerFromTfToCreateRequest(tf resource_load_balancer.LoadBalancerModel) api.CreateLoadBalancerJSONRequestBody {
	return api.CreateLoadBalancerJSONRequestBody{}
}
