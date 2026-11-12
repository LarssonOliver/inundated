package caldav

import (
	"encoding/xml"
	"net/http"
	"time"

	gocaldav "github.com/emersion/go-webdav/caldav"
	"github.com/larssonoliver/inundated/internal/service"
)

// NewHandler serves the CalDAV tree under /dav, plus /.well-known/caldav.
// now sets the window of closed tasks served.
func NewHandler(svc service.Service, now func() time.Time) http.Handler {
	dav := &gocaldav.Handler{Backend: &backend{svc: svc, now: now}, Prefix: prefix}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// go-webdav answers PROPPATCH with 501, and Apple clients send one on
		// every sync to set colour and order.
		if r.Method == "PROPPATCH" {
			refusePropPatch(w, r)
			return
		}
		dav.ServeHTTP(w, r)
	})
}

type property struct {
	XMLName xml.Name
}

type prop struct {
	Properties []property `xml:",any"`
}

// propertyUpdate reads both set and remove operations.
type propertyUpdate struct {
	XMLName    xml.Name `xml:"DAV: propertyupdate"`
	Operations []struct {
		Prop prop `xml:"DAV: prop"`
	} `xml:",any"`
}

type multistatus struct {
	XMLName  xml.Name `xml:"DAV: multistatus"`
	Href     string   `xml:"DAV: response>href"`
	Propstat propstat `xml:"DAV: response>propstat"`
}

type propstat struct {
	Prop   prop   `xml:"DAV: prop"`
	Status string `xml:"DAV: status"`
}

// refusePropPatch reports every property in the request as forbidden.
func refusePropPatch(w http.ResponseWriter, r *http.Request) {
	var update propertyUpdate
	if err := xml.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "malformed propertyupdate", http.StatusBadRequest)
		return
	}

	response := multistatus{Href: r.URL.EscapedPath(), Propstat: propstat{Status: "HTTP/1.1 403 Forbidden"}}
	for _, operation := range update.Operations {
		response.Propstat.Prop.Properties = append(response.Propstat.Prop.Properties, operation.Prop.Properties...)
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusMultiStatus)
	_, _ = w.Write([]byte(xml.Header))
	_ = xml.NewEncoder(w).Encode(response)
}
