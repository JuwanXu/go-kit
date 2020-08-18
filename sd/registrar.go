package sd

// 服务注册/取消注册接口
type Registrar interface {
	Register()
	Deregister()
}
