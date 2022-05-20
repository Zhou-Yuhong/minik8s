package object

const (
	PodPending   string = "Pending"
	PodRunning   string = "Running"
	PodSucceeded string = "Succeeded"
	PodFailed    string = "Failed"
	PodUnknown   string = "Unknown"
	PodDelete    string = "Delete"

	// SUCCESS http status code
	SUCCESS int = 200
	FAILED  int = 400

	// kind
	PodKind        string = "Pod"
	ReplicaSetKind string = "RS"

	// Node
	NodeShardFilePath string = "/home/sharedData"
)

type ObjectMeta struct {
	Name   string            `json:"name" yaml:"name"`
	Labels map[string]string `json:"labels" yaml:"labels"`
	UID    string            `json:"uid" yaml:"uid"`

	OwnerReferences []OwnerReference `json:"ownerReferences" yaml:"ownerReferences"`
	Ctime           string
}

// OwnerReference ownership for objects, e.g. replicaset and pods
type OwnerReference struct {
	Kind string `json:"kind" yaml:"kind"`
	Name string `json:"name" yaml:"name"`
	UID  string `json:"uid" yaml:"uid"`
	// If true, this reference points to the managing controller.
	Controller bool `json:"controller" yaml:"controller"`
}

/*******************ReplicaSet*************************/

// ReplicaSet ensures that a specified number of pod replicas are running at any given time.
type ReplicaSet struct {
	ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec       ReplicaSetSpec   `json:"spec" yaml:"spec"`
	Status     ReplicaSetStatus `json:"status" yaml:"status"`
}

type ReplicaSetSpec struct {
	Replicas int32       `json:"replicas" yaml:"replicas"`
	Template PodTemplate `json:"template" yaml:"template"`
}

// ReplicaSetStatus represents the current status of a ReplicaSet.
type ReplicaSetStatus struct {
	Replicas int32 `json:"replicas" yaml:"replicas"`
}

type LabelSelector struct {
}

/*******************Pod*************************/

type Pod struct {
	ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec       PodSpec   `json:"spec" yaml:"spec"`
	Status     PodStatus `json:"status" yaml:"status"`
}

type PodSpec struct {
	Volumes    []Volume    `json:"volumes" yaml:"volumes"`
	Containers []Container `json:"containers" yaml:"containers"`
	NodeName   string      `json:"nodeName" yaml:"nodeName"`
}

type PodStatus struct {
	Phase string `json:"phase" yaml:"phase"`
	// IP address when the pod is assigned
	HostIP string `json:"hostIP" yaml:"hostIP"`
	// IP address allocated to the pod. Routable at least within the cluster
	PodIP string `json:"podIP" yaml:"podIP"`
	//error message
	Err string
}

type PodTemplate struct {
	ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec       PodSpec
}

// PodList is a list of Pods.
type PodList struct {
}

type ListOptions struct {
}

type Volume struct {
	Name string `json:"name" yaml:"name"`
	Type string `json:"type" yaml:"type"`
	Path string `json:"path" yaml:"path"`
}

type Container struct {
	Name         string        `json:"name" yaml:"name"`
	Image        string        `json:"image" yaml:"image"`
	Command      []string      `json:"command" yaml:"command"`
	Args         []string      `json:"args" yaml:"args"`
	VolumeMounts []VolumeMount `json:"volumeMounts" yaml:"volumeMounts"`
	Limits       Limit         `json:"limits" yaml:"limits"`
	Ports        []Port        `json:"ports" yaml:"ports"`
	Env          []EnvEntry    `json:"env" yaml:"env"`
}

type VolumeMount struct {
	Name      string `json:"name" yaml:"name"`
	MountPath string `json:"mountPath" yaml:"mountPath"`
}

type ContainerMeta struct {
	OriginName  string
	RealName    string
	ContainerId string
}

type Limit struct {
	Cpu    string `json:"cpu" yaml:"cpu"`
	Memory string `json:"memory" yaml:"memory"`
}

type Port struct {
	ContainerPort string `json:"containerPort" yaml:"containerPort"`
}

type EnvEntry struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

/*******************Node*************************/

type Node struct {
	MetaData ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec     NodeSpec   `json:"spec" yaml:"spec"`
	Status   NodeStatus `json:"status" yaml:"status"`
}

type NodeList struct {
	Items []Node `json:"items" yaml:"items"`
}

type NodeSpec struct {
	//物理ip地址
	PhysicalIp string `json:"physicalIp" yaml:"physicalIp""`
	//为该节点分配的pod网段
	NodeIpAndMask string `json:"nodeIpAndMask" yaml:"nodeIpAndMask"`
}

type NodeStatus struct {
}
