<!--
SPDX-FileCopyrightText: 2021 Lightmeter <hello@lightmeter.io>

SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
  <button
    v-if="hasInterval(insight)"
    v-on:click="downloadRawLogsInInterval(insight)"
    class="btn btn-sm"
  >
    <i class="fas fa-download" style="margin-right: 0.25rem;"></i>
    <translate>Logs</translate>
  </button>
</template>

<script>
import { linkToRawLogsInInterval } from "@/lib/api";
import moment from "moment";
import tracking from "../../mixin/global_shared.js";

export default {
  mixins: [tracking],
  props: {
    insight: Object
  },
  methods: {
    getInterval(insight) {
      return insight.content.time_interval
        ? insight.content.time_interval
        : insight.content.interval;
    },
    hasInterval(insight) {
      let interval = this.getInterval(insight);
      return interval && interval.from && interval.to;
    },
    downloadRawLogsInInterval(insight) {
      let interval = this.getInterval(insight);

      let link = linkToRawLogsInInterval(
        moment.utc(interval.from).format("YYYY-MM-DD HH:mm:ss"),
        moment.utc(interval.to).format("YYYY-MM-DD HH:mm:ss")
      );

      this.trackEvent("DownloadInsightLogs", insight.content_type);

      window.open(link);
    }
  }
};
</script>
