package main

import (
	"context"
	"fmt"

	corelambda "github.com/jtfm/go-cdk-core/pkg/lambda"
	"github.com/rs/zerolog/log"
)

type CustomEvent struct {
	Name string `json:"name"`
}

func main() {
	err := corelambda.RunHandler(handler, CustomEvent{Name: "Test"})
	if err != nil {
		panic(err)
	}
}

func handler(ctx context.Context, event interface{}) (interface{}, error) {
	log.Info().Msg("Running handler...")
	// Ensure event is of type CustomEvent
	customEvent, ok := event.(CustomEvent)
	if !ok {
		return nil, fmt.Errorf("event is not of type CustomEvent")
	}

	log.Info().Msgf("Hello, %s!", customEvent.Name)

	return nil, nil
}
