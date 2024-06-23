package main

import (
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpczerolog "github.com/grpc-ecosystem/go-grpc-middleware/providers/zerolog/v2"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/unionj-cloud/go-doudou/v2/framework/grpcx"
	"github.com/unionj-cloud/go-doudou/v2/framework/registry/nacos"
	"github.com/unionj-cloud/go-doudou/v2/framework/rest"
	"github.com/unionj-cloud/go-doudou/v2/framework/restclient"
	logger "github.com/unionj-cloud/go-doudou/v2/toolkit/zlogger"
	"google.golang.org/grpc"
	service "statsvc"
	"statsvc/config"
	"statsvc/internal/nacosservicej"
	pb "statsvc/transport/grpc"
	"statsvc/transport/httpsrv"
)

func main() {
	conf := config.LoadFromEnv()
	svc := service.NewStatsvc(conf,
		nacosservicej.NewEcho(
			restclient.WithRootPath("/nacos-service-j"),
			restclient.WithProvider(nacos.NewWRRServiceProvider("nacos-service-j"))),
	)

	go func() {
		grpcServer := grpcx.NewGrpcServer(
			grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
				grpc_ctxtags.StreamServerInterceptor(),
				grpc_opentracing.StreamServerInterceptor(),
				grpc_prometheus.StreamServerInterceptor,
				tags.StreamServerInterceptor(tags.WithFieldExtractor(tags.CodeGenRequestFieldExtractor)),
				logging.StreamServerInterceptor(grpczerolog.InterceptorLogger(logger.Logger)),
				grpc_recovery.StreamServerInterceptor(),
			)),
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(),
				grpc_opentracing.UnaryServerInterceptor(),
				grpc_prometheus.UnaryServerInterceptor,
				tags.UnaryServerInterceptor(tags.WithFieldExtractor(tags.CodeGenRequestFieldExtractor)),
				logging.UnaryServerInterceptor(grpczerolog.InterceptorLogger(logger.Logger)),
				grpc_recovery.UnaryServerInterceptor(),
			)),
		)
		pb.RegisterStatsvcServiceServer(grpcServer, svc)
		grpcServer.Run()
	}()

	handler := httpsrv.NewStatsvcHandler(svc)
	srv := rest.NewRestServer()
	srv.AddRoute(httpsrv.Routes(handler)...)
	srv.Run()
}
