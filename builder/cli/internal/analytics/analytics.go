package analytics

import (
	"github.com/ruckstack/ruckstack/builder/cli/internal/settings"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/common/ui"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var httpClient = http.Client{
	Timeout: 1 * time.Second,
}
var tid = "UA-47553716-7"

var WaitGroup sync.WaitGroup

func Ask() {
	if !ui.IsTerminal {
		return
	}

	if settings.Settings.AskedAnalytics {
		return
	}

	settings.Settings.AskedAnalytics = true
	settings.Settings.AllowAnalytics = ui.PromptForBoolean("Help us improve Ruckstack via anonymous usage and crash reports?", nil)
	if settings.Settings.AllowAnalytics {
		ui.Println("Thanks! We appreciate the help")
		settings.Settings.AnalyticsId = uuid.NewV4().String()
	} else {
		ui.Println("We understand. We won't ask again")
	}

	if err := settings.Settings.Save(); err != nil {
		ui.VPrintf("error saving analytics settings: %s", err)
	}

	return
}

func TrackCommand(command string) {
	if !settings.Settings.AllowAnalytics {
		return
	}

	if command == "" {
		ui.VPrintf("analytics: command not set")
		return
	}

	v := url.Values{
		"v":   {"1"},
		"tid": {tid},

		"cid": {settings.Settings.AnalyticsId},

		"t":  {"pageview"},
		"dh": {"cli.ruckstack.com"},
		"dp": {strings.ReplaceAll(command, " ", "/")},
		"ua": {"Ruckstack " + global_util.RuckstackVersion},
	}

	WaitGroup.Add(1)
	go func() {
		defer WaitGroup.Done()

		resp, err := httpClient.PostForm("https://www.google-analytics.com/collect", v)
		if resp != nil {
			ui.VPrintf("analytics response: %s", resp.Status)
		}
		if err != nil {
			ui.VPrintf("error sending usage data: %s", err)
		}
	}()
}

func TrackError(seenError error) {
	if !settings.Settings.AllowAnalytics {
		return
	}

	v := url.Values{
		"v":   {"1"},
		"tid": {tid},

		"cid": {settings.Settings.AnalyticsId},

		"t":   {"exception"},
		"exd": {seenError.Error()},
		"exf": {"1"},
		"ua":  {"Ruckstack " + global_util.RuckstackVersion},
	}

	WaitGroup.Add(1)
	go func() {
		defer WaitGroup.Done()

		resp, err := httpClient.PostForm("https://www.google-analytics.com/collect", v)
		if resp != nil {
			ui.VPrintf("analytics exception response: %s", resp.Status)
		}
		if err != nil {
			ui.VPrintf("error sending exception data: %s", err)
		}
	}()
}
