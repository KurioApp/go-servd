package servd_test

import (
	"context"
	"testing"
	"time"

	servd "github.com/KurioApp/go-servd"
)

func TestUsage(t *testing.T) {

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
			t.Log("Run err:", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	status, err := d.WaitForStatus(ctx, servd.Running)
	if err != nil {
		t.Fatal("err:", err)
	}

	if got, want := status, servd.Running; got != want {
		t.Error("got", got, "want:", want)
	}

	// Stop the service
	d.Stop()

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	status, err = d.WaitForStatus(ctx, servd.Stopped)
	if err != nil {
		t.Fatal("err:", err)
	}

	if got, want := status, servd.Stopped; got != want {
		t.Error("got", got, "want:", want)
	}
}
