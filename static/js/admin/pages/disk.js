(function (global) {
  var PanAdmin = global.PanAdmin || (global.PanAdmin = {});
  PanAdmin.pages = PanAdmin.pages || {};

  var MODE_LABELS = {
    native: "本地磁盘",
    ftp: "FTP",
    webdav: "WebDav",
    cloud189: "天翼云盘",
    aliyundrive: "阿里云盘",
    "aliyundrive-share": "阿里云盘（分享）",
    teambition: "Teambition项目盘",
    "teambition-us": "Teambition国际服",
    yun139: "和彩云",
    onedrive: "微软云盘",
    "onedrive-cn": "世纪互联",
    googledrive: "谷歌云盘",
    s3: "S3",
    pikpak: "PikPak",
    "123": "123云盘",
    "115": "115云盘",
  };

  var MODE_COLORS = {
    cloud189: "#06b6d4",
    teambition: "#3b82f6",
    "teambition-us": "#3b82f6",
    onedrive: "#93c5fd",
    "onedrive-cn": "#93c5fd",
    native: "#64748b",
    aliyundrive: "#7c3aed",
    "aliyundrive-share": "#7c3aed",
    ftp: "#9ca3af",
    webdav: "#92400e",
    yun139: "#ec4899",
    googledrive: "#f59e0b",
    s3: "#f97316",
    pikpak: "#6366f1",
    "123": "#0ea5e9",
    "115": "#1e3a8a",
  };

  function emptyAccount() {
    return {
      id: "",
      name: "",
      mode: "native",
      user: "",
      password: "",
      refresh_token: "",
      redirect_uri: "",
      api_url: "",
      root_id: "/",
      site_id: "",
      path_style: "Path",
      host: "",
      down_transfer: false,
      transfer_domain: "",
      info: "",
    };
  }

  function modeMeta(mode) {
    var m = {
      showUser: true,
      showPassword: true,
      showApiUrl: false,
      showRefresh: false,
      showRedirect: false,
      showSite: false,
      showS3: false,
      showAliQr: false,
      passwordType: "password",
      userLabel: "用户名",
      passwordLabel: "密码",
      apiUrlLabel: "",
      apiUrlPlaceholder: "",
      siteLabel: "网站ID",
      redirectLabel: "重定向地址",
      redirectPlaceholder: "https://mgaa.noki.workers.dev",
      defaultRoot: "/",
    };
    switch (mode) {
      case "native":
        m.showUser = false;
        m.showPassword = false;
        m.defaultRoot = "/";
        break;
      case "cloud189":
        m.showSite = true;
        m.siteLabel = "家庭ID（Family ID）";
        m.defaultRoot = "-11";
        break;
      case "teambition":
      case "teambition-us":
        m.showSite = true;
        m.siteLabel = "项目ID（Project ID）";
        m.defaultRoot = "";
        break;
      case "aliyundrive":
        m.showUser = false;
        m.showPassword = false;
        m.showRefresh = true;
        m.showAliQr = true;
        m.defaultRoot = "root";
        break;
      case "aliyundrive-share":
        m.showUser = false;
        m.showPassword = true;
        m.showRefresh = true;
        m.showSite = true;
        m.showAliQr = true;
        m.passwordType = "text";
        m.passwordLabel = "提取码";
        m.siteLabel = "分享ID";
        m.defaultRoot = "";
        break;
      case "onedrive":
      case "onedrive-cn":
        m.showRefresh = true;
        m.showRedirect = true;
        m.showSite = true;
        m.userLabel = "客户端ID（Client ID）";
        m.passwordLabel = "客户端密码（Client Secret）";
        m.siteLabel = "网站ID（/xxx.sharepoint.com:/sites/xxx）";
        m.defaultRoot = "/";
        break;
      case "ftp":
        m.showApiUrl = true;
        m.apiUrlLabel = "FTP地址（FTP Addr）";
        m.apiUrlPlaceholder = "192.168.1.1:21";
        m.defaultRoot = "/";
        break;
      case "webdav":
        m.showApiUrl = true;
        m.apiUrlLabel = "WebDav地址（WebDav Server）";
        m.apiUrlPlaceholder = "https://webdav.mydomain.me";
        m.defaultRoot = "/";
        break;
      case "yun139":
        m.userLabel = "手机号";
        m.passwordLabel = "COOKIE";
        m.passwordType = "text";
        m.defaultRoot = "00019700101000000001";
        break;
      case "googledrive":
        m.showRefresh = true;
        m.showRedirect = true;
        m.userLabel = "客户端ID（Client ID）";
        m.passwordLabel = "客户端密码（Client Secret）";
        m.defaultRoot = "";
        break;
      case "s3":
        m.showRedirect = true;
        m.showSite = true;
        m.showApiUrl = true;
        m.showS3 = true;
        m.siteLabel = "Region（us-east-1）";
        m.userLabel = "用户ID";
        m.passwordLabel = "密钥";
        m.passwordType = "text";
        m.apiUrlLabel = "端点（Endpoint）";
        m.apiUrlPlaceholder = "https://s3.amazonaws.com";
        m.redirectLabel = "桶（Bucket）";
        m.redirectPlaceholder = "bucket";
        m.defaultRoot = "";
        break;
      case "pikpak":
        m.defaultRoot = "";
        break;
      case "123":
        m.defaultRoot = "0";
        break;
      case "115":
        m.showUser = false;
        m.passwordLabel = "COOKIE";
        m.passwordType = "text";
        m.defaultRoot = "0";
        break;
    }
    return m;
  }

  function cookieStatusText(s) {
    if (s === 1) return "未刷新";
    if (s === 2) return "正常";
    if (s === 3) return "失效";
    if (s === 4) return "登录失败";
    if (s === -1) return "刷新中";
    return "-";
  }

  function cachePolicyText(p) {
    if (p === "nc") return "API直连(No Cache)";
    if (p === "mc") return "命中缓存（Memory）";
    if (p === "dc") return "完全缓存（DB）";
    return "-";
  }

  function cacheStatusText(acc) {
    if (acc.cache_policy !== "dc") return "-";
    if (acc.status === 1) return "未缓存";
    if (acc.status === 2) return "缓存成功";
    if (acc.status === 3) return "缓存失败";
    if (acc.status === -1) return "缓存中";
    return "未知";
  }

  PanAdmin.pages.Disk = {
    name: "PageDisk",
    inject: ["store"],
    data: function () {
      return {
        selected: [],
        dialog: false,
        dialogTitle: "添加",
        form: emptyAccount(),
        meta: modeMeta("native"),
        saving: false,
        qrVisible: false,
        qrImg: "",
        qrTimer: null,
        cacheDialog: false,
        cacheFolder: "/",
        cacheAccountId: "",
        uploadDialog: false,
        uploadFolder: "/",
        uploadAccountId: "",
        uploadFiles: null,
        uploading: false,
        cacheCfgDialog: false,
        cacheCfg: {
          id: "",
          cache_policy: "nc",
          expire_time_span: 1,
          sync_cron: "",
          sync_dir: "/",
          sync_child: true,
        },
        modes: Object.keys(MODE_LABELS),
        modeLabels: MODE_LABELS,
        modeColors: MODE_COLORS,
      };
    },
    computed: {
      accounts: function () {
        return this.store.config.accounts || [];
      },
    },
    mounted: function () {
      var self = this;
      this.$nextTick(function () {
        self.initSortable();
      });
    },
    updated: function () {
      var self = this;
      this.$nextTick(function () {
        self.initSortable();
      });
    },
    beforeUnmount: function () {
      this.stopQr();
      if (this._sortable) {
        this._sortable.destroy();
        this._sortable = null;
      }
    },
    methods: {
      modeLabel: function (m) {
        return MODE_LABELS[m] || m;
      },
      cookieStatusText: cookieStatusText,
      cachePolicyText: cachePolicyText,
      cacheStatusText: cacheStatusText,
      initSortable: function () {
        var el = this.$refs.items;
        if (!el || !window.Sortable) return;
        if (this._sortable) {
          this._sortable.destroy();
          this._sortable = null;
        }
        var self = this;
        this._sortable = Sortable.create(el, {
          handle: ".handle",
          animation: 150,
          ghostClass: "sortable-ghost",
          onEnd: async function () {
            var ids = Array.prototype.map.call(el.querySelectorAll("tr[data-id]"), function (tr) {
              return tr.getAttribute("data-id");
            });
            try {
              var res = await PanAdmin.api.post("/accounts/sort", ids);
              PanAdmin.toast((res && res.msg) || "排序已保存");
              await PanAdmin.refreshConfig(self.store);
            } catch (e) {
              PanAdmin.toast(e.message || "排序失败");
            }
          },
        });
      },
      selectedIds: function () {
        return this.selected.slice();
      },
      isSelected: function (id) {
        return this.selected.indexOf(id) >= 0;
      },
      toggle: function (id, e) {
        if (e) {
          e.preventDefault();
          e.stopPropagation();
        }
        var i = this.selected.indexOf(id);
        if (i >= 0) this.selected.splice(i, 1);
        else this.selected.push(id);
      },
      clearSelection: function () {
        this.selected = [];
      },
      onModeChange: function () {
        this.meta = modeMeta(this.form.mode);
        if (!this.form.id) {
          this.form.root_id = this.meta.defaultRoot;
        }
      },
      openAdd: function () {
        this.dialogTitle = "添加";
        this.form = emptyAccount();
        this.meta = modeMeta("native");
        this.qrVisible = false;
        this.dialog = true;
      },
      openEdit: async function () {
        var ids = this.selectedIds();
        if (ids.length !== 1) {
          PanAdmin.toast("请选择一条需要修改的记录");
          return;
        }
        try {
          var account = await PanAdmin.api.get("/account?id=" + encodeURIComponent(ids[0]));
          this.dialogTitle = "修改";
          this.form = {
            id: account.id || "",
            name: account.name || "",
            mode: account.mode || "native",
            user: account.user || "",
            password: account.password || "",
            refresh_token: account.refresh_token || "",
            redirect_uri: account.redirect_uri || "",
            api_url: account.api_url || "",
            root_id: account.root_id || "",
            site_id: account.site_id || "",
            path_style: account.path_style || "Path",
            host: account.host || "",
            down_transfer: !!account.down_transfer,
            transfer_domain: account.transfer_domain || "",
            info: account.info || "",
          };
          this.meta = modeMeta(this.form.mode);
          this.qrVisible = false;
          this.dialog = true;
        } catch (e) {
          PanAdmin.toast(e.message || "加载失败");
        }
      },
      saveAccount: async function () {
        if (this.saving) return;
        this.saving = true;
        try {
          var payload = Object.assign({}, this.form, {
            down_transfer: this.form.down_transfer ? 1 : 0,
          });
          var res = await PanAdmin.api.post("/account", payload);
          PanAdmin.toast((res && res.msg) || "已保存");
          this.dialog = false;
          this.stopQr();
          await PanAdmin.refreshConfig(this.store);
          this.clearSelection();
        } catch (e) {
          PanAdmin.toast(e.message || "保存失败");
        } finally {
          this.saving = false;
        }
      },
      remove: async function () {
        var ids = this.selectedIds();
        if (!ids.length) {
          PanAdmin.toast("请选择需要删除的记录");
          return;
        }
        try {
          var res = await PanAdmin.api.del("/accounts", ids);
          PanAdmin.toast((res && res.msg) || "删除成功");
          await PanAdmin.refreshConfig(this.store);
          this.clearSelection();
        } catch (e) {
          PanAdmin.toast(e.message || "删除失败");
        }
      },
      refreshToken: async function () {
        var ids = this.selectedIds();
        if (ids.length !== 1) {
          PanAdmin.toast("请选择需要刷新的挂载盘");
          return;
        }
        try {
          var res = await PanAdmin.api.post(
            "/refresh/login?id=" + encodeURIComponent(ids[0])
          );
          PanAdmin.toast((res && res.msg) || "已刷新");
        } catch (e) {
          PanAdmin.toast(e.message || "刷新失败");
        }
      },
      refreshBatchCache: async function () {
        var ids = this.selectedIds();
        if (!ids.length) {
          PanAdmin.toast("请选择需要批量刷新的挂载盘");
          return;
        }
        try {
          var res = await PanAdmin.api.post("/cache/update/batch", ids);
          PanAdmin.toast((res && res.msg) || "已刷新");
          await PanAdmin.refreshConfig(this.store);
        } catch (e) {
          PanAdmin.toast(e.message || "刷新失败");
        }
      },
      openRefreshCache: async function () {
        var ids = this.selectedIds();
        if (ids.length !== 1) {
          PanAdmin.toast("请选择需要刷新的挂载盘");
          return;
        }
        var acc = this.accounts.find(function (a) {
          return a.id === ids[0];
        });
        this.cacheAccountId = ids[0];
        this.cacheFolder = "/" + ((acc && acc.name) || "");
        try {
          var res = await PanAdmin.api.get(
            "/bypass?account_id=" + encodeURIComponent(ids[0])
          );
          if (res && res.data && res.data.name) {
            this.cacheFolder = "/" + res.data.name;
          }
        } catch (e) {}
        this.cacheDialog = true;
      },
      confirmRefreshCache: async function () {
        var fd = new FormData();
        fd.append("accountId", this.cacheAccountId);
        fd.append("cachePath", this.cacheFolder || "/");
        try {
          var res = await PanAdmin.api.postForm("/cache/update", fd);
          PanAdmin.toast((res && res.msg) || "已刷新");
          this.cacheDialog = false;
        } catch (e) {
          PanAdmin.toast(e.message || "刷新失败");
        }
      },
      openUpload: async function () {
        var ids = this.selectedIds();
        if (ids.length !== 1) {
          PanAdmin.toast("请选择需要上传的挂载盘");
          return;
        }
        var acc = this.accounts.find(function (a) {
          return a.id === ids[0];
        });
        this.uploadAccountId = ids[0];
        this.uploadFolder = "/" + ((acc && acc.name) || "");
        this.uploadFiles = null;
        try {
          var res = await PanAdmin.api.get(
            "/bypass?account_id=" + encodeURIComponent(ids[0])
          );
          if (res && res.data && res.data.name) {
            this.uploadFolder = "/" + res.data.name;
          }
        } catch (e) {}
        this.uploadDialog = true;
      },
      onUploadPick: function (e) {
        this.uploadFiles = e.target.files;
      },
      doUpload: async function (type) {
        if (this.uploading) return;
        if (!this.uploadFiles || !this.uploadFiles.length) {
          PanAdmin.toast("请选择文件");
          return;
        }
        var fd = new FormData();
        fd.append("uploadAccount", this.uploadAccountId);
        fd.append("uploadPath", this.uploadFolder || "/");
        fd.append("type", String(type));
        Array.prototype.forEach.call(this.uploadFiles, function (f) {
          fd.append("uploadFile", f);
        });
        this.uploading = true;
        PanAdmin.toast("开始上传，请耐心等待");
        try {
          var res = await PanAdmin.api.postForm("/upload", fd);
          PanAdmin.toast((res && res.msg) || "上传完成");
          this.uploadDialog = false;
        } catch (e) {
          PanAdmin.toast(e.message || "上传失败");
        } finally {
          this.uploading = false;
        }
      },
      openCacheConfig: function () {
        var ids = this.selectedIds();
        if (ids.length !== 1) {
          PanAdmin.toast("请选择需要操作的记录");
          return;
        }
        var acc = this.accounts.find(function (a) {
          return a.id === ids[0];
        });
        if (!acc) return;
        this.cacheCfg = {
          id: acc.id,
          cache_policy: acc.cache_policy || "nc",
          expire_time_span: acc.expire_time_span || 1,
          sync_cron: acc.sync_cron || "",
          sync_dir: acc.sync_dir || "/",
          sync_child: acc.sync_child === 0 || acc.sync_child === "0",
        };
        this.cacheCfgDialog = true;
      },
      saveCacheConfig: async function (t) {
        var d = {
          id: this.cacheCfg.id,
          cache_policy: this.cacheCfg.cache_policy,
          expire_time_span: parseInt(this.cacheCfg.expire_time_span, 10) || 1,
          sync_cron: this.cacheCfg.sync_cron,
          sync_dir: this.cacheCfg.sync_dir,
          sync_child: this.cacheCfg.sync_child ? 0 : 1,
        };
        try {
          var res = await PanAdmin.api.post("/cache/config?t=" + t, d);
          PanAdmin.toast((res && res.msg) || "已保存");
          this.cacheCfgDialog = false;
          await PanAdmin.refreshConfig(this.store);
        } catch (e) {
          PanAdmin.toast(e.message || "保存失败");
        }
      },
      stopQr: function () {
        if (this.qrTimer) {
          clearInterval(this.qrTimer);
          this.qrTimer = null;
        }
      },
      genQrcode: async function () {
        var self = this;
        this.stopQr();
        try {
          var result = await PanAdmin.api.get("/ali/qrcode");
          if (result && result.qr) {
            this.qrImg = result.qr;
            this.qrVisible = true;
            var param = {};
            try {
              param =
                typeof result.param === "string"
                  ? JSON.parse(result.param)
                  : result.param || {};
            } catch (e) {
              param = {};
            }
            var times = 0;
            this.qrTimer = setInterval(async function () {
              times++;
              if (times > 60) {
                self.stopQr();
                return;
              }
              try {
                var fd = new FormData();
                Object.keys(param).forEach(function (k) {
                  fd.append(k, param[k]);
                });
                var st = await PanAdmin.api.postForm("/ali/qrcode/check", fd);
                if (st && st.qrCodeStatus === "CONFIRMED" && st.refreshToken) {
                  self.form.refresh_token = st.refreshToken;
                  self.stopQr();
                  self.qrVisible = false;
                  PanAdmin.toast("获取成功");
                } else if (st && st.qrCodeStatus === "EXPIRED") {
                  self.stopQr();
                }
              } catch (e) {}
            }, 2000);
          }
        } catch (e) {
          PanAdmin.toast(e.message || "获取二维码失败");
        }
      },
    },
    template:
      '<div>' +
      '  <page-header title="网盘挂载" subtitle="拖动头像修改显示顺序"></page-header>' +
      '  <div class="admin-card overflow-hidden">' +
      '    <div class="admin-card-body">' +
      '      <div class="admin-toolbar">' +
      '        <pa-tooltip content="添加"><button type="button" class="admin-toolbar-btn" @click="openAdd"><pa-icon name="circle-plus" :size="18"></pa-icon></button></pa-tooltip>' +
      '        <pa-tooltip content="修改"><button type="button" class="admin-toolbar-btn" @click="openEdit"><pa-icon name="pencil" :size="18"></pa-icon></button></pa-tooltip>' +
      '        <pa-tooltip content="缓存设置"><button type="button" class="admin-toolbar-btn" @click="openCacheConfig"><pa-icon name="database" :size="18"></pa-icon></button></pa-tooltip>' +
      '        <pa-tooltip content="文件上传"><button type="button" class="admin-toolbar-btn" @click="openUpload"><pa-icon name="upload" :size="18"></pa-icon></button></pa-tooltip>' +
      '        <pa-tooltip content="刷新令牌"><button type="button" class="admin-toolbar-btn" @click="refreshToken"><pa-icon name="refresh-cw" :size="18"></pa-icon></button></pa-tooltip>' +
      '        <pa-tooltip content="刷新缓存"><button type="button" class="admin-toolbar-btn" @click="openRefreshCache"><pa-icon name="hard-drive" :size="18"></pa-icon></button></pa-tooltip>' +
      '        <pa-tooltip content="批量刷新缓存"><button type="button" class="admin-toolbar-btn" @click="refreshBatchCache"><pa-icon name="layers" :size="18"></pa-icon></button></pa-tooltip>' +
      '        <pa-tooltip content="删除"><button type="button" class="admin-toolbar-btn is-danger" @click="remove"><pa-icon name="circle-minus" :size="18"></pa-icon></button></pa-tooltip>' +
      "      </div>" +
      '      <div class="admin-card-main">' +
      '      <div class="overflow-x-auto flex flex-col flex-1 min-h-0">' +
      '      <table class="admin-table" :class="{\'admin-table--empty\': !accounts.length}">' +
      "        <thead><tr><th class=\"w-10\"></th><th>网盘名称</th><th>网盘类型</th><th>登录状态</th><th>缓存策略</th><th>缓存状态</th><th>文件总数</th><th>上一次更新</th><th>耗时</th><th>备注</th></tr></thead>" +
      '        <tbody v-if="accounts.length" ref="items">' +
      '          <tr v-for="acc in accounts" :key="acc.id" :data-id="acc.id" :class="{selected: isSelected(acc.id)}" @click="toggle(acc.id)">' +
      '            <td @click.stop><pa-check :model-value="isSelected(acc.id)" @update:model-value="toggle(acc.id)"></pa-check></td>' +
      "            <td>" +
      '              <div class="admin-chip handle" @click.stop>' +
      '                <span class="admin-chip-icon" :style="{background: modeColors[acc.mode] || \'#64748b\'}"><pa-icon name="hard-drive" :size="12"></pa-icon></span>' +
      '                <span class="truncate text-sm">{{ acc.name }}</span>' +
      "              </div>" +
      "            </td>" +
      "            <td>{{ modeLabel(acc.mode) }}</td>" +
      "            <td>{{ cookieStatusText(acc.cookie_status) }}</td>" +
      "            <td>{{ cachePolicyText(acc.cache_policy) }}</td>" +
      "            <td>{{ cacheStatusText(acc) }}</td>" +
      "            <td>{{ acc.cache_policy === 'dc' ? acc.files_count : '-' }}</td>" +
      "            <td>{{ acc.cache_policy === 'dc' ? acc.last_op_time : '-' }}</td>" +
      "            <td>{{ acc.cache_policy === 'dc' ? acc.time_span : '-' }}</td>" +
      "            <td>{{ acc.info }}</td>" +
      "          </tr>" +
      "        </tbody>" +
      "      </table>" +
      '      <div v-if="!accounts.length" class="admin-empty">暂无网盘</div>' +
      "    </div>" +
      "      </div>" +
      "    </div>" +
      "  </div>" +

      '  <pa-modal :open="dialog" :title="dialogTitle" width="max-w-xl" @close="dialog=false; stopQr()">' +

      '    <pa-field label="网盘名称" help=\'<a class="underline" href="https://docs.noki.icu/#/docs/zh/question?id=%e5%a6%82%e4%bd%95%e8%8e%b7%e5%8f%96%e7%9b%ae%e5%bd%95id%ef%bc%9f" target="_blank">说明</a>\'><input class="admin-input" v-model="form.name" /></pa-field>' +
      '    <pa-field label="网盘模式"><select class="admin-select" v-model="form.mode" @change="onModeChange"><option v-for="m in modes" :key="m" :value="m">{{ modeLabels[m] }}</option></select></pa-field>' +
      '    <pa-field v-if="meta.showUser" :label="meta.userLabel"><input class="admin-input" v-model="form.user" /></pa-field>' +
      '    <pa-field v-if="meta.showPassword" :label="meta.passwordLabel"><input class="admin-input" :type="meta.passwordType" v-model="form.password" /></pa-field>' +
      '    <pa-field v-if="meta.showApiUrl" :label="meta.apiUrlLabel"><input class="admin-input" v-model="form.api_url" :placeholder="meta.apiUrlPlaceholder" /></pa-field>' +
      '    <pa-field v-if="meta.showRefresh" label="刷新令牌（Refresh Token）">' +
      '      <input class="admin-input" v-model="form.refresh_token" />' +
      '      <button v-if="meta.showAliQr" type="button" class="mt-2 text-sm underline text-[var(--admin-primary)]" @click="genQrcode">点击获取</button>' +
      '      <div v-if="qrVisible" class="mt-3 p-3 border rounded-lg text-center" style="border-color: var(--admin-border)">' +
      '        <img :src="qrImg" class="mx-auto h-36" alt="qrcode" />' +
      '        <button type="button" class="admin-btn admin-btn-ghost mt-2" @click="genQrcode">刷新二维码</button>' +
      "      </div>" +
      "    </pa-field>" +
      '    <pa-field v-if="meta.showRedirect" :label="meta.redirectLabel"><input class="admin-input" v-model="form.redirect_uri" :placeholder="meta.redirectPlaceholder" /></pa-field>' +
      '    <pa-field v-if="meta.showSite" :label="meta.siteLabel"><input class="admin-input" v-model="form.site_id" /></pa-field>' +
      '    <pa-field v-if="meta.showS3" label="S3ForcePathStyle"><select class="admin-select" v-model="form.path_style"><option value="Path">path-style bucket</option><option value="VirtualHosting">virtual hosted bucket</option></select></pa-field>' +
      '    <pa-field label="根目录ID(路径)" help=\'<a class="underline" href="https://docs.noki.icu/#/docs/zh/question?id=%e5%a6%82%e4%bd%95%e8%8e%b7%e5%8f%96%e7%9b%ae%e5%bd%95id%ef%bc%9f" target="_blank">如何获取目录ID？</a>\'><input class="admin-input" v-model="form.root_id" /></pa-field>' +
      '    <pa-field label="域名 Host" help="设置后，将根据 Host 过滤账号列表"><input class="admin-input" v-model="form.host" placeholder="localhost:5238" /></pa-field>' +
      '    <div class="mb-4"><pa-switch v-model="form.down_transfer" label="流量中转"></pa-switch></div>' +

      '    <pa-field label="中转地址" help="为空将在本机中转，多个逗号分隔"><input class="admin-input" v-model="form.transfer_domain" placeholder="原域名|中转域名" /></pa-field>' +
      '    <pa-field label="备注"><textarea class="admin-textarea" rows="2" v-model="form.info"></textarea></pa-field>' +
      '    <template #footer>' +
      '      <button type="button" class="admin-btn admin-btn-ghost" @click="dialog=false; stopQr()">关闭</button>' +
      '      <button type="button" class="admin-btn admin-btn-primary" :disabled="saving" @click="saveAccount"><pa-icon name="save" :size="16"></pa-icon> 保存</button>' +
      "    </template>" +
      "  </pa-modal>" +

      '  <pa-modal :open="cacheDialog" title="刷新缓存" @close="cacheDialog=false">' +
      '    <pa-field label="远程目录(PanIndex路径)" help="虚拟路径，格式：/{name}/a/b"><input class="admin-input" v-model="cacheFolder" /></pa-field>' +
      '    <template #footer>' +
      '      <button type="button" class="admin-btn admin-btn-ghost" @click="cacheDialog=false">关闭</button>' +
      '      <button type="button" class="admin-btn admin-btn-primary" @click="confirmRefreshCache">确定</button>' +
      "    </template>" +
      "  </pa-modal>" +

      '  <pa-modal :open="uploadDialog" title="文件上传" @close="uploadDialog=false">' +
      '    <pa-field label="选择文件"><input type="file" multiple class="admin-input" @change="onUploadPick" /></pa-field>' +
      '    <pa-field label="远程目录(PanIndex路径)" help="虚拟路径，格式：/{name}/a/b"><input class="admin-input" v-model="uploadFolder" /></pa-field>' +
      '    <template #footer>' +
      '      <button type="button" class="admin-btn admin-btn-ghost" @click="uploadDialog=false">关闭</button>' +
      '      <button type="button" class="admin-btn admin-btn-primary" :disabled="uploading" @click="doUpload(0)"><pa-icon name="upload" :size="16"></pa-icon> 上传</button>' +
      '      <button type="button" class="admin-btn admin-btn-primary" :disabled="uploading" @click="doUpload(2)"><pa-icon name="upload" :size="16"></pa-icon> 上传并刷新</button>' +
      "    </template>" +
      "  </pa-modal>" +

      '  <pa-modal :open="cacheCfgDialog" title="缓存配置" @close="cacheCfgDialog=false">' +
      '    <pa-field label="缓存策略"><select class="admin-select" v-model="cacheCfg.cache_policy"><option value="nc">Api直连（No Cache）</option><option value="mc">命中缓存（Memory Cache）</option><option value="dc">完全缓存（Database Cache）</option></select></pa-field>' +
      '    <pa-field v-if="cacheCfg.cache_policy===\'mc\'" label="内存缓存有效期" help="单位是小时"><input class="admin-input" v-model="cacheCfg.expire_time_span" /></pa-field>' +
      '    <template v-if="cacheCfg.cache_policy===\'dc\'">' +
      '      <pa-field label="定时刷新缓存" help="为空将关闭定时缓存"><input class="admin-input" v-model="cacheCfg.sync_cron" placeholder="0 0 4 1/1 * ?" /></pa-field>' +
      '      <pa-field label="定时缓存目录" help="多个目录逗号分隔，默认缓存全部"><textarea class="admin-textarea" rows="2" v-model="cacheCfg.sync_dir"></textarea></pa-field>' +
      '      <div class="mb-4"><pa-switch v-model="cacheCfg.sync_child" label="缓存是否包含子目录"></pa-switch></div>' +

      "    </template>" +
      '    <template #footer>' +
      '      <button type="button" class="admin-btn admin-btn-ghost" @click="cacheCfgDialog=false">关闭</button>' +
      '      <button type="button" class="admin-btn admin-btn-primary" @click="saveCacheConfig(0)">确定</button>' +
      '      <button type="button" class="admin-btn admin-btn-primary" @click="saveCacheConfig(1)">确定并重置缓存</button>' +
      "    </template>" +
      "  </pa-modal>" +
      "</div>",
  };
})(window);
