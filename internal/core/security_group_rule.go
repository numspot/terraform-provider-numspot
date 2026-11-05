package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var (
	// errCreatedSecurityGroupRuleNotFound is returned when a freshly created
	// security group rule cannot be located on read.
	errCreatedSecurityGroupRuleNotFound = errors.New("created security group rule could not be found in security group")
	// errSecurityGroupRuleNotFound is returned when a security group rule cannot
	// be found in its security group.
	errSecurityGroupRuleNotFound = errors.New("security group rule not found in security group")
)

// SecurityGroupRuleID derives a stable, synthetic identifier for a single
// security group rule. The Numspot API has no per-rule ID: a rule is addressed
// by its (protocol, port range, IP range) tuple within a security group, and the
// read endpoint groups every IP range sharing a protocol/port range into one
// SecurityGroupRule entry. We therefore model one Terraform resource per
// (protocol, port range, single IP range) and hash that tuple so Read/Delete can
// locate the rule again. The hash is flow-independent: an identical rule present
// both inbound and outbound collapses to the same ID, which is acceptable since
// the Flow attribute is preserved from prior state and the serialized values are
// identical either way.
//
// rule must carry exactly one IP range (see normalizeRule).
func SecurityGroupRuleID(rule api.SecurityGroupRule) string {
	from := -1
	if rule.FromPortRange != nil {
		from = *rule.FromPortRange
	}
	to := -1
	if rule.ToPortRange != nil {
		to = *rule.ToPortRange
	}
	proto := ""
	if rule.IpProtocol != nil {
		proto = *rule.IpProtocol
	}
	ipRange := ""
	if rule.IpRanges != nil && len(*rule.IpRanges) > 0 {
		ipRange = (*rule.IpRanges)[0]
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d|%d|%s|%s", from, to, proto, ipRange)))
	return hex.EncodeToString(sum[:])
}

// normalizeRule rebuilds a SecurityGroupRule carrying a single IP range, so that
// it represents exactly one logical Terraform rule (and hashes deterministically).
func normalizeRule(rule api.SecurityGroupRule, ipRange string) api.SecurityGroupRule {
	return api.SecurityGroupRule{
		FromPortRange: rule.FromPortRange,
		ToPortRange:   rule.ToPortRange,
		IpProtocol:    rule.IpProtocol,
		IpRanges:      &[]string{ipRange},
	}
}

// findRuleByID scans a flow's grouped rules and returns the single-IP-range rule
// whose synthetic ID matches, or nil when absent.
func findRuleByID(rules *[]api.SecurityGroupRule, id string) *api.SecurityGroupRule {
	if rules == nil {
		return nil
	}
	for _, rule := range *rules {
		if rule.IpRanges == nil {
			continue
		}
		for _, ipRange := range *rule.IpRanges {
			normalized := normalizeRule(rule, ipRange)
			if SecurityGroupRuleID(normalized) == id {
				return &normalized
			}
		}
	}
	return nil
}

// matchCreatedRule locates the grouped rule produced by a create call: same
// protocol and port range, containing the requested IP range. It returns the
// normalized single-IP-range rule (using the API's canonical protocol/ports) so
// the synthetic ID is computed from server-side values.
func matchCreatedRule(rules *[]api.SecurityGroupRule, body *api.CreateSecurityGroupRuleJSONRequestBody) *api.SecurityGroupRule {
	if rules == nil || body.IpRange == nil {
		return nil
	}
	for _, rule := range *rules {
		if !intPtrEqual(rule.FromPortRange, body.FromPortRange) || !intPtrEqual(rule.ToPortRange, body.ToPortRange) {
			continue
		}
		if !strPtrEqual(rule.IpProtocol, body.IpProtocol) {
			continue
		}
		if rule.IpRanges == nil {
			continue
		}
		for _, ipRange := range *rule.IpRanges {
			if ipRange == *body.IpRange {
				normalized := normalizeRule(rule, ipRange)
				return &normalized
			}
		}
	}
	return nil
}

func intPtrEqual(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func strPtrEqual(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// CreateSecurityGroupRule adds a single rule to a security group and returns the
// created rule normalized to one IP range. The generated resource then derives
// the synthetic ID from it via SecurityGroupRuleID.
func CreateSecurityGroupRule(
	ctx context.Context,
	provider *client.NumspotSDK,
	body *api.CreateSecurityGroupRuleJSONRequestBody,
	securityGroupID string,
) (*api.SecurityGroupRule, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	res, err := numspotClient.CreateSecurityGroupRuleWithResponse(ctx, provider.SpaceID, securityGroupID, *body)
	if err != nil {
		return nil, utils.WrapError(err, "failed to create security group rule")
	}
	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to parse HTTP error response")
	}

	securityGroup := res.JSON201
	if securityGroup == nil {
		securityGroup, err = ReadSecurityGroup(ctx, provider, securityGroupID)
		if err != nil {
			return nil, err
		}
	}

	rules := securityGroup.OutboundRules
	if body.Flow == "Inbound" {
		rules = securityGroup.InboundRules
	}

	created := matchCreatedRule(rules, body)
	if created == nil {
		return nil, fmt.Errorf("%w %s", errCreatedSecurityGroupRuleNotFound, securityGroupID)
	}
	return created, nil
}

// ReadSecurityGroupRule re-reads the parent security group and returns the
// single rule matching the synthetic ID. It scans both flows because the ID is
// flow-independent; the caller preserves Flow from prior state.
func ReadSecurityGroupRule(ctx context.Context, provider *client.NumspotSDK, securityGroupID, id string) (*api.SecurityGroupRule, error) {
	securityGroup, err := ReadSecurityGroup(ctx, provider, securityGroupID)
	if err != nil {
		return nil, err
	}

	if rule := findRuleByID(securityGroup.InboundRules, id); rule != nil {
		return rule, nil
	}
	if rule := findRuleByID(securityGroup.OutboundRules, id); rule != nil {
		return rule, nil
	}
	return nil, fmt.Errorf("%w %s: %s", errSecurityGroupRuleNotFound, id, securityGroupID)
}

// DeleteSecurityGroupRule removes the single rule matching the synthetic ID from
// the given flow. It re-reads the security group to recover the rule's
// protocol/port range/IP range (the delete API is addressed by rule body, not
// ID). A rule already absent is treated as successfully deleted.
func DeleteSecurityGroupRule(ctx context.Context, provider *client.NumspotSDK, securityGroupID, id, flow string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	securityGroup, err := ReadSecurityGroup(ctx, provider, securityGroupID)
	if err != nil {
		return err
	}

	rules := securityGroup.OutboundRules
	if flow == "Inbound" {
		rules = securityGroup.InboundRules
	}
	rule := findRuleByID(rules, id)
	if rule == nil {
		return nil
	}

	var ipRange *string
	if rule.IpRanges != nil && len(*rule.IpRanges) > 0 {
		ipRange = &(*rule.IpRanges)[0]
	}
	body := api.DeleteSecurityGroupRuleJSONRequestBody{
		Flow:          flow,
		FromPortRange: rule.FromPortRange,
		ToPortRange:   rule.ToPortRange,
		IpProtocol:    rule.IpProtocol,
		IpRange:       ipRange,
	}

	res, err := numspotClient.DeleteSecurityGroupRuleWithResponse(ctx, provider.SpaceID, securityGroupID, body)
	if err != nil {
		return utils.WrapError(err, "failed to delete security group rule")
	}
	return utils.WrapError(utils.ParseHTTPError(res.Body, res.StatusCode()), "failed to parse HTTP error response")
}
