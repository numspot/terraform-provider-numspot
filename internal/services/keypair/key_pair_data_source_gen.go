// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package keypair

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func KeyPairDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"items": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"fingerprint": schema.StringAttribute{
							Computed:            true,
							Description:         "The MD5 public key fingerprint as specified in section 4 of RFC 4716.",
							MarkdownDescription: "The MD5 public key fingerprint as specified in section 4 of RFC 4716.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							Description:         "The name of the keypair.",
							MarkdownDescription: "The name of the keypair.",
						},
						"type": schema.StringAttribute{
							Computed:            true,
							Description:         "The type of the keypair (`ssh-rsa`, `ssh-ed25519`, `ecdsa-sha2-nistp256`, `ecdsa-sha2-nistp384`, or `ecdsa-sha2-nistp521`).",
							MarkdownDescription: "The type of the keypair (`ssh-rsa`, `ssh-ed25519`, `ecdsa-sha2-nistp256`, `ecdsa-sha2-nistp384`, or `ecdsa-sha2-nistp521`).",
						},
					},
				}, // MANUALLY EDITED : Removed CustomType block
				Computed:            true,
				Description:         "Information about one or more keypairs.",
				MarkdownDescription: "Information about one or more keypairs.",
			},
			"keypair_fingerprints": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The fingerprints of the keypairs.",
				MarkdownDescription: "The fingerprints of the keypairs.",
			},
			"keypair_names": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The names of the keypairs.",
				MarkdownDescription: "The names of the keypairs.",
			},
			"keypair_types": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The types of the keypairs (`ssh-rsa`, `ssh-ed25519`, `ecdsa-sha2-nistp256`, `ecdsa-sha2-nistp384`, or `ecdsa-sha2-nistp521`).",
				MarkdownDescription: "The types of the keypairs (`ssh-rsa`, `ssh-ed25519`, `ecdsa-sha2-nistp256`, `ecdsa-sha2-nistp384`, or `ecdsa-sha2-nistp521`).",
			},
			// MANUALLY EDITED : SpaceId Removed
		},
		DeprecationMessage: "Managing IAAS services with Terraform is deprecated", // MANUALLY EDITED : Add Deprecation message
	}
}

type KeyPairDatasourceItemModel struct { // MANUALLY EDITED : Create Model from ItemsValue struct
	Fingerprint types.String `tfsdk:"fingerprint"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
}

// MANUALLY EDITED : Functions associated with ItemsType / ItemsValue and Tags removed
