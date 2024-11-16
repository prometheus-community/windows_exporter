//go:build windows

package httphandler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/prometheus/common/version"
)

type VersionHandler struct{}

// Same struct prometheus uses for their /version endpoint.
// Separate copy to avoid pulling all of prometheus as a dependency.
type prometheusVersion struct {
	Version   string `json:"version"`
	Revision  string `json:"revision"`
	Branch    string `json:"branch"`
	BuildUser string `json:"buildUser"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
}

// Interface guard.
var _ http.Handler = (*VersionHandler)(nil)

func NewVersionHandler() VersionHandler {
	return VersionHandler{}
}

func (h VersionHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	// we can't use "version" directly as it is a package, and not an object that
	// can be serialized.
	err := json.NewEncoder(w).Encode(prometheusVersion{
		Version:   version.Version,
		Revision:  version.Revision,
		Branch:    version.Branch,
		BuildUser: version.BuildUser,
		BuildDate: version.BuildDate,
		GoVersion: version.GoVersion,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error encoding JSON: %s", err), http.StatusInternalServerError)
	}
}
