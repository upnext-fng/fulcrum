package jwt

type Claims struct {
	UserID string                 `json:"user_id"`
	Custom map[string]interface{} `json:"custom,omitempty"`
}