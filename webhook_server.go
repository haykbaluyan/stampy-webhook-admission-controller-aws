package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

// WebhookServer encapsulates webhook server related fields
type WebhookServer struct {
	admissionController AdmissionControllerInterface

	server *http.Server

	logger *logrus.Logger

	certificateReader CertificateReader
}

// NewWebhookServer is a constructor for WebhookServer
func NewWebhookServer(admissionController AdmissionControllerInterface, logger *logrus.Logger, certificateReader CertificateReader) *WebhookServer {

	srv := &WebhookServer{
		admissionController: admissionController,
		logger:              logger,
		certificateReader:   certificateReader,
	}

	return srv
}

// Start starts webhook server
func (srv *WebhookServer) Start(port int) chan bool {
	isTLS := srv.certificateReader != nil
	serverLogger := srv.logger.
		WithField(portField, port).
		WithField(isTLSField, isTLS)

	serverLogger.Infof("starting webhook server...")

	var tlsConfig *tls.Config

	if isTLS {
		tlsConfig = &tls.Config{
			GetCertificate: srv.certificateReader.GetCertificate,
		}
	}

	srv.server = &http.Server{
		Addr:      fmt.Sprintf(":%v", port),
		TLSConfig: tlsConfig,
	}

	router := mux.NewRouter()
	router.HandleFunc("/ping", srv.handlePing)
	router.HandleFunc("/mutate", srv.handleMutate).Methods("POST")
	srv.server.Handler = router

	// Channel to indicate when the server stopped listening for some reason
	doneListeningChannel := make(chan bool)

	go func() {
		if isTLS {
			if err := srv.server.ListenAndServeTLS("", ""); err != nil {
				serverLogger.WithError(err).Errorf("Failed to listen and serve webhook TLS server.")
			}
		} else {
			if err := srv.server.ListenAndServe(); err != nil {
				serverLogger.WithError(err).Errorf("Failed to listen and serve webhook Plaintext server")
			}
		}
		doneListeningChannel <- true
	}()

	return doneListeningChannel
}

// Stop stop webhook server
func (srv *WebhookServer) Stop() {
	srv.logger.Infof("shutting down webhook server gracefully...")
	srv.server.Shutdown(context.Background())

}

func (srv *WebhookServer) handlePing(w http.ResponseWriter, r *http.Request) {
	httpLogger := srv.httpLogger(r)

	if _, err := fmt.Fprint(w, "pong\n\n"); err != nil {
		httpLogger.Errorf("Unable to write response: %v", err)
		http.Error(w, fmt.Sprintf("Unable to write response: %v", err), http.StatusInternalServerError)
	}

}

func (srv *WebhookServer) handleMutate(w http.ResponseWriter, r *http.Request) {
	handleMutateInternal(srv, w, r)
	return
}

func handleMutateInternal(srv *WebhookServer, w http.ResponseWriter, r *http.Request) {
	httpLogger := srv.httpLogger(r)
	httpLogger.Infof("handleMutate invoked.")

	if r.Method != http.MethodPost {

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requestDump, err := httputil.DumpRequest(r, true)

	httpLogger.
		WithField(requestField, string(requestDump)).
		Debugf("received a request")

	var body []byte

	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		httpLogger.Error("empty request body")
		http.Error(w, "empty request body", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		httpLogger.Error("unsupported media type")
		http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *v1beta1.AdmissionResponse

	admissionReview := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &admissionReview); err != nil {
		httpLogger.Errorf("Unable to decode request body: %v", err)
		http.Error(w, "Unable to decode request body", http.StatusBadRequest)
		return
	}

	// Mutate using the provided controller
	admissionResponse = srv.admissionController.Mutate(&admissionReview)

	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if admissionReview.Request != nil {
			admissionReview.Response.UID = admissionReview.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		httpLogger.Errorf("Unable to encode response: %v", err)
		http.Error(w, fmt.Sprintf("Unable to encode response: %v", err), http.StatusInternalServerError)
	}

	srv.logger.Infof("Responding...")

	if _, err := w.Write(resp); err != nil {
		httpLogger.Errorf("Unable to write response: %v", err)
		http.Error(w, fmt.Sprintf("Unable to write response: %v", err), http.StatusInternalServerError)
	}

	return
}

func (srv *WebhookServer) httpLogger(r *http.Request) *logrus.Entry {
	return srv.logger.
		WithField(remoteAddrField, r.RemoteAddr).
		WithField(methodField, r.Method).
		WithField(urlField, r.URL)
}
