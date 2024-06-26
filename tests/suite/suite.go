package suite

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/Raimguhinov/dav-go/internal/config"
	"github.com/Raimguhinov/dav-go/pkg/postgres"
	"github.com/ceres919/go-webdav"
	"github.com/ceres919/go-webdav/caldav"
	"github.com/google/uuid"
)

type Suite struct {
	*testing.T
	Cfg        *config.Config
	Client     *caldav.Client
	TestFolder map[string]string
}

var oncePg *postgres.Postgres
var once sync.Once

func getPg(t *testing.T, cfg *config.Config) *postgres.Postgres {
	once.Do(func() {
		var err error
		oncePg, err = postgres.New(context.Background(), nil, cfg.PG.URL,
			postgres.ConnTimeout(cfg.HTTP.IdleTimout), postgres.MaxPoolSize(1000),
		)
		if err != nil {
			t.Fatal(err)
		}
	})
	return oncePg
}

func New(t *testing.T, withAuth bool) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.GetConfig()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.Timeout)

	// Test Calendar
	testFolder := map[string]string{
		"name":        fmt.Sprintf("Private Calendar - UID:%s for (%s)", uuid.New(), t.Name()),
		"description": "Protei Calendar",
	}

	pg := getPg(t, cfg)

	_, err := pg.Pool.Exec(context.Background(), `INSERT INTO caldav.calendar_folder (name) VALUES ($1)`,
		testFolder["name"])
	if err != nil {
		t.Fatal(err)
	}

	httpClient := webdav.HTTPClient(&http.Client{})
	if withAuth {
		httpClient = webdav.HTTPClientWithBasicAuth(&http.Client{}, cfg.HTTP.User, cfg.HTTP.Password)
	}
	client, err := caldav.NewClient(httpClient, fmt.Sprintf("http://%s:%s", cfg.HTTP.IP, cfg.HTTP.Port))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		t.Helper()
		cancel()

		_ = pg
		_, err := pg.Pool.Exec(context.Background(), `DELETE FROM caldav.calendar_folder WHERE name = $1`,
			testFolder["name"])
		if err != nil {
			t.Fatal(err)
		}
	})

	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		Client:     client,
		TestFolder: testFolder,
	}
}

func (s *Suite) CreateTestCalendar(ctx context.Context, calendarHomeSet string) {
	var reportCalendarData = fmt.Sprintf(`
		<?xml version='1.0' encoding='UTF-8' ?>
		<A:mkcol xmlns:A="DAV:" xmlns:B="urn:ietf:params:xml:ns:caldav">
		  <A:set>
		    <A:prop>
		      <A:resourcetype>
		        <A:collection />
		        <B:calendar />
		      </A:resourcetype>
		      <A:displayname>%s</A:displayname>
		      <A:calendar-description>%s</A:calendar-description>
		    </A:prop>
		  </A:set>
		</A:mkcol>
		`, s.TestFolder["name"], s.TestFolder["description"],
	)

	client := &http.Client{}
	req, err := http.NewRequest("MKCOL", fmt.Sprintf(
		"http://%s:%s%s%d/", s.Cfg.HTTP.IP, s.Cfg.HTTP.Port, calendarHomeSet, 777),
		strings.NewReader(reportCalendarData),
	)
	if err != nil {
		s.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/xml")
	req.SetBasicAuth(s.Cfg.HTTP.User, s.Cfg.HTTP.Password)

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		s.Fatal(err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			s.Fatal(err)
		}
	}()
}
