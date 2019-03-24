package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"k8s.io/kubernetes/pkg/util/file"
)

// Config encapsulates configurations related to the webhook
type Config struct {
	cert     string
	key      string
	logLevel logrus.Level
	port     int
	region   string
	bucket   string
}

func readConfig() (*Config, error) {
	f := flag.NewFlagSet("", flag.ExitOnError)
	port := f.Int("port", 443, "Webhook server port.")
	logLevelStr := f.String("log-level", "debug", "Logging level.")
	tlsPairName := f.String("tlsPairName", "tls", "certificate and key pair name")
	tlsCertDir := f.String("tlsCertdir", "/var/run/stampy-webhook-admission-controller/certs", "certificate and key directory")
	region := f.String("region", "", "AWS region that stores signature files.")
	bucket := f.String("bucket", "", "AWS S3 bucket that stores signature files.")
	f.Parse(os.Args[1:])

	certPath := path.Join(*tlsCertDir, *tlsPairName+".crt")
	keyPath := path.Join(*tlsCertDir, *tlsPairName+".key")
	if certPath != ".crt" {
		if exists, _ := file.FileExists(certPath); !exists {
			return nil, fmt.Errorf("unable to find certificate file - %s", certPath)
		}
	}

	if keyPath != ".key" {
		if exists, _ := file.FileExists(keyPath); !exists {
			return nil, fmt.Errorf("unable to find key file - %s", keyPath)
		}
	}

	if *region == "" {
		return nil, fmt.Errorf("invalid region: empty")
	}

	if *bucket == "" {
		return nil, fmt.Errorf("invalid bucket: empty")
	}

	logLevel, err := logrus.ParseLevel(*logLevelStr)
	if err != nil {
		return nil, fmt.Errorf("invalid log level")
	}

	return &Config{
		port:     *port,
		cert:     certPath,
		key:      keyPath,
		region:   *region,
		bucket:   *bucket,
		logLevel: logLevel,
	}, nil
}
