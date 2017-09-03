# wsu [![Godoc](https://godoc.org/github.com/orijtech/wsu?status.svg)](https://godoc.org/github.com/orijtech/wsu)
Websockets Utilities

## Client connection
Sample to get the latest 10000 Bitcoin-USD exchange events from GDAX/Coinbase
```go
package main

import (
        "log"

        "github.com/orijtech/wsu"
)

func main() {
        conn, err := wsu.NewClientConnection(&wsu.ClientSetup{
                URL: "wss://ws-feed.gdax.com",
        })
        if err != nil {
                log.Fatal(err)
        }
        defer conn.Close()

        conn.Send(&wsu.Message{
                Frame: []byte(`{"type": "subscribe", "product_ids": ["BTC-USD"]}`),
        })

        for i := 0; i < 10000; i++ {
                msg, ok := conn.Receive()
                if !ok {
                        break
                }
                log.Printf("#%d: TimeAt: (%v) Sequence #%d Frame: %s", i, msg.TimeAt, msg.Sequence, msg.Frame)
        }
}
```
