// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package securitygroup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
)

func SecurityGroupDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"items": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Computed:            true,
							Description:         "The description of the security group.",
							MarkdownDescription: "The description of the security group.",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the security group.",
							MarkdownDescription: "The ID of the security group.",
						},
						"inbound_rules": schema.SetNestedAttribute{ // MANUALLY EDITED : Use Set type instead of List
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"from_port_range": schema.Int64Attribute{
										Computed:            true,
										Optional:            true, // MANUALLY EDITED : Add Optional attribute
										Description:         "The beginning of the port range for the TCP and UDP protocols, or an ICMP type number.",
										MarkdownDescription: "The beginning of the port range for the TCP and UDP protocols, or an ICMP type number.",
									},
									"ip_protocol": schema.StringAttribute{
										Computed:            true,
										Optional:            true, // MANUALLY EDITED : Add Optional attribute
										Description:         "The IP protocol name (`tcp`, `udp`, `icmp`, or `-1` for all protocols). By default, `-1`. In a Vpc, this can also be an IP protocol number. For more information, see the [IANA.org website](https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml).",
										MarkdownDescription: "The IP protocol name (`tcp`, `udp`, `icmp`, or `-1` for all protocols). By default, `-1`. In a Vpc, this can also be an IP protocol number. For more information, see the [IANA.org website](https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml).",
									},
									"ip_ranges": schema.ListAttribute{
										ElementType:         types.StringType,
										Computed:            true,
										Optional:            true, // MANUALLY EDITED : Add Optional attribute
										Description:         "One or more IP ranges for the security group rules, in CIDR notation (for example, `10.0.0.0/16`).",
										MarkdownDescription: "One or more IP ranges for the security group rules, in CIDR notation (for example, `10.0.0.0/16`).",
									},
									"security_groups_members": schema.ListNestedAttribute{
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"security_group_id": schema.StringAttribute{
													Computed:            true,
													Optional:            true, // MANUALLY EDITED : Add Optional attribute
													Description:         "The ID of a source or destination security group that you want to link to the security group of the rule.",
													MarkdownDescription: "The ID of a source or destination security group that you want to link to the security group of the rule.",
												},
												"security_group_name": schema.StringAttribute{
													Computed:            true,
													Optional:            true, // MANUALLY EDITED : Add Optional attribute
													Description:         "(Public Cloud only) The name of a source or destination security group that you want to link to the security group of the rule.",
													MarkdownDescription: "(Public Cloud only) The name of a source or destination security group that you want to link to the security group of the rule.",
												},
											},
											CustomType: SecurityGroupsMembersType{
												ObjectType: types.ObjectType{
													AttrTypes: SecurityGroupsMembersValue{}.AttributeTypes(ctx),
												},
											},
										},
										Computed:            true,
										Optional:            true, // MANUALLY EDITED : Add Optional attribute
										Description:         "Information about one or more source or destination security groups.",
										MarkdownDescription: "Information about one or more source or destination security groups.",
									},
									"service_ids": schema.ListAttribute{
										ElementType:         types.StringType,
										Computed:            true,
										Optional:            true, // MANUALLY EDITED : Add Optional attribute
										Description:         "One or more service IDs to allow traffic from a Vpc to access the corresponding NumSpot services.",
										MarkdownDescription: "One or more service IDs to allow traffic from a Vpc to access the corresponding NumSpot services.",
									},
									"to_port_range": schema.Int64Attribute{
										Computed:            true,
										Optional:            true, // MANUALLY EDITED : Add Optional attribute
										Description:         "The end of the port range for the TCP and UDP protocols, or an ICMP code number.",
										MarkdownDescription: "The end of the port range for the TCP and UDP protocols, or an ICMP code number.",
									},
								},
								CustomType: InboundRulesType{
									ObjectType: types.ObjectType{
										AttrTypes: InboundRulesValue{}.AttributeTypes(ctx),
									},
								},
							},
							Computed:            true,
							Optional:            true, // MANUALLY EDITED : Add Optional attribute
							Description:         "The inbound rules associated with the security group.",
							MarkdownDescription: "The inbound rules associated with the security group.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							Description:         "The name of the security group.",
							MarkdownDescription: "The name of the security group.",
						},
						"outbound_rules": schema.SetNestedAttribute{ // MANUALLY EDITED : Use Set type instead of List
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"from_port_range": schema.Int64Attribute{
										Computed:            true,
										Optional:            true, // MANUALLY EDITED : Add Optional attribute
										Description:         "The beginning of the port range for the TCP and UDP protocols, or an ICMP type number.",
										MarkdownDescription: "The beginning of the port range for the TCP and UDP protocols, or an ICMP type number.",
									},
									"ip_protocol": schema.StringAttribute{
										Computed:            true,
										Optional:            true, // MANUALLY EDITED : Add Optional attribute
										Description:         "The IP protocol name (`tcp`, `udp`, `icmp`, or `-1` for all protocols). By default, `-1`. In a Vpc, this can also be an IP protocol number. For more information, see the [IANA.org website](https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml).",
										MarkdownDescription: "The IP protocol name (`tcp`, `udp`, `icmp`, or `-1` for all protocols). By default, `-1`. In a Vpc, this can also be an IP protocol number. For more information, see the [IANA.org website](https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml).",
									},
									"ip_ranges": schema.ListAttribute{
										ElementType:         types.StringType,
										Computed:            true,
										Optional:            true, // MANUALLY EDITED : Add Optional attribute
										Description:         "One or more IP ranges for the security group rules, in CIDR notation (for example, `10.0.0.0/16`).",
										MarkdownDescription: "One or more IP ranges for the security group rules, in CIDR notation (for example, `10.0.0.0/16`).",
									},
									"security_groups_members": schema.ListNestedAttribute{
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"security_group_id": schema.StringAttribute{
													Computed:            true,
													Optional:            true, // MANUALLY EDITED : Add Optional attribute
													Description:         "The ID of a source or destination security group that you want to link to the security group of the rule.",
													MarkdownDescription: "The ID of a source or destination security group that you want to link to the security group of the rule.",
												},
												"security_group_name": schema.StringAttribute{
													Computed:            true,
													Optional:            true, // MANUALLY EDITED : Add Optional attribute
													Description:         "(Public Cloud only) The name of a source or destination security group that you want to link to the security group of the rule.",
													MarkdownDescription: "(Public Cloud only) The name of a source or destination security group that you want to link to the security group of the rule.",
												},
											},
											CustomType: SecurityGroupsMembersType{
												ObjectType: types.ObjectType{
													AttrTypes: SecurityGroupsMembersValue{}.AttributeTypes(ctx),
												},
											},
										},
										Computed:            true,
										Optional:            true, // MANUALLY EDITED : Add Optional attribute
										Description:         "Information about one or more source or destination security groups.",
										MarkdownDescription: "Information about one or more source or destination security groups.",
									},
									"service_ids": schema.ListAttribute{
										ElementType:         types.StringType,
										Computed:            true,
										Optional:            true, // MANUALLY EDITED : Add Optional attribute
										Description:         "One or more service IDs to allow traffic from a Vpc to access the corresponding NumSpot services.",
										MarkdownDescription: "One or more service IDs to allow traffic from a Vpc to access the corresponding NumSpot services.",
									},
									"to_port_range": schema.Int64Attribute{
										Computed:            true,
										Optional:            true, // MANUALLY EDITED : Add Optional attribute
										Description:         "The end of the port range for the TCP and UDP protocols, or an ICMP code number.",
										MarkdownDescription: "The end of the port range for the TCP and UDP protocols, or an ICMP code number.",
									},
								},
								CustomType: OutboundRulesType{
									ObjectType: types.ObjectType{
										AttrTypes: OutboundRulesValue{}.AttributeTypes(ctx),
									},
								},
							},
							Computed:            true,
							Optional:            true, // MANUALLY EDITED : Add Optional attribute
							Description:         "The outbound rules associated with the security group.",
							MarkdownDescription: "The outbound rules associated with the security group.",
						},
						"tags": tags.TagsSchema(ctx), // MANUALLY EDITED : Use shared tags
						"vpc_id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the Vpc for the security group.",
							MarkdownDescription: "The ID of the Vpc for the security group.",
						},
					},
				}, // MANUALLY EDITED : Removed CustomType block
				Computed:            true,
				Description:         "Information about one or more security groups.",
				MarkdownDescription: "Information about one or more security groups.",
			},
			"descriptions": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The descriptions of the security groups.",
				MarkdownDescription: "The descriptions of the security groups.",
			},
			"inbound_rule_from_port_ranges": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Optional:            true,
				Computed:            true,
				Description:         "The beginnings of the port ranges for the TCP and UDP protocols, or the ICMP type numbers.",
				MarkdownDescription: "The beginnings of the port ranges for the TCP and UDP protocols, or the ICMP type numbers.",
			},
			"inbound_rule_ip_ranges": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IP ranges that have been granted permissions, in CIDR notation (for example, `10.0.0.0/24`).",
				MarkdownDescription: "The IP ranges that have been granted permissions, in CIDR notation (for example, `10.0.0.0/24`).",
			},
			"inbound_rule_protocols": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IP protocols for the permissions (`tcp` \\| `udp` \\| `icmp`, or a protocol number, or `-1` for all protocols).",
				MarkdownDescription: "The IP protocols for the permissions (`tcp` \\| `udp` \\| `icmp`, or a protocol number, or `-1` for all protocols).",
			},
			"inbound_rule_security_group_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the security groups that have been granted permissions.",
				MarkdownDescription: "The IDs of the security groups that have been granted permissions.",
			},
			"inbound_rule_security_group_names": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The names of the security groups that have been granted permissions.",
				MarkdownDescription: "The names of the security groups that have been granted permissions.",
			},
			"inbound_rule_to_port_ranges": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Optional:            true,
				Computed:            true,
				Description:         "The ends of the port ranges for the TCP and UDP protocols, or the ICMP code numbers.",
				MarkdownDescription: "The ends of the port ranges for the TCP and UDP protocols, or the ICMP code numbers.",
			},
			"outbound_rule_from_port_ranges": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Optional:            true,
				Computed:            true,
				Description:         "The beginnings of the port ranges for the TCP and UDP protocols, or the ICMP type numbers.",
				MarkdownDescription: "The beginnings of the port ranges for the TCP and UDP protocols, or the ICMP type numbers.",
			},
			"outbound_rule_ip_ranges": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IP ranges that have been granted permissions, in CIDR notation (for example, `10.0.0.0/24`).",
				MarkdownDescription: "The IP ranges that have been granted permissions, in CIDR notation (for example, `10.0.0.0/24`).",
			},
			"outbound_rule_protocols": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IP protocols for the permissions (`tcp` \\| `udp` \\| `icmp`, or a protocol number, or `-1` for all protocols).",
				MarkdownDescription: "The IP protocols for the permissions (`tcp` \\| `udp` \\| `icmp`, or a protocol number, or `-1` for all protocols).",
			},
			"outbound_rule_security_group_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the security groups that have been granted permissions.",
				MarkdownDescription: "The IDs of the security groups that have been granted permissions.",
			},
			"outbound_rule_security_group_names": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The names of the security groups that have been granted permissions.",
				MarkdownDescription: "The names of the security groups that have been granted permissions.",
			},
			"outbound_rule_to_port_ranges": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Optional:            true,
				Computed:            true,
				Description:         "The ends of the port ranges for the TCP and UDP protocols, or the ICMP code numbers.",
				MarkdownDescription: "The ends of the port ranges for the TCP and UDP protocols, or the ICMP code numbers.",
			},
			"security_group_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the security groups.",
				MarkdownDescription: "The IDs of the security groups.",
			},
			"security_group_names": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The names of the security groups.",
				MarkdownDescription: "The names of the security groups.",
			},
			"tag_keys": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The keys of the tags associated with the security groups.",
				MarkdownDescription: "The keys of the tags associated with the security groups.",
			},
			"tag_values": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The values of the tags associated with the security groups.",
				MarkdownDescription: "The values of the tags associated with the security groups.",
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The key/value combination of the tags associated with the security groups, in the following format: \"Filters\":{\"Tags\":[\"TAGKEY=TAGVALUE\"]}.", // MANUALLY EDITED : replaced HTML encoded character
				MarkdownDescription: "The key/value combination of the tags associated with the security groups, in the following format: \"Filters\":{\"Tags\":[\"TAGKEY=TAGVALUE\"]}.", // MANUALLY EDITED : replaced HTML encoded character
			},
			"vpc_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the Vpcs specified when the security groups were created.",
				MarkdownDescription: "The IDs of the Vpcs specified when the security groups were created.",
			},
			// MANUALLY EDITED : SpaceId removed
		},
		DeprecationMessage: "Managing IAAS services with Terraform is deprecated", // MANUALLY EDITED : Add Deprecation message
	}
}

// MANUALLY EDITED : Model declaration removed

// MANUALLY EDITED : Functions associated with ItemsType / ItemsValue and Tags removed
