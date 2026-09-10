(function (global) {
  var PanAdmin = global.PanAdmin || (global.PanAdmin = {});
  var Vue = global.Vue;

  var toasts = Vue ? Vue.reactive([]) : [];
  var toastId = 1;

  function toast(message, ms) {
    ms = ms == null ? 2200 : ms;
    var id = toastId++;
    toasts.push({ id: id, message: String(message || "") });
    setTimeout(function () {
      var idx = toasts.findIndex(function (t) {
        return t.id === id;
      });
      if (idx >= 0) toasts.splice(idx, 1);
    }, ms);
  }

  PanAdmin.toast = toast;
  PanAdmin.toasts = toasts;

  PanAdmin.ToastHost = {
    name: "ToastHost",
    setup: function () {
      return { toasts: toasts };
    },
    template:
      '<div class="fixed bottom-4 right-4 z-[80] flex flex-col gap-2 max-w-sm">' +
      '  <div v-for="t in toasts" :key="t.id" class="admin-toast-enter admin-card px-4 py-3 text-sm shadow-lg">' +
      "    {{ t.message }}" +
      "  </div>" +
      "</div>",
  };

  PanAdmin.Modal = {
    name: "PaModal",
    props: {
      open: { type: Boolean, default: false },
      title: { type: String, default: "" },
      width: { type: String, default: "max-w-lg" },
    },
    emits: ["close"],
    mounted: function () {
      var self = this;
      this._onKey = function (e) {
        if (e.key === "Escape" && self.open) self.$emit("close");
      };
      document.addEventListener("keydown", this._onKey);
    },
    beforeUnmount: function () {
      document.removeEventListener("keydown", this._onKey);
    },
    template:
      '<div v-if="open" class="fixed inset-0 z-[70] flex items-center justify-center p-4">' +
      '  <div class="admin-modal-mask absolute inset-0" @click="$emit(\'close\')"></div>' +
      '  <div class="admin-card relative w-full overflow-hidden" :class="width">' +
      '    <div class="flex items-center justify-between px-5 py-4 border-b" style="border-color: var(--admin-border)">' +
      '      <h3 class="text-base font-semibold">{{ title }}</h3>' +
      '      <button type="button" class="admin-btn admin-btn-ghost !p-2" @click="$emit(\'close\')">' +
      '        <pa-icon name="x" :size="16"></pa-icon>' +
      "      </button>" +
      "    </div>" +
      '    <div class="px-5 py-4 max-h-[70vh] overflow-auto">' +
      "      <slot></slot>" +
      "    </div>" +
      '    <div v-if="$slots.footer" class="px-5 py-3 border-t flex flex-wrap gap-2 justify-end" style="border-color: var(--admin-border)">' +
      "      <slot name=\"footer\"></slot>" +
      "    </div>" +
      "  </div>" +
      "</div>",
  };

  PanAdmin.PageHeader = {
    name: "PageHeader",
    props: {
      title: String,
      subtitle: { type: String, default: "" },
    },
    template:
      '<div class="mb-5">' +
      '  <h2 class="text-xl font-semibold tracking-tight">{{ title }}</h2>' +
      '  <p v-if="subtitle" class="mt-1 text-sm" style="color: var(--admin-muted)">{{ subtitle }}</p>' +
      "</div>",
  };

  PanAdmin.Field = {
    name: "PaField",
    props: {
      label: String,
      help: { type: String, default: "" },
    },
    template:
      '<div class="mb-4">' +
      '  <label class="admin-label">{{ label }}</label>' +
      "  <slot></slot>" +
      '  <div v-if="help" class="admin-help" v-html="help"></div>' +
      "</div>",
  };

  PanAdmin.Tooltip = {
    name: "PaTooltip",
    props: {
      content: { type: String, required: true },
      placement: { type: String, default: "right" },
    },
    template:
      '<span class="admin-tooltip-wrap">' +
      "  <slot></slot>" +
      '  <span class="admin-tooltip-bubble" :class="{\'is-bottom\': placement===\'bottom\', \'is-top\': placement===\'top\'}">{{ content }}</span>' +
      "</span>",
  };

  PanAdmin.Switch = {
    name: "PaSwitch",
    props: {
      modelValue: { default: false },
      trueValue: { default: true },
      falseValue: { default: false },
      label: { type: String, default: "" },
      onLabel: { type: String, default: "" },
      offLabel: { type: String, default: "" },
    },
    emits: ["update:modelValue"],
    computed: {
      isOn: function () {
        return this.modelValue === this.trueValue || this.modelValue === true;
      },
      displayLabel: function () {
        if (this.onLabel || this.offLabel) {
          return this.isOn ? this.onLabel || this.label : this.offLabel || this.label;
        }
        return this.label;
      },
    },
    methods: {
      toggle: function () {
        this.$emit(
          "update:modelValue",
          this.isOn ? this.falseValue : this.trueValue
        );
      },
    },
    template:
      '<button type="button" class="admin-switch" :class="{\'is-on\': isOn}" role="switch" :aria-checked="isOn ? \'true\' : \'false\'" @click="toggle">' +
      '  <span class="admin-switch-track"><span class="admin-switch-thumb"></span></span>' +
      '  <span v-if="displayLabel" class="admin-switch-label">{{ displayLabel }}</span>' +
      "</button>",
  };

  PanAdmin.RadioGroup = {
    name: "PaRadioGroup",
    props: {
      modelValue: { default: "" },
      options: { type: Array, default: function () { return []; } },
    },
    emits: ["update:modelValue"],
    methods: {
      pick: function (val) {
        this.$emit("update:modelValue", val);
      },
      isActive: function (val) {
        return String(this.modelValue) === String(val);
      },
    },
    template:
      '<div class="admin-radio-group" role="radiogroup">' +
      '  <button v-for="opt in options" :key="String(opt.value)" type="button" class="admin-radio-pill" :class="{\'is-active\': isActive(opt.value)}" role="radio" :aria-checked="isActive(opt.value) ? \'true\' : \'false\'" @click="pick(opt.value)">{{ opt.label }}</button>' +
      "</div>",
  };

  PanAdmin.Check = {
    name: "PaCheck",
    props: {
      modelValue: { type: Boolean, default: false },
    },
    emits: ["update:modelValue"],
    methods: {
      toggle: function (e) {
        if (e) {
          e.preventDefault();
          e.stopPropagation();
        }
        this.$emit("update:modelValue", !this.modelValue);
      },
    },
    template:
      '<button type="button" class="admin-check" :class="{\'is-on\': modelValue}" role="checkbox" :aria-checked="modelValue ? \'true\' : \'false\'" @click="toggle">' +
      '  <pa-icon v-if="modelValue" name="check" :size="12"></pa-icon>' +
      "</button>",
  };

  PanAdmin.createSelection = function () {
    return {
      selected: [],
      isSelected: function (id) {
        return this.selected.indexOf(id) >= 0;
      },
      toggleSelect: function (id) {
        var i = this.selected.indexOf(id);
        if (i >= 0) this.selected.splice(i, 1);
        else this.selected.push(id);
      },
      clearSelection: function () {
        this.selected = [];
      },
      selectedIds: function () {
        return this.selected.slice();
      },
    };
  };

  PanAdmin.setCookie = function (name, value, days) {
    var maxAge = (days || 365) * 24 * 60 * 60;
    document.cookie =
      name +
      "=" +
      encodeURIComponent(value) +
      "; path=/; max-age=" +
      maxAge +
      "; SameSite=Lax";
  };

  PanAdmin.getCookie = function (name) {
    var m = document.cookie.match(new RegExp("(^| )" + name + "=([^;]+)"));
    return m ? decodeURIComponent(m[2]) : "";
  };

  PanAdmin.applyTheme = function (theme) {
    var dark = false;
    if (theme === "dark") dark = true;
    else if (theme === "light") dark = false;
    else
      dark =
        window.matchMedia &&
        window.matchMedia("(prefers-color-scheme: dark)").matches;
    document.documentElement.classList.toggle("dark", dark);
    return dark;
  };

  PanAdmin.formatExpire = function (ts) {
    if (!ts || Number(ts) <= 0) return "永久";
    var d = new Date(Number(ts) * 1000);
    if (isNaN(d.getTime())) return String(ts);
    var pad = function (n) {
      return n < 10 ? "0" + n : "" + n;
    };
    return (
      d.getFullYear() +
      "-" +
      pad(d.getMonth() + 1) +
      "-" +
      pad(d.getDate()) +
      " " +
      pad(d.getHours()) +
      ":" +
      pad(d.getMinutes()) +
      ":" +
      pad(d.getSeconds())
    );
  };

  PanAdmin.datetimeLocalToUnix = function (val) {
    if (!val) return 0;
    var d = new Date(val);
    if (isNaN(d.getTime())) return 0;
    return Math.floor(d.getTime() / 1000);
  };

  PanAdmin.unixToDatetimeLocal = function (ts) {
    if (!ts || Number(ts) <= 0) return "";
    var d = new Date(Number(ts) * 1000);
    if (isNaN(d.getTime())) return "";
    var pad = function (n) {
      return n < 10 ? "0" + n : "" + n;
    };
    return (
      d.getFullYear() +
      "-" +
      pad(d.getMonth() + 1) +
      "-" +
      pad(d.getDate()) +
      "T" +
      pad(d.getHours()) +
      ":" +
      pad(d.getMinutes())
    );
  };

  PanAdmin.refreshConfig = async function (store) {
    var cfg = await PanAdmin.api.getConfig();
    store.config = cfg;
    return cfg;
  };
})(window);
