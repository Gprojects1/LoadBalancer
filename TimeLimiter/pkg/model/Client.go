package model

type ClientConfig struct {
	ClientID      string  `json:"client_id"`
	Capacity      int     `json:"capacity"`
	RatePerSec    float64 `json:"rate_per_sec"`
	CurrentTokens float64 `json:"current_tokens"`
}
