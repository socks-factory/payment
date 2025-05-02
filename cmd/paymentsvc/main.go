package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"payment"

	"github.com/go-kit/kit/log"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkingo "github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"golang.org/x/net/context"
)

const (
	ServiceName = "payment"
)

func main() {
	var (
		port          = flag.String("port", "8080", "Port to bind HTTP listener")
		zip           = flag.String("zipkin", os.Getenv("ZIPKIN"), "Zipkin address")
		declineAmount = flag.Float64("decline", 105, "Decline payments over certain amount")
	)
	flag.Parse()
	var tracer stdopentracing.Tracer
	{
		// Log domain.
		var logger log.Logger
		{
			logger = log.NewLogfmtLogger(os.Stderr)
			logger = log.With(logger, "ts", log.DefaultTimestampUTC)
			logger = log.With(logger, "caller", log.DefaultCaller)
		}
		// Find service local IP.
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		_ = strings.Split(localAddr.String(), ":")[0]
		defer conn.Close()

		if *zip == "" {
			tracer = stdopentracing.NoopTracer{}
		} else {
			logger := log.With(logger, "tracer", "Zipkin")
			logger.Log("addr", zip)

			conn, err := net.Dial("udp", "8.8.8.8:80")
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
			localAddr := conn.LocalAddr().(*net.UDPAddr)

			reporter := zipkinhttp.NewReporter(*zip)
			defer reporter.Close()
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
			endpoint, err := zipkingo.NewEndpoint("catalogue", localAddr.String())
			if err != nil {
				logger.Log("unable to create local endpoint: %+v\n", err)
				os.Exit(1)
			}

			nativeTracer, err := zipkingo.NewTracer(reporter, zipkingo.WithLocalEndpoint(endpoint))
			if err != nil {
				logger.Log("unable to create tracer: %+v\n", err)
			}
			tracer = zipkin.Wrap(nativeTracer)

		}
		stdopentracing.SetGlobalTracer(tracer)

	}
	// Mechanical stuff.
	errc := make(chan error)
	ctx := context.Background()

	handler, logger := payment.WireUp(ctx, float32(*declineAmount), tracer, ServiceName)

	// Create and launch the HTTP server.
	go func() {
		logger.Log("transport", "HTTP", "port", *port)
		errc <- http.ListenAndServe(":"+*port, handler)
	}()

	// Capture interrupts.
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("exit", <-errc)
}
