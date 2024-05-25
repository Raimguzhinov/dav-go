package app

import (
	"context"
	"net/http"

	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/go-webdav/carddav"
)

type userPrincipalBackend struct{}

func (u *userPrincipalBackend) CurrentUserPrincipal(ctx context.Context) (string, error) {
	//authCtx, ok := auth.FromContext(ctx)
	//if !ok {
	//	panic("Invalid data in auth context!")
	//}
	//if authCtx == nil {
	//	return "", fmt.Errorf("unauthenticated requests are not supported")
	//}
	//
	//userDir := base64.RawStdEncoding.EncodeToString([]byte(authCtx.UserName))
	return "/" + "admin" + "/", nil
}

type davHandler struct {
	upBackend webdav.UserPrincipalBackend
	//authBackend    auth.AuthProvider
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
		path, err := d.carddavBackend.AddressbookHomeSetPath(r.Context())
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
