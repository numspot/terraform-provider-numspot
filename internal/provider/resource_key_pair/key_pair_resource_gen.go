// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_key_pair

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func KeyPairResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"fingerprint": schema.StringAttribute{
				Computed:            true,
				Description:         "The MD5 public key fingerprint as specified in section 4 of RFC 4716.",
				MarkdownDescription: "The MD5 public key fingerprint as specified in section 4 of RFC 4716.",
			},
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "ID for /keypairs",
				MarkdownDescription: "ID for /keypairs",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "A unique name for the keypair, with a maximum length of 255 [ASCII printable characters](https://en.wikipedia.org/wiki/ASCII#Printable_characters).",
				MarkdownDescription: "A unique name for the keypair, with a maximum length of 255 [ASCII printable characters](https://en.wikipedia.org/wiki/ASCII#Printable_characters).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"private_key": schema.StringAttribute{
				Computed:            true,
				Description:         "The private key. When saving the private key in a .rsa file, replace the `\\n` escape sequences with line breaks.",
				MarkdownDescription: "The private key. When saving the private key in a .rsa file, replace the `\\n` escape sequences with line breaks.",
			},
			"public_key": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The public key. It must be Base64-encoded.",
				MarkdownDescription: "The public key. It must be Base64-encoded.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		DeprecationMessage: "Managing IAAS services with Terraform is deprecated",
	}
}

type KeyPairModel struct {
	Fingerprint types.String `tfsdk:"fingerprint"`
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	PrivateKey  types.String `tfsdk:"private_key"`
	PublicKey   types.String `tfsdk:"public_key"`
}
