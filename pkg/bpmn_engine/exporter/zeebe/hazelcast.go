package zeebe

type Hazelcast struct {
	sendToRingbufferFunc func(data []byte)
}

type HazelcastClient interface {
	SendToRingbuffer(data []byte)
}

func (h *Hazelcast) SendToRingbuffer(data []byte) {
	h.sendToRingbufferFunc(data)
}
