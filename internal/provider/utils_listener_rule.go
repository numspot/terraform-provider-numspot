package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_listener_rule"
)

func ListenerRuleFromTfToHttp(tf *resource_listener_rule.ListenerRuleModel) *api.ListenerRuleSchema {
	return &api.ListenerRuleSchema{}
}

func ListenerRuleFromHttpToTf(http *api.ListenerRuleSchema) resource_listener_rule.ListenerRuleModel {
	return resource_listener_rule.ListenerRuleModel{}
}

func ListenerRuleFromTfToCreateRequest(tf *resource_listener_rule.ListenerRuleModel) api.CreateListenerRuleJSONRequestBody {
	return api.CreateListenerRuleJSONRequestBody{}
}
