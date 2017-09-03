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
	"io"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/net/websocket"

	"github.com/orijtech/wsu"
)

func TestClientConnection(t *testing.T) {
	cst := httptest.NewServer(websocket.Handler(func(conn *websocket.Conn) {
		io.Copy(conn, conn)
	}))
	defer cst.Close()

	parsedURL, err := url.Parse(cst.URL)
	if err != nil {
		t.Fatalf("parseURL: %v", err)
	}
	parsedURL.Scheme = "ws"
	t.Logf("parsedURL: %s\n", parsedURL)

	conn, err := wsu.NewClientConnection(&wsu.ClientSetup{
		URL: parsedURL.String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	n := 10
	writeLog := new(bytes.Buffer)
	go func() {
		for i := 0; i < n; i++ {
			buf := []byte(fmt.Sprintf("Sequence: %d\n", i))
			conn.Send(&wsu.Message{Frame: buf})
			writeLog.Write(buf)
		}
		conn.Send(&wsu.Message{Frame: []byte(fmt.Sprintf("Bye!"))})
		conn.Close()
	}()

	readLog := new(bytes.Buffer)
	doneChan := make(chan bool)
	go func() {
		defer close(doneChan)
		for {
			msg, ok := conn.Receive()
			if !ok {
				return
			}
			readLog.Write(msg.Frame)
		}
	}()
	<-doneChan

	r, w := readLog.Bytes(), writeLog.Bytes()
	t.Logf("readLog=%s\nwriteLog=%s", r, w)
	if g, w := len(r), 5; g < w {
		t.Errorf("got=%d want >= %d", g, w)
	}
}
