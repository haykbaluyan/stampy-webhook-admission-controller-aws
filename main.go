package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

const (
	successExitCode = 0
	errorExitCode   = 1
)

func main() {
	logger := logrus.New()
	config, err := readConfig()
	if err != nil {
		logger.Errorf("api=main, reason=readConfig, err=%v", err)
		os.Exit(errorExitCode)
	}
	logger = initializeLogger(logger, config.logLevel)

	var certificateReader CertificateReader
	if (config.cert != "") && (config.key != "") {
		logger.
			WithField(certFileField, config.cert).
			WithField(keyFileField, config.key).
			Info("Configuring certificate reader to use with the server")
		certificateReader = NewCertificateFileReader(logger, config.cert, config.key)
	} else {
		logger.Errorf("api=main, reason='certificate files were not provided'")
		os.Exit(errorExitCode)
	}

	admissionController, err := NewAdmissionController(config.region, config.bucket, logger)
	webhookServer := NewWebhookServer(admissionController, logger, certificateReader)

	doneListeningChannel := webhookServer.Start(config.port)

	// listening OS shutdown signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	stopped := false
	// Wait until we receive either a termination signal, or the server stops by itself for some reason
	select {
	case signal := <-signalChan:
		{
			logger.Infof("Received a termination signal. SIG=%s", signal)
		}
	case stopped = <-doneListeningChannel:
		{
			logger.Warn("Server has stopped on it's own... exiting.")
		}
	}

	if !stopped {
		webhookServer.Stop()
	}

	logger.Info("Webhook server exited successfully.")
	os.Exit(successExitCode)
}

func initializeLogger(logger *logrus.Logger, level logrus.Level) *logrus.Logger {
	logger.SetLevel(level)

	// By default StdErr is used. Let's change that to StdOut
	logger.SetOutput(os.Stdout)

	logger.SetFormatter(
		&logrus.JSONFormatter{
			TimestampFormat: logDateFormat,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  timeField,
				logrus.FieldKeyLevel: levelField,
				logrus.FieldKeyMsg:   messageField,
			},
		},
	)

	return logger
}
