package live

type liveEventType int

const (
	APICreate liveEventType = iota + 1
	APIUpdate
	APIDelete
	AdminCreate
	AdminUpdate
	AdminDelete
	Save
	Delete
	Approve
	Reject
	Enable
	Disable
)

type LiveEvent struct {
	Type    liveEventType
	content interface{}
}

func (l LiveEvent) Content() interface{} {
	return l.content
}
