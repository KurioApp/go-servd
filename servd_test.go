package servd_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	servd "github.com/KurioApp/go-servd"
)

func TestServd(t *testing.T) {
	// Construct the service
	d := &servd.Servd{
		Handler: servd.HandleFunc(func(ctx context.Context) error {
			// wait until shutdown initated
			<-ctx.Done()
			return nil
		}),
	}

	d.Stop()
	if got, want := d.Status(), servd.Stopped; got != want {
		t.Fatal("got:", got, "want:", want)
	}

	err := d.Run()
	if err == nil {
		t.Fatal("should cannot be run")
	}
}

func ExampleServd() {
	// Construct the service
	d := &servd.Servd{
		Handler: servd.HandleFunc(func(ctx context.Context) error {
			// wait until shutdown initated
			<-ctx.Done()
			return nil
		}),
	}

	// Run the service
	go func() {
		if err := d.Run(); err != nil {
			// TODO: handle error
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	status, err := d.WaitForStatus(ctx, servd.Running)
	if err != nil {
		// TODO: handle timeout, wait for status takes more than 1sec
		return
	}

	fmt.Println("Status:", status)

	// Stop the service
	d.Stop()

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	status, err = d.WaitForStatus(ctx, servd.Stopped)
	if err != nil {
		// TODO: handle timeout, wait for status takes more than 1sec
		return
	}

	fmt.Println("Status:", status)
	// output:
	// Status: Running
	// Status: Stopped
}
