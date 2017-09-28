package servd_test

import (
	"context"
	"fmt"
	"net/http"

	servd "github.com/KurioApp/go-servd"
)

func ExampleHandleFunc() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi there!")
	})
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	handler := servd.HandleFunc(func(ctx context.Context) error {
		go func() {
			// wait until stop signal received
			<-ctx.Done()
			if err := server.Shutdown(context.Background()); err != nil {
				// fail to shutdown
			}
		}()

		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	serv := &servd.Servd{
		Handler: handler,
	}
	_ = serv // TODO: now we can use serv
}
