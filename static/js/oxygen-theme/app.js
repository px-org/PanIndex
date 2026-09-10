(function () {
  "use strict";

  function cookieGet(name) {
    var m = document.cookie.match(new RegExp("(?:^|; )" + name.replace(/([.$?*|{}()[\]\\/+^])/g, "\\$1") + "=([^;]*)"));
    return m ? decodeURIComponent(m[1]) : undefined;
  }

  function cookieSet(name, value, days) {
    var expires = "";
    if (days) {
      var d = new Date();
      d.setTime(d.getTime() + days * 24 * 60 * 60 * 1000);
      expires = "; expires=" + d.toUTCString();
    }
    document.cookie = name + "=" + encodeURIComponent(value) + expires + "; path=/";
  }

  function toPascal(name) {
    return String(name || "")
      .split("-")
      .map(function (p) {
        return p.charAt(0).toUpperCase() + p.slice(1);
      })
      .join("");
  }

  function fileIconName(file) {
    if (file.is_folder) return "folder-open";
    var vt = file.view_type || "";
    if (vt === "img") return "image";
    if (vt === "audio") return "music";
    if (vt === "video") return "film";
    if (vt === "code" || vt === "md") return "file-code";
    if (vt === "pdf" || vt === "office" || vt === "epub") return "file-text";
    var t = (file.file_type || "").toLowerCase();
    if (t === "apk") return "smartphone";
    if (t === "exe") return "app-window";
    if ("zip,gz,7z,rar".indexOf(t) >= 0) return "file-archive";
    return "file";
  }

  function resolveDark(theme) {
    if (theme === "oxygen-dark" || theme === "vue-dark") return true;
    if (theme === "oxygen-light" || theme === "vue-light") return false;
    return window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches;
  }

  function applyDarkClass(isDark) {
    document.documentElement.classList.toggle("dark", !!isDark);
  }

  function toast(msg) {
    var el = document.createElement("div");
    el.className = "vt-toast";
    el.textContent = msg;
    document.body.appendChild(el);
    setTimeout(function () {
      el.remove();
    }, 1800);
  }

  function copyText(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      return navigator.clipboard.writeText(text).then(function () {
        toast("链接已复制到剪切板");
      });
    }
    var ta = document.createElement("textarea");
    ta.value = text;
    document.body.appendChild(ta);
    ta.select();
    document.execCommand("copy");
    ta.remove();
    toast("链接已复制到剪切板");
    return Promise.resolve();
  }

  function md5Hash(str) {
    if (typeof md5 === "function") return md5(str);
    if (typeof window.md5 === "function") return window.md5(str);
    return str;
  }

  var Icon = {
    name: "VtIcon",
    props: {
      name: { type: String, required: true },
      size: { type: [Number, String], default: 20 },
      stroke: { type: [Number, String], default: 2 },
      className: { type: String, default: "" },
    },
    mounted: function () {
      this.renderIcon();
    },
    updated: function () {
      this.renderIcon();
    },
    methods: {
      renderIcon: function () {
        var root = this.$refs.root;
        if (!root) return;
        root.innerHTML = "";
        var key = toPascal(this.name);
        if (window.lucide && lucide.icons && lucide.icons[key]) {
          var iconNode = lucide.icons[key];
          var svg;
          if (typeof lucide.createElement === "function") {
            svg = lucide.createElement(iconNode);
          } else if (Array.isArray(iconNode)) {
            svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
            svg.setAttribute("viewBox", "0 0 24 24");
            svg.setAttribute("fill", "none");
            svg.setAttribute("stroke", "currentColor");
            svg.setAttribute("stroke-linecap", "round");
            svg.setAttribute("stroke-linejoin", "round");
            iconNode.forEach(function (item) {
              if (!Array.isArray(item) || item.length < 2) return;
              var el = document.createElementNS("http://www.w3.org/2000/svg", item[0]);
              var attrs = item[1] || {};
              Object.keys(attrs).forEach(function (k) {
                el.setAttribute(k, attrs[k]);
              });
              svg.appendChild(el);
            });
          }
          if (svg) {
            svg.setAttribute("width", String(this.size));
            svg.setAttribute("height", String(this.size));
            svg.setAttribute("stroke-width", String(this.stroke));
            if (this.className) svg.setAttribute("class", this.className);
            root.appendChild(svg);
            return;
          }
        }
        root.setAttribute("data-lucide", this.name);
        if (window.lucide && typeof lucide.createIcons === "function") {
          lucide.createIcons({ nodes: [root], attrs: { width: this.size, height: this.size, "stroke-width": this.stroke } });
        }
      },
    },
    template: '<span ref="root" class="inline-flex items-center justify-center"></span>',
  };

  function createListingApp(data) {
    var prefix = (data.config && data.config.path_prefix) || "";
    var initialDark = resolveDark(data.theme);

    return Vue.createApp({
      components: { Icon: Icon },
      data: function () {
        return {
          title: data.title || "",
          siteName: (data.config && data.config.site_name) || "",
          path: data.path || "/",
          fullPath: data.full_path || "/",
          accountPath: data.account_path || "",
          accounts: data.accounts || [],
          account: data.account || {},
          files: data.files || [],
          prePaths: data.pre_paths || [],
          config: data.config || {},
          hasPwd: !!data.has_pwd,
          pwdErr: data.pwd_err_msg || "",
          pwdPath: data.pwd_path || "",
          password: "",
          hasParent: !!data.has_parent,
          parentPath: data.parent_path || "/",
          searchKey: data.search_key || "",
          searchOpen: !!(data.search_key && data.search_key.length),
          layout: data.layout || "view_comfy",
          theme: data.theme || "oxygen",
          isDark: initialDark,
          isAdmin: !!data.is_admin_login,
          version: data.version || "",
          headHtml: "",
          readmeHtml: "",
          showAccountMenu: false,
          showSortMenu: false,
          showShareMenu: false,
          shareQr: "",
          shareUrl: "",
          toastMsg: "",
          showFab: false,
          loading: false,
        };
      },
      computed: {
        displayTitle: function () {
          return this.siteName || this.title || "PanIndex";
        },
        isGrid: function () {
          return this.layout !== "view_comfy";
        },
        empty: function () {
          return !this.hasPwd && (!this.files || this.files.length === 0);
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
        applyDarkClass(this.isDark);
        var self = this;
        this._onScroll = function () {
          self.showFab = window.scrollY > 240;
        };
        window.addEventListener("scroll", this._onScroll);
        window.addEventListener("popstate", this.onPopState);
        document.addEventListener("click", this.onDocClick);
        if (!this.hasPwd && !this.searchKey) {
          this.loadMarkdown("HEAD.md", "head");
          this.loadMarkdown("README.md", "readme");
        }
      },
      beforeUnmount: function () {
        document.removeEventListener("click", this.onDocClick);
        window.removeEventListener("popstate", this.onPopState);
        if (this._onScroll) window.removeEventListener("scroll", this._onScroll);
        if (this._navAbort) this._navAbort.abort();
      },
      methods: {
        onDocClick: function () {
          this.showAccountMenu = false;
          this.showSortMenu = false;
          this.showShareMenu = false;
        },
        url: function (p) {
          return prefix + (p || "");
        },
        fileUrl: function (file) {
          return this.url(file.path);
        },
        iconOf: function (file) {
          return fileIconName(file);
        },
        pathFromLocation: function () {
          var path = location.pathname || "/";
          if (prefix && path.indexOf(prefix) === 0) {
            path = path.slice(prefix.length) || "/";
          }
          if (!path) path = "/";
          if (path !== "/" && path.charAt(path.length - 1) === "/") {
            path = path.slice(0, -1);
          }
          return path;
        },
        updateDocumentTitle: function () {
          var base = this.siteName || this.title || "PanIndex";
          if (this.searchKey) {
            document.title = base + " 搜索:" + this.searchKey;
          } else {
            document.title = base + " " + (this.path || "/");
          }
        },
        applyIndexPayload: function (data, opts) {
          opts = opts || {};
          data = data || {};
          this.path = data.path || "/";
          this.fullPath = data.full_path || data.fullPath || "/";
          this.prePaths = data.pre_paths || data.prePaths || [];
          this.hasParent = !!(data.has_parent || data.hasParent);
          this.parentPath = data.parent_path || data.parentPath || "/";
          this.accountPath = data.account_path || data.accountPath || this.accountPath;
          this.files = normalizeFiles(data.content || data.fns || data.files || []);
          if (data.account) this.account = normalizeAccount(data.account);
          if (data.title) this.title = data.title;
          this.hasPwd = !!opts.hasPwd;
          this.pwdErr = opts.pwdErr || "";
          this.pwdPath = data.pwd_path || data.pwdPath || this.pwdPath || "";
          this.password = "";
          if (opts.searchKey !== undefined) {
            this.searchKey = opts.searchKey;
            if (opts.searchKey) this.searchOpen = true;
          }
          this.updateDocumentTitle();
        },
        refreshMarkdown: function () {
          this.headHtml = "";
          this.readmeHtml = "";
          if (this.hasPwd || this.searchKey) return;
          this.loadMarkdown("HEAD.md", "head");
          this.loadMarkdown("README.md", "readme");
        },
        navigateTo: function (fullPath, opts) {
          opts = opts || {};
          var self = this;
          fullPath = fullPath || "/";
          if (fullPath !== "/" && fullPath.charAt(fullPath.length - 1) === "/") {
            fullPath = fullPath.slice(0, -1);
          }

          if (this._navAbort) this._navAbort.abort();
          this._navAbort = typeof AbortController !== "undefined" ? new AbortController() : null;
          this.loading = true;
          this.showAccountMenu = false;
          this.showSortMenu = false;
          this.showShareMenu = false;

          var formData = new FormData();
          formData.append("path", fullPath);
          var sortCol = cookieGet("sort_column");
          var sortOrd = cookieGet("sort_order");
          // Match SSR SortCheck: only override when cookies exist; otherwise server uses config defaults.
          if (sortCol) formData.append("sort_by", sortCol);
          if (sortOrd) formData.append("order", sortOrd);

          var fetchOpts = {
            method: "POST",
            body: formData,
            credentials: "same-origin",
          };
          if (this._navAbort) fetchOpts.signal = this._navAbort.signal;

          return fetch(prefix + "/api/v3/public/index", fetchOpts)
            .then(function (r) {
              return r.json();
            })
            .then(function (resp) {
              var data = resp.data || {};
              if (resp.status === 0 && data.is_folder === false) {
                location.href = self.url(data.full_path || fullPath) + "?v";
                return;
              }
              self.applyIndexPayload(data, {
                hasPwd: resp.status === 403,
                pwdErr: resp.msg || "",
                searchKey: "",
              });
              var nextUrl = self.url(self.fullPath || fullPath);
              if (opts.replace) {
                history.replaceState({ spa: 1, path: self.fullPath }, "", nextUrl);
              } else {
                history.pushState({ spa: 1, path: self.fullPath }, "", nextUrl);
              }
              if (!opts.keepScroll) {
                window.scrollTo({ top: 0, behavior: "smooth" });
              }
              self.refreshMarkdown();
            })
            .catch(function (err) {
              if (err && err.name === "AbortError") return;
              toast("加载失败，请重试");
            })
            .finally(function () {
              self.loading = false;
            });
        },
        runSearch: function (key, opts) {
          opts = opts || {};
          var self = this;
          if (this._navAbort) this._navAbort.abort();
          this._navAbort = typeof AbortController !== "undefined" ? new AbortController() : null;
          this.loading = true;
          this.showAccountMenu = false;
          this.showSortMenu = false;
          this.showShareMenu = false;

          var formData = new FormData();
          formData.append("key", key);
          var fetchOpts = {
            method: "POST",
            body: formData,
            credentials: "same-origin",
          };
          if (this._navAbort) fetchOpts.signal = this._navAbort.signal;

          return fetch(prefix + "/api/v3/public/search", fetchOpts)
            .then(function (r) {
              return r.json();
            })
            .then(function (resp) {
              var data = resp.data || {};
              self.files = normalizeFiles(data.content || []);
              self.hasPwd = false;
              self.pwdErr = "";
              self.searchKey = key;
              self.searchOpen = true;
              self.headHtml = "";
              self.readmeHtml = "";
              self.updateDocumentTitle();
              var nextUrl = (prefix || "") + "/?search=" + encodeURIComponent(key);
              if (opts.replace) {
                history.replaceState({ spa: 1, search: key }, "", nextUrl);
              } else {
                history.pushState({ spa: 1, search: key }, "", nextUrl);
              }
              if (!opts.keepScroll) {
                window.scrollTo({ top: 0, behavior: "smooth" });
              }
            })
            .catch(function (err) {
              if (err && err.name === "AbortError") return;
              toast("搜索失败，请重试");
            })
            .finally(function () {
              self.loading = false;
            });
        },
        onPopState: function () {
          var params = new URLSearchParams(location.search || "");
          var search = params.get("search");
          if (search) {
            this.runSearch(search, { replace: true, keepScroll: true });
          } else {
            this.navigateTo(this.pathFromLocation(), { replace: true, keepScroll: true });
          }
        },
        toggleTheme: function () {
          this.isDark = !this.isDark;
          applyDarkClass(this.isDark);
          var next = this.isDark ? "oxygen-dark" : "oxygen-light";
          this.theme = next;
          cookieSet("theme", next, 3650);
        },
        toggleLayout: function () {
          this.layout = this.isGrid ? "view_comfy" : "view_list";
          cookieSet("layout", this.layout, 3650);
        },
        setSort: function (column, order) {
          cookieSet("sort_column", column, 3650);
          cookieSet("sort_order", order, 3650);
          this.showSortMenu = false;
          if (this.searchKey) return;
          this.navigateTo(this.fullPath || this.pathFromLocation(), { replace: true, keepScroll: true });
        },
        resetSort: function () {
          cookieSet("sort_column", "default", 3650);
          cookieSet("sort_order", "null", 3650);
          this.showSortMenu = false;
          if (this.searchKey) return;
          this.navigateTo(this.fullPath || this.pathFromLocation(), { replace: true, keepScroll: true });
        },
        goHome: function () {
          this.navigateTo("/");
        },
        goPath: function (pathUrl) {
          this.navigateTo(pathUrl || "/");
        },
        goParent: function () {
          this.navigateTo(this.parentPath || "/");
        },
        openFile: function (file, ev) {
          if (ev && ev.target && ev.target.closest && ev.target.closest("[data-stop]")) return;
          if (file.is_folder) {
            this.navigateTo(file.path);
            return;
          }
          var u = this.fileUrl(file);
          if (String(this.config.enable_preview) === "0") {
            location.href = u;
          } else {
            location.href = u + "?v";
          }
        },
        copyFileLink: function (file) {
          var full = location.protocol + "//" + location.host + this.fileUrl(file);
          copyText(encodeURI(full));
        },
        copyAllLinks: function () {
          var self = this;
          var urls = (this.files || [])
            .filter(function (f) {
              return !f.is_folder;
            })
            .map(function (f) {
              return encodeURI(location.protocol + "//" + location.host + self.fileUrl(f));
            });
          if (!urls.length) {
            toast("没有可复制的文件链接");
            return;
          }
          copyText(urls.join("\n"));
        },
        previewImages: function () {
          var imgs = (this.files || []).filter(function (f) {
            return f.view_type === "img";
          });
          if (!imgs.length) {
            toast("本页没有图片");
            return;
          }
          var list = document.getElementById("image-preview-list");
          if (!list || typeof Viewer === "undefined") {
            location.href = this.fileUrl(imgs[0]) + "?v";
            return;
          }
          list.innerHTML = "";
          imgs.forEach(function (f) {
            var img = document.createElement("img");
            img.setAttribute("data-original", prefix + f.path);
            img.alt = f.file_name;
            list.appendChild(img);
          });
          var viewer = new Viewer(list, {
            url: "data-original",
            hidden: function () {
              viewer.destroy();
            },
            title: function (image) {
              return image.alt + " (" + (this.index + 1) + "/" + this.length + ")";
            },
          });
          viewer.show();
        },
        onSearchKey: function (e) {
          if (e.key !== "Enter") return;
          var key = (this.searchKey || "").trim();
          if (!key || key.length >= 30) return;
          this.runSearch(key);
        },
        submitPwd: function () {
          var pwd = (this.password || "").trim();
          if (!pwd || pwd.length >= 30) return;
          var fullPath = this.pwdPath;
          var result = cookieGet("file_pwd");
          var ppwd = md5Hash(fullPath) + ":" + pwd;
          if (result && result !== "undefined" && result !== "null") {
            var values = result.split(",");
            var ok = false;
            for (var i = 0; i < values.length; i++) {
              var val = values[i];
              if (val.indexOf(md5Hash(fullPath) + ":") === 0) {
                values[i] = ppwd;
                result = values.join(",");
                ok = true;
                break;
              }
            }
            if (!ok) result += "," + ppwd;
            cookieSet("file_pwd", result, 3650);
          } else {
            cookieSet("file_pwd", ppwd, 3650);
          }
          this.navigateTo(this.fullPath || this.pathFromLocation(), { replace: true });
        },
        loadMarkdown: function (name, kind) {
          if (kind === "head" && String(this.config.head) !== "1") return;
          if (kind === "readme" && String(this.config.readme) !== "1") return;
          var p = this.path === "/" ? this.path + name : this.path + "/" + name;
          var fullUrl = encodeURI(
            location.protocol + "//" + location.host + prefix + "/api/v3/public/raw" + this.accountPath + p
          );
          var key = md5Hash(fullUrl);
          var cached = localStorage.getItem(key);
          var self = this;

          function isMissingPayload(text) {
            if (!text || typeof text !== "string") return true;
            var t = text.trim();
            if (!t) return true;
            if (t.charAt(0) !== "{") return false;
            try {
              var obj = JSON.parse(t);
              return !!(obj && obj.status);
            } catch (e) {
              return false;
            }
          }

          function apply(text) {
            if (isMissingPayload(text) || typeof parseMarkdown !== "function") return false;
            var html = parseMarkdown(text);
            if (!html || !String(html).trim()) return false;
            if (kind === "head") self.headHtml = html;
            else self.readmeHtml = html;
            return true;
          }

          if (cached) {
            if (isMissingPayload(cached)) localStorage.removeItem(key);
            else apply(cached);
          }

          fetch(fullUrl)
            .then(function (r) {
              if (!r.ok) throw new Error("missing");
              return r.text();
            })
            .then(function (text) {
              if (isMissingPayload(text)) {
                localStorage.removeItem(key);
                if (kind === "head") self.headHtml = "";
                else self.readmeHtml = "";
                return;
              }
              localStorage.setItem(key, text);
              apply(text);
            })
            .catch(function () {
              localStorage.removeItem(key);
            });
        },
        openShare: function (ev) {
          if (ev) ev.stopPropagation();
          this.showShareMenu = !this.showShareMenu;
          if (!this.showShareMenu) return;
          var formData = new FormData();
          var adminPath = this.config.admin_path || "/admin";
          formData.append("prefix", location.protocol + "//" + location.host + prefix + "/s/");
          formData.append("path", this.fullPath);
          formData.append("isFile", "0");
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
              toast("获取分享信息失败");
            });
        },
        copyShare: function () {
          if (this.shareUrl) copyText(this.shareUrl);
        },
        scrollTop: function () {
          window.scrollTo({ top: 0, behavior: "smooth" });
        },
        year: function () {
          return new Date().getFullYear();
        },
      },
      template:
        '<div class="vt-container" :class="{ \'is-loading\': loading }">' +
        '  <div v-if="headHtml" class="vt-md-card">' +
        '    <div class="vt-md-card-header"><Icon name="file-text" :size="16" /><span>HEAD.md</span></div>' +
        '    <div class="vt-md" v-html="headHtml"></div>' +
        "  </div>" +
        '  <div class="vt-card">' +
        '    <div class="px-3 pt-3 pb-2 flex flex-wrap items-center gap-1">' +
        '      <div class="relative inline-flex">' +
        '        <div class="vt-chip" @click.stop="showAccountSwitcher ? (showAccountMenu=!showAccountMenu) : goHome()">' +
        '          <span class="vt-chip-icon"><Icon name="home" :size="16" /></span>' +
        '          <span class="vt-chip-title" @click.stop="goHome()">{{ displayTitle }}</span>' +
        "        </div>" +
        '        <div v-if="showAccountMenu" class="vt-menu left-0 top-10" @click.stop>' +
        '          <button type="button" v-for="a in accounts" :key="a.name" class="vt-menu-item" @click="goPath(\'/\' + a.name)">' +
        '            <Icon name="user" :size="16" /> {{ a.name }}' +
        "          </button>" +
        "        </div>" +
        "      </div>" +
        '      <template v-for="(bp, idx) in breadcrumbPaths" :key="bp.PathUrl || idx">' +
        '        <Icon name="chevron-right" :size="16" class-name="text-[var(--vt-muted)]" />' +
        '        <div class="vt-chip" @click="goPath(bp.PathUrl)">' +
        '          <span class="vt-chip-title" :title="bp.PathName">{{ bp.PathName }}</span>' +
        "        </div>" +
        "      </template>" +
        "    </div>" +
        '    <div class="flex flex-wrap items-center gap-1 px-2 py-2 border-t border-[var(--vt-border)]">' +
        '      <button v-if="hasParent" class="vt-icon-btn" title="返回上级目录" @click="goParent"><Icon name="arrow-left" /></button>' +
        '      <div class="vt-search ml-auto">' +
        '        <button class="vt-icon-btn" title="搜索" @click="searchOpen=!searchOpen"><Icon name="search" :size="18" /></button>' +
        '        <input v-show="searchOpen || searchKey" v-model="searchKey" type="text" placeholder="搜索文件（夹）" @keydown="onSearchKey" />' +
        "      </div>" +
        '      <div class="relative">' +
        '        <button class="vt-icon-btn" title="排序" @click.stop="showSortMenu=!showSortMenu"><Icon name="arrow-up-down" /></button>' +
        '        <div v-if="showSortMenu" class="vt-menu right-0 top-10 w-44" @click.stop>' +
        '          <button class="vt-menu-item" @click="setSort(\'file_name\',\'asc\')">文件名升序</button>' +
        '          <button class="vt-menu-item" @click="setSort(\'file_name\',\'desc\')">文件名降序</button>' +
        '          <button class="vt-menu-item" @click="setSort(\'file_size\',\'asc\')">大小升序</button>' +
        '          <button class="vt-menu-item" @click="setSort(\'file_size\',\'desc\')">大小降序</button>' +
        '          <button class="vt-menu-item" @click="setSort(\'last_op_time\',\'asc\')">时间升序</button>' +
        '          <button class="vt-menu-item" @click="setSort(\'last_op_time\',\'desc\')">时间降序</button>' +
        '          <button class="vt-menu-item" @click="resetSort()">默认</button>' +
        "        </div>" +
        "      </div>" +
        '      <button class="vt-icon-btn" title="预览本页所有图片" @click="previewImages"><Icon name="images" /></button>' +
        '      <div v-if="hasParent && isAdmin" class="relative">' +
        '        <button class="vt-icon-btn" title="分享" @click="openShare"><Icon name="share-2" /></button>' +
        '        <div v-if="showShareMenu" class="vt-menu vt-share-menu right-0 top-10 w-52 p-3 text-center" @click.stop>' +
        '          <img v-if="shareQr" :src="shareQr" class="w-full h-auto block rounded-md mb-2" alt="qrcode" />' +
        '          <button class="px-3 py-1.5 rounded-md text-white text-sm" style="background:var(--vt-primary)" @click="copyShare">复制短链接</button>' +
        "        </div>" +
        "      </div>" +
        '      <button class="vt-icon-btn" title="复制全部下载链接" @click="copyAllLinks"><Icon name="copy" /></button>' +
        '      <button class="vt-icon-btn" title="明暗主题切换" @click="toggleTheme"><Icon :name="isDark ? \'sun\' : \'moon\'" /></button>' +
        '      <button class="vt-icon-btn" title="布局切换" @click="toggleLayout"><Icon :name="isGrid ? \'list\' : \'layout-grid\'" /></button>' +
        "    </div>" +
        '    <div v-if="hasPwd" class="vt-pwd">' +
        '      <div class="flex items-center gap-2 mb-2 text-[var(--vt-muted)]"><Icon name="lock" /> 请输入密码</div>' +
        '      <div class="flex items-center gap-2">' +
        '        <input v-model="password" type="password" placeholder="密码" @keydown.enter="submitPwd" />' +
        '        <button class="vt-icon-btn is-active" title="确认" @click="submitPwd"><Icon name="fingerprint" /></button>' +
        "      </div>" +
        '      <div v-if="pwdErr" class="text-red-500 text-sm mt-2">{{ pwdErr }}</div>' +
        "    </div>" +
        '    <template v-else>' +
        '      <div v-if="empty" class="vt-empty">' +
        '        <Icon name="folder" :size="48" />' +
        "        <p>网盘空空如也</p>" +
        "      </div>" +
        '      <div v-else-if="isGrid" class="vt-file-grid">' +
        '        <div v-for="f in files" :key="f.file_id || f.path" class="vt-file-tile" @click="openFile(f, $event)">' +
        '          <img v-if="f.thumbnail && (f.view_type===\'img\' || f.view_type===\'video\')" :src="f.thumbnail" :alt="f.file_name" />' +
        '          <div v-else class="flex justify-center py-3 text-[var(--vt-primary)]"><Icon :name="iconOf(f)" :size="56" :stroke="1.5" /></div>' +
        '          <div class="name" :title="f.file_name">{{ f.file_name }}</div>' +
        "        </div>" +
        "      </div>" +
        '      <div v-else>' +
        '        <div v-for="f in files" :key="f.file_id || f.path" class="vt-file-row" @click="openFile(f, $event)">' +
        '          <span class="text-[var(--vt-primary)]"><Icon :name="iconOf(f)" :size="22" /></span>' +
        '          <div class="flex-1 min-w-0">' +
        '            <div class="truncate font-medium">{{ f.file_name }}</div>' +
        '            <div v-if="!f.is_folder" class="meta">{{ f.size_fmt }} / {{ f.last_op_time }}</div>' +
        '            <div v-if="searchKey" class="meta truncate">{{ url(f.path) }}</div>' +
        "          </div>" +
        '          <template v-if="!f.is_folder">' +
        '            <button data-stop class="vt-icon-btn" title="复制链接" @click="copyFileLink(f)"><Icon name="copy" :size="16" /></button>' +
        '            <a data-stop class="vt-icon-btn" title="下载" :href="fileUrl(f)" @click.stop><Icon name="download" :size="18" /></a>' +
        "          </template>" +
        "        </div>" +
        "      </div>" +
        "    </template>" +
        "  </div>" +
        '  <div v-if="readmeHtml" class="vt-md-card">' +
        '    <div class="vt-md-card-header"><Icon name="file-text" :size="16" /><span>README.md</span></div>' +
        '    <div class="vt-md" v-html="readmeHtml"></div>' +
        "  </div>" +
        '  <div class="vt-footer" v-html="config.footer || (\'©\' + year() + \' <a href=&quot;https://github.com/px-org/PanIndex&quot; target=&quot;_blank&quot;>PanIndex</a>. All rights reserved.\')"></div>' +
        '  <button v-if="showFab && !hasPwd" class="vt-fab" title="回到顶部" @click="scrollTop"><Icon name="arrow-up" /></button>' +
        '  <div id="image-preview-list" class="hidden"></div>' +
        "</div>",
    });
  }

  function publicConfig(cfg) {
    cfg = cfg || {};
    return {
      path_prefix: cfg.PathPrefix || cfg.path_prefix || "",
      site_name: cfg.SiteName || cfg.site_name || "",
      favicon_url: cfg.FaviconUrl || cfg.favicon_url || "",
      footer: cfg.Footer || cfg.footer || "",
      enable_preview: cfg.EnablePreview || cfg.enable_preview || "1",
      account_choose: cfg.AccountChoose || cfg.account_choose || "",
      head: cfg.Head || cfg.head || "0",
      readme: cfg.Readme || cfg.readme || "0",
      access: cfg.Access || cfg.access || "0",
      admin_path: cfg.AdminPath || cfg.admin_path || "/admin",
      image: cfg.Image || cfg.image || "",
      audio: cfg.Audio || cfg.audio || "",
      video: cfg.Video || cfg.video || "",
      code: cfg.Code || cfg.code || "",
      doc: cfg.Doc || cfg.doc || "",
      subtitle: cfg.Subtitle || cfg.subtitle || "",
      subtitle_path: cfg.SubtitlePath || cfg.subtitle_path || "",
      danmuku: cfg.Danmuku || cfg.danmuku || "",
      danmuku_path: cfg.DanmukuPath || cfg.danmuku_path || "",
      enable_lrc: cfg.EnableLrc || cfg.enable_lrc || "0",
      lrc_path: cfg.LrcPath || cfg.lrc_path || "",
    };
  }

  function normalizeAccounts(list) {
    return (list || []).map(function (a) {
      return {
        id: a.Id || a.id || "",
        name: a.Name || a.name || "",
        mode: a.Mode || a.mode || "",
      };
    });
  }

  function normalizeFiles(list) {
    return (list || []).map(function (f) {
      return {
        file_id: f.FileId || f.file_id || "",
        file_name: f.FileName || f.file_name || "",
        file_type: f.FileType || f.file_type || "",
        is_folder: !!(f.IsFolder || f.is_folder),
        path: f.Path || f.path || "",
        size_fmt: f.SizeFmt || f.size_fmt || "",
        last_op_time: f.LastOpTime || f.last_op_time || "",
        thumbnail: f.Thumbnail || f.thumbnail || "",
        view_type: f.ViewType || f.view_type || "",
      };
    });
  }

  function normalizeAccount(a) {
    a = a || {};
    return {
      id: a.Id || a.id || "",
      name: a.Name || a.name || "",
      mode: a.Mode || a.mode || "",
    };
  }

  window.PanOxygenTheme = {
    Icon: Icon,
    cookieGet: cookieGet,
    cookieSet: cookieSet,
    resolveDark: resolveDark,
    applyDarkClass: applyDarkClass,
    copyText: copyText,
    toast: toast,
    fileIconName: fileIconName,
    publicConfig: publicConfig,
    normalizeAccounts: normalizeAccounts,
    normalizeFiles: normalizeFiles,
    normalizeAccount: normalizeAccount,
    md5Hash: md5Hash,
    mountListing: function (raw) {
      var data = {
        title: raw.title || "",
        path: raw.path || "/",
        full_path: raw.full_path || "/",
        account_path: raw.account_path || "",
        accounts: normalizeAccounts(raw.accounts),
        account: normalizeAccount(raw.account),
        files: normalizeFiles(raw.fns || raw.files),
        pre_paths: raw.pre_paths || [],
        config: publicConfig(raw.config),
        has_pwd: raw.has_pwd,
        pwd_err_msg: raw.pwd_err_msg || "",
        pwd_path: raw.pwd_path || "",
        has_parent: raw.has_parent,
        parent_path: raw.parent_path || "/",
        search_key: raw.search_key || "",
        layout: raw.layout || "view_comfy",
        theme: raw.theme || "oxygen",
        is_admin_login: raw.is_admin_login,
        version: raw.version || "",
      };
      createListingApp(data).mount("#app");
    },
  };
  // Backward-compatible alias for older scripts/bookmarks.
  window.PanVueTheme = window.PanOxygenTheme;
})();
