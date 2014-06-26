package main

const (
	queueSize = 100
)

const (
	MT_KeyPress = iota
	MT_UpdateText
	MT_UpdateGfx
	MT_UpdateEdit
	MT_EditStart
	MT_EditStop
	MT_Quit
	MT_VisibleAreaChange
)

type Message interface {
	MessageType() int
}

type MessageBase struct {
	messageType int
}

func (this *MessageBase) MessageType() int { return this.messageType }

type Channel struct {
	name string
	qc   int
	c    chan Message
	f    []int
	p    bool
	b    *MessageBroker
}

func (this *Channel) Pause() {
	this.p = true
}

func (this *Channel) Resume() {
	this.p = false
}

func (this *Channel) Wait() Message {
	m := <-this.c
	this.qc--
	return m
}

func (this *Channel) Poll() Message {
	select {
	case m := <-this.c:
		this.qc--
		return m
	default:
	}

	return nil
}

func (this *Channel) Publish(m Message) {
	this.b.Publish(m)
}

func (this *Channel) PublishId(messageId int) {
	this.b.PublishId(messageId)
}

func filterContains(filter []int, mt int) bool {
	for _, f := range filter {
		if f == mt {
			return true
		}
	}
	return false
}

func (this *Channel) push(m Message) {

	if len(this.f) == 0 || filterContains(this.f, m.MessageType()) {
		this.qc++
		if this.qc == queueSize {
			println(this.name, "Queue filled.")
		}
		this.c <- m
	}
}

type MessageBroker struct {
	channels []*Channel
}

func CreateMessageBroker() *MessageBroker {
	return &MessageBroker{
		make([]*Channel, 0, 10)}
}

func (this *MessageBroker) Subscribe(name string, messageTypes ...int) *Channel {
	l := &Channel{name, 0, make(chan Message, queueSize), messageTypes, false, this}
	this.channels = append(this.channels, l)
	return l
}

func (this *MessageBroker) Publish(m Message) {
	go func() {
		for _, l := range this.channels {
			if !l.p {
				l.push(m)
			}
		}
	}()
}

func (this *MessageBroker) PublishId(messageType int) {
	this.Publish(&MessageBase{messageType})
}
