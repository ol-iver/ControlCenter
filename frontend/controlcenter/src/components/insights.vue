<!--
SPDX-FileCopyrightText: 2021 Lightmeter <hello@lightmeter.io>

SPDX-License-Identifier: AGPL-3.0-only
-->

<!-- TODO: clean up the mess in this file, moving each insight to its own component -->

<template>
  <div class="insights" id="insights">
    <b-modal
      ref="modal-rbl-list"
      id="modal-rbl-list"
      size="lg"
      hide-footer
      centered
      :title="insightRblCheckedIpTitle"
    >
      <p class="intro">
        <translate
          >These lists are recommending that emails from your server be blocked
          &ndash; check their messages for hints</translate
        >
      </p>
      <span id="rbl-list-content">
        <div class="card" v-for="r of rbls" v-bind:key="r.text">
          <div class="card-body">
            <h5 class="card-title">
              <span class="badge badge-pill badge-warning">List</span
              >{{ r.rbl }}
            </h5>
            <p class="card-text">
              <span class="message-label">Message:</span>
              <span v-linkified:options="{ target: { url: '_blank' } }">{{
                r.text
              }}</span>
            </p>
          </div>
        </div>
      </span>

      <b-row class="vue-modal-footer">
        <b-col>
          <b-button
            class="btn-cancel"
            variant="outline-danger"
            @click="hideRBLListModal"
          >
            <translate>Close</translate>
          </b-button>
        </b-col>
      </b-row>
    </b-modal>

    <b-modal
      ref="modal-detective-escalation"
      id="modal-detective-escalation"
      size="lg"
      hide-footer
      centered
      :title="titleForDetectiveInsightWindow"
    >
      <p
        class="detective-insight-header"
        v-translate="{
          sender: detectiveInsightSender,
          recipient: detectiveInsightRecipient,
          begin: detectiveInsightTimeBegin,
          end: detectiveInsightTimeEnd
        }"
        render-html="true"
      >
        From %{sender} to %{recipient} between %{begin} and %{end}
      </p>
      <detective-results
        :results="detectiveInsight.content.messages"
        :showQueues="true"
      ></detective-results>
      <b-row class="vue-modal-footer">
        <b-col>
          <b-button
            class="btn-cancel"
            variant="outline-danger"
            @click="hideDetectiveInsightModalWindow()"
          >
            <translate>Close</translate>
          </b-button>
        </b-col>
      </b-row>
    </b-modal>
    <b-modal
      ref="modal-msg-rbl"
      id="modal-msg-rbl"
      size="lg"
      hide-footer
      centered
      :title="insightMsgRblTitle"
    >
      <div class="modal-body">
        <blockquote>
          <span
            id="rbl-msg-rbl-content"
            v-linkified:options="{ target: { url: '_blank' } }"
          >
            {{ msgRblDetails }}
          </span>
        </blockquote>
      </div>

      <b-row class="vue-modal-footer">
        <b-col>
          <b-button
            class="btn-cancel"
            variant="outline-danger"
            @click="hideRBLMsqModal"
          >
            <translate>Close</translate>
          </b-button>
        </b-col>
      </b-row>
    </b-modal>
    <b-modal
      ref="modal-import-summary"
      id="modal-import-summary"
      size="lg"
      hide-footer
      centered
      :title="importSummaryWindowTitle()"
    >
      <div class="modal-body">
        <import-summary-insight-content
          :content="importSummaryInsight.content"
        ></import-summary-insight-content>
      </div>

      <b-row class="vue-modal-footer">
        <b-col>
          <b-button
            variant="outline-primary"
            @click="showArchivedInsightsBySummaryInsight(importSummaryInsight)"
          >
            <translate>View all archived</translate>
          </b-button>
        </b-col>
      </b-row>
    </b-modal>

    <b-modal
      ref="modal-blockedips"
      id="modal-blockedips"
      size="lg"
      hide-footer
      centered
      :title="blockedIPsWindowTitle()"
    >
      <div class="modal-body">
        <blockedips-insight-content
          :content="blockedIPsInsight.content"
        ></blockedips-insight-content>
      </div>

      <b-row class="vue-modal-footer">
        <b-col>
          <b-button
            class="btn-cancel"
            variant="outline-danger"
            @click="hideBlockedIPsListModal"
          >
            <translate>Close</translate>
          </b-button>
        </b-col>
      </b-row>
    </b-modal>

    <b-modal
      ref="modal-blockedips-summary"
      id="modal-blockedips-summary"
      size="lg"
      hide-footer
      centered
      :title="blockedIPsSummarysWindowTitle()"
    >
      <div class="modal-body">
        <blockedips-insight-summary-content
          :content="blockedIPsSummarysInsight.content"
        ></blockedips-insight-summary-content>
      </div>

      <b-row class="vue-modal-footer">
        <b-col>
          <b-button
            variant="outline-primary"
            @click="
              showArchivedInsightsByBlockedIPsSummaryInsight(
                blockedIPsSummarysInsight
              )
            "
          >
            <translate>View Details</translate>
          </b-button>
        </b-col>
      </b-row>
    </b-modal>

    <div
      class="row container d-flex justify-content-center"
      style="margin: 3rem 0;"
      v-if="insights.length == 0"
    >
      <translate>No insight matching selected filters</translate>
    </div>
    <div
      v-for="insight of insightsTransformed"
      v-bind:key="insight.id"
      class="col-card col-md-6 h-25"
    >
      <div class="card" v-bind:class="[insight.highlightedClass]">
        <div class="row">
          <div
            class="col-lg-1 col-md-2 col-sm-1 col-2 rating"
            v-bind:class="[insight.ratingClass]"
          ></div>
          <div class="col-lg-11 col-md-10 col-10">
            <div class="card-block">
              <div
                class="d-flex flex-row justify-content-between insight-header"
              >
                <p class="card-text category">{{ insight.category }}</p>
                <div class="insight-actions">
                  <span
                    v-if="insight.help_link"
                    v-on:click="onInsightInfo($event, insight.help_link)"
                    v-b-tooltip.hover
                    :title="Info"
                  >
                    <i class="fa fa-info-circle lm-info-circle-grayblue"></i>
                  </span>

                  <span
                    v-if="
                      insight.content_type === 'high_bounce_rate' ||
                        insight.content_type === 'mail_inactivity'
                    "
                    v-on:click="
                      trackClick('Settings', 'highBounceRateInsightClick')
                    "
                    v-b-tooltip.hover
                    :title="titleEditInsightSettings"
                  >
                    <router-link to="/settings">
                      <i
                        class="fas fa-cog"
                        data-toggle="tooltip"
                        data-placement="bottom"
                      ></i
                    ></router-link>
                  </span>

                  <span
                    v-if="insight.category.toLowerCase() != 'archived'"
                    v-on:click="
                      archiveInsight(insight.id, insight.content_type)
                    "
                    v-b-tooltip.hover
                    :title="titleArchiveInsight"
                  >
                    <i class="fas fa-times-circle"></i>
                  </span>
                </div>
              </div>
              <h6 class="card-title title">{{ insight.title }}</h6>

              <!-- TODO: extract each kind of card to its own component and remove this giant if-then... -->
              <p
                v-if="insight.content_type === 'high_bounce_rate'"
                class="card-text description"
              >
                <span v-html="insight.description"></span>
                <log-viewer-button :insight="insight" />
              </p>

              <p
                v-if="insight.content_type === 'newsfeed_content'"
                class="card-text description"
              >
                <span v-html="insight.description"></span>
                <button
                  v-on:click="onNewsFeedMoreInfo($event, insight)"
                  class="btn btn-sm"
                >
                  <translate>Read more</translate>
                </button>
              </p>

              <p
                v-if="insight.content_type === 'import_summary'"
                class="card-text description"
              >
                <span v-html="insight.description"></span>
                <button
                  v-b-modal.modal-import-summary
                  v-on:click="onImportSummaryDetails(insight)"
                  class="btn btn-sm"
                  v-show="insight.content.insights.length > 0"
                >
                  <translate>Details</translate>
                </button>
              </p>

              <p
                v-if="insight.content_type === 'blockedips'"
                class="card-text description"
              >
                <span v-html="insight.description"></span>
                <button
                  v-b-modal.modal-blockedips
                  v-on:click="onBruteForceDetails(insight)"
                  class="btn btn-sm"
                >
                  <translate>Details</translate>
                </button>
              </p>

              <p
                v-if="insight.content_type === 'blockedips_summary'"
                class="card-text description"
              >
                <span v-html="insight.description"></span>
                <button
                  v-b-modal.modal-blockedips-summary
                  v-on:click="onBruteForceSummaryDetails(insight)"
                  class="btn btn-sm"
                >
                  <translate>Details</translate>
                </button>
              </p>

              <p
                v-if="insight.content_type === 'mail_inactivity'"
                class="card-text description"
              >
                <span v-html="insight.description"></span>
                <log-viewer-button :insight="insight" />
              </p>
              <p
                v-if="insight.content_type === 'welcome_content'"
                class="card-text description"
              >
                <translate
                  >Insights reveal mailops problems in real time &ndash; both
                  here, and via notifications</translate
                >
              </p>
              <p
                v-if="insight.content_type === 'insights_introduction_content'"
                class="card-text description"
              >
                <translate
                  >Join us on the journey to better mailops! We're listening for
                  your feedback</translate
                >
              </p>
              <p
                v-if="insight.content_type === 'local_rbl_check'"
                class="card-text description"
              >
                <span v-html="insight.description.message"></span>

                <button
                  v-b-modal.modal-rbl-list
                  v-on:click="onBuildInsightRbl(insight.id)"
                  class="btn btn-sm"
                >
                  <translate>Details</translate>
                </button>
              </p>
              <p
                v-if="insight.content_type === 'message_rbl'"
                class="card-text description"
              >
                <span v-html="insight.description.message"></span>
                <button
                  v-b-modal.modal-msg-rbl
                  v-on:click="onBuildInsightMsgRbl(insight.id)"
                  class="btn btn-sm"
                >
                  <translate>Details</translate>
                </button>
              </p>
              <p
                v-if="insight.content_type === 'detective_escalation'"
                class="card-text description"
              >
                <span v-translate="{ count: countDetectiveIssues(insight) }"
                  >Investigation requested into failed delivery of %{count}
                  messages
                </span>
                <button
                  v-b-modal.modal-detective-escalation
                  v-on:click="onDetectiveEscalationDetails(insight)"
                  class="btn btn-sm"
                >
                  <translate>Details</translate>
                </button>
                <log-viewer-button :insight="insight" />
              </p>

              <div
                class="card-text time d-flex justify-content-between align-items-center"
              >
                <span>{{ insight.modTime }}</span>
                <div
                  v-if="showUserRating(insight)"
                  class="user-rating d-flex flex-wrap align-items-center"
                >
                  <span>
                    <translate>Useful?</translate>
                  </span>
                  <div>
                    <span
                      class="user-rating-smiley user-rating-smiley-good"
                      v-on:click="sendUserRating(insight, 2)"
                      title="Very useful"
                    >
                      <i class="far fa-2x fa-smile"></i>
                    </span>
                    <span
                      class="user-rating-smiley user-rating-smiley-neutral"
                      v-on:click="sendUserRating(insight, 1)"
                      title="Somewhat useful"
                    >
                      <i class="far fa-2x fa-meh"></i>
                    </span>
                    <span
                      class="user-rating-smiley user-rating-smiley-bad"
                      v-on:click="sendUserRating(insight, 0)"
                      title="Not useful at all"
                    >
                      <i class="far fa-2x fa-frown"></i>
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import moment from "moment";
import { getApplicationInfo, postUserRating, archiveInsight } from "@/lib/api";
import tracking from "../mixin/global_shared.js";
import linkify from "vue-linkify";
import Vue from "vue";
import { mapActions } from "vuex";

Vue.directive("linkified", linkify);

export default {
  name: "insights",
  mixins: [tracking],
  props: {
    insights: Array
  },
  computed: {
    insightsTransformed() {
      return this.transformInsights(this.insights);
    },
    Info() {
      return this.$gettext("Info");
    },
    titleArchiveInsight() {
      return this.$gettext("Archive Insight");
    },
    titleForDetectiveInsightWindow() {
      return this.$gettext("Failed deliveries reported");
    },
    titleEditInsightSettings() {
      return this.$gettext("Edit settings for insights generation");
    },
    detectiveInsightSender() {
      return `<strong>` + this.detectiveInsight.content.sender + `</strong>`;
    },
    detectiveInsightRecipient() {
      return `<strong>` + this.detectiveInsight.content.recipient + `</strong>`;
    },
    detectiveInsightTimeBegin() {
      return (
        `<strong>` +
        formatDateForDetectiveInsightModalWindow(
          this.detectiveInsight.content.time_interval.from
        ) +
        `</strong>`
      );
    },
    detectiveInsightTimeEnd() {
      return (
        `<strong>` +
        formatDateForDetectiveInsightModalWindow(
          this.detectiveInsight.content.time_interval.to
        ) +
        `</strong>`
      );
    }
  },
  updated() {
    let i = document.querySelector(".insight-highlighted");

    if (i == undefined) {
      return;
    }

    i.scrollIntoView();
  },
  mounted() {
    let vue = this;
    getApplicationInfo().then(function(response) {
      vue.applicationData = response.data;
    });
  },
  data() {
    return {
      rbls: [],
      msgRblDetails: "",
      insightRblCheckedIpTitle: "",
      insightMsgRblTitle: "",
      importSummaryInsight: {},
      blockedIPsInsight: {},
      blockedIPsSummarysInsight: {},
      applicationData: { version: "" },
      // FIXME: AAAHHH, this is really ugly! This insight component should be refactored/split ASAP!
      detectiveInsight: {
        content: { messages: [], time_interval: { from: "", to: "" } }
      }
    };
  },
  methods: {
    showUserRating(insight) {
      if (!insight.user_rating_old) return false;
      for (let i of this.insights)
        if (i.content_type == insight.content_type) return i.id == insight.id;
    },
    sendUserRating(insight, rating) {
      let vue = this;
      insight.user_rating_old = false; // hide rating div before next insights refresh
      postUserRating(insight.content_type, rating).then(function() {
        vue.newAlertSuccess(vue.$gettext("Thank you for your feedback"));
      });
    },
    countDetectiveIssues(insight) {
      return Object.keys(insight.content.messages).length;
    },
    onBuildInsightRbl: function(id) {
      this.buildInsightRblCheckedIp(id);
      this.buildInsightRblList(id);
      this.trackEvent("InsightDescription", "openRblModal");
    },
    onBuildInsightMsgRbl(id) {
      this.buildInsightMsgRblTitle(id);
      this.buildInsightMsgRblDetails(id);
      this.trackEvent("InsightDescription", "openHostBlockModal");
    },
    onDetectiveEscalationDetails(insight) {
      this.detectiveInsight = insight;
      this.trackEvent("InsightDescription", "openDetectiveEscalateMessage");
    },
    onNewsFeedMoreInfo(event, insight) {
      event.preventDefault();
      this.trackClick("InsightNewsfeed", insight.content.link);
      window.open(insight.content.link);
    },
    newsfeed_content_title(insight) {
      return insight.content.title;
    },
    blockedips_title(insight) {
      let translation = this.$gettext(`Suspicious IPs banned`);
      return this.$gettextInterpolate(translation, {
        count: insight.content.total_number
      });
    },
    blockedips_summary_title() {
      return this.$gettext(`Suspicious IPs banned this week`);
    },
    high_bounce_rate_title() {
      return this.$gettext("High Bounce Rate");
    },
    mail_inactivity_title() {
      return this.$gettext("Mail Inactivity");
    },
    welcome_content_title() {
      return this.$gettext("Your first Insight");
    },
    import_summary_title() {
      return this.$gettext("Insights were generated from your logs");
    },
    insights_introduction_content_title() {
      let translation = this.$gettext("Welcome to Lightmeter %{version}");
      return this.$gettextInterpolate(translation, {
        version: this.applicationData.version
      });
    },
    local_rbl_check_title() {
      return this.$gettext("IP on shared blocklist");
    },
    message_rbl_title(i) {
      let translation = this.$gettext("IP blocked by %{host}");
      return this.$gettextInterpolate(translation, { host: i.content.host });
    },
    detective_escalation_title() {
      return this.$gettext("Lost message(s) escalated by user");
    },
    high_bounce_rate_description(i) {
      let c = i.content;
      let translation = this.$gettext(
        "<b>%{bounceValue}%</b> bounce rate between %{intFrom} and %{intTo}"
      );

      return this.$gettextInterpolate(translation, {
        bounceValue: (c.value * 100.0).toFixed(2),
        intFrom: formatInsightDescriptionDateTime(c.interval.from),
        intTo: formatInsightDescriptionDateTime(c.interval.to)
      });
    },
    mail_inactivity_description(i) {
      let c = i.content;
      let translation = this.$gettext(
        "No emails were sent or received between %{intFrom} and %{intTo}"
      );

      return this.$gettextInterpolate(translation, {
        intFrom: formatInsightDescriptionDateTime(c.interval.from),
        intTo: formatInsightDescriptionDateTime(c.interval.to)
      });
    },
    blockedips_description(insight) {
      let translation = this.$gettext(
        `<strong>%{total_connections}</strong> connections blocked from <strong>%{total_ips}</strong> banned IPs (peer network)`
      );
      return this.$gettextInterpolate(translation, {
        total_ips: insight.content.top_ips.length,
        total_connections: new Intl.NumberFormat().format(
          insight.content.total_number
        )
      });
    },
    blockedips_summary_description(insight) {
      let translation = this.$gettext(
        `<strong>%{connections_count}</strong> connections from <strong>%{ip_count}</strong> IPs were blocked over %{days} days`
      );

      let v = insight.content.summary.reduce(function(p, c) {
        return {
          ip_count: p.ip_count + c.ip_count,
          connections_count: p.connections_count + c.connections_count
        };
      });

      return this.$gettextInterpolate(translation, {
        days:
          (moment(insight.content.time_interval.to) -
            moment(insight.content.time_interval.from)) /
          (24 * 3600 * 1000),
        connections_count: new Intl.NumberFormat().format(v.connections_count),
        ip_count: new Intl.NumberFormat().format(v.ip_count)
      });
    },

    local_rbl_check_description(i) {
      let c = i.content;
      // TODO: handle difference between singular (one RBL) and plurals
      let translation = this.$gettext(
        "The IP address %{ip} is listed by <strong>%{rblCount}</strong> <abbr title='Real-time Blackhole List'>RBL</abbr>s"
      );

      let message = this.$gettextInterpolate(translation, {
        ip: c.address,
        rblCount: c.rbls.length
      });

      return {
        id: i.id.toString(),
        message: message
      };
    },
    message_rbl_description(i) {
      let c = i.content;
      let translation = this.$gettext(
        "The IP %{ip} cannot deliver to %{recipient} (<strong>%{host}</strong>)"
      );
      let message = this.$gettextInterpolate(translation, {
        ip: c.address,
        recipient: c.recipient,
        host: c.host
      });

      return {
        id: i.id.toString(),
        message: message
      };
    },
    newsfeed_content_description(insight) {
      return insight.content.description.substr(0, 65);
    },
    import_summary_description(insight) {
      let c = insight.content;
      //let counter = Object.entries(c.ids).reduce(function(acc, v) { return acc + v[1].length }, 0)
      let counter = c.insights.length;

      let translation = this.$gettext(
        "Events since %{start} were analysed, producing %{count} Insights"
      );

      return this.$gettextInterpolate(translation, {
        start: formatInsightDescriptionDate(c.interval.from),
        count: counter
      });
    },
    transformInsights(insights) {
      let vue = this;

      if (insights === null) {
        return;
      }

      let insightsTransformed = [];

      let highlightedClass = function(id) {
        if (vue.$route.params.id == undefined) {
          return "";
        }

        return vue.$route.params.id == id ? "insight-highlighted" : "";
      };

      for (let insight of insights) {
        let transformed = insight;

        transformed.id = insight.id;
        transformed.category = vue.buildInsightCategory(insight);
        transformed.modTime = vue.buildInsightTime(insight);
        transformed.title = vue.buildInsightTitle(insight);
        transformed.ratingClass = vue.buildInsightRating(insight);
        transformed.description = vue.buildInsightDescriptionValues(insight);
        transformed.highlightedClass = highlightedClass(insight.id);

        insightsTransformed.push(transformed);
      }

      return insightsTransformed;
    },
    buildInsightCategory(insight) {
      // FIXME We shouldn't capitalise in the code -- leave that for the i18n workflow to decide
      return (
        insight.category.charAt(0).toUpperCase() + insight.category.slice(1)
      );
    },
    buildInsightTime(insight) {
      return moment(insight.time).format("DD MMM YYYY | h:mmA");
    },
    buildInsightTitle(insight) {
      const s = this[insight.content_type + "_title"];

      if (typeof s == "string") {
        return s;
      }

      if (typeof s == "function") {
        return s(insight);
      }

      let translation = this.$gettext("Title for %{content}");

      return this.$gettextInterpolate(translation, {
        content: insight.content_type
      });
    },
    buildInsightRating(insight) {
      return insight.rating;
    },
    buildInsightDescriptionValues(insight) {
      let handler = this[insight.content_type + "_description"];

      if (handler === undefined) {
        // NOTE: this string is for debug purposes only, therefore does not need to be translated
        // It happens only during the development of a new insight, as a final one should always have a description
        return "Description for " + insight.content_type;
      }

      return handler(insight);
    },
    buildInsightRblCheckedIp(insightId) {
      let insight = this.insights.find(i => i.id === insightId);

      if (insight === undefined) {
        return "";
      }

      let translation = this.$gettext("RBLS for %{address}");

      this.insightRblCheckedIpTitle = this.$gettextInterpolate(translation, {
        address: insight.content.address
      });
    },
    buildInsightRblList(insightId) {
      let insight = this.insights.find(i => i.id === insightId);

      if (insight === undefined) {
        return;
      }

      this.rbls = insight.content.rbls;
    },
    buildInsightMsgRblDetails(insightId) {
      let insight = this.insights.find(i => i.id == insightId);

      if (insight === undefined) {
        return;
      }

      this.msgRblDetails = insight.content.message;
    },
    buildInsightMsgRblTitle(insightId) {
      let insight = this.insights.find(i => i.id === insightId);

      if (insight === undefined) {
        return "";
      }

      let translation = this.$gettext(
        "Original response from %{recipient} (%{host})"
      );

      this.insightMsgRblTitle = this.$gettextInterpolate(translation, {
        recipient: insight.content.recipient,
        host: insight.content.host
      });
    },
    onInsightInfo(event, helpLink) {
      event.preventDefault();
      this.trackEvent("InsightsInfoButton", helpLink);
      window.open(helpLink);
    },
    archiveInsight(id, type) {
      let vue = this;

      archiveInsight(id).then(function() {
        vue.$emit("dateIntervalChanged"); // ask index.vue to refresh insights
        vue.trackEvent("ArchiveInsight", type);
      });
    },
    hideRBLListModal() {
      this.$refs["modal-rbl-list"].hide();
    },
    hideBlockedIPsListModal() {
      this.$refs["modal-blockedips"].hide();
    },
    hideBlockedIPsSummaryListModal() {
      this.$refs["modal-blockedips-summary"].hide();
    },
    hideRBLMsqModal() {
      this.$refs["modal-msg-rbl"].hide();
    },
    hideDetectiveInsightModalWindow() {
      this.$refs["modal-detective-escalation"].hide();
    },
    applySummaryInterval(interval) {
      this.$emit("dateIntervalChanged", {
        startDate: interval.from,
        endDate: interval.to,
        category: "archived"
      });
    },
    onImportSummaryDetails(insight) {
      this.trackEvent("InsightDescription", "openSummaryInsightModal");
      this.importSummaryInsight = insight;
    },
    onBruteForceDetails(insight) {
      this.trackEvent("InsightDescription", "openBlockedIPsInsightModal");
      this.blockedIPsInsight = insight;
    },
    onBruteForceSummaryDetails(insight) {
      this.trackEvent(
        "InsightDescription",
        "openBlockedIPsSummaryInsightModal"
      );
      this.blockedIPsSummarysInsight = insight;
    },
    importSummaryWindowTitle() {
      return this.$gettext("Mail activity imported successfully");
    },
    blockedIPsWindowTitle() {
      return this.$gettext("Blocked suspicious connection attempts");
    },
    blockedIPsSummarysWindowTitle() {
      return this.$gettext("Blocked suspicious connection attempts summary");
    },
    showArchivedInsightsBySummaryInsight(insight) {
      this.trackEvent("HistoricalInsights", "showArchivedImportedInsights");
      this.applySummaryInterval(insight.content.interval);
      this.$refs["modal-import-summary"].hide();
    },
    showArchivedInsightsByBlockedIPsSummaryInsight(insight) {
      this.trackEvent(
        "HistoricalInsights",
        "showArchivedBlockedIPsSummaryInsights"
      );
      this.applySummaryInterval(insight.content.time_interval);
      this.$refs["modal-blockedips-summary"].hide();
    },
    seeMessageDetails(insight) {
      let params = {
        sender: insight.content.sender,
        recipient: insight.content.recipient,
        interval: insight.content.time_interval
      };

      this.$router.push({ name: "detective", params: params });
    },
    ...mapActions(["newAlertSuccess"])
  }
};

function formatInsightDescriptionDateTime(d) {
  // TODO: this should be formatted according to the chosen language
  return moment(d).format("DD MMM. (h:mmA)");
}

function formatInsightDescriptionDate(d) {
  // TODO: this should be formatted according to the chosen language
  return moment(d).format("MMM. D YYYY");
}

function formatDateForDetectiveInsightModalWindow(d) {
  // TODO: this should be formatted according to the chosen language
  return moment(d).format("DD MMM YYYY");
}
</script>

<style lang="less">
.insights .card.insight-highlighted {
  border: 2px solid #5ec4eb;
  box-shadow: 0px 0px 20px #0003;
}

.insights {
  margin-bottom: 1em;
  margin-top: 0.9rem;
  align-content: start;
}

.insights .col-card {
  margin-top: 1rem;
  align-content: start;
}

.insights .card {
  background: #ffffff 0% 0% no-repeat padding-box;
  box-shadow: 0px 0px 6px #0000001a;
  border: 1px solid #e6e7e7;
  border-radius: 5px;
}
.insights .card-text.category {
  background-color: #f2f2f2;
  border-radius: 18px;
  padding: 0.4em 1.5em;
  margin-bottom: 1.2em;
  width: min-content;
  height: 100%;
}
.insights .card-text.category,
.insights .card-text.time {
  color: #414445;
  font-size: 10px;
  font-weight: bold;
}
.insights .card-block {
  padding: 0.5rem 0.4rem 0.5rem 0rem;
  text-align: left;
}
.insights .card .rating {
  background-clip: content-box;
  background-color: #f0f8fc;
}
.insights .card-text {
  font: 12px/18px "Open Sans", sans-serif;
}
.insights .card-text.time {
  background: #f0f8fc 0% 0% no-repeat padding-box;
  border-radius: 5px;
  padding: 1.5em;
  .user-rating {
    font-size: 120%;
    .user-rating-smiley {
      margin-left: 0.25rem;
      cursor: pointer;
      color: #c3c3c3;
    }
    .user-rating-smiley-good {
      &:hover,
      &:active {
        color: #28a745;
      }
    }
    .user-rating-smiley-neutral {
      &:hover,
      &:active {
        color: #ffc107;
      }
    }
    .user-rating-smiley-bad {
      &:hover,
      &:active {
        color: #dc3545;
      }
    }
  }
}
.insights .card-title {
  margin: 0 0 0.3rem 0;
  font: 18px Inter;
  font-weight: bold;
  color: #202324;
  letter-spacing: 0px;
}
.insights .card .rating {
  background-clip: content-box;
  background-color: #f0f8fc;
}
.insights .card .rating.bad {
  background: rgb(255, 92, 111);
  background: linear-gradient(
      0deg,
      rgba(255, 92, 111, 1) 85%,
      rgba(230, 231, 231, 1) 85%
    )
    content-box;
}
.insights .card .rating.ok {
  background: rgb(255, 220, 0);
  background: linear-gradient(
      0deg,
      rgba(255, 220, 0, 1) 50%,
      rgba(230, 231, 231, 1) 50%
    )
    content-box;
}
.insights .card .rating.good {
  background: rgb(135, 197, 40);
  background: linear-gradient(
      0deg,
      rgba(135, 197, 40, 1) 15%,
      rgba(230, 231, 231, 1) 15%
    )
    content-box;
}
.insights .card svg {
  margin-right: 0.05em;
}
.insights .insight-actions svg {
  font-size: 1.3em;
  color: #c5c7c6;

  &:hover,
  &:active {
    color: #2c9cd6;
  }

  &.fa-times-circle {
    &:hover,
    &:active {
      color: #e67e22;
    }
  }
}

.insights .card-text.description button {
  padding: 0.4em 1.5em;
  margin-bottom: 0;
  background: #f9f9f9 0% 0% no-repeat padding-box;
  border-radius: 10px;
  font-size: 10px;
  font-weight: bold;
  color: #1d8caf;
  line-height: 1;
  border: 1px solid #e6e7e7;
  margin-left: 1em;
}

.insights .card-text.description button:hover {
  background-color: #1d8caf;
  color: #ffffff;
}

#modal-msg-rbl blockquote {
  font-style: italic;
  color: #555555;
  padding: 1em 30px 1em 0.6em;
  border-left: 8px solid #1d8caf;
  line-height: 1.3;
  position: relative;
  background: #f9f9f9;
}

#modal-msg-rbl blockquote::before {
  font-family: Inter;
  content: "\201C";
  color: #daebf4;
  font-size: 3em;
  position: absolute;
  left: 10px;
  top: -10px;
}

#modal-msg-rbl blockquote::after {
  content: "";
}

#modal-msg-rbl blockquote span {
  display: block;
  color: #333333;
  font-style: normal;
  margin-top: 1em;
  font-size: 0.7em;
}

#modal-msg-rbl .btn-cancel,
#modal-rbl-list .btn-cancel,
#modal-blockedips .btn-cancel,
#modal-import-summary .btn-cancel,
#modal-detective-escalation .btn-cancel {
  background: #ff5c6f33 0% 0% no-repeat padding-box;
  border: 1px solid #ff5c6f;
  border-radius: 2px;
  opacity: 0.8;
  text-align: center;
  font: normal normal bold 14px/24px Open Sans;
  letter-spacing: 0px;
  color: #820d1b;
}

#modal-msg-rbl .btn-cancel:hover,
#modal-rbl-list .btn-cancel:hover,
#modal-detective-escalation .btn-cancel:hover {
  color: #212529;
  text-decoration: none;
}

#modal-rbl-list .modal-body .intro {
  font-size: 0.7em;
}

#rbl-list-content .card {
  margin-bottom: 0.8em;
  background-color: #f9f9f9;
  border: 1px solid #c5c7c6;
}
.modal-content .card .card-title {
  font-size: 19px;
  font-weight: bold;
  color: #202324;
  font-family: Inter;
}

.modal-content .card .card-text {
  font-size: 15px;
}

#rbl-list-content .card .message-label {
  color: #7a82ab;
  font-weight: bold;
  padding-right: 0.5em;
  font-size: 15px;
}

#rbl-list-content .badge {
  font-size: 80%;
}

#rbl-list-content .badge-info {
  background-color: #fce4c4;
  color: #ff5c6f;
}

#rbl-list-content .badge-warning {
  background-color: #ff5c6f;
  color: white;
  margin-right: 0.6em;
}

.modal-content {
  border-radius: 0;
}

@media (max-width: 768px) {
  .modal-content {
    width: 100%;
  }
}

/* Note: workaround for vue bootstraps weird default modal footer handling */
.vue-modal-footer {
  padding-top: 0.75rem;
  border-top: 1px solid #dee2e6;
  text-align: right;
  margin-top: 1em;
}

.detective-insight-header {
  font-size: 15px;
}
</style>
