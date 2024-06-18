package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/api/codes"
	"git.rpjosh.de/RPJosh/workout/internal/api/dashboard"
	"git.rpjosh.de/RPJosh/workout/internal/api/kubernetes"
	"git.rpjosh.de/RPJosh/workout/internal/api/middleware"
	rpRouter "git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/api/statistics"
	"git.rpjosh.de/RPJosh/workout/internal/api/user"
	"git.rpjosh.de/RPJosh/workout/internal/api/workout"
	"git.rpjosh.de/RPJosh/workout/internal/database"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
	"git.rpjosh.de/RPJosh/workout/pkg/response"

	"github.com/go-chi/chi/v5"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

// Api contains dependencies of the programm
// that are needed from the API endpoints
type Api struct {
	Config *models.AppConfig
	dev    devApi
}

// devApi contains dependencies for spanning up the dev endpoint
type devApi struct {
	closed      atomic.Bool
	connections map[*websocket.Conn]int
	mtx         sync.Mutex
}

// SetupServer mounts all routes of this application
// into the given router of chi
func (api *Api) SetupServer(router *chi.Mux) {

	// Set global variables we need to access across the whole application.
	// In the future we could add a router config which would return these global objects
	rpRouter.GlobalTranslator = translator.NewTranslator()
	rpRouter.GlobalConfig = api.Config
	rpRouter.GlobalDb = api.GetDb()

	// Global function to check if username / password is correct.
	// We cannot reference the user package from package [middleware] because
	// of an import cycle
	userRequest := rpRouter.NewApiRequestWithValues(rpRouter.Route{}, database.NewDatabaseUtils(rpRouter.GlobalDb), logger.GetGlobalLogger(), "", models.WebUser{}, *rpRouter.GlobalTranslator)
	userApi := user.Api{ApiRequest: userRequest}
	middleware.GlobalIsLoginCorrect = userApi.IsLoginCorrect

	// Add 404 custom response
	router.Mount("/notRelevant!", codes.GetRoutes().GetHandlerWithRouter(router))

	// Mount all routes
	router.Mount("/", api.configureRoutes())

	// Mount dev endpoints
	if api.Config.DevMode {
		api.dev.connections = make(map[*websocket.Conn]int)
		router.Mount("/dev", api.addHotReload())
	}

	// Mount kubernetes endpoints
	router.Mount("/kube", kubernetes.GetRoutes().GetHandler())
}

// configureRoutes configures all routes
func (api *Api) configureRoutes() http.Handler {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Mount("/kube", kubernetes.GetRoutes().GetHandler())
		r.Mount("/user", user.GetRoutes(api.Config).GetHandler())
		r.Mount("/", dashboard.GetRoutes().GetHandler())
		r.Mount("/dashboard", dashboard.GetRoutes().GetHandler())
		r.Mount("/statistic", statistics.GetRoutes().GetHandler())
		r.Mount("/workout", workout.GetRoutes(database.NewDatabaseUtils(api.GetDb())).GetHandler())
	})

	return r
}

func (api *Api) addHotReload() http.Handler {
	r := chi.NewRouter()

	// Close all connections before leaving this application
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-signalChannel
		api.dev.closed.Store(true)

		// Close all connections
		if sig != syscall.SIGABRT {
			api.dev.mtx.Lock()
			for con := range api.dev.connections {
				con.Close()
			}
			api.dev.mtx.Unlock()
		}

		os.Exit(0)
	}()

	// WebSocket endpoint
	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		if api.dev.closed.Load() {
			response.WriteText("Gone", 410, w)
			return
		}

		// Upgrade connection
		upg := websocket.NewUpgrader()
		upg.KeepaliveTime = 30 * time.Minute
		upg.CheckOrigin = func(r *http.Request) bool {
			return true
		}
		conn, err := upg.Upgrade(w, r, nil)
		api.dev.mtx.Lock()
		api.dev.connections[conn] = 0
		api.dev.mtx.Unlock()

		// Handler
		if err != nil {
			logger.Warning("Cannot upgrade to ws: %s", err)
		} else {
			conn.OnClose(func(*websocket.Conn, error) {
				logger.Debug("Closed ws connection in dev mode")

				// Remove connection
				api.dev.mtx.Lock()
				delete(api.dev.connections, conn)
				api.dev.mtx.Unlock()
			})
		}
	})

	return r
}

// GetDb returns a DB connection to the configured database.
// This function does panic if the connection failed
func (api *Api) GetDb() *sql.DB {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", api.Config.Db.User, api.Config.Db.Password, api.Config.Db.Address, api.Config.Db.Db))
	if err != nil {
		logger.Fatal("Failed to open DB connection: %s", err)
	}

	// Set performance setttings
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(6)
	db.SetMaxIdleConns(6)

	return db
}
