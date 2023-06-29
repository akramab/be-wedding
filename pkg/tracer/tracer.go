package tracer

import (
	"context"
	"fmt"
	"be-wedding/pkg/appinfo"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type Config struct {
	Enable           bool    `toml:"enable"`
	Exporter         string  `toml:"exporter"`
	ExporterEndpoint string  `toml:"exporter_endpoint"`
	Sampling         string  `toml:"sampling"`
	SamplingRatio    float64 `toml:"sampling_ratio"`
}

func (c Config) spanExporter() (sdktrace.SpanExporter, error) {
	var t sdktrace.SpanExporter
	var err error
	switch strings.ToLower(c.Exporter) {
	case "otlp-grpc":
		t, err = otlptracegrpc.New(context.Background(),
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(c.ExporterEndpoint),
			otlptracegrpc.WithDialOption(grpc.WithBlock()),
		)
	case "otlp-http":
		t, err = otlptracehttp.New(context.Background(),
			otlptracehttp.WithInsecure(),
			otlptracehttp.WithEndpoint(c.ExporterEndpoint),
		)
	default:
		return nil, fmt.Errorf("tracer: unsupported exporter: %s", c.Exporter)
	}
	return t, err
}

func (c Config) sampler() sdktrace.Sampler {
	var sampler sdktrace.Sampler
	switch strings.ToLower(c.Sampling) {
	case "off":
		sampler = sdktrace.NeverSample()
	case "always":
		sampler = sdktrace.AlwaysSample()
	case "ratio_based":
		samplingRatio := c.SamplingRatio
		if samplingRatio < 0 || samplingRatio > 1 {
			samplingRatio = 0.5
		}
		sampler = sdktrace.TraceIDRatioBased(samplingRatio)
	default:
		sampler = sdktrace.NeverSample()
	}
	return sdktrace.ParentBased(sampler)
}

// SetTracer set global open telemetry provider. This method should be called as early as possible in each server.
func SetTracer(cfg Config, appInfo appinfo.Info) error {
	if !cfg.Enable {
		otel.SetTracerProvider(trace.NewNoopTracerProvider())
		return nil
	}
	res, resErr := resource.New(context.Background(),
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceName(appInfo.NameWithEnv()),
			semconv.ServiceVersion(appInfo.GitTag),
			semconv.DeploymentEnvironment(appInfo.Env),
			semconv.TelemetrySDKLanguageGo,
			attribute.String("git.commit.sha", appInfo.GitCommitHash),
			attribute.String("git.repository_url", appInfo.GitURL),
		),
	)
	if resErr != nil {
		return resErr
	}
	spanExporter, err := cfg.spanExporter()
	if err != nil {
		return err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(spanExporter),
		sdktrace.WithSampler(cfg.sampler()),
		sdktrace.WithResource(res),
	)
	tp.Tracer(appInfo.Name)
	textMapPropagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(textMapPropagator)
	otel.SetTracerProvider(tp)
	return nil
}
