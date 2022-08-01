package zeebe

type Hazelcast struct {
	sendToRingbufferFunc func(data []byte) error
}

type HazelcastClient interface {
	SendToRingbuffer(data []byte) error
}

func (h *Hazelcast) SendToRingbuffer(data []byte) error {
	return h.sendToRingbufferFunc(data)
}
