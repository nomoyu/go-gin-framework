package errorcode

type ErrorCode struct {
	Code int
	Msg  string
}

// 定义一个常见错误码集合
var (
	Success     = ErrorCode{Code: 0, Msg: "success"}
	ServerError = ErrorCode{Code: 500, Msg: "服务器内部错误"}

	// 通用参数类
	InvalidParams = ErrorCode{Code: 1000, Msg: "请求参数不合法"}
	Unauthorized  = ErrorCode{Code: 1001, Msg: "未授权或 token 无效"}
	Forbidden     = ErrorCode{Code: 1002, Msg: "无权限访问"}

	// 用户类错误码
	UserNotFound = ErrorCode{Code: 2001, Msg: "用户不存在"}
	UserExists   = ErrorCode{Code: 2002, Msg: "用户已存在"}
	LoginFailed  = ErrorCode{Code: 2003, Msg: "用户名或密码错误"}

	// 业务类错误码示例
	OrderNotFound = ErrorCode{Code: 3001, Msg: "订单不存在"}
)
