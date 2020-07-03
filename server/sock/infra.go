package sock

import (
	"container/list"
    "log"
    "net"
	"sync"
)

const (
	OptJoin  = 0x00
	OptLeave = 0x01
)

var groups = make(map[string]*group)
var groupsChannel = make(chan *groupOperation)

type groupOperation struct {
	id   string
	client *client
}

func handleGroupOperation() {
	for opt := range groupsChannel {
		if opt.client != nil {
			g, ok := groups[opt.id]
			if !ok {
				g = &group{id: opt.id, broker: make(chan *message, 32)}
				groups[opt.id] = g
				go g.brokeMessage()
			}
			g.broker <- &message{opt.client, nil, OptJoin}
		} else {
			if g, ok := groups[opt.id]; ok {
				close(g.broker)
				delete(groups, opt.id)
			}
		}
	}
}

type message struct {
	client    *client
	content   interface{}
	operation int
}

type client struct {
	sock   *net.TCPConn
	output *list.List
	lock   *sync.Mutex
}

type group struct {
	id      string
	broker  chan *message
}

func (c *client) scheduleSend(msg *message) {
    if c.output.Len() != 0 {
        c.lock.Lock()
        defer c.lock.Unlock()
        c.output.PushBack(msg)
        if c.output.Len() == 1 {
            go c.send(msg)
        }
    } else {
        c.output.PushBack(msg)
        go c.send(msg)
    }
}

func (c *client) send(msg *message) {
    for {
        if _, err := c.sock.Write(msg.content.([]byte)); err != nil {
        }
        c.lock.Lock()
        c.output.Remove(c.output.Front())
        if c.output.Len() == 0 {
            c.lock.Unlock()
            return
        }
        msg = c.output.Front().Value.(*message)
        c.lock.Unlock()
    }
}

func (c *client) recv() {

}

func (g *group) brokeMessage() {
    clients := make(map[*client]bool)
    var orphans []*client
    closed := false
    for msg := range g.broker {
        switch msg.operation {
        case OptJoin:
            if !closed {
                clients[msg.client] = true
                go msg.client.recv()
            } else {
                orphans = append(orphans, msg.client)
            }
        case OptLeave:
            delete(clients, msg.client)
            if err := msg.client.sock.Close(); err != nil {
                log.Println(err)
            }
            if len(clients) == 0 && !closed {
                closed = true
                go func() {groupsChannel <- &groupOperation{g.id, nil}} ()
            }
        }
    }
    for _, o := range orphans {
        groupsChannel <- &groupOperation{g.id, o}
    }
}
