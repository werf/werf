package config

type GiterminismConfig struct {
	Config config `json:"config"`
}

type config struct {
	AllowUncommitted bool `json:"allowUncommitted"`
}
