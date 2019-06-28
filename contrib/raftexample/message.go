/*
 * Copyright (c) 2019, Tranquil Data, Inc. All rights reserved.
 */

package main

// MessageManager is the core manager for all messaging.
type MessageManager interface {
	Register(channelID uint64) MessageHandler
}

// MessageHandler is used to send and receive messages for a registered channel
type MessageHandler interface {
	Send(message Message) error
	Recv() (Message, error)
}

// Message is the interface for structures on channels.
type Message interface {
	From() uint64
	To() []uint64
	Content() []byte
}

// non-exported utility for a simple Message implementation
type simpleMessage struct {
	from    uint64
	to      []uint64
	content []byte
}

// NewMessage is a utility to create a direct message.
func NewMessage(to uint64, content []byte) Message {
	return &simpleMessage{from: 0, to: []uint64{to}, content: content}
}

// NewGroupMessage is a utility to create a multi-receiver message.
func NewGroupMessage(to []uint64, content []byte) Message {
	return &simpleMessage{from: 0, to: to, content: content}
}

func (m *simpleMessage) From() uint64 {
	return m.from
}

func (m *simpleMessage) To() []uint64 {
	return m.to
}

func (m *simpleMessage) Content() []byte {
	return m.content
}
