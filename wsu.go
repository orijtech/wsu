// Copyright 2017 orijtech. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wsu

import (
	"errors"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type Message struct {
	N        int    `json:"n,omitempty"`
	Err      error  `json:"error,omitepty"`
	Frame    []byte `json:"frame,omitempty"`
	Sequence uint64 `json:"seq,omitempty"`

	// WriteAck is an optional function that you can set when
	// providing a message to the Write routine, to have the write
	// byte count and error returned, after the respective write.
	WriteAck func(int, error) error

	// TimeAt is an output only field that's set
	// only when receiving to indicate the time that
	// a message was received.
	TimeAt time.Time `json:"time_at,omitempty"`
}

type ClientSetup struct {
	URL       string `json:"url,omitempty"`
	OriginURL string `json:"origin,omitempty"`
}

type Connection struct {
	writeChan chan *Message `json:"-"`
	readChan  chan *Message `json:"-"`
	cancelFn  func() error  `json:"-"`
}

var errAlreadyClosed = errors.New("already closed")

func makeCanceler() (<-chan bool, func() error) {
	cancelChan := make(chan bool)
	var cancelOnce sync.Once
	cancelFn := func() error {
		var err = errAlreadyClosed
		cancelOnce.Do(func() {
			err = nil
			close(cancelChan)
		})
		return err
	}
	return cancelChan, cancelFn
}

var errNilClientSetup = errors.New("expecting a non-nil client setup")

func NewClientConnection(cs *ClientSetup) (*Connection, error) {
	if cs == nil {
		return nil, errNilClientSetup
	}
	originURL := cs.OriginURL
	if originURL == "" {
		originURL = "http://localhost/"
	}

	conn, err := websocket.Dial(cs.URL, "", originURL)
	if err != nil {
		return nil, err
	}

	cancelChan, cancelFn := makeCanceler()
	writeChan := make(chan *Message)
	readChan := make(chan *Message)

	// Connection close goroutine
	go func() {
		<-cancelChan
		_ = conn.Close()
	}()

	// Write goroutine
	go func() {
		defer close(writeChan)

		for {
			select {
			case <-cancelChan:
				return
			case msg := <-writeChan:
				n, err := conn.Write(msg.Frame)
				if ackFn := msg.WriteAck; ackFn != nil {
					go ackFn(n, err)
				}
			}
		}
	}()

	// Read goroutine
	go func() {
		recvBuf := make([]byte, 32*1024*1024)
		sequence := uint64(0)
		var readWG sync.WaitGroup

		defer func() {
			readWG.Wait()
			close(readChan)
		}()

		for {
			select {
			case <-cancelChan:
				return
			default:
				n, err := conn.Read(recvBuf)
				recvTime := time.Now()
				sequence += 1

				var buf []byte
				if err == nil && n > 0 {
					buf = make([]byte, n)
					copy(buf, recvBuf[:n])
				}

				msg := &Message{
					N:        n,
					Sequence: sequence,
					Frame:    buf,
					TimeAt:   recvTime,
				}

				// It is essential that we don't block sending this frame
				readWG.Add(1)
				go func(rMsg *Message) {
					readChan <- rMsg
					readWG.Done()
				}(msg)
			}
		}
	}()

	aConn := &Connection{
		writeChan: writeChan,
		readChan:  readChan,
		cancelFn:  cancelFn,
	}

	return aConn, nil
}

func (c *Connection) Receive() (msg *Message, ok bool) {
	msg, ok = <-c.readChan
	return msg, ok
}

func (c *Connection) Send(msg *Message) {
	c.writeChan <- msg
}

func (c *Connection) Close() error {
	if c.cancelFn == nil {
		return nil
	}
	return c.cancelFn()
}
