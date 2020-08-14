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
	OptMessage = 0x00
	OptJoin    = 0x01
	OptLeave   = 0x02
	OptRaw     = 0x03
)

const (
	TagField = "Type"
)

var registeredHandler = make(map[string]func(*Message, chan *Message) []*Message)

func RegisterHandler(tag string, f func(*Message, chan *Message) []*Message) {
	registeredHandler[tag] = f
}

func handler(msg *Message, broker chan *Message) []*Message {
	if h, ok := registeredHandler[msg.Content.(map[string]interface{})[TagField].(string)]; ok {
		return h(msg, broker)
	}
	return nil
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

func Run() {
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
	group  *group
	Info   interface{}
	output *list.List
	lock   *sync.Mutex
}

func (c *Client) GetPeerNum() int {
	if c.group == nil {
		return 0
	}
	return len(c.group.clients)
}

type group struct {
	id      string
	broker  chan *Message
	clients map[*Client]bool
}

func (c *Client) send(content interface{}, broker chan *Message, counter *sync.WaitGroup) {
	defer counter.Done()
	for {
		if err := c.sock.WriteJSON(content); err != nil {
			broker <- &Message{operation: OptLeave, Client: c}
			return
		}
		c.lock.Lock()
		c.output.Remove(c.output.Front())
		if c.output.Len() == 0 {
			c.lock.Unlock()
			return
		}
		content = c.output.Front().Value
		c.lock.Unlock()
	}
}

func (c *Client) scheduleSend(content interface{}, broker chan *Message, counter *sync.WaitGroup) {
	c.lock.Lock()
	if c.output.Len() != 0 {
		defer c.lock.Unlock()
		c.output.PushBack(content)
	} else {
		c.lock.Unlock()
		c.output.PushBack(content)
		counter.Add(1)
		go c.send(content, broker, counter)
	}
}

func (c *Client) recv(broker chan *Message, counter *sync.WaitGroup) {
	defer counter.Done()
	for {
		content := make(map[string]interface{})
		if err := c.sock.ReadJSON(&content); err != nil {
			if err == io.EOF || websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				broker <- &Message{operation: OptLeave, Client: c}
				return
			} else {
				log.Println(err)
				broker <- &Message{operation: OptLeave, Client: c}
				return
			}
		}
		broker <- &Message{c, content, OptRaw}
	}
}

func (g *group) brokeMessage() {
	log.Printf("group %s: broker started\n", g.id)
	var orphans []*Client
	closed := false
	counter := sync.WaitGroup{}
	for msg := range g.broker {
		switch msg.operation {
		case OptMessage:
			if !closed {
				if msg.Client == nil {
					for c := range g.clients {
						c.scheduleSend(msg.Content, g.broker, &counter)
					}
				} else if _, ok := g.clients[msg.Client]; ok {
					msg.Client.scheduleSend(msg.Content, g.broker, &counter)
				}
			}
		case OptJoin:
			if !closed {
				if _, ok := g.clients[msg.Client]; !ok {
					log.Printf("group %s: client %v joined\n", g.id, msg.Client)
					msg.Client.group = g
					g.clients[msg.Client] = true
					counter.Add(1)
					go msg.Client.recv(g.broker, &counter)
				}
			} else {
				orphans = append(orphans, msg.Client)
			}
		case OptLeave:
			if _, ok := g.clients[msg.Client]; ok {
				log.Printf("group %s: client %v left\n", g.id, msg.Client)
				msg.Client.group = nil
				delete(g.clients, msg.Client)
				if err := msg.Client.sock.Close(); err != nil {
					log.Println(err)
				}
				if len(g.clients) == 0 && !closed {
					closed = true
					// Wait for all send & recv go routines, then ask the manager to close the channel.
					// This can prevent go routines from unexpectedly writing OptLeave message to a closed channel.
					log.Printf("group %s: closed, waiting for go routines\n", g.id)
					counter.Wait()
					go func() { groupsChannel <- &groupOperation{g.id, nil} }()
				}
			}
		case OptRaw:
			if !closed {
				res := handler(msg, g.broker)
				for _, msg := range res {
					if msg.Client == nil {
						for c := range g.clients {
							c.scheduleSend(msg.Content, g.broker, &counter)
						}
					} else if _, ok := g.clients[msg.Client]; ok {
						msg.Client.scheduleSend(msg.Content, g.broker, &counter)
					}
				}
			}
		}
	}
	for _, o := range orphans {
		groupsChannel <- &groupOperation{g.id, o}
	}
	log.Printf("group %s: broker stopped\n", g.id)
}
