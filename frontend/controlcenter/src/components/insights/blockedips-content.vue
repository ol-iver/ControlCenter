<!--
SPDX-FileCopyrightText: 2021 Lightmeter <hello@lightmeter.io>

SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
  <div>
    <div v-html="header" class="blockedips-header"></div>
    <table class="table blockedips-summary-table">
      <caption>
        <strong>{{ content.top_ips.length }}</strong>
        most active IPs out of
        <strong>{{ content.total_ips }}</strong>
      </caption>
      <thead class="thead">
        <th scope="col"><translate>IP Address</translate></th>
        <th scope="col"><translate>Connection attempts</translate></th>
      </thead>
      <tbody>
        <tr v-for="rec in content.top_ips" v-bind:key="rec.addr">
          <td>
            {{ rec.addr }}
          </td>
          <td>
            {{ new Intl.NumberFormat().format(rec.count) }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script>
import moment from "moment";
import tracking from "../../mixin/global_shared.js";

export default {
  mixins: [tracking],
  props: {
    content: Object
  },
  updated() {},
  mounted() {},
  data() {
    return {};
  },
  methods: {
    formatTableDate(time) {
      return moment(time).format("DD MMM YYYY");
    },
    formatTableTime(time) {
      return moment(time).format("h:mmA");
    }
  },
  computed: {
    header() {
      let translation = this.$gettext(
        `List of banned IPs on <strong>%{from}</strong> due to potential attacks (brute-force, slow-force, botnets):`
      );
      return this.$gettextInterpolate(translation, {
        from: this.formatTableDate(this.content.time_interval.from),
        to: this.formatTableDate(this.content.time_interval.to)
      });
    }
  }
};
</script>

<style scoped lang="less">
* {
  font-size: 15px;
}

.blockedips-summary-table .thead th {
  background-color: #5f689a;
  color: #ffffff;
}

.blockedips-summary-table tr td {
  background: #f9f9f9 0% 0% no-repeat padding-box;
  .time::before {
    content: " | ";
  }
  @media (max-width: 768px) {
    .time {
      display: none;
    }
  }
}

.blockedips-header strong {
  background-color: #e8e9f0;
  border-radius: 11px;
  padding-left: 0.5em;
  padding-right: 0.5em;
}

.blockedips-header {
  margin-bottom: 20px;
}
</style>
