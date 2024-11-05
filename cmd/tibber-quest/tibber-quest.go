package main

import (
	"log"
	"github.com/questdb/go-questdb-client/v3"
	"context"
	"time"
  "encoding/json"
	"github.com/nats-io/nats.go"
)


type Data struct {
	Timestamp              time.Time `json:"timestamp"`
	Power                  int       `json:"power"`
	AccumulatedConsumption float64   `json:"accumulatedConsumption"`
	AccumulatedCost        float64   `json:"accumulatedCost"`
	Currency               string    `json:"currency"`
	MinPower               int       `json:"minPower"`
	AveragePower           float64   `json:"averagePower"`
	MaxPower               float64   `json:"maxPower"`
} 

func main() {
	ctx := context.TODO()
	client, err := questdb.LineSenderFromEnv(ctx)
	if err != nil {
		panic(err)
	}
  nc, _ := nats.Connect(nats.DefaultURL)
  defer nc.Drain()

	sub, _ := nc.SubscribeSync("tibber.api")
	for true {
		msg, _ := sub.NextMsg(5 * time.Second)
		if msg != nil {
			var data Data
			err := json.Unmarshal(msg.Data, &data)

			if err != nil {
				panic(err)
			}

			err = client.Table("tibber").
				Int64Column("power", int64(data.Power)).
				Float64Column("accumulated_consumption", data.AccumulatedConsumption).
				Float64Column("accumulated_cost", data.AccumulatedCost).
				Int64Column("min_power", int64(data.MinPower)).
				Float64Column("average_power", data.AveragePower).
				Float64Column("max_power", data.MaxPower).
				StringColumn("currency", data.Currency).
				At(ctx, data.Timestamp)

			if err != nil {
				panic(err)
			}
			err = client.Flush(ctx)
			if err != nil {
				panic("Failed to flush data")
			}
			log.Println("Inserted message:",data.Power,"W")

		} else {
			log.Println("Message timeout")
		}
	}

}

  

