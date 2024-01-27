package models

type PingResponse struct {
	Ok bool `json:"ok"`
}

type Function struct {
	UUID         string `json:"uuid"`
	Name         string `json:"name"`
	CodeLocation string `json:"codeLocation"`
	Handler      string `json:"handler"`
}

type FunctionsResponse struct {
	Functions []Function `json:"functions"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type UploadHandlerArgs struct {
	Key string `json:"key"`
}

type InvokeRequest struct {
	VmId          string        `json:"vmId"`
	InvokePayload InvokePayload `json:"invokePayload"`
}
type InvokePayload struct {
	Handler string                 `json:"handler"`
	Args    map[string]interface{} `json:"args"`
}
