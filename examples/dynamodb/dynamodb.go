package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func NewDB() *dynamodb.Client {
	defaultResolver := endpoints.NewDefaultResolver()
	resolver := func(service, region string) (aws.Endpoint, error) {
		if service == dynamodb.EndpointsID {
			return aws.Endpoint{
				URL: "http://localhost:8000/",
			}, nil
		}

		return defaultResolver.ResolveEndpoint(service, region)
	}
	cfg := defaults.Config()
	cfg.Region = endpoints.UsWest2RegionID
	cfg.EndpointResolver = aws.EndpointResolverFunc(resolver)
	return dynamodb.New(cfg)
}

func CreateTables(db *dynamodb.Client) error {
	if db == nil {
		return fmt.Errorf("dynamodb client is nil")
	}

	req := db.CreateTableRequest(&dynamodb.CreateTableInput{
		TableName: aws.String("Test"),
	})
	_, err := req.Send(context.Background())
	if err != nil {
		return err
	}
	return nil
}
