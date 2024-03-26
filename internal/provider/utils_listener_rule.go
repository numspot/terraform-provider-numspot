package provider

import (
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_listener_rule"
)

func ListenerRuleFromTfToHttp(tf *resource_listener_rule.ListenerRuleModel) *iaas.ListenerRule {
	return &iaas.ListenerRule{}
}

func ListenerRuleFromHttpToTf(http *iaas.ListenerRule) resource_listener_rule.ListenerRuleModel {
	return resource_listener_rule.ListenerRuleModel{}
}

func ListenerRuleFromTfToCreateRequest(tf *resource_listener_rule.ListenerRuleModel) iaas.CreateListenerRuleJSONRequestBody {
	return iaas.CreateListenerRuleJSONRequestBody{}
}
