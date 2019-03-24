package interfaces

type AgentBehavior interface {
	OnMessage(data []byte) error
	SendMessage(data []byte) error
	OnDisconnected()
}
