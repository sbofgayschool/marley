package sock

import (
	"container/list"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	OptJoin    = 0x01
	OptLeave   = 0x02
	OptMessage = 0x00
)

const (
	TagField = "type"
)

var registeredHandler = make(map[string]func(*Message, chan *Message))

func RegisterHandler(tag string, f func(*Message, chan *Message)) {
	registeredHandler[tag] = f
}

func handler(msg *Message, broker chan *Message) {
	if h, ok := registeredHandler[msg.Content.(map[string]interface{})[TagField].(string)]; ok {
		h(msg, broker)
	}
}

var groups = make(map[string]*group)
var groupsChannel = make(chan *groupOperation, 1)
var upGrader = websocket.Upgrader{}

type groupOperation struct {
	id     string
	client *Client
}

func NewClient(c *gin.Context, id string, userInfo interface{}) error {
	sock, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return err
	}
	groupsChannel <- &groupOperation{id, &Client{sock: sock, Gid: id, Info: userInfo, output: &list.List{}, lock: &sync.Mutex{}}}
	return nil
}

func HandleGroupOperation() {
	for opt := range groupsChannel {
		if opt.client != nil {
			g, ok := groups[opt.id]
			if !ok {
				g = &group{id: opt.id, clients: make(map[*Client]bool), broker: make(chan *Message, 256)}
				groups[opt.id] = g
				go g.brokeMessage()
			}
			g.broker <- &Message{opt.client, nil, OptJoin}
		} else {
			if g, ok := groups[opt.id]; ok {
				close(g.broker)
				delete(groups, opt.id)
			}
		}
	}
}

type Message struct {
	Client    *Client
	Content   interface{}
	operation int
}

type Client struct {
	sock   *websocket.Conn
	Gid    string
	Info   interface{}
	output *list.List
	lock   *sync.Mutex
}

type group struct {
	id      string
	broker  chan *Message
	clients map[*Client]bool
}

func CountGroupClient(id string) int {
	if g, ok := groups[id]; ok {
		return len(g.clients)
	}
	return 0
}

func (c *Client) scheduleSend(content interface{}, broker chan *Message) {
	if c.output.Len() != 0 {
		c.lock.Lock()
		defer c.lock.Unlock()
		c.output.PushBack(content)
		if c.output.Len() == 1 {
			go c.send(content, broker)
		}
	} else {
		c.output.PushBack(content)
		go c.send(content, broker)
	}
}

func (c *Client) send(content interface{}, broker chan *Message) {
	for {
		if err := c.sock.WriteJSON(content); err != nil {
			broker <- &Message{operation: OptJoin, Client: c}
			return
		}
		c.lock.Lock()
		l := c.output.Len()
		c.output.Remove(c.output.Front())
		if l == 1 {
			c.lock.Unlock()
			return
		}
		content = c.output.Front().Value
		c.lock.Unlock()
	}
}

func (c *Client) recv(broker chan *Message) {
	for {
		content := make(map[string]interface{})
		if err := c.sock.ReadJSON(&content); err != nil {
			if err == io.EOF {
				broker <- &Message{operation: OptJoin, Client: c}
				return
			}
		}
		handler(&Message{c, content, OptMessage}, broker)
	}
}

func (g *group) brokeMessage() {
	var orphans []*Client
	closed := false
	for msg := range g.broker {
		switch msg.operation {
		case OptJoin:
			if !closed {
				if _, ok := g.clients[msg.Client]; !ok {
					g.clients[msg.Client] = true
					go msg.Client.recv(g.broker)
				}
			} else {
				orphans = append(orphans, msg.Client)
			}
		case OptLeave:
			if _, ok := g.clients[msg.Client]; ok {
				delete(g.clients, msg.Client)
				if err := msg.Client.sock.Close(); err != nil {
					log.Println(err)
				}
				if len(g.clients) == 0 && !closed {
					closed = true
					go func() { groupsChannel <- &groupOperation{g.id, nil} }()
				}
			}
		case OptMessage:
			if !closed {
				if msg.Client == nil {
					for c := range g.clients {
						c.scheduleSend(msg.Content, g.broker)
					}
				} else if _, ok := g.clients[msg.Client]; ok {
					msg.Client.scheduleSend(msg.Content, g.broker)
				}
			}
		}
	}
	for _, o := range orphans {
		groupsChannel <- &groupOperation{g.id, o}
	}
}
