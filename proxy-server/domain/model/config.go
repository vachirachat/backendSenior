package model_proxy

// ProxyConfig represent proxy
type ProxyConfig struct {
	EnablePlugin    bool   `json:"enable_plugin" bson:"enable_plugin"`
	EnablePluginEnc bool   `json:"enable_plugin_enc" bson:"enable_plugin_enc"`
	PluginPort      string `json:"plugin_port" bson:"plugin_port"`
	DockerID        string `json:"dockerID" bson:"dockerID"`
}

type JSONDocker struct {
	DockerID string `json:"dockerid"`
	Server   string `json:"server"`
	Status   string `json:"status"`
	IP       string `json:"ip"`
}
type JSONCODE struct {
	Code     string `json:"config"`
	Lang     string `json:"lang"`
	Filename string `json:"filename"`
}
