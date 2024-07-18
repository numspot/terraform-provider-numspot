package listenerrule

import "gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

func ListenerRuleFromTfToHttp(tf *ListenerRuleModel) *numspot.ListenerRule {
	return &numspot.ListenerRule{}
}

func ListenerRuleFromHttpToTf(http *numspot.ListenerRule) ListenerRuleModel {
	return ListenerRuleModel{}
}

func ListenerRuleFromTfToCreateRequest(tf *ListenerRuleModel) numspot.CreateListenerRuleJSONRequestBody {
	return numspot.CreateListenerRuleJSONRequestBody{}
}
