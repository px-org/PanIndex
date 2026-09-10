(function (global) {
  var Vue = global.Vue;
  var VueRouter = global.VueRouter;
  var PanAdmin = global.PanAdmin || (global.PanAdmin = {});

  function boot() {
    if (!Vue || !VueRouter || !global.__ADMIN__) {
      console.error("PanAdmin bootstrap failed: missing deps");
      return;
    }

    var raw = global.__ADMIN__;
    var store = Vue.reactive({
      config: raw.config || {},
      version: raw.version || "",
      theme: raw.theme || "auto",
      apiUrl: raw.apiUrl || "",
      adminUrl: raw.adminUrl || "",
      pathPrefix: raw.pathPrefix || "",
      dark: false,
    });

    store.dark = PanAdmin.applyTheme(store.theme);

    var routes = [
      { path: "/", redirect: "/dashboard" },
      { path: "/dashboard", component: PanAdmin.pages.Dashboard, meta: { title: "首页" } },
      { path: "/common", component: PanAdmin.pages.Common, meta: { title: "基础配置", group: "common" } },
      { path: "/appearance", component: PanAdmin.pages.Appearance, meta: { title: "外观", group: "common" } },
      { path: "/view", component: PanAdmin.pages.View, meta: { title: "文件预览", group: "common" } },
      { path: "/access", component: PanAdmin.pages.Access, meta: { title: "访问控制", group: "safe" } },
      { path: "/pwd", component: PanAdmin.pages.Pwd, meta: { title: "文件夹加密", group: "safe" } },
      { path: "/hide", component: PanAdmin.pages.Hide, meta: { title: "隐藏文件（夹）", group: "safe" } },
      { path: "/safety", component: PanAdmin.pages.Safety, meta: { title: "防盗链", group: "safe" } },
      { path: "/disk", component: PanAdmin.pages.Disk, meta: { title: "网盘挂载" } },
      { path: "/share", component: PanAdmin.pages.Share, meta: { title: "分享短链" } },
      { path: "/bypass", component: PanAdmin.pages.Bypass, meta: { title: "分流下载" } },
      { path: "/cache", component: PanAdmin.pages.Cache, meta: { title: "缓存管理" } },
      { path: "/webdav", component: PanAdmin.pages.Webdav, meta: { title: "WebDav" } },
    ];

    var router = VueRouter.createRouter({
      history: VueRouter.createWebHistory(store.adminUrl || "/admin"),
      routes: routes,
    });

    var App = {
      name: "AdminApp",
      components: {
        PaIcon: PanAdmin.Icon,
        PaModal: PanAdmin.Modal,
        PageHeader: PanAdmin.PageHeader,
        PaField: PanAdmin.Field,
        PaTooltip: PanAdmin.Tooltip,
        ToastHost: PanAdmin.ToastHost,
      },
      provide: function () {
        return { store: store };
      },
      data: function () {
        return {
          store: store,
          drawerOpen:
            typeof window !== "undefined" &&
            window.matchMedia("(min-width: 1024px)").matches,
          openCommon: true,
          openSafe: false,
          configDialog: false,
          configJson: "",
          savingConfig: false,
          navTop: [
            { path: "/disk", label: "网盘挂载", icon: "hard-drive" },
            { path: "/share", label: "分享短链", icon: "share-2" },
            { path: "/bypass", label: "分流下载", icon: "split" },
            { path: "/cache", label: "缓存管理", icon: "database" },
            { path: "/webdav", label: "WebDav", icon: "inbox" },
          ],
          navHome: { path: "/dashboard", label: "首页", icon: "layout-dashboard" },
        };
      },
      watch: {
        "$route.path": {
          immediate: true,
          handler: function (p, oldP) {
            if (["/common", "/appearance", "/view"].indexOf(p) >= 0) {
              this.openCommon = true;
            }
            if (["/access", "/pwd", "/hide", "/safety"].indexOf(p) >= 0) {
              this.openSafe = true;
            }
            // 仅移动端在路由切换后自动收起抽屉
            if (
              oldP != null &&
              typeof window !== "undefined" &&
              window.matchMedia("(max-width: 1023px)").matches
            ) {
              this.drawerOpen = false;
            }
          },
        },
      },
      methods: {
        isActive: function (path) {
          return this.$route.path === path;
        },
        toggleDrawer: function () {
          this.drawerOpen = !this.drawerOpen;
        },
        toggleTheme: function () {
          var next = this.store.dark ? "light" : "dark";
          this.store.theme = next;
          this.store.dark = PanAdmin.applyTheme(next);
          PanAdmin.setCookie("admin_theme", next, 3650);
        },
        openConfigDialog: async function () {
          try {
            var cfg = await PanAdmin.api.getConfig();
            this.configJson = JSON.stringify(cfg, null, 2);
            this.configDialog = true;
          } catch (e) {
            PanAdmin.toast(e.message || "加载配置失败");
          }
        },
        copyConfig: async function () {
          try {
            await navigator.clipboard.writeText(this.configJson);
            PanAdmin.toast("已复制到剪切板");
          } catch (e) {
            PanAdmin.toast("复制失败");
          }
        },
        importConfig: async function () {
          this.savingConfig = true;
          try {
            var res = await PanAdmin.api.uploadConfig(this.configJson);
            PanAdmin.toast((res && res.msg) || "导入成功");
            this.configDialog = false;
            await PanAdmin.refreshConfig(this.store);
          } catch (e) {
            PanAdmin.toast(e.message || "导入失败");
          } finally {
            this.savingConfig = false;
          }
        },
        goHome: function () {
          window.open((this.store.pathPrefix || "") + "/", "_blank");
        },
      },
      template:
        '<div class="h-full max-w-full overflow-hidden flex flex-col">' +
        '  <header class="shrink-0 sticky top-0 z-40 h-14 flex items-center gap-3 px-3 text-white shadow" style="background: var(--admin-topbar)">' +
        '    <button type="button" class="admin-btn !text-white !border-transparent hover:!bg-white/10 !p-2" @click="toggleDrawer" :title="drawerOpen ? \'收起导航\' : \'展开导航\'"><pa-icon :name="drawerOpen ? \'panel-left-close\' : \'panel-left-open\'" :size="20"></pa-icon></button>' +
        '    <button type="button" class="font-semibold tracking-wide truncate max-w-[12rem] sm:max-w-xs" @click="goHome">{{ store.config.site_name || "PanIndex" }}</button>' +
        '    <span class="opacity-90 text-sm">配置</span>' +
        '    <div class="flex-1"></div>' +
        '    <pa-tooltip content="配置（JSON）" placement="bottom"><button type="button" class="admin-btn !text-white !border-transparent hover:!bg-white/10 !p-2" @click="openConfigDialog"><pa-icon name="settings" :size="18"></pa-icon></button></pa-tooltip>' +
        '    <pa-tooltip content="明暗主题切换" placement="bottom"><button type="button" class="admin-btn !text-white !border-transparent hover:!bg-white/10 !p-2" @click="toggleTheme"><pa-icon :name="store.dark ? \'sun\' : \'moon\'" :size="18"></pa-icon></button></pa-tooltip>' +
        '    <pa-tooltip content="退出登录" placement="bottom"><a :href="store.adminUrl + \'/logout\'" class="admin-btn !text-white !border-transparent hover:!bg-white/10 !p-2"><pa-icon name="log-out" :size="18"></pa-icon></a></pa-tooltip>' +
        "  </header>" +

        '  <div class="flex flex-1 min-h-0 min-w-0 overflow-hidden">' +
        '    <div v-if="drawerOpen" class="fixed inset-0 z-30 bg-black/40 lg:hidden" @click="drawerOpen=false"></div>' +
        '    <aside :class="[\'admin-sidebar z-40 top-14 bottom-0 shrink-0 border-r flex flex-col fixed lg:static lg:top-auto lg:bottom-auto\', drawerOpen ? \'is-open\' : \'is-closed\']" style="background: var(--admin-sidebar); border-color: var(--admin-border)">' +
        '      <nav class="flex-1 overflow-y-auto overflow-x-hidden p-3 space-y-1 text-sm">' +
        '        <router-link :to="navHome.path" class="flex items-center gap-2 px-3 py-2 rounded-lg" :class="isActive(navHome.path) ? \'bg-[var(--admin-primary)]/10 text-[var(--admin-primary)] font-medium\' : \'hover:bg-black/5 dark:hover:bg-white/5\'">' +
        '          <pa-icon :name="navHome.icon" :size="16" class-name="text-sky-500"></pa-icon><span>{{ navHome.label }}</span>' +
        "        </router-link>" +
        '        <button type="button" class="w-full flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-black/5 dark:hover:bg-white/5" @click="openCommon=!openCommon">' +
        '          <pa-icon name="settings" :size="16" class-name="text-violet-500"></pa-icon><span class="flex-1 text-left">通用</span><pa-icon :name="openCommon ? \'chevron-down\' : \'chevron-right\'" :size="16"></pa-icon>' +
        "        </button>" +
        '        <div v-show="openCommon" class="pl-4 space-y-1">' +
        '          <router-link to="/common" class="block px-3 py-2 rounded-lg" :class="isActive(\'/common\') ? \'bg-[var(--admin-primary)]/10 text-[var(--admin-primary)] font-medium\' : \'hover:bg-black/5 dark:hover:bg-white/5\'">基础配置</router-link>' +
        '          <router-link to="/appearance" class="block px-3 py-2 rounded-lg" :class="isActive(\'/appearance\') ? \'bg-[var(--admin-primary)]/10 text-[var(--admin-primary)] font-medium\' : \'hover:bg-black/5 dark:hover:bg-white/5\'">外观</router-link>' +
        '          <router-link to="/view" class="block px-3 py-2 rounded-lg" :class="isActive(\'/view\') ? \'bg-[var(--admin-primary)]/10 text-[var(--admin-primary)] font-medium\' : \'hover:bg-black/5 dark:hover:bg-white/5\'">文件预览</router-link>' +
        "        </div>" +
        '        <button type="button" class="w-full flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-black/5 dark:hover:bg-white/5" @click="openSafe=!openSafe">' +
        '          <pa-icon name="shield" :size="16" class-name="text-rose-500"></pa-icon><span class="flex-1 text-left">安全</span><pa-icon :name="openSafe ? \'chevron-down\' : \'chevron-right\'" :size="16"></pa-icon>' +
        "        </button>" +
        '        <div v-show="openSafe" class="pl-4 space-y-1">' +
        '          <router-link to="/access" class="block px-3 py-2 rounded-lg" :class="isActive(\'/access\') ? \'bg-[var(--admin-primary)]/10 text-[var(--admin-primary)] font-medium\' : \'hover:bg-black/5 dark:hover:bg-white/5\'">访问控制</router-link>' +
        '          <router-link to="/pwd" class="block px-3 py-2 rounded-lg" :class="isActive(\'/pwd\') ? \'bg-[var(--admin-primary)]/10 text-[var(--admin-primary)] font-medium\' : \'hover:bg-black/5 dark:hover:bg-white/5\'">文件夹加密</router-link>' +
        '          <router-link to="/hide" class="block px-3 py-2 rounded-lg" :class="isActive(\'/hide\') ? \'bg-[var(--admin-primary)]/10 text-[var(--admin-primary)] font-medium\' : \'hover:bg-black/5 dark:hover:bg-white/5\'">隐藏文件（夹）</router-link>' +
        '          <router-link to="/safety" class="block px-3 py-2 rounded-lg" :class="isActive(\'/safety\') ? \'bg-[var(--admin-primary)]/10 text-[var(--admin-primary)] font-medium\' : \'hover:bg-black/5 dark:hover:bg-white/5\'">防盗链</router-link>' +
        "        </div>" +
        '        <router-link v-for="item in navTop" :key="item.path" :to="item.path" class="flex items-center gap-2 px-3 py-2 rounded-lg" :class="isActive(item.path) ? \'bg-[var(--admin-primary)]/10 text-[var(--admin-primary)] font-medium\' : \'hover:bg-black/5 dark:hover:bg-white/5\'">' +
        '          <pa-icon :name="item.icon" :size="16"></pa-icon><span>{{ item.label }}</span>' +
        "        </router-link>" +
        "      </nav>" +
        '      <div class="p-4 text-center text-sm border-t" style="border-color: var(--admin-border)">' +
        '        <a class="font-medium text-[var(--admin-primary)]" :href="\'https://github.com/px-org/PanIndex/releases/tag/\' + store.version" target="_blank">{{ store.version }}</a>' +
        "      </div>" +
        "    </aside>" +

        '    <main class="flex-1 min-w-0 overflow-y-auto overflow-x-hidden p-4 sm:p-6">' +
        '      <router-view></router-view>' +
        "    </main>" +
        "  </div>" +

        '  <pa-modal :open="configDialog" title="配置导入" width="max-w-2xl" @close="configDialog=false">' +
        '    <pa-field label="后台配置 JSON" help="配置导入会覆盖现有配置，请谨慎操作，导入之前先备份原有配置">' +
        '      <textarea class="admin-textarea font-mono text-xs" rows="12" v-model="configJson"></textarea>' +
        "    </pa-field>" +
        '    <template #footer>' +
        '      <button type="button" class="admin-btn admin-btn-ghost" @click="copyConfig"><pa-icon name="copy" :size="16"></pa-icon> 复制</button>' +
        '      <button type="button" class="admin-btn admin-btn-primary" :disabled="savingConfig" @click="importConfig"><pa-icon name="save" :size="16"></pa-icon> 确定</button>' +
        '      <button type="button" class="admin-btn admin-btn-ghost" @click="configDialog=false"><pa-icon name="x" :size="16"></pa-icon> 关闭</button>' +
        "    </template>" +
        "  </pa-modal>" +
        "  <toast-host></toast-host>" +
        "</div>",
    };

    var app = Vue.createApp(App);
    app.component("pa-icon", PanAdmin.Icon);
    app.component("pa-modal", PanAdmin.Modal);
    app.component("page-header", PanAdmin.PageHeader);
    app.component("pa-field", PanAdmin.Field);
    app.component("pa-tooltip", PanAdmin.Tooltip);
    app.component("pa-switch", PanAdmin.Switch);
    app.component("pa-radio-group", PanAdmin.RadioGroup);
    app.component("pa-check", PanAdmin.Check);
    app.component("toast-host", PanAdmin.ToastHost);
    app.use(router);
    app.mount("#app");
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", boot);
  } else {
    boot();
  }
})(window);
