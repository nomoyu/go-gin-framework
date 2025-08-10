package response

// StringResponse 示例结构体用于 Swagger 展示
type StringResponse struct {
	Code int    `json:"code" example:"200"`
	Msg  string `json:"msg" example:"OK"`
	Data string `json:"data" example:"Hello User!"`
}
