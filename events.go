package qb

type EventType string

const (
	BeforeCreate EventType = "before_create"
	AfterCreate  EventType = "after_create"
	BeforeUpdate EventType = "before_update"
	AfterUpdate  EventType = "after_update"
	BeforeDelete EventType = "before_delete"
	AfterDelete  EventType = "after_delete"
)

type EventHandler func(interface{}) error

// Events добавляет поддержку событий
type Events struct {
	handlers map[EventType][]EventHandler
}

// On регистрирует обработчик события
func (e *Events) On(event EventType, handler EventHandler) {
	if e.handlers == nil {
		e.handlers = make(map[EventType][]EventHandler)
	}
	e.handlers[event] = append(e.handlers[event], handler)
}

// Trigger вызывает обработчики события
func (e *Events) Trigger(event EventType, data interface{}) error {
	if handlers, ok := e.handlers[event]; ok {
		for _, handler := range handlers {
			if err := handler(data); err != nil {
				return err
			}
		}
	}
	return nil
}
