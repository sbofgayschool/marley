package sock

const (
    TypeField = "type"
)

var registeredHandler = make(map[string]func(*Message, chan *Message))

func RegisterHandler(tag string, f func(*Message, chan *Message)) {
    registeredHandler[tag] = f
}

func handler(msg *Message, broker chan *Message) {
    if h, ok := registeredHandler[msg.Content.(map[string]interface{})[TypeField].(string)]; ok {
        h(msg, broker)
    }
}
