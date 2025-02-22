// SPDX-FileCopyrightText: 2021 Lightmeter <hello@lightmeter.io>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/lightmeter/controlcenter/api"
	httpauth "gitlab.com/lightmeter/controlcenter/httpauth"
	auth "gitlab.com/lightmeter/controlcenter/httpauth/auth"
	"gitlab.com/lightmeter/controlcenter/httpmiddleware"
	"gitlab.com/lightmeter/controlcenter/httpsettings"
	"gitlab.com/lightmeter/controlcenter/i18n"
	"gitlab.com/lightmeter/controlcenter/metadata"
	"gitlab.com/lightmeter/controlcenter/newsletter"
	"gitlab.com/lightmeter/controlcenter/po"
	"gitlab.com/lightmeter/controlcenter/settings"
	"gitlab.com/lightmeter/controlcenter/settings/globalsettings"
	"gitlab.com/lightmeter/controlcenter/staticdata"
	"gitlab.com/lightmeter/controlcenter/util/errorutil"
	"gitlab.com/lightmeter/controlcenter/workspace"
)

type HttpServer struct {
	Workspace            *workspace.Workspace
	WorkspaceDirectory   string
	Timezone             *time.Location
	Address              string
	FrontendDev          bool
	IsBehindReverseProxy bool
}

func (s *HttpServer) Start() error {
	if s.Workspace == nil {
		return errorutil.Wrap(errors.New("Workspace is nil"))
	}

	if s.WorkspaceDirectory == "" {
		return errorutil.Wrap(errors.New("WorkspaceDirectory is empty string"))
	}

	if s.Timezone == nil {
		return errorutil.Wrap(errors.New("Timezone is nil"))
	}

	if s.Address == "" {
		return errorutil.Wrap(errors.New("Address is empty string"))
	}

	initialSetupSettings := settings.NewInitialSetupSettings(newsletter.NewSubscriber("https://phplist.lightmeter.io/"))

	writer, reader := s.Workspace.SettingsAcessors()

	setup := httpsettings.NewSettings(writer, reader, initialSetupSettings, s.Workspace.NotificationCenter, s.Workspace.InsightsEngine())

	publicURL, err := getPublicURL(context.Background(), reader)

	if err != nil {
		return errorutil.Wrap(err)
	}

	auth := auth.NewAuthenticator(s.Workspace.Auth(), s.WorkspaceDirectory, publicURL)

	mux := http.NewServeMux()

	i18nService := i18n.NewService(po.DefaultCatalog)

	mux.Handle("/language/metadata", httpmiddleware.WithDefaultStackWithoutAuth().
		WithEndpoint(httpmiddleware.CustomHTTPHandler(i18nService.LanguageMetaDataHandler)))

	mux.Handle("/", http.StripPrefix("/", http.FileServer(staticdata.HttpAssets)))

	exposeApiExplorer(mux)

	exposeProfiler(mux)

	api.HttpDashboard(auth, mux, s.Timezone, s.Workspace.Dashboard())
	api.HttpInsights(auth, mux, s.Timezone, s.Workspace.InsightsFetcher(), s.Workspace.InsightsEngine())
	api.HttpInsightsProgress(auth, mux, s.Workspace.InsightsProgressFetcher())
	api.HttpDetective(auth, mux, s.Timezone, s.Workspace.Detective(), s.Workspace.DetectiveEscalationRequester(), reader, s.IsBehindReverseProxy)
	api.HttpConnectionsDashboard(auth, mux, s.Timezone, s.Workspace.ConnectionStatsAccessor())
	api.HttpReports(auth, mux, s.Timezone, s.Workspace.IntelAccessor())
	api.HttpRawLogs(auth, mux, s.Timezone, s.Workspace.RawLogsAccessor())
	api.HttpStatusMessage(auth, mux, s.Workspace.IntelAccessor())

	setup.HttpSetup(mux, auth)

	httpauth.HttpAuthenticator(mux, auth, reader, s.IsBehindReverseProxy)

	server := http.Server{Handler: wrap(mux)}

	ln, err := net.Listen("tcp", s.Address)

	if err != nil {
		return errorutil.Wrap(err)
	}

	log.Info().Msgf("Lightmeter ControlCenter is running on http://%s", ln.Addr().String())

	return server.Serve(ln)
}

func getPublicURL(ctx context.Context, reader metadata.Reader) (*url.URL, error) {
	globSettings, err := globalsettings.GetSettings(ctx, reader)
	if err != nil && errors.Is(err, metadata.ErrNoSuchKey) {
		return nil, nil
	}

	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	if globSettings.PublicURL == "" {
		return nil, nil
	}

	return url.Parse(globSettings.PublicURL)
}
