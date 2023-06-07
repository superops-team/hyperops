// Package event 事件转发基本的生产消费模型
package event

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrBusClosed 事件关闭
	ErrBusClosed = fmt.Errorf("event bus is closed")
	// NowFunc 时间戳生成
	NowFunc = time.Now
)

// Type 订阅的事件类型
type Type string

// Event 事件消息体
type Event struct {
	Type      Type
	Timestamp int64
	SessionID string
	Payload   interface{}
}

// MakeEvent 产生event
func MakeEvent(typ Type, sessionID string, payload interface{}) Event {
	e := Event{
		Type:      typ,
		Timestamp: NowFunc().UnixNano(),
		SessionID: sessionID,
		Payload:   payload,
	}
	return e
}

// Handler 事件处理方法定义
type Handler func(ctx context.Context, e Event) error

// Publisher 发布器接口声明
type Publisher interface {
	Publish(ctx context.Context, typ Type, payload interface{}) error
	PublishID(ctx context.Context, typ Type, sessionID string, payload interface{}) error
}

// Bus 事件订阅生产转发集中处理
type Bus interface {
	// Publish 发布事件
	Publish(ctx context.Context, typ Type, data interface{}) error
	// PublishID 发布ID
	PublishID(ctx context.Context, typ Type, sessionID string, data interface{}) error
	// Subscribe 订阅一个或者多个事件类型
	SubscribeTypes(handler Handler, eventTypes ...Type)
	// SubscribeID 订阅某个ID的事件
	SubscribeID(handler Handler, sessionID string)
	// SubscribeAll 订阅所有事件
	SubscribeAll(handler Handler)
	// NumSubscriptions 返回订阅者数量
	NumSubscribers() int
}

// NilBus 空接口实现
var NilBus = nilBus{}

type nilBus struct{}

// assert at compile time that nilBus implements the Bus interface
var _ Bus = (*nilBus)(nil)

// Publish 不处理消息
func (nilBus) Publish(_ context.Context, _ Type, _ interface{}) error {
	return nil
}

// PublishID 不处理
func (nilBus) PublishID(_ context.Context, _ Type, _ string, _ interface{}) error {
	return nil
}

// SubscribeTypes 什么都不做
func (nilBus) SubscribeTypes(handler Handler, eventTypes ...Type) {}

func (nilBus) SubscribeID(handler Handler, id string) {}

func (nilBus) SubscribeAll(handler Handler) {}

func (nilBus) NumSubscribers() int {
	return 0
}

type bus struct {
	lk      sync.RWMutex
	closed  bool
	subs    map[Type][]Handler
	allSubs []Handler
	idSubs  map[string][]Handler
}

// assert at compile time that bus implements the Bus interface
var _ Bus = (*bus)(nil)

// NewBus creates a new event bus. Event busses should be instantiated as a
// singleton. If the passed in context is cancelled, the bus will stop emitting
// events and close all subscribed channels
func NewBus(ctx context.Context) Bus {
	b := &bus{
		subs:    map[Type][]Handler{},
		idSubs:  map[string][]Handler{},
		allSubs: []Handler{},
	}

	go func(b *bus) {
		<-ctx.Done()
		b.lk.Lock()
		b.closed = true
		b.lk.Unlock()
	}(b)

	return b
}

// Publish sends an event to the bus
func (b *bus) Publish(ctx context.Context, typ Type, payload interface{}) error {
	return b.publish(ctx, typ, "", payload)
}

// PublishID sends an event with a given sessionID to the bus
func (b *bus) PublishID(ctx context.Context, typ Type, sessionID string, payload interface{}) error {
	return b.publish(ctx, typ, sessionID, payload)
}

func (b *bus) publish(ctx context.Context, typ Type, sessionID string, payload interface{}) error {
	b.lk.RLock()
	defer b.lk.RUnlock()

	if b.closed {
		return ErrBusClosed
	}

	e := Event{
		Type:      typ,
		Timestamp: NowFunc().UnixNano(),
		SessionID: sessionID,
		Payload:   payload,
	}

	for _, handler := range b.subs[typ] {
		if err := handler(ctx, e); err != nil {
			return err
		}
	}

	if sessionID != "" {
		for _, handler := range b.idSubs[sessionID] {
			if err := handler(ctx, e); err != nil {
				return err
			}
		}
	}

	for _, handler := range b.allSubs {
		if err := handler(ctx, e); err != nil {
			return err
		}
	}

	return nil
}

// Subscribe 初始化处理器
func (b *bus) SubscribeTypes(handler Handler, eventTypes ...Type) {
	b.lk.Lock()
	defer b.lk.Unlock()

	for _, typ := range eventTypes {
		b.subs[typ] = append(b.subs[typ], handler)
	}
}

// SubscribeID 指定ID的消费者
func (b *bus) SubscribeID(handler Handler, sessionID string) {
	b.lk.Lock()
	defer b.lk.Unlock()
	b.idSubs[sessionID] = append(b.idSubs[sessionID], handler)
}

// SubscribeAll 订阅全量
func (b *bus) SubscribeAll(handler Handler) {
	b.lk.Lock()
	defer b.lk.Unlock()
	b.allSubs = append(b.allSubs, handler)
}

// NumSubscribers 返回订阅数量
func (b *bus) NumSubscribers() int {
	b.lk.RLock()
	defer b.lk.RUnlock()
	total := 0
	for _, handlers := range b.subs {
		total += len(handlers)
	}
	for _, handlers := range b.idSubs {
		total += len(handlers)
	}
	total += len(b.allSubs)
	return total
}
