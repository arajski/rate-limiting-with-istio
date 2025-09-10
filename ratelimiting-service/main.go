package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/arajski/rate-limiting-with-istio/ratelimiting-service/db"
	"github.com/arajski/rate-limiting-with-istio/ratelimiting-service/pkg/handlers"
	"github.com/arajski/rate-limiting-with-istio/ratelimiting-service/pkg/repository"
	"github.com/arajski/rate-limiting-with-istio/ratelimiting-service/pkg/service"
	"github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/test/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	grpcKeepaliveTime        = 30 * time.Second
	grpcKeepaliveTimeout     = 5 * time.Second
	grpcKeepaliveMinTime     = 30 * time.Second
	grpcMaxConcurrentStreams = 1000000
)

func main() {
	port := 18000
	db.InitDB()
	db.SeedDB()

	snapshotCache := cache.NewSnapshotCache(false, cache.IDHash{}, nil)

	usersRepo := repository.Users{
		DB: db.DB,
	}

	rateLimitsRepo := repository.RateLimits{
		DB: db.DB,
	}

	rateLimitService := service.RateLimit{
		UsersRepository:      usersRepo,
		RateLimitsRepository: rateLimitsRepo,
		Cache:                snapshotCache,
	}

	httpHandlers := handlers.HTTPHandlers{
		RateLimitService: rateLimitService,
	}

	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("GET /", httpHandlers.Users)
	mux.HandleFunc("GET /ratelimits", httpHandlers.RateLimits)
	mux.HandleFunc("GET /ratelimits/new", httpHandlers.NewRateLimit)
	mux.HandleFunc("POST /ratelimits/create", httpHandlers.CreateRateLimit)
	mux.HandleFunc("GET /ratelimits/edit", httpHandlers.EditRateLimit)
	mux.HandleFunc("POST /ratelimits/edit", httpHandlers.UpdateRateLimit)
	mux.HandleFunc("GET /ratelimits/delete", httpHandlers.DeleteRateLimit)

	go func() {
		log.Fatal(http.ListenAndServe(":8080", mux))
	}()

	xDSServer := server.NewServer(context.Background(), snapshotCache, &test.Callbacks{Debug: false})

	grpcServer := grpc.NewServer(
		grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    grpcKeepaliveTime,
			Timeout: grpcKeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             grpcKeepaliveMinTime,
			PermitWithoutStream: true,
		}),
	)

	discoveryv3.RegisterAggregatedDiscoveryServiceServer(grpcServer, xDSServer)

	log.Printf("tcp server listening on %d\n", port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Printf("could not start a tcp server: %+v", err)
		os.Exit(1)
	}
	log.Printf("serving grpc...")
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Printf("could not start a GRPC server: %+v", err)
		os.Exit(1)
	}
}
