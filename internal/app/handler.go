package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Raimguzhinov/dav-go/internal/auth"
	"github.com/ceres919/go-webdav"
	"github.com/ceres919/go-webdav/caldav"
	"github.com/ceres919/go-webdav/carddav"
)

type userPrincipalBackend struct{}

func (u *userPrincipalBackend) CurrentUserPrincipal(ctx context.Context) (string, error) {
	authCtx, ok := auth.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("unauthenticated requests are not supported")
	}
	if authCtx == nil {
		return "", fmt.Errorf("unauthenticated requests are not supported")
	}

	userDir := authCtx.UserName
	//userDir := base64.RawStdEncoding.EncodeToString([]byte(authCtx.UserName))
	return "/" + userDir + "/", nil
}

type davHandler struct {
	upBackend      webdav.UserPrincipalBackend
	authBackend    auth.AuthProvider
	caldavBackend  caldav.Backend
	carddavBackend carddav.Backend
}

func (d *davHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userPrincipalPath, err := d.upBackend.CurrentUserPrincipal(r.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	var homeSets []webdav.BackendSuppliedHomeSet
	if d.caldavBackend != nil {
		path, err := d.caldavBackend.CalendarHomeSetPath(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		} else {
			homeSets = append(homeSets, caldav.NewCalendarHomeSet(path))
		}
	}
	if d.carddavBackend != nil {
		path, err := d.carddavBackend.AddressBookHomeSetPath(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		} else {
			homeSets = append(homeSets, carddav.NewAddressBookHomeSet(path))
		}
	}

	if userPrincipalPath != "" {
		opts := webdav.ServePrincipalOptions{
			CurrentUserPrincipalPath: userPrincipalPath,
			HomeSets:                 homeSets,
			Capabilities: []webdav.Capability{
				carddav.CapabilityAddressBook,
				caldav.CapabilityCalendar,
			},
		}

		webdav.ServePrincipal(w, r, &opts)
		return
	}

	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}
