package suite

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"path"
	"testing"

	"github.com/Raimguhinov/dav-go/internal/config"
	"github.com/ceres919/go-webdav"
	"github.com/ceres919/go-webdav/caldav"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Suite struct {
	*testing.T
	Cfg              *config.Config
	Client           *caldav.Client
	TestCalendarData map[string]any
}

func New(t *testing.T, withAuth bool) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.GetConfig()
	ctx, cancel := context.WithCancel(context.Background()) //, cfg.HTTP.Timeout)

	// Test Calendar
	testCalendarData := map[string]any{
		"name":        fmt.Sprintf("Private Calendar - UID:%s for (%s)", uuid.New(), t.Name()),
		"description": base64.RawStdEncoding.EncodeToString([]byte(cfg.PG.URL + cfg.HTTP.Password)),
		"types":       []string{"VEVENT", "VTODO", "VJOURNAL"},
		"max_size":    7777,
	}

	t.Cleanup(func() {
		t.Helper()
		cancel()

		conn, err := pgx.Connect(context.Background(), cfg.PG.URL)
		if err != nil {
			t.Fatal(err)
		}
		_, err = conn.Exec(context.Background(), `
			INSERT INTO caldav.test_cases (calendar_folder_id, calendar_file_uid, event_component_id, calendar_property_id,
			                               custom_property_id, recurrence_id, recurrence_exception_id)
			SELECT cfo.id, cf.uid, ec.id, cp.id, cup.id, rc.id, re.id
			FROM caldav.calendar_folder cfo
			         JOIN caldav.calendar_file cf ON cfo.id = cf.calendar_folder_id
			         LEFT JOIN caldav.event_component ec ON cf.uid = ec.calendar_file_uid
			         JOIN caldav.calendar_property cp on cf.uid = cp.calendar_file_uid
			         LEFT JOIN caldav.custom_property cup on cf.uid = cup.calendar_file_uid
			         LEFT JOIN caldav.recurrence rc on ec.id = rc.event_component_id
			         LEFT JOIN caldav.recurrence_exception re on rc.id = re.recurrence_id
			WHERE cfo.name = $1
			  AND cfo.description = $2
			  AND cfo.types = $3 
			  AND cfo.max_size = $4
		`, testCalendarData["name"], testCalendarData["description"],
			testCalendarData["types"], testCalendarData["max_size"],
		)
		if err != nil {
			t.Fatal(err)
		}
	})

	httpClient := webdav.HTTPClient(&http.Client{})
	if withAuth {
		httpClient = webdav.HTTPClientWithBasicAuth(&http.Client{}, cfg.HTTP.User, cfg.HTTP.Password)
	}
	client, err := caldav.NewClient(httpClient, fmt.Sprintf("http://%s:%s", cfg.HTTP.IP, cfg.HTTP.Port))
	if err != nil {
		t.Fatal(err)
	}

	return ctx, &Suite{
		T:                t,
		Cfg:              cfg,
		Client:           client,
		TestCalendarData: testCalendarData,
	}
}

func (s *Suite) CreateTestCalendar(ctx context.Context, calendarHomeSet string) {
	err := s.Client.Mkdir(ctx, path.Join(calendarHomeSet, s.TestCalendarData["name"].(string)))
	if err != nil {
		s.Fatal(err)
	}
}
