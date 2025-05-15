package api

type PostgresIncompatibleStatusDeleteProblem409ResponseType = PostgresProblem

type PostgresIncompatibleStatusProblem409ResponseType = PostgresProblem

type PostgresInvalidAuthenticationProblem403ResponseType = PostgresProblem

type PostgresResourceNotFoundProblem404ResponseType = PostgresProblem

type PostgresServiceMalfunctionProblem500ResponseType = PostgresProblem

// PostgresClusterCreationRequestWithVolume defines model for PostgresClusterCreationRequest.
// This type fix the union field of volume
type PostgresClusterCreationRequestWithVolume struct {
	AllowedIpRanges PostgresAllowedIpRanges `json:"allowedIpRanges"`

	// AutomaticBackup Whether automatic backup is enabled for this cluster.
	AutomaticBackup *bool `json:"automaticBackup"`

	// IsPublic Whether public exposition is enabled for this cluster.
	IsPublic *bool           `json:"isPublic,omitempty"`
	Name     StrictSlugMax63 `json:"name"`

	// NetCidr The CIDR of the network where the cluster will be created.
	//
	// **Warning**: The CIDR must be in the following three blocks:
	// - 10.*.0.0/16
	// - 172.(16-31).0.0/16
	// - 192.168.0.0/16
	// The mask mut not be greater than /24.
	NetCidr *PostgresClusterNetCIDR `json:"netCidr,omitempty"`

	// NodeConfiguration The configuration used to provision the cluster nodes.
	NodeConfiguration PostgresNodeConfiguration `json:"nodeConfiguration"`

	// SourceBackupId A backup unique identifier.
	SourceBackupId *PostgresBackupId `json:"sourceBackupId,omitempty"`

	// Tags Tag to identify resources
	Tags *PostgresTags `json:"tags,omitempty"`

	// User The name of the user with administration privileges on the cluster.
	User PostgresUser `json:"user"`

	// Volume The configuration for a data storage volume.
	Volume PostgresAllVolumes `json:"volume"`
}

// PostgresAllVolumes The configuration for a data storage volume.
type PostgresAllVolumes struct {
	// SizeGiB The size of the volume in GiB.
	SizeGiB int `json:"sizeGiB"`

	// Type The type of the volume.
	Type string `json:"type"`

	// Iops The number of IOPS to allocate to the volume.
	Iops *int `json:"iops,omitempty"`
}
