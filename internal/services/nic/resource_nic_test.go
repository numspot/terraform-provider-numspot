//go:build acc

package nic_test

import (
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type Tag struct {
	key   string `hcl:"key"`
	value string `hcl:"value"`
}

// This struct will store the input data that will be used in your tests (all fields as string)
type vpc struct {
	*utils.BlockDefinition
	ipRange string `hcl:"ip_range,noref"`
}

// This struct will store the input data that will be used in your tests (all fields as string)
type subnet struct {
	*utils.BlockDefinition
	vpcID   string `hcl:"vpc_id"`
	ipRange string `hcl:"ip_range,noref"`
}

type vm struct {
	*utils.BlockDefinition
	imageID  string `hcl:"image_id"`
	vmType   string `hcl:"type"`
	subnetID string `hcl:"subnet_id"`
}

type linkNIC struct {
	vmID         string `hcl:"vm_id"`
	deviceNumber int    `hcl:"device_number"`
}

type SGRule struct {
	fromPortRange int      `hcl:"from_port_range"`
	toPortRange   int      `hcl:"to_port_range"`
	ipRanges      []string `hcl:"ip_ranges,noref"`
	ipProtocol    string   `hcl:"ip_protocol"`
}
type securityGroup struct {
	*utils.BlockDefinition
	vpcID        string   `hcl:"vpc_id"`
	name         string   `hcl:"name"`
	description  string   `hcl:"description"`
	inboundRules []SGRule `hcl:"inbound_rules"`
}

type nic struct {
	*utils.BlockDefinition
	subnetID    string  `hcl:"subnet_id"`
	description string  `hcl:"description"`
	linkNIC     linkNIC `hcl:"link_nic"`
	// vmName       string `hcl:"vm_name"`
	deviceNumber     int      `hcl:"device_number"`
	securityGroupIDs []string `hcl:"security_group_ids"`
	tags             []Tag    `hcl:"tags"`
}

var nicResourceID, linkNicID string

// Generate checks to validate that resource 'numspot_nic.test' has input data values
func assertionNIC(data nic) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_nic.test", "description", data.description),
		resource.TestCheckResourceAttr("numspot_nic.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "tags.*", map[string]string{
			"key":   data.tags[0].key,
			"value": data.tags[0].value,
		}),
	}
}

// Generate checks to validate that resource 'numspot_nic.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func dependencyAssertionNIC(dependenciesSuffix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_nic.test", "subnet_id", "numspot_subnet.test"+dependenciesSuffix, "id"),
		resource.TestCheckTypeSetElemAttrPair("numspot_nic.test", "security_group_ids.*", "numspot_security_group.test"+dependenciesSuffix, "id"),
	}
}

func TestAccNicResource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	createPlan, createChecks := createNIC(t)
	updatePlan, updateChecks := updateNIC(t)
	updateLinkUnlinkVMPlan, updateLinkUnlinkVMChecks := updateNICLinkUnlinkVM(t)
	replacePlan, replaceChecks := replaceNIC(t)

	/////////////////////////////////////////////////////////////////////////////////////
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // Create testing
				Config: createPlan,
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					dependencyAssertionNIC(provider.BASE_SUFFIX),
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_nic.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: updatePlan,
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					dependencyAssertionNIC(provider.BASE_SUFFIX),
				)...),
			},
			//// Update testing with unlink old VM and link new VM
			{
				Config: updateLinkUnlinkVMPlan,
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateLinkUnlinkVMChecks,
					dependencyAssertionNIC(provider.BASE_SUFFIX),
				)...),
			},
			{
				Config: replacePlan,
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					dependencyAssertionNIC(provider.NEW_SUFFIX),
				)...),
			},
		},
	})
}

func createNIC(t *testing.T) (string, []resource.TestCheckFunc) {
	vpcObj := vpc{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_vpc",
			Labels: []string{"test"},
		},
		ipRange: "10.101.0.0/16",
	}
	subnetObj := subnet{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_subnet",
			Labels: []string{"test"},
		},
		vpcID:   "numspot_vpc.test.id",
		ipRange: "10.101.1.0/24",
	}
	sgObj := securityGroup{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_security_group",
			Labels: []string{"test"},
		},
		vpcID:       "numspot_vpc.test.id",
		name:        "security_group",
		description: "numspot_security_group description",
		inboundRules: []SGRule{
			{
				fromPortRange: 80,
				toPortRange:   80,
				ipRanges:      []string{"0.0.0.0/0"},
				ipProtocol:    "tcp",
			},
		},
	}
	vmObj := vm{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_vm",
			Labels: []string{"test"},
		},
		subnetID: "numspot_subnet.test.id",
		vmType:   "ns-cus6-2c4r",
		imageID:  "ami-0b7df82c",
	}
	nicObj := nic{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_nic",
			Labels: []string{"test"},
		},
		linkNIC: linkNIC{
			deviceNumber: 1,
			vmID:         "numspot_vm.test.id",
		},
		subnetID:         "numspot_subnet.test.id",
		description:      "The nic",
		deviceNumber:     1,
		securityGroupIDs: []string{"numspot_security_group.test.id"},
		tags: []Tag{
			{
				key:   "name",
				value: "Terraform-Test-Volume",
			},
		},
	}

	checks := append(
		assertionNIC(nicObj),

		resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			nicResourceID = v
			return nil
		}),
		resource.TestCheckResourceAttrWith("numspot_nic.test", "link_nic.id", func(v string) error {
			linkNicID = v
			require.NotEmpty(t, v)
			return nil
		}),
	)

	vpcHCL, err := utils.Marshal(&vpcObj)
	require.NoError(t, err)
	subnetHCL, err := utils.Marshal(&subnetObj)
	require.NoError(t, err)
	sgHCL, err := utils.Marshal(&sgObj)
	require.NoError(t, err)
	vmHCL, err := utils.Marshal(&vmObj)
	require.NoError(t, err)
	nicHCL, err := utils.Marshal(&nicObj)
	require.NoError(t, err)

	plan := string(vpcHCL) + string(subnetHCL) + string(sgHCL) + string(vmHCL) + string(nicHCL)
	return plan, checks
}

func updateNIC(t *testing.T) (string, []resource.TestCheckFunc) {
	vpcObj := vpc{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_vpc",
			Labels: []string{"test"},
		},
		ipRange: "10.101.0.0/16",
	}
	subnetObj := subnet{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_subnet",
			Labels: []string{"test"},
		},
		vpcID:   "numspot_vpc.test.id",
		ipRange: "10.101.1.0/24",
	}
	sgObj := securityGroup{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_security_group",
			Labels: []string{"test"},
		},
		vpcID:       "numspot_vpc.test.id",
		name:        "security_group",
		description: "numspot_security_group description",
		inboundRules: []SGRule{
			{
				fromPortRange: 80,
				toPortRange:   80,
				ipRanges:      []string{"0.0.0.0/0"},
				ipProtocol:    "tcp",
			},
		},
	}
	vmObj := vm{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_vm",
			Labels: []string{"test"},
		},
		subnetID: "numspot_subnet.test.id",
		vmType:   "ns-cus6-2c4r",
		imageID:  "ami-0b7df82c",
	}
	nicObj := nic{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_nic",
			Labels: []string{"test"},
		},
		linkNIC: linkNIC{
			deviceNumber: 1,
			vmID:         "numspot_vm.test.id",
		},
		subnetID:         "numspot_subnet.test.id",
		description:      "The better nic",
		deviceNumber:     1,
		securityGroupIDs: []string{"numspot_security_group.test.id"},
		tags: []Tag{
			{
				key:   "name",
				value: "Terraform-Test-Volume-Update",
			},
		},
	}

	checks := append(
		assertionNIC(nicObj),

		resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, nicResourceID)
			return nil
		}),
	)

	vpcHCL, err := utils.Marshal(&vpcObj)
	require.NoError(t, err)
	subnetHCL, err := utils.Marshal(&subnetObj)
	require.NoError(t, err)
	sgHCL, err := utils.Marshal(&sgObj)
	require.NoError(t, err)
	vmHCL, err := utils.Marshal(&vmObj)
	require.NoError(t, err)
	nicHCL, err := utils.Marshal(&nicObj)
	require.NoError(t, err)

	plan := string(vpcHCL) + string(subnetHCL) + string(sgHCL) + string(vmHCL) + string(nicHCL)
	return plan, checks
}

func updateNICLinkUnlinkVM(t *testing.T) (string, []resource.TestCheckFunc) {
	vpcObj := vpc{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_vpc",
			Labels: []string{"test"},
		},
		ipRange: "10.101.0.0/16",
	}
	subnetObj := subnet{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_subnet",
			Labels: []string{"test"},
		},
		vpcID:   "numspot_vpc.test.id",
		ipRange: "10.101.1.0/24",
	}
	sgObj := securityGroup{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_security_group",
			Labels: []string{"test"},
		},
		vpcID:       "numspot_vpc.test.id",
		name:        "security_group",
		description: "numspot_security_group description",
		inboundRules: []SGRule{
			{
				fromPortRange: 80,
				toPortRange:   80,
				ipRanges:      []string{"0.0.0.0/0"},
				ipProtocol:    "tcp",
			},
		},
	}
	vmObj := vm{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_vm",
			Labels: []string{"test-updated"},
		},
		subnetID: "numspot_subnet.test.id",
		vmType:   "ns-cus6-2c4r",
		imageID:  "ami-0b7df82c",
	}
	nicObj := nic{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_nic",
			Labels: []string{"test"},
		},
		linkNIC: linkNIC{
			deviceNumber: 1,
			vmID:         "numspot_vm.test-updated.id",
		},
		subnetID:         "numspot_subnet.test.id",
		description:      "The better nic",
		deviceNumber:     1,
		securityGroupIDs: []string{"numspot_security_group.test.id"},
		tags: []Tag{
			{
				key:   "name",
				value: "Terraform-Test-Volume-Update",
			},
		},
	}

	checks := append(
		assertionNIC(nicObj),

		resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, nicResourceID)
			return nil
		}),
	)

	vpcHCL, err := utils.Marshal(&vpcObj)
	require.NoError(t, err)
	subnetHCL, err := utils.Marshal(&subnetObj)
	require.NoError(t, err)
	sgHCL, err := utils.Marshal(&sgObj)
	require.NoError(t, err)
	vmHCL, err := utils.Marshal(&vmObj)
	require.NoError(t, err)
	nicHCL, err := utils.Marshal(&nicObj)
	require.NoError(t, err)

	plan := string(vpcHCL) + string(subnetHCL) + string(sgHCL) + string(vmHCL) + string(nicHCL)
	return plan, checks
}

func replaceNIC(t *testing.T) (string, []resource.TestCheckFunc) {
	vpcObj := vpc{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_vpc",
			Labels: []string{"testnew"},
		},
		ipRange: "10.101.0.0/16",
	}
	subnetObj := subnet{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_subnet",
			Labels: []string{"testnew"},
		},
		vpcID:   "numspot_vpc.testnew.id",
		ipRange: "10.101.1.0/24",
	}
	sgObj := securityGroup{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_security_group",
			Labels: []string{"testnew"},
		},
		vpcID:       "numspot_vpc.testnew.id",
		name:        "security_group",
		description: "numspot_security_group description",
		inboundRules: []SGRule{
			{
				fromPortRange: 80,
				toPortRange:   80,
				ipRanges:      []string{"0.0.0.0/0"},
				ipProtocol:    "tcp",
			},
		},
	}
	vmObj := vm{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_vm",
			Labels: []string{"testnew"},
		},
		subnetID: "numspot_subnet.testnew.id",
		vmType:   "ns-cus6-2c4r",
		imageID:  "ami-0b7df82c",
	}
	nicObj := nic{
		BlockDefinition: &utils.BlockDefinition{
			Type:   "resource",
			Name:   "numspot_nic",
			Labels: []string{"test"},
		},
		linkNIC: linkNIC{
			deviceNumber: 1,
			vmID:         "numspot_vm.testnew.id",
		},
		subnetID:         "numspot_subnet.testnew.id",
		description:      "The better nic",
		deviceNumber:     1,
		securityGroupIDs: []string{"numspot_security_group.testnew.id"},
		tags: []Tag{
			{
				key:   "name",
				value: "Terraform-Test-Volume-Update",
			},
		},
	}

	checks := append(
		assertionNIC(nicObj),

		resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.NotEqual(t, v, nicResourceID)
			return nil
		}),
		resource.TestCheckResourceAttrWith("numspot_nic.test", "link_nic.id", func(v string) error {
			require.NotEmpty(t, v)
			require.NotEqual(t, v, linkNicID)
			return nil
		}),
	)

	vpcHCL, err := utils.Marshal(&vpcObj)
	require.NoError(t, err)
	subnetHCL, err := utils.Marshal(&subnetObj)
	require.NoError(t, err)
	sgHCL, err := utils.Marshal(&sgObj)
	require.NoError(t, err)
	vmHCL, err := utils.Marshal(&vmObj)
	require.NoError(t, err)
	nicHCL, err := utils.Marshal(&nicObj)
	require.NoError(t, err)

	plan := string(vpcHCL) + string(subnetHCL) + string(sgHCL) + string(vmHCL) + string(nicHCL)
	return plan, checks
}
