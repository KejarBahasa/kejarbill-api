package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KejarBahasa/kejarbill-api/internal/bootstrap"
)

func main() {
	// DEPENDENCY
	dep, err := bootstrap.BuildDependency()
	if err != nil {
		log.Fatal(err)
	}

	// APP
	app := bootstrap.BuildApp(dep)

	// START SERVER
	go func() {
		log.Printf("server running on port %s", dep.Config.App.Port)
		if err := app.Listen(":" + dep.Config.App.Port); err != nil {
			log.Fatal(err)
		}
	}()

	// GRACEFUL SHUTDOWN
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Fiber Shutdown
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatal(err)
	}

	// DB Close
	dep.DB.Close()

	// Redis Close
	if err := dep.Redis.Close(); err != nil {
		log.Println(err)
	}

	log.Println("server exited properly")
}
