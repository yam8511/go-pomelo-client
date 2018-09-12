package main

import (
	"encoding/json"
	"log"

	"github.com/yam8511/go-pomelo-client"
)

// PomeloAddress Pomelo位址
const PomeloAddress = "ip"

func main() {
	c := client.NewConnector()
	req := map[string]interface{}{
		"sys": map[string]string{
			"version": "1.1.1",
			"type":    "js-websocket",
		},
	}

	err := c.SetHandshake(req)
	if err != nil {
		log.Fatal("設定握手協定失敗 : ", err)
	}

	go func() {
		err = c.Run(PomeloAddress)
		log.Fatal(err)
	}()

	data := map[string]interface{}{
		"username": "zuolar",
	}
	bata, err := json.Marshal(data)
	if err != nil {
		defer c.Close()
		log.Fatal(err)
		return
	}

	done := make(chan byte)
	c.Request("entry", bata, func(data []byte) {
		log.Println("接收回傳資料 : ", string(data))
		done <- 0
	})
	<-done
}
