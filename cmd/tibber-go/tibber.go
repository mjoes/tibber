package main

import (
	"log"
	"net/http"
	"os"
	"time"
  "encoding/json"
	"github.com/nats-io/nats.go"

	jsonutil "github.com/hasura/go-graphql-client/pkg/jsonutil"
	graphql "github.com/hasura/go-graphql-client"
)


const (
	subscriptionEndpoint = "wss://websocket-api.tibber.com/v1-beta/gql/subscriptions"
)


func main() {
	if err := startSubscription(); err != nil {
		panic(err)
	}
}

func startSubscription() error {
	demoToken := os.Getenv("TOKEN")
	if demoToken == "" {
		panic("TOKEN env variable is required")
	}

	client := graphql.NewSubscriptionClient(subscriptionEndpoint).
		WithProtocol(graphql.GraphQLWS).
		WithWebSocketOptions(graphql.WebsocketOptions{
			HTTPClient: &http.Client{
				Transport: headerRoundTripper{
					setHeaders: func(req *http.Request) {
						req.Header.Set("User-Agent", "go-graphql-client/0.9.0")
					},
					rt: http.DefaultTransport,
				},
			},
		}).
		WithConnectionParams(map[string]interface{}{
			"token": demoToken,
		}).//WithLog(log.Println).
		OnError(func(sc *graphql.SubscriptionClient, err error) error {
			panic(err)
		})

	defer client.Close()

  
  nc, _ := nats.Connect(nats.DefaultURL)
  defer nc.Drain()

	var sub struct {
		LiveMeasurement struct {
			Timestamp              time.Time `graphql:"timestamp"`
			Power                  int       `graphql:"power"`
			AccumulatedConsumption float64   `graphql:"accumulatedConsumption"`
			AccumulatedCost        float64   `graphql:"accumulatedCost"`
			Currency               string    `graphql:"currency"`
			MinPower               int       `graphql:"minPower"`
			AveragePower           float64   `graphql:"averagePower"`
			MaxPower               float64   `graphql:"maxPower"`
		} `graphql:"liveMeasurement(homeId: $homeId)"`
	}

	variables := map[string]interface{}{
		"homeId": graphql.ID(os.Getenv("HOME_ID")),
	}
	_, err := client.Subscribe(sub, variables, func(data []byte, err error) error {

		if err != nil {
			log.Println("ERROR: ", err)
			return nil
		}

		if data == nil {
			return nil
		}
		err = jsonutil.UnmarshalGraphQL(data, &sub)
    b, err := json.Marshal(sub.LiveMeasurement)
    nc.Publish("tibber.api", b)
    log.Println("Published message:", sub.LiveMeasurement.Power, "W")
		return nil
	})

	if err != nil {
		panic(err)
	}

	return client.Run()
}

type headerRoundTripper struct {
	setHeaders func(req *http.Request)
	rt         http.RoundTripper
}

func (h headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	h.setHeaders(req)
	return h.rt.RoundTrip(req)
}
