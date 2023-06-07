package event

const (
	ETOplog = Type("op:Oplog")
	ETPrint = Type("op:Print")
	ETTask  = Type("op:Task")
	ETData  = Type("op:Data")
)

// DataEvent kv数据存档事件
type DataEvent struct {
	ID   string                 `json:"id"`
	Data map[string]interface{} `json:"data"`
}

// PrintEvent 打印事件
type PrintEvent struct {
	ID  string `json:"id"`
	Msg string `json:"message"`
}

// TaskEvent 任务事件
type TaskEvent struct {
	ID   string `json:"id"`
	From string `json:"from"`
	To   string `json:"to"`
}

// OplogEvent op相关的event
type OplogEvent struct {
	ID         string                 `json:"id"`
	Ctx        map[string]interface{} `json:"ctx"`
	RemoteAddr string                 `json:"remoteAddr"`
	Name       string                 `json:"name"`
	Status     string                 `json:"status"`
	StartTime  int64                  `json:"start_time"`
	EndTime    int64                  `json:"end_time"`
	TimeUsed   int64                  `json:"time_used"`
	Details    []string               `json:"detail"`
	Exception  string                 `json:"exception,omitempty"`
}
