package v1

import (
	"net"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resourceName=network-attachment-definitions

type NetworkAttachmentDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkAttachmentDefinitionSpec   `json:"spec"`
	Status NetworkAttachmentDefinitionStatus `json:"status,omitempty"`
}

// StateType contains a valid NetworkAttachmentDefinition state
type StateType string

const (
	// PendingState indicates NetworkAttachmentDefinition waiting to reconcile
	PendingState StateType = "Pending"
	// PendingObservationMessage is the default message for a
	// NetworkAttachmentDefinition pending to be reconciled
	PendingObservationMessage = "NetworkAttachmentDefinition waiting to be reconciled"
	// SuccessState indicates NetworkAttachmentDefinition successfully reconcile
	SuccessState StateType = "Success"
	// FailureState indicates NetworkAttachmentDefinition has failed to reconcile
	// for one or more reasons
	FailureState StateType = "Failure"
)

type NetworkAttachmentDefinitionSpec struct {
	Config string `json:"config"`
}

// NetworkAttachmentDefinitionStatus contains information for Status
type NetworkAttachmentDefinitionStatus struct {
	// ReconcilerState fields
	ReconcilerState `json:",inline"`
}

// ReconcilerState permits to know if NetworkAttachmentDefinition was reconciled
type ReconcilerState struct {
	// State is the current state of the reconciliation. The state is updated
	// during the process. See: State's type.
	State StateType `json:"state"`

	// Observation provides relative information related to the signature and
	// state, example if an error happened.
	Observation string `json:"observation"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NetworkAttachmentDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []NetworkAttachmentDefinition `json:"items"`
}

// DNS contains values interesting for DNS resolvers
// +k8s:deepcopy-gen=false
type DNS struct {
	Nameservers []string `json:"nameservers,omitempty"`
	Domain      string   `json:"domain,omitempty"`
	Search      []string `json:"search,omitempty"`
	Options     []string `json:"options,omitempty"`
}

const (
	DeviceInfoTypePCI       = "pci"
	DeviceInfoTypeVHostUser = "vhost-user"
	DeviceInfoTypeMemif     = "memif"
	DeviceInfoTypeVDPA      = "vdpa"
	DeviceInfoVersion       = "1.0.0"
)

// DeviceInfo contains the information of the device associated
// with this network (if any)
type DeviceInfo struct {
	Type      string       `json:"type,omitempty"`
	Version   string       `json:"version,omitempty"`
	Pci       *PciDevice   `json:"pci,omitempty"`
	Vdpa      *VdpaDevice  `json:"vdpa,omitempty"`
	VhostUser *VhostDevice `json:"vhost-user,omitempty"`
	Memif     *MemifDevice `json:"memif,omitempty"`
}

type PciDevice struct {
	PciAddress   string `json:"pci-address,omitempty"`
	Vhostnet     string `json:"vhost-net,omitempty"`
	RdmaDevice   string `json:"rdma-device,omitempty"`
	PfPciAddress string `json:"pf-pci-address,omitempty"`
}

type VdpaDevice struct {
	ParentDevice string `json:"parent-device,omitempty"`
	Driver       string `json:"driver,omitempty"`
	Path         string `json:"path,omitempty"`
	PciAddress   string `json:"pci-address,omitempty"`
	PfPciAddress string `json:"pf-pci-address,omitempty"`
}

const (
	VhostDeviceModeClient = "client"
	VhostDeviceModeServer = "server"
)

type VhostDevice struct {
	Mode string `json:"mode,omitempty"`
	Path string `json:"path,omitempty"`
}

const (
	MemifDeviceRoleMaster   = "master"
	MemitDeviceRoleSlave    = "slave"
	MemifDeviceModeEthernet = "ethernet"
	MemitDeviceModeIP       = "ip"
	MemitDeviceModePunt     = "punt"
)

type MemifDevice struct {
	Role string `json:"role,omitempty"`
	Path string `json:"path,omitempty"`
	Mode string `json:"mode,omitempty"`
}

// NetworkStatus is for network status annotation for pod
// +k8s:deepcopy-gen=false
type NetworkStatus struct {
	Name       string      `json:"name"`
	Interface  string      `json:"interface,omitempty"`
	IPs        []string    `json:"ips,omitempty"`
	Mac        string      `json:"mac,omitempty"`
	Default    bool        `json:"default,omitempty"`
	DNS        DNS         `json:"dns,omitempty"`
	DeviceInfo *DeviceInfo `json:"device-info,omitempty"`
}

// PortMapEntry for CNI PortMapEntry
// +k8s:deepcopy-gen=false
type PortMapEntry struct {
	HostPort      int    `json:"hostPort"`
	ContainerPort int    `json:"containerPort"`
	Protocol      string `json:"protocol,omitempty"`
	HostIP        string `json:"hostIP,omitempty"`
}

// BandwidthEntry for CNI BandwidthEntry
// +k8s:deepcopy-gen=false
type BandwidthEntry struct {
	IngressRate  int `json:"ingressRate"`
	IngressBurst int `json:"ingressBurst"`

	EgressRate  int `json:"egressRate"`
	EgressBurst int `json:"egressBurst"`
}

// NetworkSelectionElement represents one element of the JSON format
// Network Attachment Selection Annotation as described in section 4.1.2
// of the CRD specification.
// +k8s:deepcopy-gen=false
type NetworkSelectionElement struct {
	// Name contains the name of the Network object this element selects
	Name string `json:"name"`
	// Namespace contains the optional namespace that the network referenced
	// by Name exists in
	Namespace string `json:"namespace,omitempty"`
	// IPRequest contains an optional requested IP addresses for this network
	// attachment
	IPRequest []string `json:"ips,omitempty"`
	// MacRequest contains an optional requested MAC address for this
	// network attachment
	MacRequest string `json:"mac,omitempty"`
	// InfinibandGUIDRequest contains an optional requested Infiniband GUID
	// address for this network attachment
	InfinibandGUIDRequest string `json:"infiniband-guid,omitempty"`
	// InterfaceRequest contains an optional requested name for the
	// network interface this attachment will create in the container
	InterfaceRequest string `json:"interface,omitempty"`
	// PortMappingsRequest contains an optional requested port mapping
	// for the network
	PortMappingsRequest []*PortMapEntry `json:"portMappings,omitempty"`
	// BandwidthRequest contains an optional requested bandwidth for
	// the network
	BandwidthRequest *BandwidthEntry `json:"bandwidth,omitempty"`
	// CNIArgs contains additional CNI arguments for the network interface
	CNIArgs *map[string]interface{} `json:"cni-args"`
	// GatewayRequest contains default route IP address for the pod
	GatewayRequest []net.IP `json:"default-route,omitempty"`
}

const (
	// Pod annotation for network-attachment-definition
	NetworkAttachmentAnnot = "k8s.v1.cni.cncf.io/networks"
	// Pod annotation for network status
	NetworkStatusAnnot = "k8s.v1.cni.cncf.io/network-status"
	// Old Pod annotation for network status (which is used before but it will be obsolated)
	OldNetworkStatusAnnot = "k8s.v1.cni.cncf.io/networks-status"
)

// NoK8sNetworkError indicates error, no network in kubernetes
// +k8s:deepcopy-gen=false
type NoK8sNetworkError struct {
	Message string
}

func (e *NoK8sNetworkError) Error() string { return string(e.Message) }
