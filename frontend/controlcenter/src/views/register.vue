<!--
SPDX-FileCopyrightText: 2021 Lightmeter <hello@lightmeter.io>

SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
  <panel-page>
    <div id="registration-page">
      <h2>
        <translate>Welcome</translate>
      </h2>
      <p class="align-left" render-html="true" v-translate>
        Please create an admin account. This creates a user ID on lightmeter.io
        too. %{openHelpLink}Get help%{closeHelpLink} to avoid repeating this
        step if you've done it before.
      </p>

      <div class="field-group">
        <h4>
          <translate>User details</translate>
        </h4>
        <b-form @submit.stop.prevent="onSubmit">
          <b-form-group>
            <b-form-input
              name="name"
              id="name"
              v-model="form.name"
              type="text"
              required
              aria-describedby="nameHelp"
              :placeholder="NameInputPlaceholder"
              maxlength="255"
            ></b-form-input>
            <b-form-input
              name="email"
              id="email"
              v-model="form.email"
              type="email"
              required
              aria-describedby="emailHelp"
              :placeholder="EmailInputPlaceholder"
              maxlength="255"
            ></b-form-input>
            <b-input-group id="show_hide_password">
              <b-form-input
                name="password"
                id="password"
                v-model="form.password"
                type="password"
                required
                aria-describedby="passwordHelp"
                :placeholder="PasswordInputPlaceholder"
                maxlength="255"
              ></b-form-input>
              <div class="input-group-addon" v-on:click="onTogglePasswordShow">
                <a href=""><i class="fa fa-eye" aria-hidden="true"></i></a>
              </div>

              <b-form-checkbox
                id="subscribe_newsletter"
                v-model="form.subscribe_newsletter"
                name="subscribe_newsletter"
                value="on"
                unchecked-value="off"
                class="custom-form-check-label"
              >
                <translate>Subscribe to newsletter</translate>
              </b-form-checkbox>
            </b-input-group>
            <b-form-group>
              <h4><translate>System details</translate></h4>
              <div class="input-group">
                <b-form-input
                  name="postfix_public_ip"
                  id="postfixPublicIP"
                  v-model="$v.form.postfix_public_ip.$model"
                  type="text"
                  aria-describedby="publicIPHelp"
                  :placeholder="PostfixPublicIPInputPlaceholder"
                  maxlength="255"
                  :state="validateState('postfix_public_ip')"
                ></b-form-input>
                <div class="input-group-append">
                  <button
                    class="btn btn-outline-secondary"
                    type="button"
                    data-toggle="tooltip"
                    data-placement="top"
                    v-b-tooltip.hover
                    :title="PublicIPHelpText"
                  >
                    <i class="far fa-question-circle"></i>
                  </button>
                </div>
              </div>
              <b-form-invalid-feedback>
                <translate
                  >The Ip Address is invalid</translate
                ></b-form-invalid-feedback
              >
            </b-form-group>
          </b-form-group>
          <b-button variant="primary" class="w-100" type="submit">
            <translate>Register</translate>
          </b-button>
        </b-form>
        <div class="card info">
          <div class="card-body">
            <h5 class="card-title">
              <i class="fa fa-info-circle"></i>

              <translate class="text-blue">Terms of service</translate>
            </h5>

            <p class="card-text" v-translate render-html="true">
              %{openPeerNetworkLink}Peer networking%{closePeerNetworkLink}
              features require exchanging
              %{openPrivacyLink}data%{closePrivacyLink} with lightmeter.io
              servers. By registering you agree to the %{openTosLink}Terms of
              Service%{closeTosLink}. Feature usage data is
              %{openPrivacyLink}collected%{closePrivacyLink} to improve your
              experience.
            </p>
          </div>
        </div>
      </div>
    </div>
    <b-toast
      id="progress-toast"
      :visible="!isImportProgressFinished"
      :title="progressIndicatorTitle"
      toaster="b-toaster-bottom-right progress-indicator-toast"
      no-auto-hide
      no-close-button
    >
      <template #toast-title>
        <span class="progress-toast-title">
          <translate>Generating Insights</translate>
        </span>
        <span class="progress-toast-collapse">
          <b-icon v-b-toggle.collapse-progress icon="arrows-collapse"></b-icon>
        </span>
      </template>
      <b-collapse visible id="collapse-progress">
        <div class="collapse-body">
          <import-progress-indicator
            :showLabel="false"
            @finished="handleProgressFinished"
          ></import-progress-indicator>
        </div>
      </b-collapse>
    </b-toast>
  </panel-page>
</template>

<script>
import { submitRegisterForm } from "../lib/api.js";
import { togglePasswordShow } from "../lib/util.js";
import { mapState, mapActions } from "vuex";
import { ipAddress } from "vuelidate/lib/validators";
import shared_texts from "../mixin/shared_texts.js";

import linkify from "vue-linkify";
import Vue from "vue";

Vue.directive("linkified", linkify);

export default {
  name: "register",
  components: {},
  mixins: [shared_texts],
  data() {
    return {
      form: {
        email: "",
        password: "",
        name: ``,
        subscribe_newsletter: null,
        postfix_public_ip: ""
      }
    };
  },
  validations: {
    form: {
      postfix_public_ip: {
        ipAddress
      }
    }
  },
  computed: {
    openPrivacyLink() {
      return `<a target="_blank" href="https://lightmeter.io/privacy-policy/">`;
    },
    closePrivacyLink() {
      return `</a>`;
    },
    openTosLink() {
      return `<a target="_blank" href="https://lightmeter.io/terms/">`;
    },
    closeTosLink() {
      return `</a>`;
    },
    openPeerNetworkLink() {
      return `<a target="_blank" href="https://lightmeter.io/network-faq/">`;
    },
    closePeerNetworkLink() {
      return `</a>`;
    },
    openHelpLink() {
      return `<a href="https://gitlab.com/lightmeter/controlcenter#upgrade"><translate class="get-help">`;
    },
    closeHelpLink() {
      return `</a>`;
    },
    PostfixPublicIPInputPlaceholder() {
      return this.$gettext("Mail server public IP");
    },
    NameInputPlaceholder: function() {
      return this.$gettext("Name");
    },
    EmailInputPlaceholder: function() {
      return this.$gettext("Email");
    },
    PasswordInputPlaceholder: function() {
      return this.$gettext("Password");
    },
    progressIndicatorTitle() {
      return this.$gettext(`Generating Insights`);
    },
    ...mapState(["language", "isImportProgressFinished"])
  },
  methods: {
    validateState(name) {
      const { $dirty, $error } = this.$v.form[name];
      return $dirty ? !$error : null;
    },
    onSubmit(event) {
      this.$v.form.$touch();
      if (this.$v.form.$anyError) {
        return;
      }
      event.preventDefault();
      let vue = this;

      let settingsData = {
        email: this.form.email,
        app_language: this.language,
        postfix_public_ip: this.form.postfix_public_ip
      };

      if (this.form.subscribe_newsletter === "on") {
        settingsData.subscribe_newsletter = this.form.subscribe_newsletter;
      }

      const registrationData = {
        email: this.form.email,
        name: this.form.name,
        password: this.form.password
      };

      const redirect = () => {
        vue.$router.push({ name: "index" });
      };

      submitRegisterForm(registrationData, settingsData, redirect);
    },
    onTogglePasswordShow(event) {
      event.preventDefault();
      togglePasswordShow(event);
    },
    handleProgressFinished() {
      this.setInsightsImportProgressFinished({ wait: 3 });
    },
    ...mapActions(["setInsightsImportProgressFinished"])
  },
  mounted() {
    const el = document.body;
    el.classList.add("login-gradient");
    this.$bvToast.show("progress-toast");
  },
  destroyed() {
    const el = document.body;
    el.classList.remove("login-gradient");
  }
};
</script>

<style lang="less">
#registration-page .card {
  margin-top: 1em;
  text-align: left;
}

#registration-page .get-help {
  margin-left: 0.2em;
  margin-right: 0.2em;
}

#registration-page .card .card-text {
  color: #00689d;
  font-size: 12px;
}

#registration-page .card.info {
  background: #daebf4;
  border: none;
}

#registration-page .card.info .card-title {
  font-size: 12px;
  letter-spacing: 0px;
  font-weight: bold;
  margin-bottom: 0.8em;
  color: #00689d;
}

#registration-page .card.info .card-title .text-blue {
  color: #00689d;
  margin-left: 0.2em;
}

#registration-page .card.info .card-body {
  padding: 0.8em;
}

#registration-page .card .fa {
  padding-right: 0.8em;
}
#auth-page-footer .container .sub-container {
  margin: 0 auto;
}

#registration-page .btn-primary:hover {
  color: #fff;
  background-color: #0069d9;
  border-color: #0062cc;
}

.custom-control-label {
  padding-top: 0.25rem;
  font-size: 14px;
}

.progress-indicator-toast .b-toaster-slot {
  max-width: 200px !important;
}

.toast-header {
  display: flex;
  flex-flow: row;
  justify-content: space-between;
}

.toast-body {
  padding: 0 !important;
}

.collapse-body {
  padding: 1.7em;
}

/* Position toast above language select menu to avoid obscuring it */
@media (max-width: 768px) {
  .b-toaster-slot {
    bottom: 3.2rem !important;
  }
}
</style>
