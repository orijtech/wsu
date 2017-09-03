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

package wsu_test

import (
	"bytes"
	"fmt"
	"log"

	"github.com/orijtech/wsu"
)

func Example_GDAX_bitcoinEvents() {
	conn, err := wsu.NewClientConnection(&wsu.ClientSetup{
		URL: "wss://ws-feed.gdax.com",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Subscribe to your desired events
	conn.Send(&wsu.Message{
		Frame: []byte(`{"type": "subscribe", "product_ids": ["BTC-USD"]}`),
		WriteAck: func(n int, err error) error {
			log.Printf("WriteAck n: %d err: %v", n, err)
			return nil
		},
	})

	for i := 0; i < 10000; i++ {
		msg, ok := conn.Receive()
		if !ok {
			break
		}
		log.Printf("#%d: TimeAt: (%v) Sequence #%d Frame: %s", i, msg.TimeAt, msg.Sequence, msg.Frame)
	}
}

func Example_Echo() {
	conn, err := wsu.NewClientConnection(&wsu.ClientSetup{
		URL: "ws://127.0.0.1:62954",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	go func() {
		for i := 0; i < 10; i++ {
			conn.Send(&wsu.Message{
				Frame: []byte(fmt.Sprintf(`This is sequence number: %d`, i)),
			})
		}
		conn.Send(&wsu.Message{
			Frame: []byte("Bye bye!"),
		})

	}()

	doneChan := make(chan bool)
	go func() {
		defer close(doneChan)
		for {
			msg, ok := conn.Receive()
			if !ok {
				break
			}
			log.Printf("TimeAt: (%v) Sequence #%d Frame: %s", msg.TimeAt, msg.Sequence, msg.Frame)
			if bytes.Contains(msg.Frame, []byte("Bye")) {
				return
			}
		}
	}()
	<-doneChan
}
