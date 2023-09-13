package config

import "fmt"

// Config is the history stack configuration
type Config struct {
	StreamName string
	AWSRegion  string
	RoleARN    string
}

const (
	defaultRegion = "us-west-2"
)

var (
	stagingConfig = Config{
		AWSRegion:  defaultRegion,
		RoleARN:    "arn:aws:iam::005087123760:role/history-v3-staging-ingest",
		StreamName: "history-v3-staging-stream",
	}

	stagingCanaryConfig = Config{
		AWSRegion:  defaultRegion,
		RoleARN:    "arn:aws:iam::005087123760:role/history-v3-staging-ingest",
		StreamName: "history-v3-staging-stream",
	}

	prodConfig = Config{
		AWSRegion:  defaultRegion,
		RoleARN:    "arn:aws:iam::958416494912:role/history-v3-prod-ingest",
		StreamName: "history-v3-prod-stream",
	}

	prodCanaryConfig = Config{
		AWSRegion:  defaultRegion,
		RoleARN:    "arn:aws:iam::958416494912:role/history-v3-prod-ingest",
		StreamName: "history-v3-prod-stream",
	}
)

// Environment gets a config for an environment
func Environment(environment string) (Config, error) {
	switch environment {
	case "staging":
		return stagingConfig, nil
	case "staging-canary":
		return stagingCanaryConfig, nil
	// TODO: v3 stack is currently considered as canary.
	case "prod-canary":
		return prodCanaryConfig, nil
	case "":
		fallthrough
	case "prod":
		return prodConfig, nil
	}
	return Config{}, fmt.Errorf("invalid history environment: %s", environment)
}
