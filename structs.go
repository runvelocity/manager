package main

type PingResponse struct {
	Ok bool `json:"ok"`
}

type Function struct {
	UUID           string `json:"uuid"`
	Name           string `json:"name"`
	CodeLocation   string `json:"codeLocation"`
	RootFsLocation string `json:"rootFsLocation"`
	Status         Status `json:"status"`
	Handler        string `json:"handler"`
}

type FunctionsResponse struct {
	Functions []Function `json:"functions"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type Status string

const (
	PENDING Status = "Pending"
	CREATED Status = "Created"
	FAILED  Status = "Failed"
)

type UploadHandlerArgs struct {
	Key string `json:"key"`
}

type StepFunctionsRequest struct {
	UUID         string `json:"uuid"`
	CodeLocation string `json:"codeLocation"`
}

type InvokeRequest struct {
	VmId          string        `json:"vmId"`
	InvokePayload InvokePayload `json:"invokePayload"`
}
type InvokePayload struct {
	Handler string                 `json:"handler"`
	Args    map[string]interface{} `json:"args"`
}
