package sd

// 若实例发生了变更，则封装成Event推送给Cache
type Event struct {
	Instances []string
	Err       error
}

type Instancer interface {
	Register(chan<- Event)
	Deregister(chan<- Event)
	Stop()
}

type FixedInstancer []string

func (d FixedInstancer) Register(ch chan<- Event) { ch <- Event{Instances: d} }

func (d FixedInstancer) Deregister(ch chan<- Event) {}

func (d FixedInstancer) Stop() {}
