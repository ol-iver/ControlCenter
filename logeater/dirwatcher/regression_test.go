// SPDX-FileCopyrightText: 2021 Lightmeter <hello@lightmeter.io>
//
// SPDX-License-Identifier: AGPL-3.0-only

package dirwatcher

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"gitlab.com/lightmeter/controlcenter/pkg/postfix"
	parsertimeutil "gitlab.com/lightmeter/controlcenter/pkg/postfix/logparser/timeutil"
	"gitlab.com/lightmeter/controlcenter/util/testutil"
	"gitlab.com/lightmeter/controlcenter/util/timeutil"
)

func TestRegressionIssue368(t *testing.T) {
	/*
		Lessons learned:
		- the timezone information is lost from the filesystem when reading from a docker mounted volume
			therefore all timezones for the parsed lines in the log files are considered to be in UTC,
			which we use as a reference time.
	*/
	Convey("Regression Tests issue 368", t, func() {
		clock := &timeutil.FakeClock{Time: timeutil.MustParseTime(`2030-01-01 10:00:00 +0000`)}

		timeFormat, err := parsertimeutil.Get("default")
		So(err, ShouldBeNil)

		Convey("Guessing initial time from files", func() {
			Convey("mail.err.2.gz", func() {
				reader := plainDataReader(`Feb 28 22:08:56 ubuntu-2gb-nbg1-1 postfix/postmap[1400]: fatal: open /h-5e9ec2de88d0040a44ee23d5867b3c12b58fd34f/: No such file or directory
Feb 28 22:39:44 ubuntu-2gb-nbg1-1 postfix/smtpd[4470]: error: open database /h-9e096e99702f280aef3bad9c4f6a462df2670537/: No such file or directory
Feb 28 22:43:31 ubuntu-2gb-nbg1-1 postfix/smtpd[4677]: error: open database /h-9e096e99702f280aef3bad9c4f6a462df2670537/: No such file or directory`)
				date, err := guessInitialDateForFile(reader, testutil.MustParseTime(`2019-02-28 22:43:31 +0100`), timeFormat)
				So(err, ShouldBeNil)
				So(date, ShouldEqual, testutil.MustParseTime(`2019-02-28 22:08:56 +0000`))
			})

			Convey("mail.log.4.gz", func() {
				reader := plainDataReader(`Dec  6 06:25:06 cloud2 postfix/pickup[22197]: D4D433E8C6: uid=0 from=<root>
Dec  6 06:25:06 cloud2 postfix/cleanup[23434]: D4D433E8C6: message-id=<h-e24810f14bc82f4c71d942d6e@h-32c0e75797df5c34bbefdfa.com>
Dec 14 06:24:27 cloud2 postfix/anvil[15757]: statistics: max cache size 1 at Dec 14 06:21:07`)
				date, err := guessInitialDateForFile(reader, testutil.MustParseTime(`2020-12-14 06:24:27 +0100`), timeFormat)
				So(err, ShouldBeNil)
				So(date, ShouldEqual, testutil.MustParseTime(`2020-12-06 06:25:06 +0000`))
			})
		})

		Convey("Logs importing", func() {
			dirContent := FakeDirectoryContent{
				entries: fileEntryList{
					fileEntry{filename: "mail.err", modificationTime: testutil.MustParseTime(`2020-06-26 06:25:01 +0200`)},
					fileEntry{filename: "mail.err.1", modificationTime: testutil.MustParseTime(`2020-06-25 16:40:09 +0200`)},
					// this file is in early 2019, much older than the others!!
					fileEntry{filename: "mail.err.2.gz", modificationTime: testutil.MustParseTime(`2019-02-28 22:43:31 +0100`)},
					fileEntry{filename: "mail.log", modificationTime: testutil.MustParseTime(`2021-01-05 14:05:47 +0100`)},
					fileEntry{filename: "mail.log.1", modificationTime: testutil.MustParseTime(`2021-01-03 06:22:59 +0100`)},
					fileEntry{filename: "mail.log.2.gz", modificationTime: testutil.MustParseTime(`2020-12-28 06:25:04 +0100`)},
					fileEntry{filename: "mail.log.3.gz", modificationTime: testutil.MustParseTime(`2020-12-20 06:22:40 +0100`)},
					fileEntry{filename: "mail.log.4.gz", modificationTime: testutil.MustParseTime(`2020-12-14 06:24:27 +0100`)},
					fileEntry{filename: "nonsense", modificationTime: testutil.MustParseTime(`2019-02-28 22:43:31 +0200`)},
				},
				contents: map[string]fakeFileData{
					"mail.err":   plainCurrentDataFile(``, ``),
					"mail.err.1": plainDataFile(`Jun 25 16:40:09 cloud2 postfix/postfix-script[31421]: fatal: unknown command: 'reloadd'. Usage: postfix start (or stop, reload, abort, flush, check, status, set-permissions, upgrade-configuration)`),
					"mail.err.2.gz": gzippedDataFile(`Feb 28 22:08:56 ubuntu-2gb-nbg1-1 postfix/postmap[1400]: fatal: open /h-5e9ec2de88d0040a44ee23d5867b3c12b58fd34f/: No such file or directory
Feb 28 22:39:44 ubuntu-2gb-nbg1-1 postfix/smtpd[4470]: error: open database /h-9e096e99702f280aef3bad9c4f6a462df2670537/: No such file or directory
Feb 28 22:43:31 ubuntu-2gb-nbg1-1 postfix/smtpd[4677]: error: open database /h-9e096e99702f280aef3bad9c4f6a462df2670537/: No such file or directory`),
					"mail.log": plainCurrentDataFile(`Jan  3 06:25:07 cloud2 postfix/pickup[25779]: DD78F3E8C1: uid=0 from=<root>
Jan  3 06:25:07 cloud2 postfix/cleanup[26489]: DD78F3E8C1: message-id=<h-02419a263e156696315f34ffa@h-32c0e75797df5c34bbefdfa.com>
Jan  5 14:05:47 cloud2 postfix/qmgr[1428]: 5EEC73E8C6: removed`, ``),
					"mail.log.1": plainDataFile(`Dec 28 06:25:07 cloud2 postfix/pickup[24537]: 572F93E8B7: uid=0 from=<root>
Dec 28 06:25:07 cloud2 postfix/cleanup[27496]: 572F93E8B7: message-id=<h-52b7359754499a54086fa9b88@h-32c0e75797df5c34bbefdfa.com>
Jan  3 06:22:59 cloud2 postfix/smtpd[26341]: disconnect from h-1c62d4153b2e275bb625c632[26.93.33.217] commands=0/0`),
					"mail.log.2.gz": gzippedDataFile(`Dec 20 06:25:07 cloud2 postfix/pickup[15941]: AF96E3E8C6: uid=0 from=<root>
Dec 20 06:25:07 cloud2 postfix/cleanup[16236]: AF96E3E8C6: message-id=<h-006d72b7798821336acf614a7@h-32c0e75797df5c34bbefdfa.com>
Dec 28 06:25:04 cloud2 postfix/smtpd[27432]: disconnect from h-1c62d4153b2e275bb625c632[26.93.33.217] commands=0/0`),
					"mail.log.3.gz": gzippedDataFile(`Dec 14 06:25:07 cloud2 postfix/pickup[14915]: E75F43E8C5: uid=0 from=<root>
Dec 14 06:25:07 cloud2 postfix/cleanup[16017]: E75F43E8C5: message-id=<h-ec65578888181a672d81cd9d7@h-32c0e75797df5c34bbefdfa.com>
Dec 20 06:22:40 cloud2 postfix/smtpd[16077]: disconnect from h-1c62d4153b2e275bb625c632[26.93.33.217] commands=0/0`),
					"mail.log.4.gz": gzippedDataFile(`Dec  6 06:25:06 cloud2 postfix/pickup[22197]: D4D433E8C6: uid=0 from=<root>
Dec  6 06:25:06 cloud2 postfix/cleanup[23434]: D4D433E8C6: message-id=<h-e24810f14bc82f4c71d942d6e@h-32c0e75797df5c34bbefdfa.com>
Dec 14 06:24:27 cloud2 postfix/anvil[15757]: statistics: max cache size 1 at Dec 14 06:21:07`),
				},
			}
			pub := fakePublisher{}
			announcer := &fakeAnnouncer{}
			importer := NewDirectoryImporter(dirContent, &pub, announcer, postfix.SumPair{Time: testutil.MustParseTime(`1970-01-01 00:00:00 +0100`)}, timeFormat, DefaultLogPatterns, clock)
			err := importer.Run()
			So(err, ShouldBeNil)
			So(len(pub.logs), ShouldEqual, 19)

			So(pub.logs[0].Time, ShouldResemble, testutil.MustParseTime(`2019-02-28 22:08:56 +0000`))
			So(pub.logs[len(pub.logs)-1].Time, ShouldResemble, testutil.MustParseTime(`2021-01-05 14:05:47 +0000`))
		})
	})
}

func TestRegressionIssue463(t *testing.T) {
	Convey("Regression Tests issue #463", t, func() {
		clock := &timeutil.FakeClock{Time: timeutil.MustParseTime(`2030-01-01 10:00:00 +0000`)}

		timeFormat, err := parsertimeutil.Get("default")
		So(err, ShouldBeNil)

		// the log file starts empty, but in some point is updated with some content.
		// We should not crash due an invalid time converter, obviously :-)
		dirContent := FakeDirectoryContent{
			entries: fileEntryList{
				fileEntry{filename: "mail.err", modificationTime: testutil.MustParseTime(`2021-04-27 08:00:20 +0000`)},
			},
			contents: map[string]fakeFileData{
				"mail.err": plainCurrentDataFile(``, `Apr 27 08:00:21 cloud2 postfix/pickup[15941]: AF96E3E8C6: uid=0 from=<root>`),
			},
		}

		pub := fakePublisher{}
		importer := NewDirectoryImporter(dirContent, &pub, &fakeAnnouncer{}, postfix.SumPair{Time: testutil.MustParseTime(`1970-01-01 00:00:00 +0000`)}, timeFormat, DefaultLogPatterns, clock)
		err = importer.Run()
		So(err, ShouldBeNil)

		So(len(pub.logs), ShouldEqual, 1)
		So(pub.logs[0].Time, ShouldResemble, testutil.MustParseTime(`2021-04-27 08:00:21 +0000`))
	})
}

func TestRegressionIssue644(t *testing.T) {
	/*
		Lessons learned:
			- Due to some unknown existing bug (and unknown bugs to come), sometimes logs from the past are read from an updated file.
				- In those cases, we need to prevent the year to be bumped when it happens, otherwise we'll be processing logs from the future,
					which makes no sense.
				- I suspect this is a bug in the code that handles the rsync'd files, and this might also happen in case the user accidentally
					sends an old version of a log file.
					- What we need to to in such cases is ignoring any old logs until they reach the time we expect (equal or newer than the latest log published)
				- For now what we have are workarounds. Workarounds everywhere.
			- This issue can also be intentionally caused by the user, by rsync'ng an unrelated file which contains old log lines, which cannot be computed
			- Or by messing up with the logs...
	*/
	Convey("Regression Tests issue 644", t, func() {
		clock := &timeutil.FakeClock{Time: timeutil.MustParseTime(`2021-12-10 20:00:00 +0000`)}

		timeFormat, err := parsertimeutil.Get("default")
		So(err, ShouldBeNil)

		pub := fakePublisher{}

		Convey("Test1", func() {
			{
				dirContent := FakeDirectoryContent{
					entries: fileEntryList{
						fileEntry{filename: "mail.log", modificationTime: testutil.MustParseTime(`2021-03-05 14:05:47 +0000`)},
					},
					contents: map[string]fakeFileData{
						"mail.log": plainCurrentDataFile(`Jan  2 06:25:07 cloud2 postfix/pickup[25779]: DD78F3E8C1: uid=0 from=<root>
Jan  3 06:25:07 cloud2 postfix/cleanup[26489]: DD78F3E8C1: message-id=<h-02419a263e156696315f34ffa@h-32c0e75797df5c34bbefdfa.com>
Jan  4 07:00:00 cloud2 postfix/cleanup[26489]: Something not supported
Mar  5 14:05:47 cloud2 postfix/qmgr[1428]: 5EEC73E8C6: removed
`, ``),
					},
				}

				// first run
				announcer := &fakeAnnouncer{}
				importer := NewDirectoryImporter(dirContent, &pub, announcer, postfix.SumPair{Time: testutil.MustParseTime(`1970-01-01 00:00:00 +0100`)}, timeFormat, DefaultLogPatterns, clock)
				err := importer.Run()
				So(err, ShouldBeNil)
			}

			{
				dirContent := FakeDirectoryContent{
					entries: fileEntryList{
						fileEntry{filename: "mail.log", modificationTime: testutil.MustParseTime(`2021-03-08 14:05:47 +0000`)},
					},

					// we get two new lines, and after that, two repeated ones from the past, and then some new lines again. The repeated ones are ignored, not changing the year
					contents: map[string]fakeFileData{
						"mail.log": plainCurrentDataFile(`Mar  6 00:00:00 cloud2 postfix/qmgr[1428]: 5EEC73E8C6: removed
Mar  6 10:00:00 cloud2 postfix/qmgr[1428]: 5EEC73E8C6: removed
Jan  3 06:25:07 cloud2 postfix/cleanup[26489]: DD78F3E8C1: message-id=<h-02419a263e156696315f34ffa@h-32c0e75797df5c34bbefdfa.com>
Jan  4 07:00:00 cloud2 postfix/cleanup[26489]: Something not supported
Mar  7 10:11:12 cloud2 postfix/qmgr[1428]: 5EEC73E8C6: removed
Mar  8 10:11:12 cloud2 postfix/qmgr[1428]: 5EEC73E8C6: removed`, ``),
					},
				}

				// second run
				announcer := &fakeAnnouncer{}
				importer := NewDirectoryImporter(dirContent, &pub, announcer, postfix.SumPair{Time: testutil.MustParseTime(`1970-01-01 00:00:00 +0100`)}, timeFormat, DefaultLogPatterns, clock)
				err := importer.Run()
				So(err, ShouldBeNil)

			}

			So(len(pub.logs), ShouldEqual, 8)

			So(pub.logs[0].Time, ShouldResemble, testutil.MustParseTime(`2021-01-02 06:25:07 +0000`))
			So(pub.logs[len(pub.logs)-1].Time, ShouldResemble, testutil.MustParseTime(`2021-03-08 10:11:12 +0000`))
		})
	})
}

func TestDuplicatedFiles(t *testing.T) {
	Convey("If the same file seems to be present in both compressed and non compressed form, choose the compressed ont", t, func() {
		Convey("Some files are duplicated, due to rsync'ng generated artifacts", func() {
			f := fileEntryList{
				fileEntry{filename: "mail.log", modificationTime: testutil.MustParseTime(`2023-03-02 19:38:51 +0100`)},
				fileEntry{filename: "mail.log-20220722", modificationTime: testutil.MustParseTime(`2022-07-25 15:07:12 +0200`)},
				fileEntry{filename: "mail.log-20220722.bz2", modificationTime: testutil.MustParseTime(`2022-07-25 15:07:12 +0200`)},
				fileEntry{filename: "mail.log-20220728", modificationTime: testutil.MustParseTime(`2022-07-29 16:14:44 +0200`)},
				fileEntry{filename: "mail.log-20220728.bz2", modificationTime: testutil.MustParseTime(`2022-07-29 16:14:44 +0200`)},
				fileEntry{filename: "mail.log-20220731", modificationTime: testutil.MustParseTime(`2022-08-02 14:20:14 +0200`)},
				fileEntry{filename: "mail.log-20220731.bz2", modificationTime: testutil.MustParseTime(`2022-08-02 14:20:14 +0200`)},
				fileEntry{filename: "mail.log-20220804", modificationTime: testutil.MustParseTime(`2022-08-22 11:20:08 +0200`)},
			}

			// We ignore all files that have a correspondent compressed version of it
			So(buildFilesToImport(f, BuildLogPatterns([]string{"mail.log", "mail.err", "mail.warn"}), time.Time{}), ShouldResemble,
				fileQueues{
					"mail.log": fileEntryList{
						fileEntry{filename: "mail.log-20220722.bz2", modificationTime: testutil.MustParseTime(`2022-07-25 15:07:12 +0200`)},
						fileEntry{filename: "mail.log-20220728.bz2", modificationTime: testutil.MustParseTime(`2022-07-29 16:14:44 +0200`)},
						fileEntry{filename: "mail.log-20220731.bz2", modificationTime: testutil.MustParseTime(`2022-08-02 14:20:14 +0200`)},
						fileEntry{filename: "mail.log-20220804", modificationTime: testutil.MustParseTime(`2022-08-22 11:20:08 +0200`)},
						fileEntry{filename: "mail.log", modificationTime: testutil.MustParseTime(`2023-03-02 19:38:51 +0100`)},
					},
					"mail.err":  {},
					"mail.warn": {},
				})
		})
	})
}
