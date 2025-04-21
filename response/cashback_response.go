package response

type CashbackResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data,omitempty"`
}
