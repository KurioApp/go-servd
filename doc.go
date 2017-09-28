/*
Package servd provides standard API for service like daemon.

Standard way to run:

	go func() {
		if err := d.Run(); err != nil {
			// fail to run
		}
	}

Standard way to stop:

	if ok := d.Stop(); !ok {
		// already been stopped
	}

Standard way to wait for service to be shutdown gracefully:

	// wait until reach Stopped state
	s := d.WaitForStatus(servd.Stopped)

Implementation Example

What we need is to implement servd.Handler. For shortcut we can use servd.HandleFunc.
Servd will pass cancellable context.Context, listen to the ctx.Done() channel as shutdown signal.

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


*/
package servd
