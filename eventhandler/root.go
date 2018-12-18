package eventhandler

// EventHandler is used for two-way communication between the Discord and Webrcon clients
type EventHandler struct {
	Name      string
	Listeners map[string][]chan Message
}

// Message is used for emitting data through the EventHandler
type Message struct {
	Event   string
	User    string
	Message string
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
