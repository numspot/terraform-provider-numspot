package provider

import (
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_listener_rule"
)

func ListenerRuleFromTfToHttp(tf *resource_listener_rule.ListenerRuleModel) *numspot.ListenerRule {
	return &numspot.ListenerRule{}
}

func ListenerRuleFromHttpToTf(http *numspot.ListenerRule) resource_listener_rule.ListenerRuleModel {
	return resource_listener_rule.ListenerRuleModel{}
}

func ListenerRuleFromTfToCreateRequest(tf *resource_listener_rule.ListenerRuleModel) numspot.CreateListenerRuleJSONRequestBody {
	return numspot.CreateListenerRuleJSONRequestBody{}
}
