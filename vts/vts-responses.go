package vts

type VTSResponse struct {
	MessageType string         `json:"messageType"`
	Data        map[string]any `json:"data"`
}
