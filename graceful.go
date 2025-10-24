package graceful

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type ctxValueKey string
type ctxTopContextKey string

const cancelFuncKey = ctxValueKey("cancelFunc")
const signalContextKey = ctxTopContextKey("signalContext")

var logger *slog.Logger 

func init() {
	logger = slog.New(slog.DiscardHandler)
}

// SetLogger sets logger
func SetLogger(l *slog.Logger) {
	logger = l
}

func Context(parentCtx context.Context, shutdownWait time.Duration) (context.Context, context.CancelFunc, func(*http.Server)) {
	signalCtx, stop := signal.NotifyContext(parentCtx, os.Interrupt)
	ctx := context.WithValue(signalCtx, signalContextKey, signalCtx)
	ctx = context.WithValue(ctx, cancelFuncKey, stop)

	waitShutdown := func(server *http.Server) {
		<-ctx.Done()
		timeoutCtx, cancel := context.WithTimeout(context.Background(), shutdownWait)
		defer cancel()

		logger.Info("graceful,T26: Gracefull Shutdown: started")
		err := server.Shutdown(timeoutCtx)
		logger.Info("graceful,T27: Gracefull Shutdowned:", slog.Any("error", err))

	}
	return ctx, stop, waitShutdown
}

func Cancel(ctx context.Context) {
	if f := ctx.Value(cancelFuncKey); f != nil {
		if cancel, ok := f.(context.CancelFunc); ok {
			cancel()
		}
	}
}

func TopContext(ctx context.Context) context.Context {
	if c := ctx.Value(signalContextKey); c != nil {
		if topCtx, ok := c.(context.Context); ok {
			logger.Info("graceful,T47: TopContext")
			return topCtx
		}
	}
	return context.Background()
}
