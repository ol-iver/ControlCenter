<!--
SPDX-FileCopyrightText: 2021 Lightmeter <hello@lightmeter.io>

SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
  <b-container class="mt-5 detective-body">
    <b-form
      ref="searchForm"
      @submit.prevent="
        page = 1;
        updateResults();
      "
      class="detective-form d-flex flex-wrap"
    >
      <div class="col p-2">
        <label>
          <translate>Sender Email Address</translate>
        </label>

        <b-form-input
          type="text"
          name="mail_from"
          maxlength="255"
          v-model="mail_from"
          :v-state="isEmailFrom"
          placeholder="sender@example.org"
        />
      </div>

      <div class="p-2 d-flex align-items-center" @click="swapEmails()">
        <i class="fas fa-exchange-alt"></i>
      </div>

      <div class="col p-2">
        <label>
          <translate>Recipient Email Address</translate>
        </label>

        <b-form-input
          type="text"
          name="mail_to"
          maxlength="255"
          v-model="mail_to"
          :v-state="isEmailTo"
          placeholder="recipient@example.org"
        />
      </div>

      <div class="col p-2">
        <label>
          <translate>Message ID</translate>
        </label>

        <b-form-input
          type="text"
          name="some_id"
          v-model="some_id"
          placeholder="Message/Queue ID"
        />
      </div>

      <div class="col p-2">
        <label>
          <translate>Sent Between</translate>
        </label>

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
        >
        </DateRangePicker>
      </div>

      <div class="col p-2">
        <label>
          <translate>Status</translate>
        </label>

        <select
          class="form-control custom-select"
          name="status"
          v-model="statusSelected"
        >
          <option value="-1"><translate>Any status</translate></option>
          <option value="0"><translate>Sent</translate></option>
          <option value="42"><translate>Received</translate></option>
          <option value="43"><translate>Replied</translate></option>
          <option value="1"><translate>Bounced</translate></option>
          <option value="2"><translate>Deferred</translate></option>
          <option value="3"><translate>Expired</translate></option>
          <option value="4"><translate>Returned</translate></option>
        </select>
      </div>

      <div class="col p-2 ml-auto">
        <b-button type="submit" variant="primary" class="btn-block">
          <translate>Search</translate>
        </b-button>
      </div>
    </b-form>

    <b-container ref="searchResultText" class="search-result-text mt-4">
      <p :class="searchResultClass">
        {{ searchResultText }}
        <b-button
          v-if="rawLogsEnabled"
          v-show="showLogsDownloadButton"
          v-on:click="downloadRawLogsInInterval()"
          variant="primary"
          size="sm"
          style="margin-left: 1rem;"
          v-b-tooltip.hover
          :title="titleDownloadLogsTimeInterval"
        >
          <i class="fas fa-download"></i>
          <translate>Logs</translate>
        </b-button>
        <b-button
          v-show="showLogsDownloadButton"
          v-on:click="exportCSV()"
          variant="secondary"
          size="sm"
          style="margin-left: 1rem;"
          v-b-tooltip.hover
          :title="titleExportCSV"
        >
          <i class="fas fa-download"></i>
          <translate>CSV</translate>
        </b-button>
      </p>
    </b-container>

    <detective-results
      :rawLogsEnabled="rawLogsEnabled"
      :results="results.messages"
      :showQueues="!forEndUsers"
      :showFromTo="!forEndUsers"
    ></detective-results>

    <b-container class="pages mt-4 mb-4" v-show="results.last_page > 1">
      <button
        type="button"
        class="btn btn-outline-primary"
        v-for="p in results.last_page"
        :key="p"
        :disabled="p == results.page"
        @click="
          page = p;
          updateResults();
        "
      >
        {{ p }}
      </button>
    </b-container>
  </b-container>
</template>

<script>
import axios from "axios";
axios.defaults.withCredentials = true;

import {
  checkMessageDelivery,
  escalateMessage,
  oldestAvailableTimeForMessageDetective,
  linkToRawLogsInInterval,
  getSettings
} from "@/lib/api.js";

import tracking from "@/mixin/global_shared.js";
import auth from "@/mixin/auth.js";
import datepicker from "@/mixin/datepicker.js";
import DateRangePicker from "vue2-daterange-picker";
import "vue2-daterange-picker/dist/vue2-daterange-picker.css";

function isEmail(forEndUsers, email) {
  if (!forEndUsers) return true;
  // NOTE: regexp also used in util/emailutil/email.go
  if (email == "") return null;
  return email.match(/^[^@\s]+@[^@\s]+$/) !== null;
}

export default {
  name: "detective",
  components: { DateRangePicker },
  mixins: [tracking, auth, datepicker],
  props: {
    rawLogsEnabled: {
      type: Boolean,
      default: true
    },
    forEndUsers: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      // detective-specific
      mail_from: "",
      mail_to: "",
      searchResultText: this.$gettext("No results yet"),
      searchResultClass: "text-muted",
      results: [],
      statusSelected: "-1",
      some_id: "",
      page: 1,

      // specific auth
      neededAuth: this.forEndUsers ? "detective" : "auth"

      // TODO: restrict timeInterval to 1 day if forEndUsers?
    };
  },
  computed: {
    isEmailFrom: function() {
      return isEmail(this.forEndUsers, this.mail_from);
    },
    isEmailTo: function() {
      return isEmail(this.forEndUsers, this.mail_to);
    },
    showLogsDownloadButton: function() {
      if (this.forEndUsers || this.results.length == 0) {
        return false;
      }

      return this.results.messages.length > 0;
    },
    titleDownloadLogsTimeInterval: function() {
      return this.$gettext(
        "Download mail server logs for selected time interval"
      );
    },
    titleExportCSV: function() {
      return this.$gettext("Export search results to a CSV file");
    }
  },
  methods: {
    swapEmails() {
      let temp = this.mail_from;
      this.mail_from = this.mail_to;
      this.mail_to = temp;
    },
    updateSelectedInterval(obj) {
      let vue = this;
      vue.formatDatePickerValue(obj);
    },
    onUpdateDateRangePicker: function(obj) {
      this.trackEvent(
        "MessageDetectiveDatePicker",
        obj.startDate + "-" + obj.endDate
      );

      this.updateSelectedInterval(obj);
    },
    updateResults: function() {
      let vue = this;

      if (!vue.some_id && (!vue.isEmailFrom || !vue.isEmailTo)) {
        vue.searchResultClass = "text-warning";
        vue.searchResultText = vue.$gettext(
          "Please check the given email addresses"
        );
        return;
      }

      vue.searchResultClass = "text-muted";
      vue.searchResultText = "...";

      let interval = vue.buildDateInterval();

      checkMessageDelivery(
        vue.mail_from,
        vue.mail_to,
        interval.startDate,
        interval.endDate,
        vue.statusSelected,
        vue.some_id,
        vue.page,
        false
      ).then(function(response) {
        vue.results = response.data;

        vue.trackEvent(
          "MessageDetectiveSearch" + (vue.forEndUsers ? "EndUser" : "Admin"),
          vue.results.total
        );

        vue.$emit(
          "onResults",
          response.data,
          vue.mail_from,
          vue.mail_to,
          interval,
          vue.some_id
        );

        let pageNb =
          vue.page > 1 ? " - " + vue.$gettext("Page") + " " + vue.page : "";

        vue.searchResultClass = vue.results.total
          ? "text-primary"
          : "text-secondary";
        vue.searchResultText = vue.results.total
          ? new Intl.NumberFormat().format(vue.results.total) +
            " " +
            vue.$gettext("message(s) found") +
            pageNb
          : vue.$gettext("No message found");
        vue.$refs.searchForm.scrollIntoView();
      });
    },
    escalateMessage() {
      let interval = this.buildDateInterval();
      escalateMessage(
        this.mail_from,
        this.mail_to,
        interval.startDate,
        interval.endDate
      ).then(function() {
        console.log("All good");
      });
    },
    downloadRawLogsInInterval() {
      let interval = this.buildDateInterval();
      let link = linkToRawLogsInInterval(interval.startDate, interval.endDate);
      window.open(link);
    },
    exportCSV() {
      let vue = this;
      let interval = this.buildDateInterval();

      let link = checkMessageDelivery(
        vue.mail_from,
        vue.mail_to,
        interval.startDate,
        interval.endDate,
        vue.statusSelected,
        vue.some_id,
        vue.page,
        true
      );
      window.open(link);
    }
  },
  mounted() {
    let vue = this;

    let runSearch = false;
    let dateFromParams = false;

    let hash = window.location.hash.split("?");
    if (hash.length > 1) {
      let get = hash[1];
      let params = {};
      get.split("&").forEach(function(gp) {
        let [k, v] = gp.split("=");
        params[k] = decodeURIComponent(v);
      });

      ["mail_from", "mail_to", "statusSelected", "some_id"].forEach(function(
        k
      ) {
        if (params[k]) {
          vue[k] = params[k];
          runSearch = true;
        }
      });

      if (params.startDate && params.endDate) {
        vue.dateRange = {
          startDate: params.startDate,
          endDate: params.endDate
        };
        dateFromParams = true;
        runSearch = true;
      }
    }

    this.updateSelectedInterval(this.dateRange);

    if (runSearch) {
      vue.updateResults();
    }

    if (!dateFromParams) {
      oldestAvailableTimeForMessageDetective().then(r => {
        if (r.data.time != null) {
          this.dateRange = {
            startDate: r.data.time,
            endDate: this.dateRange.endDate
          };
          this.updateSelectedInterval(this.dateRange);
        }
      });
    }

    // FIXME: this code is in the wrong place and only increasing the technical debt.
    // this component should not depend on any external value other than its props!!!
    getSettings().then(function(response) {
      vue.rawLogsEnabled = !response.data.feature_flags.disable_raw_logs;
    });
  }
};
</script>

<style lang="less">
/* don't squeeze the inputs or datepicker too much, so they'll flex-wrap on smaller screens */
input,
.vue-daterange-picker {
  min-width: 180px !important;
  display: block !important;
}

button.btn-primary {
  background-color: #227aaf;
}

.pages {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;

  button {
    margin-top: 0.5em;
    & + button {
      margin-left: 0.5em;
    }
  }
}

.detective-form {
}

.detective-form label {
  display: none;
}

.detective-form .col {
}

.detective-body {
  padding-right: 0px;
  padding-left: 0px;
  margin-bottom: 20px;
}
</style>
