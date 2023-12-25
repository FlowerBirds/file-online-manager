package model

type PodInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Ready     string `json:"ready"`
	Restarts  int32  `json:"restarts"`
	Age       string `json:"age"`
	IP        string `json:"ip"`
	Node      string `json:"node"`
}

type Namespace struct {
	Name string `json:"name"`
}
