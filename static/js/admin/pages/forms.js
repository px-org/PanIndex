(function (global) {
  var PanAdmin = global.PanAdmin || (global.PanAdmin = {});
  PanAdmin.pages = PanAdmin.pages || {};

  PanAdmin.pages.Common = {
    name: "PageCommon",
    inject: ["store"],
    data: function () {
      var c = this.store.config || {};
      return {
        form: {
          site_name: c.site_name || "",
          path_prefix: c.path_prefix || "",
          admin_path: c.admin_path || "/admin",
          account_choose: c.account_choose || "default",
          admin_user: c.admin_user || "",
          admin_password: c.admin_password || "",
          s_column: c.s_column || "default",
          s_order: c.s_order || "asc",
          cdn: String(c.cdn == null ? "0" : c.cdn),
          proxy: c.proxy || "",
          enable_download_statistics: c.enable_download_statistics === "1" || c.enable_download_statistics === 1,
        },
        saving: false,
        accountChooseOpts: [
          { value: "default", label: "默认账号" },
          { value: "display", label: "全部账号（WebDav必选）" },
        ],
        cdnOpts: [
          { value: "0", label: "本机" },
          { value: "1", label: "国内 CDN" },
          { value: "2", label: "jsDelivr" },
        ],
      };
    },
    methods: {
      save: async function () {
        this.saving = true;
        try {
          var payload = Object.assign({}, this.form, {
            enable_download_statistics: this.form.enable_download_statistics ? "1" : "0",
          });
          var res = await PanAdmin.api.saveConfig(payload);
          PanAdmin.toast((res && res.msg) || "已保存");
          Object.assign(this.store.config, payload);
        } catch (e) {
          PanAdmin.toast(e.message || "保存失败");
        } finally {
          this.saving = false;
        }
      },
    },
    template:
      '<div>' +
      '  <page-header title="基础配置"></page-header>' +
      '  <div class="admin-card p-5 max-w-3xl">' +
      '    <pa-field label="网站标题" help="显示优先于网盘名称">' +
      '      <input class="admin-input" v-model="form.site_name" />' +
      "    </pa-field>" +
      '    <pa-field label="网站路径前缀" help="配合反代可实现多层级跳转">' +
      '      <input class="admin-input" v-model="form.path_prefix" />' +
      "    </pa-field>" +
      '    <pa-field label="后台管理地址" help="自定义后台地址，默认为 /admin，重启生效">' +
      '      <input class="admin-input" v-model="form.admin_path" />' +
      "    </pa-field>" +
      '    <pa-field label="首页账号切换">' +
      '      <pa-radio-group v-model="form.account_choose" :options="accountChooseOpts"></pa-radio-group>' +
      "    </pa-field>" +
      '    <pa-field label="登录账号">' +
      '      <input class="admin-input" v-model="form.admin_user" required />' +
      "    </pa-field>" +
      '    <pa-field label="登录密码" help="如果是第一次运行，请务必修改默认密码！">' +
      '      <input class="admin-input" type="password" v-model="form.admin_password" required />' +
      "    </pa-field>" +
      '    <pa-field label="排序">' +
      '      <div class="flex flex-wrap gap-3">' +
      '        <select class="admin-select w-auto" v-model="form.s_column">' +
      '          <option value="default">网盘默认</option>' +
      '          <option value="file_name">文件名</option>' +
      '          <option value="file_size">文件大小</option>' +
      '          <option value="last_op_time">修改时间</option>' +
      "        </select>" +
      '        <select class="admin-select w-auto" v-model="form.s_order">' +
      '          <option value="asc">升序</option>' +
      '          <option value="desc">降序</option>' +
      "        </select>" +
      "      </div>" +
      "    </pa-field>" +
      '    <pa-field label="静态资源 CDN">' +
      '      <pa-radio-group v-model="form.cdn" :options="cdnOpts"></pa-radio-group>' +
      "    </pa-field>" +
      '    <pa-field label="代理" help="代理地址，仅对流量中转及 GoogleDrive 生效">' +
      '      <input class="admin-input" v-model="form.proxy" placeholder="socks5://127.0.0.1:10808" />' +
      "    </pa-field>" +
      '    <pa-field label="下载统计" help="开启后记录文件下载次数，供后台首页仪表盘展示">' +
      '      <pa-switch v-model="form.enable_download_statistics" label="启用下载统计"></pa-switch>' +
      "    </pa-field>" +
      '    <button type="button" class="admin-btn admin-btn-primary" :disabled="saving" @click="save">' +
      '      <pa-icon name="save" :size="16"></pa-icon> 保存' +
      "    </button>" +
      "  </div>" +
      "</div>",
  };

  function normalizeThemeName(theme) {
    if (theme === "vue") return "oxygen";
    if (theme === "vue-light") return "oxygen-light";
    if (theme === "vue-dark") return "oxygen-dark";
    return theme || "mdui";
  }

  PanAdmin.pages.Appearance = {
    name: "PageAppearance",
    inject: ["store"],
    data: function () {
      var c = this.store.config || {};
      return {
        form: {
          theme: normalizeThemeName(c.theme),
          favicon_url: c.favicon_url || "",
          footer: c.footer || "",
          css: c.css || "",
          js: c.js || "",
          readme: String(c.readme == null ? "0" : c.readme),
          head: String(c.head == null ? "0" : c.head),
          hide_readme_files: String(c.hide_readme_files == null ? "0" : c.hide_readme_files),
        },
        saving: false,
      };
    },
    methods: {
      save: async function () {
        this.saving = true;
        try {
          var res = await PanAdmin.api.saveConfig(this.form);
          PanAdmin.toast((res && res.msg) || "已保存");
          Object.assign(this.store.config, this.form);
        } catch (e) {
          PanAdmin.toast(e.message || "保存失败");
        } finally {
          this.saving = false;
        }
      },
    },
    template:
      '<div>' +
      '  <page-header title="外观"></page-header>' +
      '  <div class="admin-card p-5 max-w-3xl">' +
      '    <pa-field label="主题">' +
      '      <select class="admin-select" v-model="form.theme">' +
      '        <option value="mdui">mdui</option>' +
      '        <option value="mdui-light">mdui-light</option>' +
      '        <option value="mdui-dark">mdui-dark</option>' +
      '        <option value="oxygen">oxygen</option>' +
      '        <option value="oxygen-light">oxygen-light</option>' +
      '        <option value="oxygen-dark">oxygen-dark</option>' +
      '        <option value="bootstrap">bootstrap</option>' +
      '        <option value="classic">classic</option>' +
      "      </select>" +
      "    </pa-field>" +
      '    <pa-field label="网站图标 (Favicon)">' +
      '      <input class="admin-input" v-model="form.favicon_url" placeholder="网站图标 Url" />' +
      "    </pa-field>" +
      '    <pa-field label="自定义底部信息 (Footer)">' +
      '      <textarea class="admin-textarea" rows="3" v-model="form.footer" placeholder="支持 html 代码"></textarea>' +
      "    </pa-field>" +
      '    <pa-field label="自定义 CSS">' +
      '      <textarea class="admin-textarea" rows="3" v-model="form.css"></textarea>' +
      "    </pa-field>" +
      '    <pa-field label="自定义 Javascript">' +
      '      <textarea class="admin-textarea" rows="3" v-model="form.js"></textarea>' +
      "    </pa-field>" +
      '    <pa-field label="README.md">' +
      '      <pa-switch v-model="form.readme" true-value="1" false-value="0" on-label="渲染" off-label="不渲染"></pa-switch>' +
      "    </pa-field>" +
      '    <pa-field label="HEAD.md">' +
      '      <pa-switch v-model="form.head" true-value="1" false-value="0" on-label="渲染" off-label="不渲染"></pa-switch>' +
      "    </pa-field>" +
      '    <pa-field label="列表中隐藏 README.md / HEAD.md" help="开启后文件列表与搜索中不再显示，仍可用于页面 Markdown 渲染">' +
      '      <pa-switch v-model="form.hide_readme_files" true-value="1" false-value="0" on-label="隐藏" off-label="显示"></pa-switch>' +
      "    </pa-field>" +
      '    <button type="button" class="admin-btn admin-btn-primary" :disabled="saving" @click="save">' +
      '      <pa-icon name="save" :size="16"></pa-icon> 保存' +
      "    </button>" +
      "  </div>" +
      "</div>",
  };

  PanAdmin.pages.View = {
    name: "PageView",
    inject: ["store"],
    data: function () {
      var c = this.store.config || {};
      return {
        form: {
          enable_preview: c.enable_preview === "1" || c.enable_preview === 1,
          image: c.image || "",
          video: c.video || "",
          audio: c.audio || "",
          code: c.code || "",
          doc: c.doc || "",
          other: c.other || "",
          enable_lrc: c.enable_lrc === "1" || c.enable_lrc === 1,
          lrc_path: c.lrc_path || "",
          subtitle: c.subtitle || "",
          subtitle_path: c.subtitle_path || "",
          danmuku: String(c.danmuku == null ? "0" : c.danmuku),
          danmuku_path: c.danmuku_path || "",
        },
        saving: false,
        subtitleOpts: [
          { value: "", label: "关闭" },
          { value: "srt", label: "srt" },
          { value: "vtt", label: "vtt" },
          { value: "ass", label: "ass" },
        ],
      };
    },
    methods: {
      resetDefaults: function () {
        this.form.enable_preview = true;
        this.form.subtitle = "";
        this.form.danmuku = "0";
        this.form.enable_lrc = false;
        this.form.image = "png,gif,jpg,bmp,jpeg,ico,svg";
        this.form.video = "mp4,mkv,m3u8,ts,avi";
        this.form.audio = "mp3,wav,ape,flac";
        this.form.code =
          "txt,go,html,js,java,json,css,lua,sh,sql,py,cpp,xml,jsp,properties,yaml,ini";
        this.form.other = "*";
        this.form.lrc_path = "";
        this.form.subtitle_path = "";
        this.form.danmuku_path = "";
      },
      save: async function () {
        this.saving = true;
        try {
          var payload = Object.assign({}, this.form, {
            enable_preview: this.form.enable_preview ? "1" : "0",
            enable_lrc: this.form.enable_lrc ? "1" : "0",
          });
          var res = await PanAdmin.api.saveConfig(payload);
          PanAdmin.toast((res && res.msg) || "已保存");
          Object.assign(this.store.config, payload);
        } catch (e) {
          PanAdmin.toast(e.message || "保存失败");
        } finally {
          this.saving = false;
        }
      },
    },
    template:
      '<div>' +
      '  <page-header title="文件预览" subtitle="多个后缀名，半角英文逗号分隔"></page-header>' +
      '  <div class="admin-card p-5 max-w-3xl">' +
      '    <p class="text-sm font-medium mb-3" style="color: var(--admin-muted)">文件后缀</p>' +
      '    <pa-field label="是否开启预览">' +
      '      <pa-switch v-model="form.enable_preview" label="开启预览"></pa-switch>' +
      "    </pa-field>" +
      '    <pa-field label="图片 (image)"><input class="admin-input" v-model="form.image" /></pa-field>' +
      '    <pa-field label="视频 (video)"><input class="admin-input" v-model="form.video" /></pa-field>' +
      '    <pa-field label="音频 (audio)"><input class="admin-input" v-model="form.audio" /></pa-field>' +
      '    <pa-field label="代码 (code)"><input class="admin-input" v-model="form.code" /></pa-field>' +
      '    <pa-field label="文档 (doc)"><input class="admin-input" v-model="form.doc" /></pa-field>' +
      '    <pa-field label="其他 (other)" help="这部分文件不支持预览，但也可以跳转到预览页"><input class="admin-input" v-model="form.other" /></pa-field>' +
      '    <p class="text-sm font-medium mb-3 mt-6" style="color: var(--admin-muted)">播放增强</p>' +
      '    <pa-field label="歌词 (lrc)"><pa-switch v-model="form.enable_lrc" label="开启歌词"></pa-switch></pa-field>' +
      '    <pa-field label="歌词路径" help="相对路径，开启后将自动加载歌词，歌词文件需要和歌曲文件同名"><input class="admin-input" v-model="form.lrc_path" placeholder="lrc/" /></pa-field>' +
      '    <pa-field label="字幕"><pa-radio-group v-model="form.subtitle" :options="subtitleOpts"></pa-radio-group></pa-field>' +
      '    <pa-field label="字幕路径" help="相对路径，字幕文件需要和视频文件同名"><input class="admin-input" v-model="form.subtitle_path" placeholder="subtitle/" /></pa-field>' +
      '    <pa-field label="弹幕"><pa-switch v-model="form.danmuku" true-value="1" false-value="0" on-label="已开启" off-label="已关闭"></pa-switch></pa-field>' +
      '    <pa-field label="弹幕路径" help="相对路径，弹幕文件需要和视频文件同名，video.xml"><input class="admin-input" v-model="form.danmuku_path" placeholder="danmuku/" /></pa-field>' +
      '    <div class="flex flex-wrap gap-2">' +
      '      <button type="button" class="admin-btn admin-btn-primary" :disabled="saving" @click="save"><pa-icon name="save" :size="16"></pa-icon> 保存</button>' +
      '      <button type="button" class="admin-btn admin-btn-ghost" @click="resetDefaults"><pa-icon name="rotate-ccw" :size="16"></pa-icon> 恢复默认</button>' +
      "    </div>" +
      "  </div>" +
      "</div>",
  };

  PanAdmin.pages.Access = {
    name: "PageAccess",
    inject: ["store"],
    data: function () {
      var c = this.store.config || {};
      return {
        form: {
          short_action: String(c.short_action == null ? "0" : c.short_action),
          access: String(c.access == null ? "0" : c.access),
        },
        saving: false,
        shortOpts: [
          { value: "0", label: "预览" },
          { value: "1", label: "下载" },
        ],
        accessOpts: [
          { value: "0", label: "公开" },
          { value: "1", label: "仅直链" },
          { value: "2", label: "直链 + 预览" },
          { value: "3", label: "登录" },
        ],
      };
    },
    methods: {
      save: async function () {
        this.saving = true;
        try {
          var res = await PanAdmin.api.saveConfig(this.form);
          PanAdmin.toast((res && res.msg) || "已保存");
          Object.assign(this.store.config, this.form);
        } catch (e) {
          PanAdmin.toast(e.message || "保存失败");
        } finally {
          this.saving = false;
        }
      },
    },
    template:
      '<div>' +
      '  <page-header title="访问控制" subtitle="列表、预览页面的访问控制"></page-header>' +
      '  <div class="admin-card p-5 max-w-3xl">' +
      '    <pa-field label="短链行为"><pa-radio-group v-model="form.short_action" :options="shortOpts"></pa-radio-group></pa-field>' +
      '    <pa-field label="模式（重启生效）"><pa-radio-group v-model="form.access" :options="accessOpts"></pa-radio-group></pa-field>' +
      '    <button type="button" class="admin-btn admin-btn-primary" :disabled="saving" @click="save"><pa-icon name="save" :size="16"></pa-icon> 保存</button>' +
      "  </div>" +
      "</div>",
  };

  PanAdmin.pages.Safety = {
    name: "PageSafety",
    inject: ["store"],
    data: function () {
      var c = this.store.config || {};
      return {
        form: {
          only_referrer: c.only_referrer || "",
          enable_safety_link: c.enable_safety_link === "1",
          is_null_referrer: c.is_null_referrer === "1",
        },
        saving: false,
      };
    },
    methods: {
      save: async function () {
        this.saving = true;
        try {
          var payload = {
            only_referrer: this.form.only_referrer,
            enable_safety_link: this.form.enable_safety_link ? "1" : "0",
            is_null_referrer: this.form.is_null_referrer ? "1" : "0",
          };
          var res = await PanAdmin.api.saveConfig(payload);
          PanAdmin.toast((res && res.msg) || "已保存");
          Object.assign(this.store.config, payload);
        } catch (e) {
          PanAdmin.toast(e.message || "保存失败");
        } finally {
          this.saving = false;
        }
      },
    },
    template:
      '<div>' +
      '  <page-header title="防盗链" subtitle="通过 Referrer 策略实现外链访问控制"></page-header>' +
      '  <div class="admin-card p-5 max-w-3xl">' +
      '    <pa-field label="许可域名" help="允许的 Referrer，英文逗号分隔，支持正则">' +
      '      <input class="admin-input" v-model="form.only_referrer" placeholder="baidu.com,google.com" />' +
      "    </pa-field>" +
      '    <div class="flex flex-col gap-3 mb-4">' +
      '      <pa-switch v-model="form.enable_safety_link" label="启用防盗链"></pa-switch>' +
      '      <pa-switch v-model="form.is_null_referrer" label="允许空 Referrer 请求"></pa-switch>' +
      "    </div>" +
      '    <button type="button" class="admin-btn admin-btn-primary" :disabled="saving" @click="save"><pa-icon name="save" :size="16"></pa-icon> 保存</button>' +
      "  </div>" +
      "</div>",
  };

  PanAdmin.pages.Webdav = {
    name: "PageWebdav",
    inject: ["store"],
    data: function () {
      var c = this.store.config || {};
      return {
        form: {
          enable_dav: c.enable_dav === "1",
          dav_path: c.dav_path || "/dav",
          dav_mode: String(c.dav_mode == null ? "0" : c.dav_mode),
          dav_down_mode: String(c.dav_down_mode == null ? "0" : c.dav_down_mode),
          dav_user: c.dav_user || "",
          dav_password: c.dav_password || "",
        },
        saving: false,
        modeOpts: [
          { value: "0", label: "只读" },
          { value: "1", label: "读写" },
        ],
        downOpts: [
          { value: "0", label: "302 直链" },
          { value: "1", label: "流量中转" },
        ],
      };
    },
    methods: {
      save: async function () {
        this.saving = true;
        try {
          var payload = {
            enable_dav: this.form.enable_dav ? "1" : "0",
            dav_path: this.form.dav_path,
            dav_mode: this.form.dav_mode,
            dav_down_mode: this.form.dav_down_mode,
            dav_user: this.form.dav_user,
            dav_password: this.form.dav_password,
          };
          var res = await PanAdmin.api.saveConfig(payload);
          PanAdmin.toast((res && res.msg) || "已保存");
          Object.assign(this.store.config, payload);
        } catch (e) {
          PanAdmin.toast(e.message || "保存失败");
        } finally {
          this.saving = false;
        }
      },
    },
    template:
      '<div>' +
      '  <page-header title="WebDav" subtitle="可通过 RaiDrive 等工具挂载成本地磁盘"></page-header>' +
      '  <div class="admin-card p-5 max-w-3xl">' +
      '    <pa-field label="是否开启 WebDav"><pa-switch v-model="form.enable_dav" label="开启 WebDav"></pa-switch></pa-field>' +
      '    <pa-field label="路径" help="请求路径，开启后密码文件将无需授权，重启生效"><input class="admin-input" v-model="form.dav_path" placeholder="/dav" /></pa-field>' +
      '    <pa-field label="访问权限"><pa-radio-group v-model="form.dav_mode" :options="modeOpts"></pa-radio-group></pa-field>' +
      '    <pa-field label="下载行为"><pa-radio-group v-model="form.dav_down_mode" :options="downOpts"></pa-radio-group></pa-field>' +
      '    <pa-field label="用户名"><input class="admin-input" v-model="form.dav_user" /></pa-field>' +
      '    <pa-field label="密码" help="默认密码：1234"><input class="admin-input" type="password" v-model="form.dav_password" /></pa-field>' +
      '    <button type="button" class="admin-btn admin-btn-primary" :disabled="saving" @click="save"><pa-icon name="save" :size="16"></pa-icon> 保存</button>' +
      "  </div>" +
      "</div>",
  };
})(window);
