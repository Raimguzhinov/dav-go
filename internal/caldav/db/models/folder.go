package models

import (
	"strconv"

	"github.com/ceres919/go-webdav/caldav"
)

type Folder struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Types       []string `json:"types"`
	Size        int64    `json:"size"`
}

func (f *Folder) ToDomain() caldav.Calendar {
	return caldav.Calendar{
		Path:                  strconv.Itoa(f.ID),
		Name:                  f.Name,
		Description:           f.Description,
		SupportedComponentSet: f.Types,
		MaxResourceSize:       f.Size,
	}
}
