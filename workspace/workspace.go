// SPDX-FileCopyrightText: 2021 Lightmeter <hello@lightmeter.io>
//
// SPDX-License-Identifier: AGPL-3.0-only

package workspace

import (
	"context"
	"errors"
	"os"
	"path"
	"time"

	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"gitlab.com/lightmeter/controlcenter/auth"
	"gitlab.com/lightmeter/controlcenter/connectionstats"
	"gitlab.com/lightmeter/controlcenter/dashboard"
	"gitlab.com/lightmeter/controlcenter/deliverydb"
	"gitlab.com/lightmeter/controlcenter/detective"
	"gitlab.com/lightmeter/controlcenter/detective/escalator"
	"gitlab.com/lightmeter/controlcenter/domainmapping"
	"gitlab.com/lightmeter/controlcenter/featureflags"
	"gitlab.com/lightmeter/controlcenter/i18n/translator"
	"gitlab.com/lightmeter/controlcenter/insights"
	insightsCore "gitlab.com/lightmeter/controlcenter/insights/core"
	"gitlab.com/lightmeter/controlcenter/intel"
	"gitlab.com/lightmeter/controlcenter/intel/collector"
	"gitlab.com/lightmeter/controlcenter/lmsqlite3/dbconn"
	"gitlab.com/lightmeter/controlcenter/lmsqlite3/migrator"
	"gitlab.com/lightmeter/controlcenter/localrbl"
	"gitlab.com/lightmeter/controlcenter/logeater/announcer"
	"gitlab.com/lightmeter/controlcenter/messagerbl"
	"gitlab.com/lightmeter/controlcenter/metadata"
	"gitlab.com/lightmeter/controlcenter/notification"
	"gitlab.com/lightmeter/controlcenter/notification/email"
	"gitlab.com/lightmeter/controlcenter/notification/slack"
	"gitlab.com/lightmeter/controlcenter/pkg/closers"
	"gitlab.com/lightmeter/controlcenter/pkg/postfix"
	"gitlab.com/lightmeter/controlcenter/pkg/runner"
	"gitlab.com/lightmeter/controlcenter/po"
	"gitlab.com/lightmeter/controlcenter/postfixversion"
	"gitlab.com/lightmeter/controlcenter/rawlogsdb"
	"gitlab.com/lightmeter/controlcenter/settings/globalsettings"
	"gitlab.com/lightmeter/controlcenter/tracking"
	"gitlab.com/lightmeter/controlcenter/util/errorutil"
	"gitlab.com/lightmeter/controlcenter/util/settingsutil"
)

type Workspace struct {
	runner.CancellableRunner
	closers.Closers

	deliveries              *deliverydb.DB
	rawLogs                 *rawlogsdb.DB
	tracker                 *tracking.Tracker
	connStats               *connectionstats.Stats
	insightsEngine          *insights.Engine
	insightsFetcher         insightsCore.Fetcher
	auth                    auth.RegistrarWithSessionKeys
	rblDetector             *messagerbl.Detector
	rblChecker              localrbl.Checker
	intelRunner             *intel.Runner
	logsLineCountPublisher  postfix.Publisher
	postfixVersionPublisher postfixversion.Publisher

	dashboard dashboard.Dashboard
	detective detective.Detective
	escalator escalator.Escalator

	NotificationCenter *notification.Center

	settingsMetaHandler *metadata.Handler
	settingsRunner      *metadata.SerialWriteRunner

	importAnnouncer         *announcer.SynchronizingAnnouncer
	connectionStatsAccessor *connectionstats.Accessor
	intelAccessor           *collector.Accessor

	rawLogsAcessor rawlogsdb.Accessor

	databases databases
}

type databases struct {
	closers.Closers

	Auth           *dbconn.PooledPair
	Connections    *dbconn.PooledPair
	Insights       *dbconn.PooledPair
	IntelCollector *dbconn.PooledPair
	Logs           *dbconn.PooledPair
	LogTracker     *dbconn.PooledPair
	Master         *dbconn.PooledPair
	RawLogs        *dbconn.PooledPair
}

func newDb(directory string, databaseName string, shouldVacuum bool) (*dbconn.PooledPair, error) {
	dbFilename := path.Join(directory, databaseName+".db")
	connPair, err := dbconn.Open(dbFilename, 10)

	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	if err := migrator.Run(connPair.RwConn.DB, databaseName); err != nil {
		return nil, errorutil.Wrap(err)
	}

	if shouldVacuum {
		// We must execute the vacuum outside of a transaction :-(
		log.Debug().Msgf("Vacuuming database %s. It'll take a while...", databaseName)

		if _, err := connPair.RwConn.DB.Exec("vacuum"); err != nil {
			return nil, errorutil.Wrap(err)
		}
	}

	return connPair, nil
}

type Options struct {
	IsUsingRsyncedLogs    bool
	DefaultSettings       metadata.DefaultValues
	AuthOptions           auth.Options
	NodeTypeHandler       tracking.NodeTypeHandler
	DataRetentionDuration time.Duration
}

var DefaultOptions = &Options{
	IsUsingRsyncedLogs:    false,
	DefaultSettings:       metadata.DefaultValues{},
	AuthOptions:           auth.Options{AllowMultipleUsers: false, PlainAuthOptions: nil},
	NodeTypeHandler:       &tracking.SingleNodeTypeHandler{},
	DataRetentionDuration: time.Hour * 24 * 30 * 3,
}

func buildFilters(reader metadata.Reader) (tracking.Filters, error) {
	settings, err := settingsutil.Get[tracking.Settings](context.Background(), reader, tracking.SettingsKey)
	if err != nil && errors.Is(err, metadata.ErrNoSuchKey) {
		return tracking.NoFilters, nil
	}

	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	filters, err := tracking.BuildFilters(settings.Filters)
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	return filters, nil
}

// FIXME: yes, I know this function is big. Splitting it into small pieces should eventually be done!
//
//nolint:maintidx
func NewWorkspace(workspaceDirectory string, options *Options) (*Workspace, error) {
	if options == nil {
		options = DefaultOptions
	}

	if err := os.MkdirAll(workspaceDirectory, os.ModePerm); err != nil {
		return nil, errorutil.Wrap(err, "Error creating working directory ", workspaceDirectory)
	}

	allDatabases := databases{Closers: closers.New()}

	for _, s := range []struct {
		name         string
		db           **dbconn.PooledPair
		shouldVacuum bool
	}{
		{"auth", &allDatabases.Auth, false},
		{"connections", &allDatabases.Connections, false},
		{"insights", &allDatabases.Insights, false},
		{"intel-collector", &allDatabases.IntelCollector, true},
		{"logs", &allDatabases.Logs, false},
		{"logtracker", &allDatabases.LogTracker, false},
		{"master", &allDatabases.Master, false},
		{"rawlogs", &allDatabases.RawLogs, false},
	} {
		db, err := newDb(workspaceDirectory, s.name, s.shouldVacuum)

		if err != nil {
			return nil, errorutil.Wrap(err, "Error opening databases in directory ", workspaceDirectory)
		}

		*s.db = db

		allDatabases.Closers.Add(db)
	}

	m, err := metadata.NewDefaultedHandler(allDatabases.Master, options.DefaultSettings)
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	filters, err := buildFilters(m.Reader)
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	deliveries, err := deliverydb.New(allDatabases.Logs, &domainmapping.DefaultMapping, deliverydb.Options{RetentionDuration: options.DataRetentionDuration})
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	tracker, err := tracking.New(allDatabases.LogTracker, &tracking.FilteredPublisher{Publisher: deliveries.ResultsPublisher(), Filters: filters}, options.NodeTypeHandler)
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	rawLogsAccessor := rawlogsdb.NewAccessor(allDatabases.RawLogs.RoConnPool)

	rawLogsDb, err := rawlogsdb.New(allDatabases.RawLogs.RwConn, rawlogsdb.Options{RetentionDuration: options.DataRetentionDuration})
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	auth, err := auth.NewAuth(allDatabases.Auth, options.AuthOptions)
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	settingsRunner := metadata.NewSerialWriteRunner(m)

	// determine instance ID from the database, or create one
	var instanceID string
	err = m.Reader.RetrieveJson(context.Background(), metadata.UuidMetaKey, &instanceID)

	if err != nil && !errors.Is(err, metadata.ErrNoSuchKey) {
		return nil, errorutil.Wrap(err)
	}

	if errors.Is(err, metadata.ErrNoSuchKey) {
		instanceID = uuid.NewV4().String()
		err := m.Writer.StoreJson(context.Background(), metadata.UuidMetaKey, instanceID)

		if err != nil {
			return nil, errorutil.Wrap(err)
		}
	}

	dashboard, err := dashboard.New(allDatabases.Logs.RoConnPool)

	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	messageDetective, err := detective.New(allDatabases.Logs.RoConnPool, rawLogsAccessor)

	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	detectiveEscalator := escalator.New()

	translators := translator.New(po.DefaultCatalog)

	notificationPolicies := notification.Policies{insights.DefaultNotificationPolicy{}}

	notifiers := map[string]notification.Notifier{
		slack.SettingsKey: slack.New(notificationPolicies, m.Reader),
		email.SettingsKey: email.New(notificationPolicies, m.Reader),
	}

	policy := &insights.DefaultNotificationPolicy{}

	notificationCenter := notification.New(m.Reader, translators, policy, notifiers)

	rblChecker := localrbl.NewChecker(m.Reader, localrbl.Options{
		NumberOfWorkers:  10,
		Lookup:           localrbl.RealLookup,
		RBLProvidersURLs: localrbl.DefaultRBLs,
	})

	rblDetector := messagerbl.New(globalsettings.New(m.Reader))

	insightsAccessor, err := insights.NewAccessor(allDatabases.Insights)
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	insightsFetcher, err := insightsCore.NewFetcher(allDatabases.Insights.RoConnPool)
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	intelOptions := intel.Options{
		InstanceID:           instanceID,
		CycleInterval:        time.Second * 30,
		ReportInterval:       time.Minute * 30,
		ReportDestinationURL: IntelReportDestinationURL,
		EventsDestinationURL: IntelEventsDestinationURL,
		IsUsingRsyncedLogs:   options.IsUsingRsyncedLogs,
	}

	intelRunner, logsLineCountPublisher, blockedipsChecker, err := intel.New(
		allDatabases.IntelCollector, allDatabases.Logs.RoConnPool, insightsFetcher,
		m.Reader, auth, allDatabases.Connections.RoConnPool, intelOptions)
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	intelAccessor, err := collector.NewAccessor(allDatabases.IntelCollector.RoConnPool)
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	insightsEngine, err := insights.NewEngine(
		&m.Reader,
		insightsAccessor,
		insightsFetcher,
		notificationCenter,
		insightsOptions(dashboard, rblChecker, rblDetector, detectiveEscalator, allDatabases.Logs.RoConnPool, blockedipsChecker, insightsFetcher))
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	connStats, err := connectionstats.New(allDatabases.Connections, connectionstats.Options{RetentionDuration: options.DataRetentionDuration})
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	connectionStatsAccessor, err := connectionstats.NewAccessor(allDatabases.Connections.RoConnPool)
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	logsRunner := runner.NewDependantPairCancellableRunner(tracker, deliveries)

	importAnnouncer := announcer.NewSynchronizingAnnouncer(insightsEngine.ImportAnnouncer(), deliveries.MostRecentLogTime, tracker.MostRecentLogTime)

	rblCheckerCancellableRunner := runner.NewCancellableRunner(func(done runner.DoneChan, cancel runner.CancelChan) {
		// TODO: Convert this one here to a proper CancellableRunner that can be cancelled...
		rblChecker.StartListening()

		go func() {
			<-cancel
			done <- nil
		}()
	})

	return &Workspace{
		deliveries:              deliveries,
		rawLogs:                 rawLogsDb,
		tracker:                 tracker,
		insightsEngine:          insightsEngine,
		insightsFetcher:         insightsFetcher,
		connStats:               connStats,
		auth:                    auth,
		rblDetector:             rblDetector,
		rblChecker:              rblChecker,
		dashboard:               dashboard,
		detective:               messageDetective,
		escalator:               detectiveEscalator,
		settingsMetaHandler:     m,
		settingsRunner:          settingsRunner,
		importAnnouncer:         importAnnouncer,
		intelRunner:             intelRunner,
		intelAccessor:           intelAccessor,
		logsLineCountPublisher:  logsLineCountPublisher,
		postfixVersionPublisher: postfixversion.NewPublisher(settingsRunner.Writer()),
		connectionStatsAccessor: connectionStatsAccessor,
		databases:               allDatabases,
		Closers: closers.New(
			connStats,
			deliveries,
			rawLogsDb,
			tracker,
			insightsEngine,
			intelRunner,
			allDatabases,
		),
		NotificationCenter: notificationCenter,
		rawLogsAcessor:     rawLogsAccessor,
		CancellableRunner: runner.NewCombinedCancellableRunners(
			insightsEngine, settingsRunner, rblDetector, logsRunner, importAnnouncer,
			intelRunner, connStats, rblCheckerCancellableRunner, rawLogsDb),
	}, nil
}

func (ws *Workspace) SettingsAcessors() (*metadata.AsyncWriter, metadata.Reader) {
	return ws.settingsRunner.Writer(), ws.settingsMetaHandler.Reader
}

func (ws *Workspace) InsightsEngine() *insights.Engine {
	return ws.insightsEngine
}

func (ws *Workspace) InsightsFetcher() insightsCore.Fetcher {
	return ws.insightsFetcher
}

func (ws *Workspace) InsightsProgressFetcher() insightsCore.ProgressFetcher {
	return ws.insightsEngine.ProgressFetcher()
}

func (ws *Workspace) Dashboard() dashboard.Dashboard {
	return ws.dashboard
}

func (ws *Workspace) ConnectionStatsAccessor() *connectionstats.Accessor {
	return ws.connectionStatsAccessor
}

func (ws *Workspace) IntelAccessor() *collector.Accessor {
	return ws.intelAccessor
}

func (ws *Workspace) Detective() detective.Detective {
	return ws.detective
}

func (ws *Workspace) DetectiveEscalationRequester() escalator.Requester {
	return ws.escalator
}

func (ws *Workspace) ImportAnnouncer() (announcer.ImportAnnouncer, error) {
	sum, err := ws.MostRecentLogTimeAndSum()
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	// first execution. Must import historical insights
	if sum.Time.IsZero() {
		return ws.importAnnouncer, nil
	}

	// otherwise skip the historical insights import
	return announcer.Skipper(ws.importAnnouncer), nil
}

func (ws *Workspace) Auth() auth.RegistrarWithSessionKeys {
	return ws.auth
}

func (ws *Workspace) MostRecentLogTimeAndSum() (postfix.SumPair, error) {
	mostRecentDeliverTime, err := ws.deliveries.MostRecentLogTime()
	if err != nil {
		return postfix.SumPair{}, errorutil.Wrap(err)
	}

	mostRecentTrackerTime, err := ws.tracker.MostRecentLogTime()
	if err != nil {
		return postfix.SumPair{}, errorutil.Wrap(err)
	}

	mostRecentConnStatsTime, err := ws.connStats.MostRecentLogTime()
	if err != nil {
		return postfix.SumPair{}, errorutil.Wrap(err)
	}

	rawLogsSum, err := rawlogsdb.MostRecentLogTimeAndSum(context.Background(), ws.databases.RawLogs.RoConnPool)
	if err != nil {
		return postfix.SumPair{}, errorutil.Wrap(err)
	}

	// if raw logs has any logs, just use it
	if rawLogsSum.Sum != nil {
		return rawLogsSum, nil
	}

	// In case we have no raw logs unavailable yet (the first execution after rawlogsdb is introduced),
	// try the old way to compute the most recent time from the other databases

	times := [3]time.Time{mostRecentConnStatsTime, mostRecentTrackerTime, mostRecentDeliverTime}

	mostRecent := time.Time{}

	for _, t := range times {
		if t.After(mostRecent) {
			mostRecent = t
		}
	}

	return postfix.SumPair{Time: mostRecent, Sum: nil}, nil
}

func (ws *Workspace) NewPublisher() postfix.Publisher {
	pub := postfix.ComposedPublisher{
		ws.tracker.Publisher(),
		ws.rblDetector.NewPublisher(),
		ws.logsLineCountPublisher,
		ws.postfixVersionPublisher,
		ws.connStats.Publisher(),
	}

	flags, err := featureflags.GetSettings(context.Background(), ws.settingsMetaHandler.Reader)

	if err != nil && !errors.Is(err, metadata.ErrNoSuchKey) {
		log.Fatal().Err(err).Msg("Should never have failed on retrieving feature flags!")
	}

	if flags != nil && flags.DisableRawLogs {
		log.Debug().Msg("Disable raw logs!")
		return pub
	}

	return append(pub, ws.rawLogs.Publisher())
}

func (ws *Workspace) HasLogs() bool {
	return ws.deliveries.HasLogs()
}

func (ws *Workspace) RawLogsAccessor() rawlogsdb.Accessor {
	return ws.rawLogsAcessor
}
