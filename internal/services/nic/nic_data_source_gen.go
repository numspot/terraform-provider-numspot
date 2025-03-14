// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package nic

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/services/tags"
)

func NicDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"availability_zone_names": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The Subregions where the NICs are located.",
				MarkdownDescription: "The Subregions where the NICs are located.",
			},
			"descriptions": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The descriptions of the NICs.",
				MarkdownDescription: "The descriptions of the NICs.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the NICs.",
				MarkdownDescription: "The IDs of the NICs.",
			},
			"is_source_dest_check": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the source/destination checking is enabled (true) or disabled (false).",
				MarkdownDescription: "Whether the source/destination checking is enabled (true) or disabled (false).",
			},
			"items": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"availability_zone_name": schema.StringAttribute{
							Computed:            true,
							Description:         "The Subregion in which the NIC is located.",
							MarkdownDescription: "The Subregion in which the NIC is located.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							Description:         "The description of the NIC.",
							MarkdownDescription: "The description of the NIC.",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the NIC.",
							MarkdownDescription: "The ID of the NIC.",
						},
						"is_source_dest_checked": schema.BoolAttribute{
							Computed:            true,
							Description:         "(Vpc only) If true, the source/destination check is enabled. If false, it is disabled. This value must be false for a NAT VM to perform network address translation (NAT) in a Vpc.",
							MarkdownDescription: "(Vpc only) If true, the source/destination check is enabled. If false, it is disabled. This value must be false for a NAT VM to perform network address translation (NAT) in a Vpc.",
						},
						"link_nic": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"delete_on_vm_deletion": schema.BoolAttribute{
									Computed:            true,
									Description:         "If true, the NIC is deleted when the VM is terminated.",
									MarkdownDescription: "If true, the NIC is deleted when the VM is terminated.",
								},
								"device_number": schema.Int64Attribute{
									Computed:            true,
									Description:         "The device index for the NIC attachment (between `1` and `7`, both included).",
									MarkdownDescription: "The device index for the NIC attachment (between `1` and `7`, both included).",
								},
								"id": schema.StringAttribute{
									Computed:            true,
									Description:         "The ID of the NIC to attach.",
									MarkdownDescription: "The ID of the NIC to attach.",
								},
								"state": schema.StringAttribute{
									Computed:            true,
									Description:         "The state of the attachment (`attaching` \\| `attached` \\| `detaching` \\| `detached`).",
									MarkdownDescription: "The state of the attachment (`attaching` \\| `attached` \\| `detaching` \\| `detached`).",
								},
								"vm_id": schema.StringAttribute{
									Computed:            true,
									Description:         "The ID of the VM.",
									MarkdownDescription: "The ID of the VM.",
								},
							},
							CustomType: LinkNicType{
								ObjectType: types.ObjectType{
									AttrTypes: LinkNicValue{}.AttributeTypes(ctx),
								},
							},
							Computed:            true,
							Description:         "Information about the NIC attachment.",
							MarkdownDescription: "Information about the NIC attachment.",
						},
						"link_public_ip": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Computed:            true,
									Description:         "(Required in a Vpc) The ID representing the association of the public IP with the VM or the NIC.",
									MarkdownDescription: "(Required in a Vpc) The ID representing the association of the public IP with the VM or the NIC.",
								},
								"public_dns_name": schema.StringAttribute{
									Computed:            true,
									Description:         "The name of the public DNS.",
									MarkdownDescription: "The name of the public DNS.",
								},
								"public_ip": schema.StringAttribute{
									Computed:            true,
									Description:         "The public IP associated with the NIC.",
									MarkdownDescription: "The public IP associated with the NIC.",
								},
								"public_ip_id": schema.StringAttribute{
									Computed:            true,
									Description:         "The allocation ID of the public IP.",
									MarkdownDescription: "The allocation ID of the public IP.",
								},
							},
							CustomType: LinkPublicIpType{
								ObjectType: types.ObjectType{
									AttrTypes: LinkPublicIpValue{}.AttributeTypes(ctx),
								},
							},
							Computed:            true,
							Description:         "Information about the public IP association.",
							MarkdownDescription: "Information about the public IP association.",
						},
						"mac_address": schema.StringAttribute{
							Computed:            true,
							Description:         "The Media Access Control (MAC) address of the NIC.",
							MarkdownDescription: "The Media Access Control (MAC) address of the NIC.",
						},
						"private_dns_name": schema.StringAttribute{
							Computed:            true,
							Description:         "The name of the private DNS.",
							MarkdownDescription: "The name of the private DNS.",
						},
						"private_ips": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"is_primary": schema.BoolAttribute{
										Computed:            true,
										Description:         "If true, the IP is the primary private IP of the NIC.",
										MarkdownDescription: "If true, the IP is the primary private IP of the NIC.",
									},
									"link_public_ip": schema.SingleNestedAttribute{
										Attributes: map[string]schema.Attribute{
											"id": schema.StringAttribute{
												Computed:            true,
												Description:         "(Required in a Vpc) The ID representing the association of the public IP with the VM or the NIC.",
												MarkdownDescription: "(Required in a Vpc) The ID representing the association of the public IP with the VM or the NIC.",
											},
											"public_dns_name": schema.StringAttribute{
												Computed:            true,
												Description:         "The name of the public DNS.",
												MarkdownDescription: "The name of the public DNS.",
											},
											"public_ip": schema.StringAttribute{
												Computed:            true,
												Description:         "The public IP associated with the NIC.",
												MarkdownDescription: "The public IP associated with the NIC.",
											},
											"public_ip_id": schema.StringAttribute{
												Computed:            true,
												Description:         "The allocation ID of the public IP.",
												MarkdownDescription: "The allocation ID of the public IP.",
											},
										},
										CustomType: LinkPublicIpType{
											ObjectType: types.ObjectType{
												AttrTypes: LinkPublicIpValue{}.AttributeTypes(ctx),
											},
										},
										Computed:            true,
										Description:         "Information about the public IP association.",
										MarkdownDescription: "Information about the public IP association.",
									},
									"private_dns_name": schema.StringAttribute{
										Computed:            true,
										Description:         "The name of the private DNS.",
										MarkdownDescription: "The name of the private DNS.",
									},
									"private_ip": schema.StringAttribute{
										Computed:            true,
										Description:         "The private IP of the NIC.",
										MarkdownDescription: "The private IP of the NIC.",
									},
								},
								CustomType: PrivateIpsType{
									ObjectType: types.ObjectType{
										AttrTypes: PrivateIpsValue{}.AttributeTypes(ctx),
									},
								},
							},
							Computed:            true,
							Description:         "The private IPs of the NIC.",
							MarkdownDescription: "The private IPs of the NIC.",
						},
						"security_groups": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"security_group_id": schema.StringAttribute{
										Computed:            true,
										Description:         "The ID of the security group.",
										MarkdownDescription: "The ID of the security group.",
									},
									"security_group_name": schema.StringAttribute{
										Computed:            true,
										Description:         "The name of the security group.",
										MarkdownDescription: "The name of the security group.",
									},
								},
								CustomType: SecurityGroupsType{
									ObjectType: types.ObjectType{
										AttrTypes: SecurityGroupsValue{}.AttributeTypes(ctx),
									},
								},
							},
							Computed:            true,
							Description:         "One or more IDs of security groups for the NIC.",
							MarkdownDescription: "One or more IDs of security groups for the NIC.",
						},
						"state": schema.StringAttribute{
							Computed:            true,
							Description:         "The state of the NIC (`available` \\| `attaching` \\| `in-use` \\| `detaching`).",
							MarkdownDescription: "The state of the NIC (`available` \\| `attaching` \\| `in-use` \\| `detaching`).",
						},
						"subnet_id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the Subnet.",
							MarkdownDescription: "The ID of the Subnet.",
						},
						"tags": tags.TagsSchema(ctx), // MANUALLY EDITED : Use shared tags

						"vpc_id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the Vpc for the NIC.",
							MarkdownDescription: "The ID of the Vpc for the NIC.",
						},
					},
				}, // MANUALLY EDITED : Removed CustomType block
				Computed:            true,
				Description:         "Information about one or more NICs.",
				MarkdownDescription: "Information about one or more NICs.",
			},
			"link_nic_delete_on_vm_deletion": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the NICs are deleted when the VMs they are attached to are terminated.",
				MarkdownDescription: "Whether the NICs are deleted when the VMs they are attached to are terminated.",
			},
			"link_nic_device_numbers": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Optional:            true,
				Computed:            true,
				Description:         "The device numbers the NICs are attached to.",
				MarkdownDescription: "The device numbers the NICs are attached to.",
			},
			"link_nic_link_nic_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The attachment IDs of the NICs.",
				MarkdownDescription: "The attachment IDs of the NICs.",
			},
			"link_nic_states": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The states of the attachments.",
				MarkdownDescription: "The states of the attachments.",
			},
			"link_nic_vm_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the VMs the NICs are attached to.",
				MarkdownDescription: "The IDs of the VMs the NICs are attached to.",
			},
			"link_public_ip_link_public_ip_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The association IDs returned when the public IPs were associated with the NICs.",
				MarkdownDescription: "The association IDs returned when the public IPs were associated with the NICs.",
			},
			"link_public_ip_public_ip_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The allocation IDs returned when the public IPs were allocated to their accounts.",
				MarkdownDescription: "The allocation IDs returned when the public IPs were allocated to their accounts.",
			},
			"link_public_ip_public_ips": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The public IPs associated with the NICs.",
				MarkdownDescription: "The public IPs associated with the NICs.",
			},
			"mac_addresses": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The Media Access Control (MAC) addresses of the NICs.",
				MarkdownDescription: "The Media Access Control (MAC) addresses of the NICs.",
			},
			"private_dns_names": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The private DNS names associated with the primary private IPs.",
				MarkdownDescription: "The private DNS names associated with the primary private IPs.",
			},
			"private_ips_link_public_ip_public_ips": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The public IPs associated with the private IPs.",
				MarkdownDescription: "The public IPs associated with the private IPs.",
			},
			"private_ips_primary_ip": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the private IP is the primary IP associated with the NIC.",
				MarkdownDescription: "Whether the private IP is the primary IP associated with the NIC.",
			},
			"private_ips_private_ips": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The private IPs of the NICs.",
				MarkdownDescription: "The private IPs of the NICs.",
			},
			"security_group_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the security groups associated with the NICs.",
				MarkdownDescription: "The IDs of the security groups associated with the NICs.",
			},
			"security_group_names": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The names of the security groups associated with the NICs.",
				MarkdownDescription: "The names of the security groups associated with the NICs.",
			},
			"states": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The states of the NICs.",
				MarkdownDescription: "The states of the NICs.",
			},
			"subnet_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the Subnets for the NICs.",
				MarkdownDescription: "The IDs of the Subnets for the NICs.",
			},
			"tag_keys": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The keys of the tags associated with the NICs.",
				MarkdownDescription: "The keys of the tags associated with the NICs.",
			},
			"tag_values": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The values of the tags associated with the NICs.",
				MarkdownDescription: "The values of the tags associated with the NICs.",
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The key/value combination of the tags associated with the NICs, in the following format: \"Filters\":{\"Tags\":[\"TAGKEY=TAGVALUE\"]}.", // MANUALLY EDITED : replaced HTML encoded character
				MarkdownDescription: "The key/value combination of the tags associated with the NICs, in the following format: \"Filters\":{\"Tags\":[\"TAGKEY=TAGVALUE\"]}.", // MANUALLY EDITED : replaced HTML encoded character
			},
			"vpc_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the Vpcs where the NICs are located.",
				MarkdownDescription: "The IDs of the Vpcs where the NICs are located.",
			},
			// MANUALLY EDITED : SpaceId Removed
		},
	}
}

type NicModelDatasource struct { // MANUALLY EDITED : Create Model from ItemsValue struct
	AvailabilityZoneName types.String      `tfsdk:"availability_zone_name"`
	Description          types.String      `tfsdk:"description"`
	Id                   types.String      `tfsdk:"id"`
	IsSourceDestChecked  types.Bool        `tfsdk:"is_source_dest_checked"`
	LinkNic              LinkNicValue      `tfsdk:"link_nic"`
	LinkPublicIp         LinkPublicIpValue `tfsdk:"link_public_ip"`
	MacAddress           types.String      `tfsdk:"mac_address"`
	PrivateDnsName       types.String      `tfsdk:"private_dns_name"`
	PrivateIps           types.List        `tfsdk:"private_ips"`
	SecurityGroups       types.List        `tfsdk:"security_groups"`
	State                types.String      `tfsdk:"state"`
	SubnetId             types.String      `tfsdk:"subnet_id"`
	Tags                 types.List        `tfsdk:"tags"`
	VpcId                types.String      `tfsdk:"vpc_id"`
}

// MANUALLY EDITED : Functions associated with ItemsType / ItemsValue and Tags removed
