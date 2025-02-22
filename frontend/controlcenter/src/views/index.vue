<!--
SPDX-FileCopyrightText: 2021 Lightmeter <hello@lightmeter.io>

SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
  <div id="insights-page" class="d-flex flex-column min-vh-100">
    <mainheader></mainheader>
    <div class="container main-content">
      <div class="row align-items-center">
        <div class="col-7 col-sm-8 col-md-9 col-lg-10">
          <h1 class="row-title">
            <translate>Observatory</translate>
          </h1>
        </div>
        <div class="col-5 col-sm-4 col-md-3 col-lg-2" v-if="!simpleViewEnabled">
          <a
            :href="FeedbackMailtoLink"
            :title="FeedbackButtonTitle"
            class="btn btn-sm btn-block btn-outline-info"
            ><i
              class="fas fa-star"
              style="margin-right: 0.25rem;"
              data-toggle="tooltip"
              data-placement="bottom"
            ></i>
            <translate>Feedback</translate>
          </a>
        </div>
      </div>

      <div class="row">
        <div class="col-md-12">
          <div class="panel panel-default greeting">
            <div class="row">
              <div class="col-md-3 align-center">
                <img
                  class="hero"
                  src="@/assets/greeting-observatory.svg"
                  alt="Observatory illustration"
                />
              </div>

              <div class="col-md-9 d-flex align-items-center">
                <div class="row">
                  <div class="container">
                    <h3>{{ greetingText }}</h3>
                    <p>{{ welcomeUserText }}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <b-toaster
        ref="statusMessage"
        name="statusMessage"
        class="status-message"
        v-if="!simpleViewEnabled"
      >
      </b-toaster>

      <div
        class="row container time-interval card-section-heading sticky-date-select"
      >
        <div class="col-lg-6 col-md-6 col-9 p-2 d-flex">
          <label class="col-md-2 col-form-label sr-only">
            <translate>Time interval</translate>:
          </label>
          <div class="p-1">
            <DateRangePicker
              @update="onUpdateDateRangePicker"
              :autoApply="autoApply"
              :opens="opens"
              :singleDatePicker="singleDatePicker"
              :alwaysShowCalendars="alwaysShowCalendars"
              :ranges="ranges"
              v-model="dateRange"
              :showCustomRangeCalendars="false"
              :max-date="new Date()"
              v-b-tooltip.hover.left
              :title="titleDatepicker"
            >
            </DateRangePicker>
          </div>
          <div class="p-1" v-if="rawLogsIsEnabled">
            <b-button
              variant="primary"
              size="sm"
              @click="downloadRawLogsInInterval"
              :disabled="rawLogsDownloadsDisable"
              v-b-tooltip.hover
              :title="titleDownloadLogs"
              ><i class="fas fa-download" style="margin-right: 0.25rem;"></i
              ><translate>Logs</translate></b-button
            >
          </div>
        </div>
      </div>

      <maindashboard
        :graphDateRange="dashboardInterval"
        v-if="dashboardV2Enabled && !simpleViewEnabled"
      ></maindashboard>

      <maindashboard-simple-view
        :graphDateRange="dashboardInterval"
        v-if="dashboardV2Enabled && simpleViewEnabled"
      ></maindashboard-simple-view>

      <graphdashboard
        :graphDateRange="dashboardInterval"
        v-if="dashboardV1Enabled"
      ></graphdashboard>

      <div
        class="row container d-flex align-items-center card-section-heading"
        v-if="insightsEnabled"
      >
        <div class="col-lg-2 col-md-2 col-3 p-2">
          <h2 class="insights-title">
            <translate>Insights</translate>
          </h2>
        </div>
        <div class="col-lg-4 col-md-4 col-9 ml-auto p-2">
          <form id="insights-form">
            <div
              class="form-group d-flex justify-content-end align-items-center"
            >
              <label class="sr-only">
                <translate>Filter</translate>
              </label>
              <select
                id="insights-filter"
                class="form-control custom-select custom-select-sm"
                name="filter"
                form="insights-form"
                v-model="insightsFilter"
                v-on:change="updateInsights"
              >
                <!-- todo remove in style -->
                <option
                  selected
                  v-on:click="
                    trackClick('InsightsFilterCategoryHomepage', 'Active')
                  "
                  value="category-active"
                >
                  <translate>Active</translate>
                </option>
                <option value="nofilter">
                  <translate>All</translate>
                </option>
                <!--    " -->
                <option
                  v-on:click="
                    trackClick('InsightsFilterCategoryHomepage', 'Local')
                  "
                  value="category-local"
                >
                  <translate>Local</translate>
                </option>
                <option
                  v-on:click="
                    trackClick('InsightsFilterCategoryHomepage', 'News')
                  "
                  value="category-news"
                >
                  <translate>News</translate>
                </option>
                <option
                  v-on:click="
                    trackClick('InsightsFilterCategoryHomepage', 'Intel')
                  "
                  value="category-intel"
                >
                  <translate>Intel</translate>
                </option>
                <option
                  v-on:click="
                    trackClick('InsightsFilterCategoryHomepage', 'Archived')
                  "
                  value="category-archived"
                >
                  <translate>Archived</translate>
                </option>
              </select>
              <select
                id="insights-sort"
                class="form-control custom-select custom-select-sm"
                name="order"
                form="insights-form"
                v-model="insightsSort"
                v-on:change="updateInsights"
              >
                <!-- todo remove in style -->
                <option
                  v-on:click="
                    trackClick('InsightsFilterOrderHomepage', 'Newest')
                  "
                  selected
                  value="creationDesc"
                >
                  <translate>Newest</translate>
                </option>
                <option
                  v-on:click="
                    trackClick('InsightsFilterOrderHomepage', 'Oldest')
                  "
                  value="creationAsc"
                >
                  <translate>Oldest</translate>
                </option>
              </select>
            </div>
          </form>
        </div>
      </div>

      <insights
        v-if="insightsEnabled"
        class="row"
        v-show="shouldShowInsights"
        :insights="insights"
        @dateIntervalChanged="handleExternalDateIntervalChanged"
      ></insights>

      <import-progress-indicator
        :label="generatingInsights"
        @finished="handleProgressFinished"
        v-if="insightsViewEnabled"
      ></import-progress-indicator>
    </div>
    <mainfooter></mainfooter>
  </div>
</template>

<script>
import axios from "axios";
axios.defaults.withCredentials = true;

import {
  linkToRawLogsInInterval,
  countLogLinesInInterval,
  fetchInsights,
  getIsNotLoginOrNotRegistered,
  getUserInfo,
  getStatusMessage,
  getSettings
} from "../lib/api.js";

import tracking from "../mixin/global_shared.js";
import shared_texts from "../mixin/shared_texts.js";
import auth from "../mixin/auth.js";
import datepicker from "@/mixin/datepicker.js";
import { mapActions, mapState } from "vuex";
import DateRangePicker from "vue2-daterange-picker";
import "vue2-daterange-picker/dist/vue2-daterange-picker.css";

export default {
  name: "insight",
  components: { DateRangePicker },
  mixins: [tracking, shared_texts, auth, datepicker],
  data() {
    return {
      username: "",
      updateDashboardAndInsightsIntervalID: null,
      dashboardInterval: this.buildDefaultInterval(),
      insightsFilter: "category-active",
      insightsSort: "creationDesc",
      insights: [],

      // log import progress
      generatingInsights: this.$gettext("Generating insights"),

      rawLogsDownloadsDisable: true,

      statusMessage: null,
      statusMessageId: null,
      dashboardV2IsEnabled: false,
      dashboardV1IsEnabled: false,
      simpleViewEnabled: true,
      insightsViewEnabled: false,
      rawLogsEnabled: true
    };
  },
  created() {},
  computed: {
    titleDatepicker() {
      return this.$gettext(
        "Choose date interval - applies to all graphs and insights"
      );
    },
    titleDownloadLogs() {
      return this.$gettext("Download server logs for selected date interval");
    },
    shouldShowInsights() {
      return this.isImportProgressFinished;
    },
    dashboardV2Enabled() {
      return this.dashboardV2IsEnabled;
    },
    dashboardV1Enabled() {
      return this.dashboardV1IsEnabled;
    },
    insightsEnabled() {
      return this.insightsViewEnabled;
    },
    rawLogsIsEnabled() {
      return this.rawLogsEnabled;
    },
    greetingText() {
      // todo use better translate function for weekdays
      let dateObj = new Date();
      let weekday = dateObj.toLocaleString("default", { weekday: "long" });
      let translation = this.$gettext("Happy %{weekday}");
      let message = this.$gettextInterpolate(translation, { weekday: weekday });
      return message;
    },
    welcomeUserText() {
      let translation = this.$gettext("and welcome back, %{username}");
      let message = this.$gettextInterpolate(translation, {
        username: this.username
      });
      return message;
    },
    ...mapState(["isImportProgressFinished"])
  },
  methods: {
    updateSelectedInterval(obj) {
      let vue = this;
      vue.updateDashboardAndInsights();
      vue.formatDatePickerValue(obj);
      vue.updateRawLogsDownloadButton();
    },
    handleProgressFinished() {
      this.setInsightsImportProgressFinished();
      this.updateDashboardAndInsights();
    },
    updateDashboardAndInsights() {
      let vue = this;
      vue.updateInsights();
      vue.updateDashboard();

      if (!vue.simpleViewEnabled) {
        vue.updateStatusMessage();
      }
    },
    handleExternalDateIntervalChanged(obj) {
      if (obj === undefined) {
        this.updateSelectedInterval(this.dateRange);
        return;
      }

      this.dateRange = obj;
      this.insightsFilter = "category-" + obj.category;
      this.updateSelectedInterval(obj);
    },
    onUpdateDateRangePicker: function(obj) {
      this.trackEvent(
        "onUpdateDateRangePicker",
        obj.startDate + "-" + obj.endDate
      );

      this.updateSelectedInterval(obj);
    },
    updateRawLogsDownloadButton: function() {
      let vue = this;
      let interval = vue.buildDateInterval();

      countLogLinesInInterval(interval.startDate, interval.endDate).then(
        function(response) {
          vue.rawLogsDownloadsDisable = response.data.count == 0;
        }
      );
    },
    downloadRawLogsInInterval() {
      let interval = this.buildDateInterval();
      let link = linkToRawLogsInInterval(interval.startDate, interval.endDate);
      let range = interval.startDate + "_" + interval.endDate;

      this.trackEvent("DownloadDatePickerLogs", range);

      window.open(link);
    },
    onStatusMessageClosed() {
      this.trackEvent("CloseStatusMessage", this.statusMessageId);
    },
    updateStatusMessage: function() {
      let vue = this;

      getStatusMessage().then(function(response) {
        let notification =
          response.data !== null ? response.data.notification : null;

        if (notification === null || notification.title == "") {
          return;
        }

        let id = response.data.id;

        let isNew =
          vue.statusMessage === null ||
          vue.statusMessage.message != notification.message ||
          vue.statusMessage.title != notification.title;

        vue.statusMessage = notification;
        vue.statusMessageId = id;

        if (!isNew) {
          return;
        }

        const e = vue.$createElement;

        const msg = [
          e("p", vue.statusMessage.message),
          e(
            "a",
            { attrs: { href: vue.statusMessage.action.link } },
            vue.statusMessage.action.label
          )
        ];

        vue.$bvToast.toast([msg], {
          variant: vue.statusMessage.severity,
          title: vue.statusMessage.title,
          noAutoHide: true,
          toaster: "statusMessage",
          solid: true
        });
      });
    },
    updateInsights: function() {
      let vue = this;
      let interval = vue.buildDateInterval();

      fetchInsights(
        interval.startDate,
        interval.endDate,
        vue.insightsFilter,
        vue.insightsSort
      ).then(function(response) {
        vue.insights = response.data;
      });
    },
    updateDashboard: function() {
      let vue = this;
      let interval = vue.buildDateInterval();
      vue.dashboardInterval = interval;
    },
    initIndex: function() {
      let vue = this;

      vue.updateSelectedInterval(vue.dateRange);

      this.updateDashboardAndInsightsIntervalID = window.setInterval(
        function() {
          getIsNotLoginOrNotRegistered().then(vue.updateDashboardAndInsights);
        },
        30000
      );
    },
    ...mapActions(["setInsightsImportProgressFinished"])
  },
  mounted() {
    this.initIndex();

    let vue = this;

    vue.$root.$on("bv::toast:hidden", event => {
      vue.onStatusMessageClosed(event);
    });

    getUserInfo().then(function(response) {
      vue.username = response.data.user.name;
    });

    getSettings().then(function(response) {
      vue.simpleViewEnabled = response.data.feature_flags.enable_simple_view;

      vue.dashboardV1IsEnabled = !response.data.feature_flags
        .disable_v1_dashboard;
      vue.dashboardV2IsEnabled =
        response.data.feature_flags.enable_v2_dashboard;
      vue.insightsViewEnabled = !response.data.feature_flags
        .disable_insights_view;
      vue.rawLogsEnabled = !response.data.feature_flags.disable_raw_logs;
    });
  },
  destroyed() {
    window.clearInterval(this.updateDashboardAndInsightsIntervalID);
  }
};
</script>

<style lang="less">
.sticky-date-select {
  position: sticky;
  top: 0;
  z-index: 50;
}

#insights-page .greeting h3 {
  font: 22px/32px Inter;
  font-weight: bold;
  margin: 0;
  text-align: left;
  color: white;
}

#insights-page .greeting p {
  text-align: left;
}

#insights-page .card-section-heading {
  background-color: #f9f9f9;
}

#insights-page .time-interval {
  margin: 0.6rem 0 0 0;
  border-radius: 10px;
}

#insights-page .card-section-heading h2 {
  font-size: 24px;
  font-weight: bold;
  margin: 0;
}

#insights-page .time-interval .form-group {
  margin: 0;
  padding: 0;
}

#insights-page #insights-form select {
  font-size: 12px;
  border-radius: 5px;
  margin-right: 0.2rem;
}

#insights-page .form-control.custom-select {
  margin: 0;
  background-color: #e6e7e7;
  color: #202324;
}

#insights-page .greeting {
  background: url(~@/assets/greeting-lensflare.svg) no-repeat right top,
    linear-gradient(104deg, #2a93d6 0%, #3dd9d6 100%) 0% 0% padding-box;
  color: white;
  padding: 0.5rem;
  border-radius: 7px;
  margin-bottom: 30px;
}

#insights-page h1.row-title {
  font-size: 32px;
  font-weight: bold;
  margin: 0.7em 0 0.8em 0;
  text-align: left;
}

#insights-page .vue-daterange-picker .reportrange-text {
  background: #daebf4;
  cursor: pointer;
  padding: 0.3rem 1rem;
  border: none;
  font-size: 12px;
  color: #00689d;
  font-weight: bold;
  border-radius: 5px;
  text-align: center;
  cursor: pointer;
  display: flex;
  justify-content: center;

  span {
    order: 1;
    margin-top: 0.25em;
  }
  svg {
    order: 2;
    margin-left: 1em;
    margin-top: 0.45em;
  }
}

#insights-page .insights-title {
  text-align: left;
}

@media (min-width: 768px) {
  #insights-page .greeting {
    height: 150px;
  }
}

@media (max-width: 768px) {
  #insights-page #insights {
    min-height: 100vh;
  }
}

@media (min-width: 768px) and (max-width: 1024px) {
  #insights-page #insights {
    min-height: 60vh;
  }
}

.progress-indicator-area {
  margin-top: 60px;
  margin-bottom: 60px;
}

.b-toaster.status-message {
  max-width: 100%;
  width: 100%;

  .b-toast,
  .toast {
    max-width: 100%;
    width: 100%;
    flex-basis: 100%;
    margin-top: 1rem;
    margin-bottom: 1.9rem;
  }
  .toast-body {
    text-align: left;
    > * {
      margin: 1em;
      display: block;
    }
  }
}
</style>
