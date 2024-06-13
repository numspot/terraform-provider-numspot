// Code generated from opl_model.ts file. DO NOT EDIT.
package lib

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

// List of all [Service] names in the OPL.
var nameToService = map[string]Service{
	"compute":           ServiceCompute,
	"datamanagement":    ServiceDatamanagement,
	"dhcp":              ServiceDhcp,
	"iam":               ServiceIam,
	"identity":          ServiceIdentity,
	"internal":          ServiceInternal,
	"kubernetes":        ServiceKubernetes,
	"loadbalancing":     ServiceLoadbalancing,
	"marketplace":       ServiceMarketplace,
	"metadata":          ServiceMetadata,
	"monitoringlogging": ServiceMonitoringlogging,
	"network":           ServiceNetwork,
	"openshift":         ServiceOpenshift,
	"postgres":          ServicePostgres,
	"region":            ServiceRegion,
	"routing":           ServiceRouting,
	"security":          ServiceSecurity,
	"serviceusage":      ServiceServiceusage,
	"storageblock":      ServiceStorageblock,
	"storageobject":     ServiceStorageobject,
}

// List of all [Service] in the OPL.
var (
	ServiceCompute           = Service{map[string]Resource{"flexibleGpu": ResourceComputeFlexibleGpu, "link": ResourceComputeLink, "vm": ResourceComputeVm, "vmGroup": ResourceComputeVmGroup, "vmTemplate": ResourceComputeVmTemplate}}
	ServiceDatamanagement    = Service{map[string]Resource{"machineImage": ResourceDatamanagementMachineImage}}
	ServiceDhcp              = Service{map[string]Resource{"dhcpOption": ResourceDhcpDhcpOption}}
	ServiceIam               = Service{map[string]Resource{"permission": ResourceIamPermission, "role": ResourceIamRole, "tenant": ResourceIamTenant}}
	ServiceIdentity          = Service{map[string]Resource{"group": ResourceIdentityGroup, "serviceAccount": ResourceIdentityServiceAccount, "user": ResourceIdentityUser}}
	ServiceInternal          = Service{map[string]Resource{"root": ResourceInternalRoot, "serviceAccountNS": ResourceInternalServiceAccountNS}}
	ServiceKubernetes        = Service{map[string]Resource{"cluster": ResourceKubernetesCluster}}
	ServiceLoadbalancing     = Service{map[string]Resource{"internetGateway": ResourceLoadbalancingInternetGateway, "link": ResourceLoadbalancingLink, "listener": ResourceLoadbalancingListener, "listenerRule": ResourceLoadbalancingListenerRule, "loadBalancer": ResourceLoadbalancingLoadBalancer, "policy": ResourceLoadbalancingPolicy}}
	ServiceMarketplace       = Service{map[string]Resource{"catalog": ResourceMarketplaceCatalog, "productType": ResourceMarketplaceProductType, "publicCatalog": ResourceMarketplacePublicCatalog}}
	ServiceMetadata          = Service{map[string]Resource{"tag": ResourceMetadataTag}}
	ServiceMonitoringlogging = Service{map[string]Resource{"apiLog": ResourceMonitoringloggingApiLog}}
	ServiceNetwork           = Service{map[string]Resource{"clientGateway": ResourceNetworkClientGateway, "directLink": ResourceNetworkDirectLink, "directLinkInterface": ResourceNetworkDirectLinkInterface, "link": ResourceNetworkLink, "natGateway": ResourceNetworkNatGateway, "nic": ResourceNetworkNic, "publicIp": ResourceNetworkPublicIp, "subnet": ResourceNetworkSubnet, "virtualGateway": ResourceNetworkVirtualGateway, "vpc": ResourceNetworkVpc, "vpcAccessPoint": ResourceNetworkVpcAccessPoint, "vpcPeering": ResourceNetworkVpcPeering, "vpnConnection": ResourceNetworkVpnConnection, "vpnConnectionRoute": ResourceNetworkVpnConnectionRoute}}
	ServiceOpenshift         = Service{map[string]Resource{"cluster": ResourceOpenshiftCluster}}
	ServicePostgres          = Service{map[string]Resource{"cluster": ResourcePostgresCluster}}
	ServiceRegion            = Service{map[string]Resource{"location": ResourceRegionLocation, "region": ResourceRegionRegion, "subRegion": ResourceRegionSubRegion}}
	ServiceRouting           = Service{map[string]Resource{"link": ResourceRoutingLink, "route": ResourceRoutingRoute, "routeTable": ResourceRoutingRouteTable}}
	ServiceSecurity          = Service{map[string]Resource{"certAuthority": ResourceSecurityCertAuthority, "keyPair": ResourceSecurityKeyPair, "securityGroup": ResourceSecuritySecurityGroup, "securityGroupRule": ResourceSecuritySecurityGroupRule, "serverCert": ResourceSecurityServerCert}}
	ServiceServiceusage      = Service{map[string]Resource{"quota": ResourceServiceusageQuota}}
	ServiceStorageblock      = Service{map[string]Resource{"link": ResourceStorageblockLink, "snapshot": ResourceStorageblockSnapshot, "volume": ResourceStorageblockVolume}}
	ServiceStorageobject     = Service{map[string]Resource{"bucket": ResourceStorageobjectBucket}}
)

// List of all [Resource] in the OPL.
var (
	ResourceComputeFlexibleGpu           = Resource{[]SubResource{}}
	ResourceComputeLink                  = Resource{[]SubResource{}}
	ResourceComputeVm                    = Resource{[]SubResource{}}
	ResourceComputeVmGroup               = Resource{[]SubResource{}}
	ResourceComputeVmTemplate            = Resource{[]SubResource{}}
	ResourceDatamanagementMachineImage   = Resource{[]SubResource{"exportTask"}}
	ResourceDhcpDhcpOption               = Resource{[]SubResource{}}
	ResourceIamPermission                = Resource{[]SubResource{}}
	ResourceIamRole                      = Resource{[]SubResource{}}
	ResourceIamTenant                    = Resource{[]SubResource{}}
	ResourceIdentityGroup                = Resource{[]SubResource{}}
	ResourceIdentityServiceAccount       = Resource{[]SubResource{}}
	ResourceIdentityUser                 = Resource{[]SubResource{}}
	ResourceInternalRoot                 = Resource{[]SubResource{}}
	ResourceInternalServiceAccountNS     = Resource{[]SubResource{}}
	ResourceKubernetesCluster            = Resource{[]SubResource{}}
	ResourceLoadbalancingInternetGateway = Resource{[]SubResource{}}
	ResourceLoadbalancingLink            = Resource{[]SubResource{}}
	ResourceLoadbalancingListener        = Resource{[]SubResource{}}
	ResourceLoadbalancingListenerRule    = Resource{[]SubResource{}}
	ResourceLoadbalancingLoadBalancer    = Resource{[]SubResource{}}
	ResourceLoadbalancingPolicy          = Resource{[]SubResource{}}
	ResourceMarketplaceCatalog           = Resource{[]SubResource{}}
	ResourceMarketplaceProductType       = Resource{[]SubResource{}}
	ResourceMarketplacePublicCatalog     = Resource{[]SubResource{}}
	ResourceMetadataTag                  = Resource{[]SubResource{}}
	ResourceMonitoringloggingApiLog      = Resource{[]SubResource{}}
	ResourceNetworkClientGateway         = Resource{[]SubResource{}}
	ResourceNetworkDirectLink            = Resource{[]SubResource{}}
	ResourceNetworkDirectLinkInterface   = Resource{[]SubResource{}}
	ResourceNetworkLink                  = Resource{[]SubResource{}}
	ResourceNetworkNatGateway            = Resource{[]SubResource{}}
	ResourceNetworkNic                   = Resource{[]SubResource{}}
	ResourceNetworkPublicIp              = Resource{[]SubResource{}}
	ResourceNetworkSubnet                = Resource{[]SubResource{}}
	ResourceNetworkVirtualGateway        = Resource{[]SubResource{}}
	ResourceNetworkVpc                   = Resource{[]SubResource{}}
	ResourceNetworkVpcAccessPoint        = Resource{[]SubResource{}}
	ResourceNetworkVpcPeering            = Resource{[]SubResource{}}
	ResourceNetworkVpnConnection         = Resource{[]SubResource{}}
	ResourceNetworkVpnConnectionRoute    = Resource{[]SubResource{}}
	ResourceOpenshiftCluster             = Resource{[]SubResource{}}
	ResourcePostgresCluster              = Resource{[]SubResource{}}
	ResourceRegionLocation               = Resource{[]SubResource{}}
	ResourceRegionRegion                 = Resource{[]SubResource{}}
	ResourceRegionSubRegion              = Resource{[]SubResource{}}
	ResourceRoutingLink                  = Resource{[]SubResource{}}
	ResourceRoutingRoute                 = Resource{[]SubResource{}}
	ResourceRoutingRouteTable            = Resource{[]SubResource{}}
	ResourceSecurityCertAuthority        = Resource{[]SubResource{}}
	ResourceSecurityKeyPair              = Resource{[]SubResource{}}
	ResourceSecuritySecurityGroup        = Resource{[]SubResource{}}
	ResourceSecuritySecurityGroupRule    = Resource{[]SubResource{}}
	ResourceSecurityServerCert           = Resource{[]SubResource{}}
	ResourceServiceusageQuota            = Resource{[]SubResource{}}
	ResourceStorageblockLink             = Resource{[]SubResource{}}
	ResourceStorageblockSnapshot         = Resource{[]SubResource{"exportTask"}}
	ResourceStorageblockVolume           = Resource{[]SubResource{}}
	ResourceStorageobjectBucket          = Resource{[]SubResource{"object"}}
)

// Service in the OPL file.
type Service struct {
	resources map[string]Resource
}

// Resources list of a service.
func (s Service) Resources() []Resource {
	res := make([]Resource, len(s.resources))
	for _, resource := range s.resources {
		res = append(res, resource)
	}

	return res
}

// Resource of a [Service] in the OPL file.
type Resource struct {
	subResources []SubResource
}

// SubResources list of a service.
func (s Resource) SubResources() []SubResource {
	return s.subResources
}

// SubResource of a [Resource] in the OPL file.
type SubResource string

// ErrNoKetoNamespace is returned when no namespace matches given namespace.
var ErrNoKetoNamespace = errors.New("namespace doesn't exist")

// Namespace as represented in keto.
type Namespace string

func firstToLower(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}
	lc := unicode.ToLower(r)
	if r == lc {
		return s
	}
	return string(lc) + s[size:]
}

// PermissionString from Service, Resource and Subresource
func (ns Namespace) PermissionString() string {
	b := new(strings.Builder)
	components := strings.Split(string(ns), "_")
	b.WriteString(strings.ToLower(components[0]))
	if len(components) >= 2 {
		b.WriteRune('.')
		b.WriteString(firstToLower(components[1]))
	}
	if len(components) == 3 {
		b.WriteRune('.')
		b.WriteString(firstToLower(components[2]))
	}

	return b.String()
}

// Components as a service.[resource].[subresource] tuple.
func (ns Namespace) Components() (service string, resource, subresource *string) {
	components := strings.Split(string(ns), "_")
	service = strings.ToLower(components[0])
	if len(components) >= 2 {
		res := firstToLower(components[1])
		resource = &res
	}
	if len(components) == 3 {
		subRes := firstToLower(components[2])
		subresource = &subRes
	}

	return
}

// String representation.
func (ns Namespace) String() string {
	return string(ns)
}

// KetoNamespace contains info about relations and actions.
type KetoNamespace struct {
	relations []PermissionRelation
	actions   []PermissionAction
}

// Relations in the OPL file.
func (ns KetoNamespace) Relations() []PermissionRelation {
	return ns.relations
}

// Actions in the OPL file.
func (ns KetoNamespace) Actions() []PermissionAction {
	return ns.actions
}

// List of all Namespaces in the OPL.
const (
	NamespaceCompute_FlexibleGpu                    Namespace = "Compute_FlexibleGpu"
	NamespaceCompute_Link                           Namespace = "Compute_Link"
	NamespaceCompute_Vm                             Namespace = "Compute_Vm"
	NamespaceCompute_VmGroup                        Namespace = "Compute_VmGroup"
	NamespaceCompute_VmTemplate                     Namespace = "Compute_VmTemplate"
	NamespaceDataManagement_MachineImage            Namespace = "DataManagement_MachineImage"
	NamespaceDataManagement_MachineImage_ExportTask Namespace = "DataManagement_MachineImage_ExportTask"
	NamespaceDhcp_DhcpOption                        Namespace = "Dhcp_DhcpOption"
	NamespaceIAM_Permission                         Namespace = "IAM_Permission"
	NamespaceIAM_Role                               Namespace = "IAM_Role"
	NamespaceIAM_Tenant                             Namespace = "IAM_Tenant"
	NamespaceIdentity_Group                         Namespace = "Identity_Group"
	NamespaceIdentity_ServiceAccount                Namespace = "Identity_ServiceAccount"
	NamespaceIdentity_User                          Namespace = "Identity_User"
	NamespaceInternal_Root                          Namespace = "Internal_Root"
	NamespaceInternal_ServiceAccountNS              Namespace = "Internal_ServiceAccountNS"
	NamespaceKubernetes_Cluster                     Namespace = "Kubernetes_Cluster"
	NamespaceLoadBalancing_InternetGateway          Namespace = "LoadBalancing_InternetGateway"
	NamespaceLoadBalancing_Link                     Namespace = "LoadBalancing_Link"
	NamespaceLoadBalancing_Listener                 Namespace = "LoadBalancing_Listener"
	NamespaceLoadBalancing_ListenerRule             Namespace = "LoadBalancing_ListenerRule"
	NamespaceLoadBalancing_LoadBalancer             Namespace = "LoadBalancing_LoadBalancer"
	NamespaceLoadBalancing_Policy                   Namespace = "LoadBalancing_Policy"
	NamespaceMarketplace_Catalog                    Namespace = "Marketplace_Catalog"
	NamespaceMarketplace_ProductType                Namespace = "Marketplace_ProductType"
	NamespaceMarketplace_PublicCatalog              Namespace = "Marketplace_PublicCatalog"
	NamespaceMetadata_Tag                           Namespace = "Metadata_Tag"
	NamespaceMonitoringLogging_ApiLog               Namespace = "MonitoringLogging_ApiLog"
	NamespaceNetwork_ClientGateway                  Namespace = "Network_ClientGateway"
	NamespaceNetwork_DirectLink                     Namespace = "Network_DirectLink"
	NamespaceNetwork_DirectLinkInterface            Namespace = "Network_DirectLinkInterface"
	NamespaceNetwork_Link                           Namespace = "Network_Link"
	NamespaceNetwork_NatGateway                     Namespace = "Network_NatGateway"
	NamespaceNetwork_Nic                            Namespace = "Network_Nic"
	NamespaceNetwork_PublicIp                       Namespace = "Network_PublicIp"
	NamespaceNetwork_Subnet                         Namespace = "Network_Subnet"
	NamespaceNetwork_VirtualGateway                 Namespace = "Network_VirtualGateway"
	NamespaceNetwork_Vpc                            Namespace = "Network_Vpc"
	NamespaceNetwork_VpcAccessPoint                 Namespace = "Network_VpcAccessPoint"
	NamespaceNetwork_VpcPeering                     Namespace = "Network_VpcPeering"
	NamespaceNetwork_VpnConnection                  Namespace = "Network_VpnConnection"
	NamespaceNetwork_VpnConnectionRoute             Namespace = "Network_VpnConnectionRoute"
	NamespaceOpenshift_Cluster                      Namespace = "Openshift_Cluster"
	NamespacePostgres_Cluster                       Namespace = "Postgres_Cluster"
	NamespaceRegion_Location                        Namespace = "Region_Location"
	NamespaceRegion_Region                          Namespace = "Region_Region"
	NamespaceRegion_SubRegion                       Namespace = "Region_SubRegion"
	NamespaceRouting_Link                           Namespace = "Routing_Link"
	NamespaceRouting_Route                          Namespace = "Routing_Route"
	NamespaceRouting_RouteTable                     Namespace = "Routing_RouteTable"
	NamespaceSecurity_CertAuthority                 Namespace = "Security_CertAuthority"
	NamespaceSecurity_KeyPair                       Namespace = "Security_KeyPair"
	NamespaceSecurity_SecurityGroup                 Namespace = "Security_SecurityGroup"
	NamespaceSecurity_SecurityGroupRule             Namespace = "Security_SecurityGroupRule"
	NamespaceSecurity_ServerCert                    Namespace = "Security_ServerCert"
	NamespaceServiceUsage_Quota                     Namespace = "ServiceUsage_Quota"
	NamespaceStorageBlock_Link                      Namespace = "StorageBlock_Link"
	NamespaceStorageBlock_Snapshot                  Namespace = "StorageBlock_Snapshot"
	NamespaceStorageBlock_Snapshot_ExportTask       Namespace = "StorageBlock_Snapshot_ExportTask"
	NamespaceStorageBlock_Volume                    Namespace = "StorageBlock_Volume"
	NamespaceStorageObject_Bucket                   Namespace = "StorageObject_Bucket"
	NamespaceStorageObject_Bucket_Object            Namespace = "StorageObject_Bucket_Object"
)

// Slice of all Namespaces in the OPL.
var Namespaces = []Namespace{
	NamespaceCompute_FlexibleGpu,
	NamespaceCompute_Link,
	NamespaceCompute_Vm,
	NamespaceCompute_VmGroup,
	NamespaceCompute_VmTemplate,
	NamespaceDataManagement_MachineImage,
	NamespaceDataManagement_MachineImage_ExportTask,
	NamespaceDhcp_DhcpOption,
	NamespaceIAM_Permission,
	NamespaceIAM_Role,
	NamespaceIAM_Tenant,
	NamespaceIdentity_Group,
	NamespaceIdentity_ServiceAccount,
	NamespaceIdentity_User,
	NamespaceInternal_Root,
	NamespaceInternal_ServiceAccountNS,
	NamespaceKubernetes_Cluster,
	NamespaceLoadBalancing_InternetGateway,
	NamespaceLoadBalancing_Link,
	NamespaceLoadBalancing_Listener,
	NamespaceLoadBalancing_ListenerRule,
	NamespaceLoadBalancing_LoadBalancer,
	NamespaceLoadBalancing_Policy,
	NamespaceMarketplace_Catalog,
	NamespaceMarketplace_ProductType,
	NamespaceMarketplace_PublicCatalog,
	NamespaceMetadata_Tag,
	NamespaceMonitoringLogging_ApiLog,
	NamespaceNetwork_ClientGateway,
	NamespaceNetwork_DirectLink,
	NamespaceNetwork_DirectLinkInterface,
	NamespaceNetwork_Link,
	NamespaceNetwork_NatGateway,
	NamespaceNetwork_Nic,
	NamespaceNetwork_PublicIp,
	NamespaceNetwork_Subnet,
	NamespaceNetwork_VirtualGateway,
	NamespaceNetwork_Vpc,
	NamespaceNetwork_VpcAccessPoint,
	NamespaceNetwork_VpcPeering,
	NamespaceNetwork_VpnConnection,
	NamespaceNetwork_VpnConnectionRoute,
	NamespaceOpenshift_Cluster,
	NamespacePostgres_Cluster,
	NamespaceRegion_Location,
	NamespaceRegion_Region,
	NamespaceRegion_SubRegion,
	NamespaceRouting_Link,
	NamespaceRouting_Route,
	NamespaceRouting_RouteTable,
	NamespaceSecurity_CertAuthority,
	NamespaceSecurity_KeyPair,
	NamespaceSecurity_SecurityGroup,
	NamespaceSecurity_SecurityGroupRule,
	NamespaceSecurity_ServerCert,
	NamespaceServiceUsage_Quota,
	NamespaceStorageBlock_Link,
	NamespaceStorageBlock_Snapshot,
	NamespaceStorageBlock_Snapshot_ExportTask,
	NamespaceStorageBlock_Volume,
	NamespaceStorageObject_Bucket,
	NamespaceStorageObject_Bucket_Object,
}

// GetNamespace from service, subresource and resource.
func GetNamespace(service string, resource, subResource *string) (Namespace, error) {
	ns := service
	if resource != nil {
		ns = strings.Join([]string{ns, *resource}, ".")
	}
	if subResource != nil {
		ns = strings.Join([]string{ns, *subResource}, ".")
	}
	res, ok := permToNamespace[ns]
	if !ok {
		return "", &BadNamespaceError{
			service:     service,
			resource:    resource,
			subresource: subResource,
		}
	}

	return res, nil
}

var permToNamespace = map[string]Namespace{
	"compute.flexibleGpu":                    NamespaceCompute_FlexibleGpu,
	"compute.link":                           NamespaceCompute_Link,
	"compute.vm":                             NamespaceCompute_Vm,
	"compute.vmGroup":                        NamespaceCompute_VmGroup,
	"compute.vmTemplate":                     NamespaceCompute_VmTemplate,
	"datamanagement.machineImage":            NamespaceDataManagement_MachineImage,
	"datamanagement.machineImage.exportTask": NamespaceDataManagement_MachineImage_ExportTask,
	"dhcp.dhcpOption":                        NamespaceDhcp_DhcpOption,
	"iam.permission":                         NamespaceIAM_Permission,
	"iam.role":                               NamespaceIAM_Role,
	"iam.tenant":                             NamespaceIAM_Tenant,
	"identity.group":                         NamespaceIdentity_Group,
	"identity.serviceAccount":                NamespaceIdentity_ServiceAccount,
	"identity.user":                          NamespaceIdentity_User,
	"internal.root":                          NamespaceInternal_Root,
	"internal.serviceAccountNS":              NamespaceInternal_ServiceAccountNS,
	"kubernetes.cluster":                     NamespaceKubernetes_Cluster,
	"loadbalancing.internetGateway":          NamespaceLoadBalancing_InternetGateway,
	"loadbalancing.link":                     NamespaceLoadBalancing_Link,
	"loadbalancing.listener":                 NamespaceLoadBalancing_Listener,
	"loadbalancing.listenerRule":             NamespaceLoadBalancing_ListenerRule,
	"loadbalancing.loadBalancer":             NamespaceLoadBalancing_LoadBalancer,
	"loadbalancing.policy":                   NamespaceLoadBalancing_Policy,
	"marketplace.catalog":                    NamespaceMarketplace_Catalog,
	"marketplace.productType":                NamespaceMarketplace_ProductType,
	"marketplace.publicCatalog":              NamespaceMarketplace_PublicCatalog,
	"metadata.tag":                           NamespaceMetadata_Tag,
	"monitoringlogging.apiLog":               NamespaceMonitoringLogging_ApiLog,
	"network.clientGateway":                  NamespaceNetwork_ClientGateway,
	"network.directLink":                     NamespaceNetwork_DirectLink,
	"network.directLinkInterface":            NamespaceNetwork_DirectLinkInterface,
	"network.link":                           NamespaceNetwork_Link,
	"network.natGateway":                     NamespaceNetwork_NatGateway,
	"network.nic":                            NamespaceNetwork_Nic,
	"network.publicIp":                       NamespaceNetwork_PublicIp,
	"network.subnet":                         NamespaceNetwork_Subnet,
	"network.virtualGateway":                 NamespaceNetwork_VirtualGateway,
	"network.vpc":                            NamespaceNetwork_Vpc,
	"network.vpcAccessPoint":                 NamespaceNetwork_VpcAccessPoint,
	"network.vpcPeering":                     NamespaceNetwork_VpcPeering,
	"network.vpnConnection":                  NamespaceNetwork_VpnConnection,
	"network.vpnConnectionRoute":             NamespaceNetwork_VpnConnectionRoute,
	"openshift.cluster":                      NamespaceOpenshift_Cluster,
	"postgres.cluster":                       NamespacePostgres_Cluster,
	"region.location":                        NamespaceRegion_Location,
	"region.region":                          NamespaceRegion_Region,
	"region.subRegion":                       NamespaceRegion_SubRegion,
	"routing.link":                           NamespaceRouting_Link,
	"routing.route":                          NamespaceRouting_Route,
	"routing.routeTable":                     NamespaceRouting_RouteTable,
	"security.certAuthority":                 NamespaceSecurity_CertAuthority,
	"security.keyPair":                       NamespaceSecurity_KeyPair,
	"security.securityGroup":                 NamespaceSecurity_SecurityGroup,
	"security.securityGroupRule":             NamespaceSecurity_SecurityGroupRule,
	"security.serverCert":                    NamespaceSecurity_ServerCert,
	"serviceusage.quota":                     NamespaceServiceUsage_Quota,
	"storageblock.link":                      NamespaceStorageBlock_Link,
	"storageblock.snapshot":                  NamespaceStorageBlock_Snapshot,
	"storageblock.snapshot.exportTask":       NamespaceStorageBlock_Snapshot_ExportTask,
	"storageblock.volume":                    NamespaceStorageBlock_Volume,
	"storageobject.bucket":                   NamespaceStorageObject_Bucket,
	"storageobject.bucket.object":            NamespaceStorageObject_Bucket_Object,
}

// KetoNamespace from Namespace.
func (ns Namespace) KetoNamespace() (KetoNamespace, error) {
	res, ok := namespaceToKetoNamespace[ns]
	if !ok {
		return KetoNamespace{}, ErrNoKetoNamespace
	}

	return res, nil
}

var namespaceToKetoNamespace = map[Namespace]KetoNamespace{
	NamespaceCompute_FlexibleGpu:                    {relations: []PermissionRelation{"owners", "parents", "getters", "updaters", "deleters", "linkers", "unlinkers", "gettersCatalog", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "link", "unlink", "getCatalog", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceCompute_Link:                           {relations: []PermissionRelation{"owners", "getters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceCompute_Vm:                             {relations: []PermissionRelation{"owners", "parents", "getters", "updaters", "deleters", "starters", "rebooters", "stoppers", "gettersAdminPassword", "gettersConsoleOutput", "gettersVmTypes", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "start", "reboot", "stop", "getAdminPassword", "getConsoleOutput", "getVmTypes", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceCompute_VmGroup:                        {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "upscalers", "downscalers", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "upscale", "downscale", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceCompute_VmTemplate:                     {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceDataManagement_MachineImage:            {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceDataManagement_MachineImage_ExportTask: {relations: []PermissionRelation{"owners", "parents", "getters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceDhcp_DhcpOption:                        {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceIAM_Permission:                         {relations: []PermissionRelation{"creatorsPolicy", "gettersPolicy", "deletersPolicy", "creators", "creatorsURL", "getters", "updaters", "deleters", "setters", "unsetters", "importers", "upscalers", "downscalers", "gettersCatalog", "acceptors", "rejectors", "linkers", "unlinkers", "disablers", "enablers", "recoverers", "impersonators", "users", "starters", "stoppers", "rebooters", "gettersAdminPassword", "gettersConsoleOutput", "gettersVmTypes", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "create", "createPolicy", "getPolicy", "deletePolicy", "createURL", "get", "update", "delete", "set", "unset", "import", "upscale", "downscale", "getCatalog", "accept", "reject", "link", "unlink", "disable", "enable", "recover", "impersonate", "use", "start", "stop", "reboot", "getAdminPassword", "getConsoleOutput", "getVmTypes", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceIAM_Role:                               {relations: []PermissionRelation{"owners", "assignees", "getters", "updaters", "deleters", "disablers", "enablers", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "disable", "enable", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceIAM_Tenant:                             {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "disablers", "enablers", "settersIAMPolicy", "gettersIAMPolicy", "parents", "parentsRoot"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "disable", "enable", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceIdentity_Group:                         {relations: []PermissionRelation{}, actions: []PermissionAction{}},
	NamespaceIdentity_ServiceAccount:                {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "disablers", "enablers", "impersonators", "settersIAMPolicy", "gettersIAMPolicy", "members"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "disable", "enable", "impersonate", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceIdentity_User:                          {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "disablers", "enablers", "recoverers", "impersonators", "settersIAMPolicy", "gettersIAMPolicy", "members"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "disable", "enable", "recover", "impersonate", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceInternal_Root:                          {relations: []PermissionRelation{"san_creator", "san_policy_viewer", "san_policy_creator", "user_creator", "\"identity.san.create\"", "\"identity.san.getIAMPolicy\"", "\"identity.san.setIAMPolicy\"", "\"identity.openId.create\"", "\"identity.user.create\"", "\"identity.user.get\"", "\"identity.user.update\"", "\"identity.user.delete\"", "\"identity.user.disable\"", "\"identity.user.enable\"", "\"identity.user.recover\"", "\"identity.user.impersonate\"", "\"identity.user.getIAMPolicy\"", "\"identity.user.setIAMPolicy\"", "\"identity.serviceAccount.create\"", "\"identity.serviceAccount.get\"", "\"identity.serviceAccount.update\"", "\"identity.serviceAccount.delete\"", "\"identity.serviceAccount.disable\"", "\"identity.serviceAccount.enable\"", "\"identity.serviceAccount.impersonate\"", "\"identity.serviceAccount.getIAMPolicy\"", "\"identity.serviceAccount.setIAMPolicy\"", "\"iam.organization.create\"", "\"iam.organization.delete\"", "\"iam.organization.disable\"", "\"iam.organization.enable\"", "\"iam.organization.getIAMPolicy\"", "\"iam.organization.setIAMPolicy\"", "\"iam.permission.get\"", "\"iam.role.get\"", "\"iam.organisation.create\"", "\"iam.organisation.delete\"", "\"iam.organisation.disable\"", "\"iam.organisation.enable\"", "\"iam.organisation.getIAMPolicy\"", "\"iam.organisation.setIAMPolicy\"", "\"iam.admin.role\"", "\"iam.admin.permission\"", "\"iam.admin.replicate\""}, actions: []PermissionAction{}},
	NamespaceInternal_ServiceAccountNS:              {relations: []PermissionRelation{}, actions: []PermissionAction{}},
	NamespaceKubernetes_Cluster:                     {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceLoadBalancing_InternetGateway:          {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "linkers", "unlinkers", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "link", "unlink", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceLoadBalancing_Link:                     {relations: []PermissionRelation{"owners", "getters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceLoadBalancing_Listener:                 {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceLoadBalancing_ListenerRule:             {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceLoadBalancing_LoadBalancer:             {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "linkers", "unlinkers", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "link", "unlink", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceLoadBalancing_Policy:                   {relations: []PermissionRelation{"owners", "getters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceMarketplace_Catalog:                    {relations: []PermissionRelation{"getters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceMarketplace_ProductType:                {relations: []PermissionRelation{"getters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceMarketplace_PublicCatalog:              {relations: []PermissionRelation{"getters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceMetadata_Tag:                           {relations: []PermissionRelation{"owners", "getters", "setters", "unsetters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "set", "unset", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceMonitoringLogging_ApiLog:               {relations: []PermissionRelation{"getters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceNetwork_ClientGateway:                  {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceNetwork_DirectLink:                     {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceNetwork_DirectLinkInterface:            {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceNetwork_Link:                           {relations: []PermissionRelation{"owners", "getters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceNetwork_NatGateway:                     {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceNetwork_Nic:                            {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "linkers", "unlinkers", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "link", "unlink", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceNetwork_PublicIp:                       {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "linkers", "unlinkers", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "link", "unlink", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceNetwork_Subnet:                         {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceNetwork_VirtualGateway:                 {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "linkers", "unlinkers", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "link", "unlink", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceNetwork_Vpc:                            {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceNetwork_VpcAccessPoint:                 {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceNetwork_VpcPeering:                     {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "acceptors", "rejectors", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "accept", "reject", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceNetwork_VpnConnection:                  {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceNetwork_VpnConnectionRoute:             {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceOpenshift_Cluster:                      {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespacePostgres_Cluster:                       {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy", "resetersPassword"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy", "resetPassword"}},
	NamespaceRegion_Location:                        {relations: []PermissionRelation{"getters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceRegion_Region:                          {relations: []PermissionRelation{"getters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceRegion_SubRegion:                       {relations: []PermissionRelation{"getters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceRouting_Link:                           {relations: []PermissionRelation{"owners", "getters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceRouting_Route:                          {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceRouting_RouteTable:                     {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "linkers", "unlinkers", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "link", "unlink", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceSecurity_CertAuthority:                 {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceSecurity_KeyPair:                       {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceSecurity_SecurityGroup:                 {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceSecurity_SecurityGroupRule:             {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceSecurity_ServerCert:                    {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceServiceUsage_Quota:                     {relations: []PermissionRelation{"getters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceStorageBlock_Link:                      {relations: []PermissionRelation{"owners", "getters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceStorageBlock_Snapshot:                  {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceStorageBlock_Snapshot_ExportTask:       {relations: []PermissionRelation{"owners", "parents", "getters", "deleters", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "delete", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceStorageBlock_Volume:                    {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "linkers", "unlinkers", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "link", "unlink", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceStorageObject_Bucket:                   {relations: []PermissionRelation{"owners", "getters", "updaters", "deleters", "disablers", "enablers", "settersIAMPolicy", "gettersIAMPolicy"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "disable", "enable", "setIAMPolicy", "getIAMPolicy"}},
	NamespaceStorageObject_Bucket_Object:            {relations: []PermissionRelation{"owners", "parents", "getters", "updaters", "deleters", "settersIAMPolicy", "gettersIAMPolicy", "creatorsURL"}, actions: []PermissionAction{"belongs", "get", "update", "delete", "setIAMPolicy", "getIAMPolicy", "createURL"}},
}
