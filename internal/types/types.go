package types

type PodListFiltered struct {
	Items []Pod `json:"items"`
}

type Pod struct {
	Metadata Metadata `json:"metadata"`
	Spec     Spec     `json:"spec"`
}

type Metadata struct {
	Name      string `json:"name"`
	Namespace string `json: "namespace"`
}

type Spec struct {
	Containers []Container `json:"containers"`
}

type Container struct {
	Name      string    `json:"name"`
	Resources Resources `json:"resources"`
}

type Resources struct {
	Requests Resource `json:"requests"`
	Limits   Resource `json:"limits"`
}

type Resource struct {
	Cpu    string `json:"cpu"`
	Memory string `json:"memory"`
}
