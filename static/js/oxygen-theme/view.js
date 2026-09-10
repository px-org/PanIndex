(function () {
  "use strict";

  var T = window.PanOxygenTheme || window.PanVueTheme;
  if (!T || typeof Vue === "undefined") return;

  function createViewShell(data, contentSelector) {
    var prefix = (data.config && data.config.path_prefix) || "";
    var file = data.file || {};
    var initialDark = T.resolveDark(data.theme);

    return Vue.createApp({
      components: { Icon: T.Icon },
      data: function () {
        return {
          title: data.title || "",
          siteName: (data.config && data.config.site_name) || "",
          path: data.path || "/",
          fullPath: data.full_path || "/",
          accounts: data.accounts || [],
          account: data.account || {},
          prePaths: data.pre_paths || [],
          config: data.config || {},
          hasPwd: !!data.has_pwd,
          pwdErr: data.pwd_err_msg || "",
          pwdPath: data.pwd_path || "",
          password: "",
          hasParent: !!data.has_parent,
          parentPath: data.parent_path || "/",
          theme: data.theme || "oxygen",
          isDark: initialDark,
          isAdmin: !!data.is_admin_login,
          access: (data.config && data.config.access) || "0",
          file: file,
          lastFile: data.last_file || "",
          nextFile: data.next_file || "",
          showInfo: (T.cookieGet("show_info") || "0") === "1",
          showAccountMenu: false,
          showShareMenu: false,
          shareQr: "",
          shareUrl: "",
          fileLink: "",
        };
      },
      computed: {
        displayTitle: function () {
          return this.siteName || this.title || "PanIndex";
        },
        canNavigate: function () {
          return this.isAdmin || this.access === "0" || this.access === "3";
        },
        showAccountSwitcher: function () {
          return this.accounts && this.accounts.length > 1;
        },
        breadcrumbPaths: function () {
          var list = this.prePaths || [];
          var singleDefault = this.accounts.length === 1 && this.config.account_choose === "default";
          if (singleDefault) {
            return list.filter(function (_, i) {
              return i > 0;
            });
          }
          return list;
        },
      },
      mounted: function () {
        T.applyDarkClass(this.isDark);
        document.addEventListener("click", this.onDocClick);
        this.fileLink = encodeURI(location.protocol + "//" + location.host + prefix + (this.file.path || ""));
        var slot = document.querySelector(contentSelector || "#oxygen-preview-content");
        var target = this.$refs.contentHost;
        if (slot && target) {
          while (slot.firstChild) {
            target.appendChild(slot.firstChild);
          }
          slot.remove();
        }
      },
      beforeUnmount: function () {
        document.removeEventListener("click", this.onDocClick);
      },
      methods: {
        onDocClick: function () {
          this.showAccountMenu = false;
          this.showShareMenu = false;
        },
        url: function (p) {
          return prefix + (p || "");
        },
        goHome: function () {
          location.href = this.url("/") || "/";
        },
        goPath: function (pathUrl, isLast) {
          var u = this.url(pathUrl);
          if (isLast && !this.file.is_folder) u += "?v";
          location.href = u;
        },
        goParent: function () {
          location.href = this.url(this.parentPath);
        },
        toggleTheme: function () {
          this.isDark = !this.isDark;
          T.applyDarkClass(this.isDark);
          var next = this.isDark ? "oxygen-dark" : "oxygen-light";
          this.theme = next;
          T.cookieSet("theme", next, 3650);
        },
        toggleInfo: function () {
          this.showInfo = !this.showInfo;
          T.cookieSet("show_info", this.showInfo ? "1" : "0", 3650);
        },
        download: function () {
          location.href = this.url(this.file.path);
        },
        copyLink: function () {
          T.copyText(this.fileLink);
        },
        submitPwd: function () {
          var pwd = (this.password || "").trim();
          if (!pwd || pwd.length >= 30) return;
          var fullPath = this.pwdPath;
          var result = T.cookieGet("file_pwd");
          var ppwd = T.md5Hash(fullPath) + ":" + pwd;
          if (result && result !== "undefined" && result !== "null") {
            var values = result.split(",");
            var ok = false;
            for (var i = 0; i < values.length; i++) {
              if (values[i].indexOf(T.md5Hash(fullPath) + ":") === 0) {
                values[i] = ppwd;
                result = values.join(",");
                ok = true;
                break;
              }
            }
            if (!ok) result += "," + ppwd;
            T.cookieSet("file_pwd", result, 3650);
          } else {
            T.cookieSet("file_pwd", ppwd, 3650);
          }
          location.reload();
        },
        openShare: function (ev) {
          if (ev) ev.stopPropagation();
          this.showShareMenu = !this.showShareMenu;
          if (!this.showShareMenu) return;
          var formData = new FormData();
          var adminPath = this.config.admin_path || "/admin";
          formData.append("prefix", location.protocol + "//" + location.host + prefix + "/s/");
          formData.append("path", this.fullPath);
          formData.append("isFile", "1");
          var self = this;
          fetch(prefix + "/api/v3" + adminPath + "/short/info", {
            method: "POST",
            body: formData,
          })
            .then(function (r) {
              return r.json();
            })
            .then(function (d) {
              self.shareQr = d.qr_code || "";
              self.shareUrl = d.short_url || "";
            })
            .catch(function () {
              T.toast("获取分享信息失败");
            });
        },
        copyShare: function () {
          if (this.shareUrl) T.copyText(this.shareUrl);
        },
        year: function () {
          return new Date().getFullYear();
        },
      },
      template:
        '<div class="vt-container">' +
        '  <div class="vt-card">' +
        '    <div class="px-3 pt-3 pb-2 flex flex-wrap items-center gap-1">' +
        '      <div class="relative inline-flex">' +
        '        <div class="vt-chip" @click.stop="showAccountSwitcher ? (showAccountMenu=!showAccountMenu) : goHome()">' +
        '          <span class="vt-chip-icon"><Icon name="home" :size="16" /></span>' +
        '          <span class="vt-chip-title" @click.stop="goHome()">{{ displayTitle }}</span>' +
        "        </div>" +
        '        <div v-if="showAccountMenu" class="vt-menu left-0 top-10" @click.stop>' +
        '          <a v-for="a in accounts" :key="a.name" class="vt-menu-item" :href="url(\'/\' + a.name)">' +
        '            <Icon name="user" :size="16" /> {{ a.name }}' +
        "          </a>" +
        "        </div>" +
        "      </div>" +
        '      <template v-if="canNavigate" v-for="(bp, idx) in breadcrumbPaths" :key="bp.PathUrl || idx">' +
        '        <Icon name="chevron-right" :size="16" class-name="text-[var(--vt-muted)]" />' +
        '        <div class="vt-chip" @click="goPath(bp.PathUrl, idx === breadcrumbPaths.length - 1)">' +
        '          <span class="vt-chip-title" :title="bp.PathName">{{ bp.PathName }}</span>' +
        "        </div>" +
        "      </template>" +
        "    </div>" +
        '    <div class="flex flex-wrap items-center gap-1 px-2 py-2 border-t border-[var(--vt-border)]">' +
        '      <button v-if="hasParent && canNavigate" class="vt-icon-btn" title="返回上级目录" @click="goParent"><Icon name="arrow-left" /></button>' +
        '      <button v-if="!hasPwd" class="vt-icon-btn" :class="{\'is-active\': showInfo}" title="文件信息" @click="toggleInfo"><Icon name="info" /></button>' +
        '      <div v-if="isAdmin" class="relative">' +
        '        <button class="vt-icon-btn" title="分享" @click="openShare"><Icon name="share-2" /></button>' +
        '        <div v-if="showShareMenu" class="vt-menu vt-share-menu left-0 top-10 w-52 p-3 text-center" @click.stop>' +
        '          <img v-if="shareQr" :src="shareQr" class="w-full h-auto block rounded-md mb-2" alt="qrcode" />' +
        '          <button class="px-3 py-1.5 rounded-md text-white text-sm" style="background:var(--vt-primary)" @click="copyShare">复制短链接</button>' +
        "        </div>" +
        "      </div>" +
        '      <a v-if="!hasPwd" id="view_down_link" class="vt-icon-btn" title="点击下载" :data-path="url(file.path)" :href="url(file.path)"><Icon name="download" /></a>' +
        '      <a v-if="lastFile && canNavigate" :href="url(lastFile) + \'?v\'" class="vt-icon-btn" title="上一个"><Icon name="chevron-left" /></a>' +
        '      <a v-if="nextFile && canNavigate" :href="url(nextFile) + \'?v\'" class="vt-icon-btn" title="下一个"><Icon name="chevron-right" /></a>' +
        '      <button class="vt-icon-btn ml-auto" title="明暗主题切换" @click="toggleTheme"><Icon :name="isDark ? \'sun\' : \'moon\'" /></button>' +
        "    </div>" +
        '    <div v-if="hasPwd" class="vt-pwd">' +
        '      <div class="flex items-center gap-2 mb-2 text-[var(--vt-muted)]"><Icon name="lock" /> 请输入密码</div>' +
        '      <div class="flex items-center gap-2">' +
        '        <input id="input-password" v-model="password" type="password" placeholder="密码" :data-file-path="pwdPath" @keydown.enter="submitPwd" />' +
        '        <button class="vt-icon-btn is-active" title="确认" @click="submitPwd"><Icon name="fingerprint" /></button>' +
        "      </div>" +
        '      <div v-if="pwdErr" class="text-red-500 text-sm mt-2">{{ pwdErr }}</div>' +
        "    </div>" +
        '    <template v-else>' +
        '      <div v-show="showInfo" class="vt-info-panel">' +
        "        <div><b>名称：</b>{{ file.file_name }}</div>" +
        "        <div><b>大小：</b>{{ file.size_fmt }}</div>" +
        "        <div><b>日期：</b>{{ file.last_op_time }}</div>" +
        '        <div class="flex items-center gap-1 flex-wrap"><b>链接：</b>' +
        '          <a id="file_link" href="javascript:void(0)" :data-path="url(file.path)" @click="download">{{ fileLink }}</a>' +
        '          <button class="vt-icon-btn" title="复制链接" @click="copyLink"><Icon name="copy" :size="14" /></button>' +
        "        </div>" +
        "      </div>" +
        '      <div class="vt-preview-body" ref="contentHost"></div>' +
        "    </template>" +
        "  </div>" +
        '  <div class="vt-footer" v-html="config.footer || (\'©\' + year() + \' <a href=&quot;https://github.com/px-org/PanIndex&quot; target=&quot;_blank&quot;>PanIndex</a>. All rights reserved.\')"></div>' +
        "</div>",
    });
  }

  T.mountView = function (raw) {
    var files = T.normalizeFiles(raw.fns || raw.files || []);
    var data = {
      title: raw.title || "",
      path: raw.path || "/",
      full_path: raw.full_path || "/",
      accounts: T.normalizeAccounts(raw.accounts),
      account: T.normalizeAccount(raw.account),
      pre_paths: raw.pre_paths || [],
      config: T.publicConfig(raw.config),
      has_pwd: raw.has_pwd,
      pwd_err_msg: raw.pwd_err_msg || "",
      pwd_path: raw.pwd_path || "",
      has_parent: raw.has_parent,
      parent_path: raw.parent_path || "/",
      theme: raw.theme || "oxygen",
      is_admin_login: raw.is_admin_login,
      last_file: raw.last_file || "",
      next_file: raw.next_file || "",
      file: files[0] || {},
    };
    createViewShell(data, "#oxygen-preview-content").mount("#app");
  };
})();
