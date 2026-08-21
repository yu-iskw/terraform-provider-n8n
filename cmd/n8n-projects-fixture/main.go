package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/yu-iskw/terraform-provider-n8n/internal/n8n/apitest"
)

func main() {
	listen := flag.String("listen", ":5678", "listen address")
	check := flag.Bool("check", false, "GET /healthz/readiness on listen address and exit")
	flag.Parse()

	apiKey := os.Getenv("N8N_API_KEY")
	if apiKey == "" {
		apiKey = apitest.DefaultAPIKey
	}

	if *check {
		if err := healthcheck(*listen); err != nil {
			log.Fatal(err)
		}
		return
	}

	server := &http.Server{
		Addr:              *listen,
		Handler:           apitest.NewProjectsServer(apiKey).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("n8n projects API fixture listening on %s", *listen)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func healthcheck(listen string) error {
	host := listen
	if len(host) > 0 && host[0] == ':' {
		host = "127.0.0.1" + host
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + host + "/healthz/readiness")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("readiness status %d", resp.StatusCode)
	}
	return nil
}
