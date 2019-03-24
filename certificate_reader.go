package main

import (
	"crypto/tls"

	"github.com/sirupsen/logrus"
)

// CertificateReader interface to read a certificate
type CertificateReader interface {
	GetCertificate(clientHelloInfo *tls.ClientHelloInfo) (*tls.Certificate, error)
}

type certificateFileReader struct {
	CertificateReader
	logger   *logrus.Logger
	certFile string
	keyFile  string
}

// NewCertificateFileReader is a constructor for certificateFileReader
func NewCertificateFileReader(logger *logrus.Logger, certFile string, keyFile string) CertificateReader {
	certReader := &certificateFileReader{
		logger:   logger,
		certFile: certFile,
		keyFile:  keyFile,
	}

	return certReader
}

// GetCertificate loads and returns certificate object
func (cw *certificateFileReader) GetCertificate(clientHelloInfo *tls.ClientHelloInfo) (*tls.Certificate, error) {
	cw.logger.
		WithField(certFileField, cw.certFile).
		WithField(keyFileField, cw.keyFile).
		Tracef("reloading certificates...")

	cert, err := tls.LoadX509KeyPair(cw.certFile, cw.keyFile)

	if err != nil {
		cw.logger.
			WithField(certFileField, cw.certFile).
			WithField(keyFileField, cw.keyFile).
			WithError(err).
			Errorf("certificates reloading failed.")

		return nil, err
	}

	return &cert, nil
}
