// SPDX-FileCopyrightText: 2021 Lightmeter <hello@lightmeter.io>
//
// SPDX-License-Identifier: AGPL-3.0-only

package dirlogsource

import (
	"github.com/rs/zerolog/log"
	"gitlab.com/lightmeter/controlcenter/logeater/announcer"
	"gitlab.com/lightmeter/controlcenter/logeater/dirwatcher"
	"gitlab.com/lightmeter/controlcenter/logeater/logsource"
	"gitlab.com/lightmeter/controlcenter/pkg/postfix"
	parsertimeutil "gitlab.com/lightmeter/controlcenter/pkg/postfix/logparser/timeutil"
	"gitlab.com/lightmeter/controlcenter/util/errorutil"
	"gitlab.com/lightmeter/controlcenter/util/timeutil"
)

type Source struct {
	sum       postfix.SumPair
	dir       dirwatcher.DirectoryContent
	announcer announcer.ImportAnnouncer
	patterns  dirwatcher.LogPatterns
	format    parsertimeutil.TimeFormat
	clock     timeutil.Clock

	// should continue waiting for new results (tail -f)?
	follow bool
}

func New(dirname string, sum postfix.SumPair, announcer announcer.ImportAnnouncer, follow bool, rsynced bool, logFormat string, patterns dirwatcher.LogPatterns, clock timeutil.Clock) (logsource.Source, error) {
	timeFormat, err := parsertimeutil.Get(logFormat)
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	dir, err := func() (dirwatcher.DirectoryContent, error) {
		if rsynced {
			return dirwatcher.NewDirectoryContentForRsync(dirname, timeFormat, patterns)
		}

		return dirwatcher.NewDirectoryContent(dirname, timeFormat)
	}()

	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	format, err := parsertimeutil.Get(logFormat)
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	func() {
		if sum.Time.IsZero() {
			log.Info().Msg("Start importing Postfix logs directory into a new workspace")
			return
		}

		log.Info().Msgf("Importing Postfix logs directory from time %v", sum.Time)
	}()

	return &Source{
		sum:       sum,
		dir:       dir,
		follow:    follow,
		announcer: announcer,
		patterns:  patterns,
		format:    format,
		clock:     clock,
	}, nil
}

func (s *Source) PublishLogs(p postfix.Publisher) error {
	watcher := dirwatcher.NewDirectoryImporter(s.dir, p, s.announcer, s.sum, s.format, s.patterns, s.clock)

	f := func() func() error {
		if s.follow {
			return watcher.Run
		}

		return watcher.ImportOnly
	}()

	if err := f(); err != nil {
		return errorutil.Wrap(err)
	}

	return nil
}
