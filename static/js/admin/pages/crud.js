(function (global) {
  var PanAdmin = global.PanAdmin || (global.PanAdmin = {});
  PanAdmin.pages = PanAdmin.pages || {};

  PanAdmin.pages.Hide = {
    name: "PageHide",
    inject: ["store"],
    data: function () {
      return {
        selected: [],
        dialog: false,
        filePath: "",
        saving: false,
      };
    },
    computed: {
      rows: function () {
        var hf = this.store.config.hide_files || {};
        return Object.keys(hf).sort();
      },
    },
    methods: {
      isSelected: function (id) {
        return this.selected.indexOf(id) >= 0;
      },
      toggle: function (id) {
        var i = this.selected.indexOf(id);
        if (i >= 0) this.selected.splice(i, 1);
        else this.selected.push(id);
      },
      clearSelection: function () {
        this.selected = [];
      },
      openAdd: function () {
        this.filePath = "";
        this.dialog = true;
      },
      save: async function () {
        if (!this.filePath) {
          PanAdmin.toast("请输入路径");
          return;
        }
        this.saving = true;
        try {
          var res = await PanAdmin.api.post("/hide/file", {
            hide_path: this.filePath,
          });
          PanAdmin.toast((res && res.msg) || "已保存");
          this.dialog = false;
          await PanAdmin.refreshConfig(this.store);
          this.clearSelection();
        } catch (e) {
          PanAdmin.toast(e.message || "保存失败");
        } finally {
          this.saving = false;
        }
      },
      remove: async function () {
        var paths = this.selected.slice();
        if (!paths.length) {
          PanAdmin.toast("请选择需要删除的记录");
          return;
        }
        try {
          var res = await PanAdmin.api.del("/hide/file", paths);
          PanAdmin.toast((res && res.msg) || "删除成功");
          await PanAdmin.refreshConfig(this.store);
          this.clearSelection();
        } catch (e) {
          PanAdmin.toast(e.message || "删除失败");
        }
      },
    },
    template:
      '<div>' +
      '  <page-header title="隐藏文件（夹）" subtitle="隐藏部分文件或文件夹"></page-header>' +
      '  <div class="admin-card overflow-hidden">' +
      '    <div class="admin-card-body">' +
      '      <div class="admin-toolbar">' +
      '        <pa-tooltip content="添加"><button type="button" class="admin-toolbar-btn" @click="openAdd"><pa-icon name="circle-plus" :size="18"></pa-icon></button></pa-tooltip>' +
      '        <pa-tooltip content="删除"><button type="button" class="admin-toolbar-btn is-danger" @click="remove"><pa-icon name="circle-minus" :size="18"></pa-icon></button></pa-tooltip>' +
      "      </div>" +
      '      <div class="admin-card-main">' +
      '        <table class="admin-table" :class="{\'admin-table--empty\': !rows.length}">' +
      "          <thead><tr><th class=\"w-10\"></th><th>文件（夹）路径</th></tr></thead>" +
      '          <tbody v-if="rows.length">' +
      '            <tr v-for="p in rows" :key="p" :class="{selected: isSelected(p)}" @click="toggle(p)">' +
      '              <td @click.stop><pa-check :model-value="isSelected(p)" @update:model-value="toggle(p)"></pa-check></td>' +
      "              <td>{{ p }}</td>" +
      "            </tr>" +
      "          </tbody>" +
      "        </table>" +
      '        <div v-if="!rows.length" class="admin-empty">暂无数据</div>' +
      "      </div>" +
      "    </div>" +
      "  </div>" +
      '  <pa-modal :open="dialog" title="添加" @close="dialog=false">' +
      '    <pa-field label="文件（夹）路径" help="PanIndex 的虚拟路径，格式：/{name}/a/b">' +
      '      <input class="admin-input" v-model="filePath" />' +
      "    </pa-field>" +
      '    <template #footer>' +
      '      <button type="button" class="admin-btn admin-btn-ghost" @click="dialog=false">取消</button>' +
      '      <button type="button" class="admin-btn admin-btn-primary" :disabled="saving" @click="save"><pa-icon name="save" :size="16"></pa-icon> 保存</button>' +
      "    </template>" +
      "  </pa-modal>" +
      "</div>",
  };

  PanAdmin.pages.Share = {
    name: "PageShare",
    inject: ["store"],
    computed: {
      rows: function () {
        return this.store.config.share_info_list || [];
      },
      pathPrefix: function () {
        return this.store.config.path_prefix || "";
      },
    },
    methods: {
      formatExpire: PanAdmin.formatExpire,
      copyShare: function (info, pwd) {
        var prefix = window.location.protocol + "//" + window.location.host;
        var path = info.file_path || "";
        var fileName = path.substring(path.lastIndexOf("/") + 1);
        var sharePath = this.pathPrefix + "/s/" + info.short_code;
        var msg = "「" + fileName + "」" + prefix + sharePath;
        if (pwd && pwd.password) {
          msg += " 密码: " + pwd.password;
        }
        if (!navigator.clipboard) {
          PanAdmin.toast("该浏览器不支持复制操作");
          return;
        }
        navigator.clipboard.writeText(msg).then(
          function () {
            PanAdmin.toast("链接和密码已复制到剪切板");
          },
          function () {
            PanAdmin.toast("复制失败");
          }
        );
      },
      remove: async function (path) {
        try {
          var res = await PanAdmin.api.del("/share/info", [path]);
          PanAdmin.toast((res && res.msg) || "删除成功");
          await PanAdmin.refreshConfig(this.store);
        } catch (e) {
          PanAdmin.toast(e.message || "删除失败");
        }
      },
    },
    template:
      '<div>' +
      '  <page-header title="分享短链"></page-header>' +
      '  <div class="admin-card overflow-hidden">' +
      '    <table class="admin-table">' +
      "      <thead><tr><th>文件（夹）路径</th><th>短链编码</th><th>访问控制</th><th>操作</th></tr></thead>" +
      "      <tbody>" +
      '        <tr v-for="info in rows" :key="info.short_code">' +
      "          <td>{{ info.file_path }}</td>" +
      '          <td><a class="text-[var(--admin-primary)] underline" :href="pathPrefix + \'/s/\' + info.short_code" target="_blank">{{ info.short_code }}</a></td>' +
      "          <td>" +
      '            <div v-for="(pwd, idx) in (info.pwd_info || [])" :key="idx" class="flex items-center gap-2 text-sm mb-1">' +
      '              <code class="px-1.5 py-0.5 rounded bg-black/5 dark:bg-white/10">{{ pwd.password }}</code>' +
      "              <span style=\"color: var(--admin-muted)\">{{ formatExpire(pwd.expire_at) }}</span>" +
      '              <pa-tooltip content="复制链接" placement="top"><button type="button" class="admin-btn admin-btn-ghost !p-1" @click="copyShare(info, pwd)"><pa-icon name="copy" :size="14"></pa-icon></button></pa-tooltip>' +
      "            </div>" +
      "          </td>" +
      "          <td>" +
      '            <pa-tooltip content="删除" placement="top"><button type="button" class="admin-btn admin-btn-ghost !p-2" @click="remove(info.file_path)"><pa-icon name="trash-2" :size="16"></pa-icon></button></pa-tooltip>' +
      "          </td>" +
      "        </tr>" +
      '        <tr v-if="!rows.length"><td colspan="4" class="text-center py-8" style="color: var(--admin-muted)">暂无数据</td></tr>' +
      "      </tbody>" +
      "    </table>" +
      "  </div>" +
      "</div>",
  };

  PanAdmin.pages.Bypass = {
    name: "PageBypass",
    inject: ["store"],
    data: function () {
      return {
        selected: [],
        dialog: false,
        title: "添加",
        form: { id: "", name: "", accountIds: {} },
        saving: false,
      };
    },
    computed: {
      rows: function () {
        return this.store.config.bypass_list || [];
      },
      accounts: function () {
        return this.store.config.accounts || [];
      },
    },
    methods: {
      isSelected: function (id) {
        return this.selected.indexOf(id) >= 0;
      },
      toggle: function (id) {
        var i = this.selected.indexOf(id);
        if (i >= 0) this.selected.splice(i, 1);
        else this.selected.push(id);
      },
      clearSelection: function () {
        this.selected = [];
      },
      toggleAccount: function (id, on) {
        if (on) {
          var o = Object.assign({}, this.form.accountIds);
          o[id] = true;
          this.form.accountIds = o;
        } else {
          var n = Object.assign({}, this.form.accountIds);
          delete n[id];
          this.form.accountIds = n;
        }
      },
      openAdd: function () {
        this.title = "添加";
        this.form = { id: "", name: "", accountIds: {} };
        this.dialog = true;
      },
      openEdit: function () {
        var ids = this.selected.slice();
        if (ids.length !== 1) {
          PanAdmin.toast("请选择一条需要修改的记录");
          return;
        }
        var row = this.rows.find(function (r) {
          return r.id === ids[0];
        });
        if (!row) return;
        var accountIds = {};
        (row.accounts || []).forEach(function (a) {
          accountIds[a.id] = true;
        });
        this.title = "修改";
        this.form = { id: row.id, name: row.name, accountIds: accountIds };
        this.dialog = true;
      },
      save: async function () {
        var accounts = Object.keys(this.form.accountIds)
          .filter(function (id) {
            return !!this.form.accountIds[id];
          }.bind(this))
          .map(function (id) {
            return { id: id };
          });
        if (!accounts.length) {
          PanAdmin.toast("请勾选需要分流的网盘");
          return;
        }
        this.saving = true;
        try {
          var res = await PanAdmin.api.post("/bypass", {
            id: this.form.id,
            name: this.form.name,
            accounts: accounts,
          });
          PanAdmin.toast((res && res.msg) || "已保存");
          this.dialog = false;
          await PanAdmin.refreshConfig(this.store);
          this.clearSelection();
        } catch (e) {
          PanAdmin.toast(e.message || "保存失败");
        } finally {
          this.saving = false;
        }
      },
      remove: async function () {
        var ids = this.selected.slice();
        if (!ids.length) {
          PanAdmin.toast("请选择需要删除的记录");
          return;
        }
        try {
          var res = await PanAdmin.api.del("/bypass", ids);
          PanAdmin.toast((res && res.msg) || "删除成功");
          await PanAdmin.refreshConfig(this.store);
          this.clearSelection();
        } catch (e) {
          PanAdmin.toast(e.message || "删除失败");
        }
      },
    },
    template:
      '<div>' +
      '  <page-header title="分流下载" subtitle="多账号分流访问与下载"></page-header>' +
      '  <div class="admin-card overflow-hidden">' +
      '    <div class="admin-card-body">' +
      '      <div class="admin-toolbar">' +
      '        <pa-tooltip content="添加"><button type="button" class="admin-toolbar-btn" @click="openAdd"><pa-icon name="circle-plus" :size="18"></pa-icon></button></pa-tooltip>' +
      '        <pa-tooltip content="修改"><button type="button" class="admin-toolbar-btn" @click="openEdit"><pa-icon name="pencil" :size="18"></pa-icon></button></pa-tooltip>' +
      '        <pa-tooltip content="删除"><button type="button" class="admin-toolbar-btn is-danger" @click="remove"><pa-icon name="circle-minus" :size="18"></pa-icon></button></pa-tooltip>' +
      "      </div>" +
      '      <div class="admin-card-main">' +
      '        <table class="admin-table" :class="{\'admin-table--empty\': !rows.length}">' +
      "          <thead><tr><th class=\"w-10\"></th><th>分流名称</th><th>绑定网盘</th></tr></thead>" +
      '          <tbody v-if="rows.length">' +
      '            <tr v-for="bp in rows" :key="bp.id" :class="{selected: isSelected(bp.id)}" @click="toggle(bp.id)">' +
      '              <td @click.stop><pa-check :model-value="isSelected(bp.id)" @update:model-value="toggle(bp.id)"></pa-check></td>' +
      "              <td>{{ bp.name }}</td>" +
      "              <td>" +
      '                <span v-for="ac in (bp.accounts || [])" :key="ac.id" class="inline-flex mr-2 mb-1 px-2 py-0.5 rounded-full text-xs border" style="border-color: var(--admin-border)">{{ ac.name }}</span>' +
      "              </td>" +
      "            </tr>" +
      "          </tbody>" +
      "        </table>" +
      '        <div v-if="!rows.length" class="admin-empty">暂无数据</div>' +
      "      </div>" +
      "    </div>" +
      "  </div>" +
      '  <pa-modal :open="dialog" :title="title" @close="dialog=false">' +
      '    <pa-field label="分流名称" help="用于替代访问路径中的网盘名称，添加后 DB 缓存策略的网盘需重新缓存">' +
      '      <input class="admin-input" v-model="form.name" />' +
      "    </pa-field>" +
      '    <pa-field label="绑定网盘">' +
      '      <div class="max-h-48 overflow-auto space-y-2">' +
      '        <label v-for="ac in accounts" :key="ac.id" class="flex items-center gap-2 text-sm cursor-pointer">' +
      '          <pa-check :model-value="!!form.accountIds[ac.id]" @update:model-value="toggleAccount(ac.id, $event)"></pa-check> {{ ac.name }}' +
      "        </label>" +
      "      </div>" +
      "    </pa-field>" +
      '    <template #footer>' +
      '      <button type="button" class="admin-btn admin-btn-ghost" @click="dialog=false">取消</button>' +
      '      <button type="button" class="admin-btn admin-btn-primary" :disabled="saving" @click="save"><pa-icon name="save" :size="16"></pa-icon> 保存</button>' +
      "    </template>" +
      "  </pa-modal>" +
      "</div>",
  };

  PanAdmin.pages.Pwd = {
    name: "PagePwd",
    inject: ["store"],
    data: function () {
      return {
        selected: [],
        dialog: false,
        title: "添加",
        form: { id: "", file_path: "", password: "", expire_at_local: "", info: "" },
        saving: false,
      };
    },
    computed: {
      rows: function () {
        return this.store.config.pwd_files || [];
      },
    },
    methods: {
      formatExpire: PanAdmin.formatExpire,
      isSelected: function (id) {
        return this.selected.indexOf(id) >= 0;
      },
      toggle: function (id) {
        var i = this.selected.indexOf(id);
        if (i >= 0) this.selected.splice(i, 1);
        else this.selected.push(id);
      },
      clearSelection: function () {
        this.selected = [];
      },
      openAdd: function () {
        this.title = "添加";
        this.form = {
          id: "",
          file_path: "",
          password: "",
          expire_at_local: "",
          info: "",
        };
        this.dialog = true;
      },
      openEdit: function () {
        var ids = this.selected.slice();
        if (ids.length !== 1) {
          PanAdmin.toast("请选择一条需要修改的记录");
          return;
        }
        var row = this.rows.find(function (r) {
          return r.id === ids[0];
        });
        if (!row) return;
        this.title = "修改";
        this.form = {
          id: row.id,
          file_path: row.file_path,
          password: row.password,
          expire_at_local: PanAdmin.unixToDatetimeLocal(row.expire_at),
          info: row.info || "",
        };
        this.dialog = true;
      },
      save: async function () {
        this.saving = true;
        try {
          var payload = {
            id: this.form.id,
            file_path: this.form.file_path,
            password: this.form.password,
            info: this.form.info,
            expire_at: PanAdmin.datetimeLocalToUnix(this.form.expire_at_local),
          };
          var res = await PanAdmin.api.post("/password/file", payload);
          PanAdmin.toast((res && res.msg) || "已保存");
          this.dialog = false;
          await PanAdmin.refreshConfig(this.store);
          this.clearSelection();
        } catch (e) {
          PanAdmin.toast(e.message || "保存失败");
        } finally {
          this.saving = false;
        }
      },
      remove: async function () {
        var ids = this.selected.slice();
        if (!ids.length) {
          PanAdmin.toast("请选择需要删除的记录");
          return;
        }
        try {
          var res = await PanAdmin.api.del("/password/file", ids);
          PanAdmin.toast((res && res.msg) || "删除成功");
          await PanAdmin.refreshConfig(this.store);
          this.clearSelection();
        } catch (e) {
          PanAdmin.toast(e.message || "删除失败");
        }
      },
      share: async function () {
        var ids = this.selected.slice();
        if (ids.length !== 1) {
          PanAdmin.toast("请选择需要分享的文件（夹）");
          return;
        }
        try {
          var prefix =
            window.location.protocol +
            "//" +
            window.location.host +
            (this.store.config.path_prefix || "") +
            "/s/";
          var res = await PanAdmin.api.post("/password/file/share/info", {
            prefix: prefix,
            id: ids[0],
          });
          if (navigator.clipboard && res && res.msg) {
            await navigator.clipboard.writeText(res.msg);
            PanAdmin.toast("链接和密码已复制到剪切板");
          } else {
            PanAdmin.toast((res && res.msg) || "已生成");
          }
          await PanAdmin.refreshConfig(this.store);
        } catch (e) {
          PanAdmin.toast(e.message || "分享失败");
        }
      },
      importFile: async function (e) {
        var file = e.target.files && e.target.files[0];
        e.target.value = "";
        if (!file) return;
        var fd = new FormData();
        fd.append("file", file);
        try {
          var res = await PanAdmin.api.postForm("/password/file/upload", fd);
          PanAdmin.toast((res && res.msg) || "导入成功");
          await PanAdmin.refreshConfig(this.store);
        } catch (err) {
          PanAdmin.toast(err.message || "导入失败");
        }
      },
    },
    template:
      '<div>' +
      '  <page-header title="密码文件（夹）" subtitle="文件夹设置密码访问，设置后的文件将无法被搜索到"></page-header>' +
      '  <div class="admin-card overflow-hidden">' +
      '    <div class="admin-card-body">' +
      '      <div class="admin-toolbar">' +
      '        <pa-tooltip content="添加"><button type="button" class="admin-toolbar-btn" @click="openAdd"><pa-icon name="circle-plus" :size="18"></pa-icon></button></pa-tooltip>' +
      '        <pa-tooltip content="修改"><button type="button" class="admin-toolbar-btn" @click="openEdit"><pa-icon name="pencil" :size="18"></pa-icon></button></pa-tooltip>' +
      '        <pa-tooltip content="删除"><button type="button" class="admin-toolbar-btn is-danger" @click="remove"><pa-icon name="circle-minus" :size="18"></pa-icon></button></pa-tooltip>' +
      '        <pa-tooltip content="创建短链并复制"><button type="button" class="admin-toolbar-btn" @click="share"><pa-icon name="share-2" :size="18"></pa-icon></button></pa-tooltip>' +
      '        <pa-tooltip content="导入"><label class="admin-toolbar-btn cursor-pointer"><pa-icon name="upload" :size="18"></pa-icon><input type="file" class="hidden" @change="importFile" /></label></pa-tooltip>' +
      "      </div>" +
      '      <div class="admin-card-main">' +
      '        <table class="admin-table" :class="{\'admin-table--empty\': !rows.length}">' +
      "          <thead><tr><th class=\"w-10\"></th><th>文件（夹）路径</th><th>访问密码</th><th>有效期</th><th>备注</th></tr></thead>" +
      '          <tbody v-if="rows.length">' +
      '            <tr v-for="row in rows" :key="row.id" :class="{selected: isSelected(row.id)}" @click="toggle(row.id)">' +
      '              <td @click.stop><pa-check :model-value="isSelected(row.id)" @update:model-value="toggle(row.id)"></pa-check></td>' +
      "              <td>{{ row.file_path }}</td>" +
      "              <td>{{ row.password }}</td>" +
      "              <td>{{ formatExpire(row.expire_at) }}</td>" +
      "              <td>{{ row.info }}</td>" +
      "            </tr>" +
      "          </tbody>" +
      "        </table>" +
      '        <div v-if="!rows.length" class="admin-empty">暂无数据</div>' +
      "      </div>" +
      "    </div>" +
      "  </div>" +
      '  <pa-modal :open="dialog" :title="title" @close="dialog=false">' +
      '    <pa-field label="文件（夹）路径" help="PanIndex 虚拟路径，格式：/{name}/a/b"><input class="admin-input" v-model="form.file_path" /></pa-field>' +
      '    <pa-field label="访问密码" help="为空将生成随机密码"><input class="admin-input" v-model="form.password" /></pa-field>' +
      '    <pa-field label="有效期至" help="为空则表示永不过期"><input class="admin-input" type="datetime-local" v-model="form.expire_at_local" /></pa-field>' +
      '    <pa-field label="备注"><textarea class="admin-textarea" rows="2" v-model="form.info"></textarea></pa-field>' +
      '    <template #footer>' +
      '      <button type="button" class="admin-btn admin-btn-ghost" @click="dialog=false">取消</button>' +
      '      <button type="button" class="admin-btn admin-btn-primary" :disabled="saving" @click="save"><pa-icon name="save" :size="16"></pa-icon> 保存</button>' +
      "    </template>" +
      "  </pa-modal>" +
      "</div>",
  };

  PanAdmin.pages.Cache = {
    name: "PageCache",
    inject: ["store"],
    data: function () {
      return {
        searchKey: "",
        list: [],
        loading: false,
        clearDialog: false,
        clearForm: { path: "", is_loop_children: "0" },
      };
    },
    mounted: function () {
      this.load();
    },
    methods: {
      load: async function () {
        this.loading = true;
        try {
          var q = this.searchKey
            ? "?path=" + encodeURIComponent(this.searchKey)
            : "";
          var res = await PanAdmin.api.get("/cache/list" + q);
          this.list = (res && res.data) || [];
        } catch (e) {
          PanAdmin.toast(e.message || "加载失败");
        } finally {
          this.loading = false;
        }
      },
      onSearchKey: function (e) {
        if (e.key === "Enter") this.load();
      },
      openClear: function (path) {
        this.clearForm = { path: path || "", is_loop_children: "0" };
        this.clearDialog = true;
      },
      clear: async function () {
        try {
          var res = await PanAdmin.api.post("/cache/clear", {
            path: this.clearForm.path,
            is_loop_children: this.clearForm.is_loop_children,
          });
          PanAdmin.toast((res && res.msg) || "清理成功");
          this.clearDialog = false;
          this.load();
        } catch (e) {
          PanAdmin.toast(e.message || "清理失败");
        }
      },
      viewJson: function (path) {
        var url =
          this.store.apiUrl + "/cache?path=" + encodeURIComponent(path);
        window.open(url, "_blank");
      },
    },
    template:
      '<div>' +
      '  <page-header title="缓存管理" subtitle="列表只展示部分信息，更详细的信息请使用搜索功能"></page-header>' +
      '  <div class="admin-card overflow-hidden">' +
      '    <div class="flex justify-end p-3 border-b" style="border-color: var(--admin-border)">' +
      '      <div class="relative w-full max-w-xs">' +
      '        <input class="admin-input pr-10" v-model="searchKey" placeholder="请输入路径" @keydown="onSearchKey" />' +
      '        <button type="button" class="absolute right-2 top-1/2 -translate-y-1/2 admin-btn admin-btn-ghost !p-1" @click="load"><pa-icon name="search" :size="16"></pa-icon></button>' +
      "      </div>" +
      "    </div>" +
      '    <table class="admin-table">' +
      "      <thead><tr><th>文件（夹）路径</th><th>缓存策略</th><th>缓存时间</th><th>操作</th></tr></thead>" +
      "      <tbody>" +
      '        <tr v-for="(item, idx) in list" :key="idx">' +
      "          <td class=\"break-all\">{{ item.file_path }}</td>" +
      "          <td>{{ item.cache_policy }}</td>" +
      "          <td>{{ item.cache_time }}</td>" +
      "          <td class=\"whitespace-nowrap\">" +
      '            <pa-tooltip content="清理" placement="top"><button type="button" class="admin-btn admin-btn-ghost !p-2" @click="openClear(item.file_path)"><pa-icon name="eraser" :size="16"></pa-icon></button></pa-tooltip>' +
      '            <pa-tooltip content="查看 JSON" placement="top"><button type="button" class="admin-btn admin-btn-ghost !p-2" @click="viewJson(item.file_path)"><pa-icon name="file-json" :size="16"></pa-icon></button></pa-tooltip>' +
      "          </td>" +
      "        </tr>" +
      '        <tr v-if="!list.length && !loading"><td colspan="4" class="text-center py-8" style="color: var(--admin-muted)">暂无数据</td></tr>' +
      "      </tbody>" +
      "    </table>" +
      "  </div>" +
      '  <pa-modal :open="clearDialog" title="清理条件" @close="clearDialog=false">' +
      '    <pa-field label="文件（夹）路径" help="db 缓存策略，清理后会重建"><input class="admin-input" v-model="clearForm.path" /></pa-field>' +
      '    <div class="mb-4"><pa-switch v-model="clearForm.is_loop_children" true-value="0" false-value="1" on-label="递归清理子目录" off-label="仅清理当前路径"></pa-switch></div>' +
      '    <template #footer>' +
      '      <button type="button" class="admin-btn admin-btn-ghost" @click="clearDialog=false">取消</button>' +
      '      <button type="button" class="admin-btn admin-btn-primary" @click="clear"><pa-icon name="save" :size="16"></pa-icon> 确定</button>' +
      "    </template>" +
      "  </pa-modal>" +
      "</div>",
  };
})(window);
