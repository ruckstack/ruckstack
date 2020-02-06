package webserver

import (
	"github.com/ruckstack/ruckstack/internal/system-control/util"
	"io"
	"net/http"
	"os"
)

func showSiteDownPage(res http.ResponseWriter, req *http.Request) {
	siteDownFile, err := os.Open(util.InstallDir() + "/data/web/site-down.html")
	defer siteDownFile.Close()
	util.Check(err)

	io.Copy(res, siteDownFile)
}

type page struct {
	Title      string
	Message    string
	SubMessage string
}
