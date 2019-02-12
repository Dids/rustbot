package eventhandler

// EventHandler is used for two-way communication between the Discord and Webrcon clients
type EventHandler struct {
	Name      string
	Listeners map[string][]chan Message
}

// MessageType represents the type of a message
type MessageType string

const (
	// DefaultType is a message type
	DefaultType MessageType = "Default"
	// JoinType is a message type
	JoinType MessageType = "Join"
	// DisconnectType is a message type
	DisconnectType MessageType = "Disconnect"
	// OtherKillType is a message type
	OtherKillType MessageType = "Other"
	// PvPKillType is a message type
	PvPKillType MessageType = "PvP"
	// StatusType is a message type
	StatusType MessageType = "Status"
	// ServerConnectedType is a message type
	ServerConnectedType MessageType = "ServerConnected"
	// ServerDisconnectedType is a message type
	ServerDisconnectedType MessageType = "ServerDisconnected"
	// TraceLogType is a message type
	TraceLogType MessageType = "TraceLog"
	// InfoLogType is a message type
	InfoLogType MessageType = "InfoLog"
	// WarningLogType is a message type
	WarningLogType MessageType = "WarningLog"
	// ErrorLogType is a message type
	ErrorLogType MessageType = "ErrorLog"
	// PanicLogType is a message type
	PanicLogType MessageType = "PanicLog"
)

// Message is used for emitting data through the EventHandler
type Message struct {
	Event   string
	User    string
	Message string
	Type    MessageType
}

// AddListener adds an event listener to the EventHandler struct instance
func (handler *EventHandler) AddListener(name string, channel chan Message) {
	// Create the listeners if they don't yet exist
	if handler.Listeners == nil {
		handler.Listeners = make(map[string][]chan Message)
	}
	// Find/set the listener
	if _, ok := handler.Listeners[name]; ok {
		handler.Listeners[name] = append(handler.Listeners[name], channel)
	} else {
		handler.Listeners[name] = []chan Message{channel}
	}
}

// RemoveListener removes an event listener from the EventHandler struct instance
func (handler *EventHandler) RemoveListener(name string, channel chan Message) {
	// Find the listener
	if _, ok := handler.Listeners[name]; ok {
		for i := range handler.Listeners[name] {
			if handler.Listeners[name][i] == channel {
				handler.Listeners[name] = append(handler.Listeners[name][:i], handler.Listeners[name][i+1:]...)
				break
			}
		}
	}
}

// Emit emits an event on the EventHandler struct instance
func (handler *EventHandler) Emit(message Message) {
	// Find the listener
	if _, ok := handler.Listeners[message.Event]; ok {
		for _, handler := range handler.Listeners[message.Event] {
			go func(handler chan Message) {
				handler <- message
			}(handler)
		}
	}
}
