package main

import (
	"flag"
	"net/http"

	"k8s.io/klog/v2"

	"github.com/lingdie/image-policy-webhook/pkg/server"
)

var certFile, keyFile string
var debug bool

func main() {
	flag.StringVar(&certFile, "tls-cert-file", "/etc/webhook/certs/tls.crt", "File containing the x509 certificate.")
	flag.StringVar(&keyFile, "tls-key-file", "/etc/webhook/certs/tls.key", "File containing the x509 private key.")
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.Parse()

	whsvr := &server.WebhookServer{
		Server: &http.Server{
			Addr: ":8443",
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/validate", whsvr.Serve)
	whsvr.Server.Handler = mux

	if debug {
		klog.Info("debug mode, listen on :8443")
		klog.Fatal(whsvr.Server.ListenAndServe())
	} else {
		klog.Info("listen on :8443 with TLS")
		klog.Fatal(whsvr.Server.ListenAndServeTLS(certFile, keyFile))
	}
}
