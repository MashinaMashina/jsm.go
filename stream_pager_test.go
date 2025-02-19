package jsm_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/jsm.go"
)

func TestPager(t *testing.T) {
	srv, nc, mgr := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	str, err := mgr.NewStream("PAGERTEST", jsm.Subjects("js.in.pager"))
	if err != nil {
		t.Fatalf("stream create failed: %s", err)
	}

	for i := 1; i <= 200; i++ {
		_, err = nc.Request("js.in.pager", []byte(fmt.Sprintf("message %d", i)), time.Second)
		if err != nil {
			t.Fatalf("publish failed: %s", err)
		}
	}

	pgr, err := str.PageContents(jsm.PagerSize(25))
	if err != nil {
		t.Fatalf("pager creation failed: %s", err)
	}

	seen := 0
	pages := 0
	for {
		_, last, err := pgr.NextMsg(context.Background())
		if err != nil && last && seen == 200 && pages == 8 {
			break
		}

		if err != nil {
			t.Fatalf("next failed seen %d pages %d: %s", seen, pages, err)
		}

		seen++
		if last {
			pages++
		}
	}

	err = pgr.Close()
	if err != nil {
		t.Fatalf("close failed")
	}

	known, err := str.ConsumerNames()
	if err != nil {
		t.Fatalf("consumer named failed: %s", err)
	}
	if len(known) != 0 {
		t.Fatalf("expected no consumers got %v", known)
	}
}

func TestPagerWQ(t *testing.T) {
	srv, nc, mgr := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	str, err := mgr.NewStream("PAGERTEST", jsm.Subjects("js.in.pager"), jsm.WorkQueueRetention(), jsm.AllowDirect())
	if err != nil {
		t.Fatalf("stream create failed: %s", err)
	}

	_, err = str.NewConsumer(jsm.ConsumerName("PULL"))
	if err != nil {
		t.Fatalf("consumer create failed: %s", err)
	}

	for i := 1; i <= 200; i++ {
		_, err = nc.Request("js.in.pager", []byte(fmt.Sprintf("message %d", i)), time.Second)
		if err != nil {
			t.Fatalf("publish failed: %s", err)
		}
	}

	_, err = str.PageContents(jsm.PagerSize(25), jsm.PagerStartDelta(time.Hour))
	if err == nil || err.Error() != "workqueue paging does not support time delta starting positions" {
		t.Fatalf("pager creation did not fail for time delta: %v", err)
	}

	pgr, err := str.PageContents(jsm.PagerSize(25))
	if err != nil {
		t.Fatalf("pager creation failed: %s", err)
	}

	seen := 0
	pages := 0
	for {
		_, last, err := pgr.NextMsg(context.Background())
		if err != nil && last && seen == 200 && pages == 8 {
			break
		}

		if err != nil {
			t.Fatalf("next failed seen %d pages %d: %s", seen, pages, err)
		}

		seen++
		if last {
			pages++
		}
	}

	err = pgr.Close()
	if err != nil {
		t.Fatalf("close failed")
	}
}
