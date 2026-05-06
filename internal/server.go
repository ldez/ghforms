package internal

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ldez/githubformpreview/internal/form"
	"github.com/ldez/githubformpreview/internal/render"
	"github.com/ldez/githubformpreview/internal/store"
	"github.com/ldez/githubformpreview/internal/watcher"
)

// Run starts the HTTP server.
func Run(addr, dir string) error {
	loader, err := form.New()
	if err != nil {
		return err
	}

	st := store.New(dir, loader)

	logSnapshot(st.Get(), dir)

	mux := http.NewServeMux()

	staticSubtree, err := fs.Sub(render.StaticFS, "static")
	if err != nil {
		return err
	}

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSubtree))))

	r, err := render.New(st)
	if err != nil {
		return err
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}

		errI := r.Index(w)
		if errI != nil {
			http.Error(w, errI.Error(), http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/forms/", func(w http.ResponseWriter, req *http.Request) {
		slug := strings.TrimPrefix(req.URL.Path, "/forms/")

		slug = strings.Trim(slug, "/")
		if slug == "" {
			http.Redirect(w, req, "/", http.StatusFound)
			return
		}

		errF := r.Form(w, slug)
		if errF != nil {
			http.Error(w, errF.Error(), http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/__events", sseHandler(st))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err = watcher.Watch(ctx, dir, func() {
		slog.InfoContext(ctx, "change detected: reloading",
			slog.String("actor", "watcher"),
			slog.String("dir", dir),
		)

		st.Reload()

		logSnapshot(st.Get(), dir)
	})
	if err != nil {
		return fmt.Errorf("start watcher: %w", err)
	}

	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer shutdownCancel()

		_ = srv.Shutdown(shutdownCtx)
	}()

	slog.InfoContext(ctx, "Listening (live-reload enabled)",
		slog.String("addr", fmt.Sprintf("http://localhost%s", addr)),
	)

	err = srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func logSnapshot(snap store.Snapshot, dir string) {
	if snap.Err != nil {
		slog.Error("Loading error",
			slog.String("dir", dir),
			slog.Any("error", snap.Err),
		)

		return
	}

	slog.Debug("Forms loaded",
		slog.Int("count", len(snap.Forms)),
		slog.String("dir", dir),
	)

	for _, f := range snap.Forms {
		slog.Debug(f.Name,
			slog.String("filename", f.Filename),
			slog.String("slug", f.Slug),
		)
	}
}

// sseHandler streams reload events to the browser.
// It sends an initial `hello` event when the connection is established
// and a `reload` event on every store change.
func sseHandler(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		ch, cancel := st.Subscribe()
		defer cancel()

		_, _ = fmt.Fprintf(w, "event: hello\ndata: connected\n\n")

		flusher.Flush()

		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-req.Context().Done():
				return

			case _, ok := <-ch:
				if !ok {
					return
				}

				_, _ = fmt.Fprintf(w, "event: reload\ndata: %d\n\n", st.Get().Version)

				flusher.Flush()

			case <-ticker.C:
				// keep-alive comment so proxies don't drop the connection.
				_, _ = fmt.Fprintf(w, ": ping\n\n")

				flusher.Flush()
			}
		}
	}
}
