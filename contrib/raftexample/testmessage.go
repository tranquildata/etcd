/*
 * Copyright (c) 2019, Tranquil Data, Inc. All rights reserved.
 */

package main

import (
	//	"fmt"
	"sync"
)

type testFabric struct {
	managers *sync.Map //[uint64]*testMessageManager
}

type fabricMessage struct {
	channel uint64
	message Message
}

type testMessageManager struct {
	nodeID   uint64
	channels map[uint64][]*testMessageHandler
	handlerC chan *testMessageHandler
	fabricC  chan *fabricMessage
}

type testMessageHandler struct {
	channel    uint64
	send, recv chan Message
}

// the "network fabric" to communicate between nodes
var fabric = &testFabric{managers: &sync.Map{}}

// NewMessageManager returns a new instance of MessageManager.
func NewMessageManager(nodeID uint64) MessageManager {
	manager := &testMessageManager{
		nodeID:   nodeID,
		channels: map[uint64][]*testMessageHandler{},
		handlerC: make(chan *testMessageHandler),
		fabricC:  make(chan *fabricMessage),
	}
	go manager.managerListener()
	fabric.managers.Store(nodeID, manager)
	return manager
}

func (tmm *testMessageManager) managerListener() {
	for {
		select {
		case m := <-tmm.fabricC:
			if handlers, exists := tmm.channels[m.channel]; exists {
				for _, h := range handlers {
					//fmt.Printf("Receiving from %d to %d: %s\n", m.message.From(), m.message.To()[0], string(m.message.Content()))
					if h.recv == nil {
						return
					}
					h.recv <- m.message
				}
			}
		case h := <-tmm.handlerC:
			if _, exists := tmm.channels[h.channel]; !exists {
				tmm.channels[h.channel] = []*testMessageHandler{}
			}
			tmm.channels[h.channel] = append(tmm.channels[h.channel], h)
		}
	}
}

/* Implement MessageManager */

func (tmm *testMessageManager) Register(channelID uint64) MessageHandler {
	tmc := &testMessageHandler{channel: channelID, send: make(chan Message, 1024), recv: make(chan Message, 1024)}
	tmm.handlerC <- tmc
	go tmm.handlerListener(channelID, tmc.send)
	return tmc
}

func (tmm *testMessageManager) handlerListener(channelID uint64, channel chan Message) {
	for {
		message := <-channel
		if message == nil {
			return
		}
		for _, target := range message.To() {
			//fmt.Printf("Sending from %d to %d: %s\n", tmm.nodeID, target, string(message.Content()))
			if remote, exists := fabric.managers.Load(target); exists {
				remote.(*testMessageManager).fabricC <- tmm.fabricMessage(channelID, message)
			}
		}
	}
}

func (tmm *testMessageManager) fabricMessage(channelID uint64, m Message) *fabricMessage {
	message := &simpleMessage{from: tmm.nodeID, to: m.To(), content: m.Content()}
	return &fabricMessage{channel: channelID, message: message}
}

/* Implement MessageHandler */

func (tmh *testMessageHandler) Send(message Message) error {
	if tmh.send != nil {
		tmh.send <- message
	}
	return nil
}

func (tmh *testMessageHandler) Recv() (Message, error) {
	if tmh.recv == nil {
		return nil, nil
	}
	return <-tmh.recv, nil
}

func (tmh *testMessageHandler) Stop() {
	close(tmh.send)
	tmh.send = nil
	close(tmh.recv)
	tmh.recv = nil
}
