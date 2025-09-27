package graceful

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type ctxValueKey string
type ctxTopContextKey string

const cancelFuncKey = ctxValueKey("cancelFunc")
const signalContextKey = ctxTopContextKey("signalContext")

func Context(parentCtx context.Context, shutdownWait time.Duration) (context.Context, context.CancelFunc, func(*http.Server)) {
	signalCtx, stop := signal.NotifyContext(parentCtx, os.Interrupt)
	ctx := context.WithValue(signalCtx, signalContextKey, signalCtx)
	ctx = context.WithValue(ctx, cancelFuncKey, stop)

	waitShutdown := func(server *http.Server) {
		<-ctx.Done()
		timeoutCtx, cancel := context.WithTimeout(context.Background(), shutdownWait)
		defer cancel()

		log.Println("T26: Gracefull Shutdown: started")
		err := server.Shutdown(timeoutCtx)
		log.Println("T27: Gracefull Shutdowned:", err)

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
			log.Println("T47: TopContext")
			return topCtx
		}
	}
	return context.Background()
}
