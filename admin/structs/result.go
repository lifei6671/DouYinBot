package structs

type JsonResult[T any] struct {
	ErrCode int    `json:"errcode"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}
