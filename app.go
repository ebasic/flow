package flow

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sedind/flow/auth/jwtauth"

	"github.com/sedind/flow/dotenv"

	"github.com/pkg/errors"
	"github.com/sedind/flow/config"
	"github.com/sedind/flow/dbe"
	"github.com/sedind/flow/logger"
)

// App is where everything is connected
type App struct {
	Context Context
	Router  http.Handler
}

// New creates instance of application Context
func New(configFile string) *App {
	//load .env file
	dotenv.Load()

	//load application config
	appConfig := Config{}

	err := config.LoadFromPath(configFile, &appConfig)
	if err != nil {
		panic(err)
	}

	// initialize logger
	appLogger := logger.New(appConfig.LogLevel)

	// set database log level
	dbe.LogLevel(appConfig.LogLevel)

	//create application DB connections
	connections := map[string]*dbe.Connection{}
	for k, d := range appConfig.ConnectionStrings {
		c, err := dbe.NewConnection(*d)
		if err != nil {
			appLogger.Panic(err)
		}
		err = c.Open()
		if err != nil {
			appLogger.Error(errors.Wrapf(err, "Unable to connect to %s connection", k))
		}
		connections[k] = c
	}

	var auth *jwtauth.JWTAuth

	if appConfig.JWTAuth {
		secret, ok := appConfig.AppSettings["jwt_secret"]
		if !ok {
			panic(errors.New("jwt_secret key not provided in app_settings"))
		}

		alg, ok := appConfig.AppSettings["jwt_algorithm"]
		if !ok {
			appLogger.Warn("jwt_algorithm key not provided in app_settings, HS256 algorithm will be used")
			alg = "HS256"
		}

		auth = jwtauth.New(alg, []byte(secret), nil)

	}

	// create application context object
	ctx := Context{
		Config:        appConfig,
		DBConnections: connections,
		Logger:        appLogger,
		jwtauth:       auth,
	}

	return &App{
		Context: ctx,
	}
}

// RegisterRouter register application router
func (a *App) RegisterRouter(fn func(ctx *Context) http.Handler) {
	a.Router = fn(&a.Context)
}

// Serve the application at the specified address/port and listen for OS
// interrupt and kill signals and will attempt to stop the application
// gracefully.
func (a *App) Serve() error {
	a.Context.Logger.Infof("Starting Application at %s", a.Context.Addr)
	if a.Router == nil {
		return errors.New("Application Router not initialized")
	}
	server := http.Server{
		Handler: a.Router,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, os.Interrupt)

	go func() {
		select {
		case <-c:
			a.Stop(nil)
			signal.Stop(c)
			os.Exit(1)
		}
	}()

	var err error
	if strings.HasPrefix(a.Context.Addr, "unix:") {
		lis, err := net.Listen("unix", a.Context.Addr[5:])
		if err != nil {
			return a.Stop(err)
		}

		err = server.Serve(lis)
	} else {
		server.Addr = a.Context.Addr
		err = server.ListenAndServe()
	}

	if err != nil {
		return a.Stop(err)
	}
	return nil
}

// Stop the application
func (a *App) Stop(err error) error {
	a.Context.Logger.Info("Stopping application...")
	if err != nil && errors.Cause(err) != context.Canceled {
		return err
	}
	return nil
}
