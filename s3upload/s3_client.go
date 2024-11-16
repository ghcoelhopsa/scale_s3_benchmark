// s3upload/s3_client.go
package s3upload

import (
    "fmt"
    "net/http" // Added import for net/http
    "time"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"

    "scale_s3_benchmark/config"
)

// InitializeS3Clients initializes and returns a slice of S3 clients based on the provided endpoints.
func InitializeS3Clients(cfg *config.Config) ([]*s3.S3, error) {
    var s3Clients []*s3.S3
    for _, endpoint := range cfg.EndpointURLs {
        sess, err := session.NewSession(&aws.Config{
            Region:           aws.String("us-east-1"), // Consider making region configurable
            Endpoint:         aws.String(endpoint),
            Credentials:      credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
            S3ForcePathStyle: aws.Bool(true),
            HTTPClient: &http.Client{
                Transport: &http.Transport{
                    MaxIdleConns:        cfg.MaxIdleConns,
                    MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
                },
                Timeout: time.Duration(cfg.HttpTimeout) * time.Second,
            },
        })

        if err != nil {
            fmt.Printf("Error creating S3 session for endpoint %s: %v\n", endpoint, err)
            continue
        }

        s3Clients = append(s3Clients, s3.New(sess))
    }

    if len(s3Clients) == 0 {
        return nil, fmt.Errorf("no S3 clients were created. Check endpoints and credentials")
    }

    return s3Clients, nil
}

