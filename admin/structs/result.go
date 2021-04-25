package structs

type JsonResult struct {
	ErrCode int         `json:"errcode"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
