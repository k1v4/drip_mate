package entity

type Message struct {
	Key       []byte
	Value     []byte
	Headers   map[string][]byte
	Topic     string
	Partition int
	Offset    int64
}
