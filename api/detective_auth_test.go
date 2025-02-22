// SPDX-FileCopyrightText: 2021 Lightmeter <hello@lightmeter.io>
//
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"gitlab.com/lightmeter/controlcenter/detective"
	mock_detective "gitlab.com/lightmeter/controlcenter/detective/mock"
	"gitlab.com/lightmeter/controlcenter/httpauth"
	"gitlab.com/lightmeter/controlcenter/httpauth/auth"
	"gitlab.com/lightmeter/controlcenter/lmsqlite3"
	"gitlab.com/lightmeter/controlcenter/metadata"
	"gitlab.com/lightmeter/controlcenter/pkg/runner"
	detectivesettings "gitlab.com/lightmeter/controlcenter/settings/detective"
	"gitlab.com/lightmeter/controlcenter/util/testutil"
)

func init() {
	lmsqlite3.Initialize(lmsqlite3.Options{})
}

func buildCookieClient() *http.Client {
	jar, err := cookiejar.New(&cookiejar.Options{})
	So(err, ShouldBeNil)
	return &http.Client{Jar: jar}
}

func buildTestEnv(t *testing.T) (*httptest.Server, *mock_detective.MockDetective, *metadata.AsyncWriter, func()) {
	ctrl := gomock.NewController(t)

	dir, clearDir := testutil.TempDir(t)

	registrar := &auth.FakeRegistrar{
		SessionKey: []byte("some_key"),
		Email:      "alice@example.com",
		Name:       "Alice",
		Password:   "super-secret",
	}

	detective := mock_detective.NewMockDetective(ctrl)

	auth := auth.NewAuthenticator(registrar, dir, nil)
	mux := http.NewServeMux()

	settingdDB, removeDB := testutil.TempDBConnectionMigrated(t, "master")

	handler, err := metadata.NewHandler(settingdDB)
	So(err, ShouldBeNil)

	writeRunner := metadata.NewSerialWriteRunner(handler)

	done, cancel := runner.Run(writeRunner)

	settingsWriter := writeRunner.Writer()
	settingsReader := handler.Reader

	HttpDetective(auth, mux, time.UTC, detective, &fakeEscalateRequester{}, settingsReader, true)

	httpauth.HttpAuthenticator(mux, auth, settingsReader, true)

	s := httptest.NewServer(mux)

	return s, detective, settingsWriter, func() {
		cancel()
		So(done(), ShouldBeNil)
		removeDB()
		clearDir()
		ctrl.Finish()
	}
}

func TestDetectiveAuth(t *testing.T) {
	Convey("Detective auth", t, func() {
		detectiveURL := "/api/v0/checkMessageDeliveryStatus?mail_from=a@b.c&mail_to=d@e.f&from=2020-01-01&to=2020-12-31&status=-1&some_id=&page=1"

		detectiveURLPartialMailFrom := "/api/v0/checkMessageDeliveryStatus?mail_from=b.c&mail_to=d@e.f&from=2020-01-01&to=2020-12-31&status=-1&some_id=&page=1"
		detectiveURLPartialMailTo := "/api/v0/checkMessageDeliveryStatus?mail_from=a@b.c&mail_to=e.f&from=2020-01-01&to=2020-12-31&status=-1&some_id=&page=1"

		detectiveURLEmptyMailFrom := "/api/v0/checkMessageDeliveryStatus?mail_to=d@e.f&from=2020-01-01&to=2020-12-31&status=-1&some_id=&page=1"
		detectiveURLEmptyMailTo := "/api/v0/checkMessageDeliveryStatus?mail_from=a@b.c&from=2020-01-01&to=2020-12-31&status=-1&some_id=&page=1"

		detectiveURLSomeID := "/api/v0/checkMessageDeliveryStatus?from=2020-01-01&to=2020-12-31&status=-1&some_id=1A2B3C4D&page=1"
		detectiveURLSomeIDEmpty := "/api/v0/checkMessageDeliveryStatus?from=2020-01-01&to=2020-12-31&status=-1&some_id=&page=1"
		detectiveURLSomeIDWhitespaceOnly := "/api/v0/checkMessageDeliveryStatus?from=2020-01-01&to=2020-12-31&status=-1&some_id=%20&page=1"

		detectiveURLExportCSV := detectiveURL + "&csv=true"

		c := buildCookieClient()

		s, d, settingsWriter, clear := buildTestEnv(t)
		defer clear()

		expect := func(d *mock_detective.MockDetective) {
			d.EXPECT().
				CheckMessageDelivery(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&detective.MessagesPage{}, nil)
		}

		Convey("Detective API not accessible to non-authenticated user", func() {
			r, err := c.Get(s.URL + detectiveURL)
			So(err, ShouldBeNil)
			So(r.StatusCode, ShouldEqual, http.StatusUnauthorized)

			Convey("Once we are logged in, detective API is accessible", func() {
				expect(d)
				r, err = c.PostForm(s.URL+"/login", url.Values{"email": {"alice@example.com"}, "password": {"super-secret"}})
				So(err, ShouldBeNil)
				So(r.StatusCode, ShouldEqual, http.StatusOK)
				{
					r, err := c.Get(s.URL + detectiveURL)
					So(err, ShouldBeNil)
					So(r.StatusCode, ShouldEqual, http.StatusOK)
				}
			})

			Convey("Partial searches available to authenticated users", func() {
				r, err = c.PostForm(s.URL+"/login", url.Values{"email": {"alice@example.com"}, "password": {"super-secret"}})
				So(err, ShouldBeNil)
				So(r.StatusCode, ShouldEqual, http.StatusOK)

				expect(d)
				r, err := c.Get(s.URL + detectiveURLPartialMailFrom)
				So(err, ShouldBeNil)
				So(r.StatusCode, ShouldEqual, http.StatusOK)

				expect(d)
				r, err = c.Get(s.URL + detectiveURLPartialMailTo)
				So(err, ShouldBeNil)
				So(r.StatusCode, ShouldEqual, http.StatusOK)

				expect(d)
				r, err = c.Get(s.URL + detectiveURLEmptyMailFrom)
				So(err, ShouldBeNil)
				So(r.StatusCode, ShouldEqual, http.StatusOK)

				expect(d)
				r, err = c.Get(s.URL + detectiveURLEmptyMailTo)
				So(err, ShouldBeNil)
				So(r.StatusCode, ShouldEqual, http.StatusOK)
			})

			Convey("Some ID search (no mailfrom/to) available to authenticated users", func() {
				r, err = c.PostForm(s.URL+"/login", url.Values{"email": {"alice@example.com"}, "password": {"super-secret"}})
				So(err, ShouldBeNil)
				So(r.StatusCode, ShouldEqual, http.StatusOK)

				expect(d)
				r, err = c.Get(s.URL + detectiveURLSomeID)
				So(err, ShouldBeNil)
				So(r.StatusCode, ShouldEqual, http.StatusOK)
			})
		})

		Convey("Detective API only accessible to end-users if setting is enabled", func() {
			r, err := c.Get(s.URL + detectiveURL)
			So(err, ShouldBeNil)
			So(r.StatusCode, ShouldEqual, http.StatusUnauthorized)

			Convey("Once we enable the setting, detective API is accessible to end-users", func() {
				settings := detectivesettings.Settings{}
				settings.EndUsersEnabled = true
				detectivesettings.SetSettings(context.Background(), settingsWriter, settings)

				expect(d)
				r, err := c.Get(s.URL + detectiveURL)
				So(err, ShouldBeNil)
				So(r.StatusCode, ShouldEqual, http.StatusOK)

				Convey("Partial searches unavailable to unauthenticated users", func() {
					r, err := c.Get(s.URL + detectiveURLPartialMailFrom)
					So(err, ShouldBeNil)
					So(r.StatusCode, ShouldEqual, http.StatusUnauthorized)

					r, err = c.Get(s.URL + detectiveURLPartialMailTo)
					So(err, ShouldBeNil)
					So(r.StatusCode, ShouldEqual, http.StatusUnauthorized)

					r, err = c.Get(s.URL + detectiveURLEmptyMailFrom)
					So(err, ShouldBeNil)
					So(r.StatusCode, ShouldEqual, http.StatusUnauthorized)

					r, err = c.Get(s.URL + detectiveURLEmptyMailTo)
					So(err, ShouldBeNil)
					So(r.StatusCode, ShouldEqual, http.StatusUnauthorized)
				})

				Convey("Some ID search (no mailfrom/to) available to unauthenticated users", func() {
					expect(d)
					r, err = c.Get(s.URL + detectiveURLSomeID)
					So(err, ShouldBeNil)
					So(r.StatusCode, ShouldEqual, http.StatusOK)
				})

				Convey("Empty some ID search unavailable to unauthenticated users", func() {
					r, err = c.Get(s.URL + detectiveURLSomeIDEmpty)
					So(err, ShouldBeNil)
					So(r.StatusCode, ShouldEqual, http.StatusUnauthorized)

					r, err = c.Get(s.URL + detectiveURLSomeIDWhitespaceOnly)
					So(err, ShouldBeNil)
					So(r.StatusCode, ShouldEqual, http.StatusUnauthorized)
				})
			})
		})

		Convey("CSV export only available to authenticated users", func() {
			r, err := c.Get(s.URL + detectiveURL)
			So(err, ShouldBeNil)
			So(r.StatusCode, ShouldEqual, http.StatusUnauthorized)

			settings := detectivesettings.Settings{}
			settings.EndUsersEnabled = true
			detectivesettings.SetSettings(context.Background(), settingsWriter, settings)

			// end-user can now access regular search…
			expect(d)
			r, err = c.Get(s.URL + detectiveURL)
			So(err, ShouldBeNil)
			So(r.StatusCode, ShouldEqual, http.StatusOK)

			// … but not CSV export
			Convey("CSV export still UNavailable when end-users search is enabled", func() {
				r, err := c.Get(s.URL + detectiveURLExportCSV)
				So(err, ShouldBeNil)
				So(r.StatusCode, ShouldEqual, http.StatusUnauthorized)
			})

			Convey("Once logged in, CSV export is available", func() {
				r, err = c.PostForm(s.URL+"/login", url.Values{"email": {"alice@example.com"}, "password": {"super-secret"}})
				So(err, ShouldBeNil)
				So(r.StatusCode, ShouldEqual, http.StatusOK)
				{
					expect(d)
					r, err := c.Get(s.URL + detectiveURLExportCSV)
					So(err, ShouldBeNil)
					So(r.StatusCode, ShouldEqual, http.StatusOK)
				}
			})
		})
	})
}
