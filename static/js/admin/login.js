(function (global) {
  var Vue = global.Vue;
  var PanAdmin = global.PanAdmin || (global.PanAdmin = {});
  var data = global.__LOGIN__ || {};

  function boot() {
    if (!Vue) return;
    var theme = data.theme || "auto";
    var dark = PanAdmin.applyTheme(theme);

    var app = Vue.createApp({
      components: {
        PaIcon: PanAdmin.Icon,
        ToastHost: PanAdmin.ToastHost,
      },
      data: function () {
        return {
          dark: dark,
          user: "",
          password: "",
          adminUrl: data.adminUrl || "/admin",
          pathPrefix: data.pathPrefix || "",
        };
      },
      mounted: function () {
        if (data.error && data.msg) {
          PanAdmin.toast(data.msg);
        }
      },
      methods: {
        toggleTheme: function () {
          var next = this.dark ? "light" : "dark";
          this.dark = PanAdmin.applyTheme(next);
          PanAdmin.setCookie("admin_theme", next, 3650);
        },
      },
      template:
        '<div class="min-h-screen flex items-center justify-center p-4 relative" style="background: var(--admin-surface)">' +
        '  <button type="button" class="absolute top-4 right-4 admin-btn admin-btn-ghost !p-2" @click="toggleTheme" title="明暗切换">' +
        '    <pa-icon :name="dark ? \'sun\' : \'moon\'" :size="18"></pa-icon>' +
        "  </button>" +
        '  <div class="admin-card w-full max-w-md p-8">' +
        '    <div class="text-center mb-8">' +
        '      <a :href="(pathPrefix || \'/\')" class="text-2xl font-semibold tracking-tight text-[var(--admin-primary)]">PanIndex</a>' +
        '      <p class="mt-2 text-sm" style="color: var(--admin-muted)">管理员登录</p>' +
        "    </div>" +
        '    <form :action="adminUrl + \'/login\'" method="post" class="space-y-5">' +
        '      <div>' +
        '        <label class="admin-label">账号</label>' +
        '        <div class="relative">' +
        '          <span class="absolute left-3 top-1/2 -translate-y-1/2 opacity-50"><pa-icon name="user" :size="16"></pa-icon></span>' +
        '          <input class="admin-input !pl-10" name="user" v-model="user" type="text" required autocomplete="username" />' +
        "        </div>" +
        "      </div>" +
        '      <div>' +
        '        <label class="admin-label">密码</label>' +
        '        <div class="relative">' +
        '          <span class="absolute left-3 top-1/2 -translate-y-1/2 opacity-50"><pa-icon name="lock" :size="16"></pa-icon></span>' +
        '          <input class="admin-input !pl-10" name="password" v-model="password" type="password" required autocomplete="current-password" />' +
        "        </div>" +
        "      </div>" +
        '      <button type="submit" class="admin-btn admin-btn-primary w-full justify-center py-2.5">' +
        '        <pa-icon name="log-in" :size="16"></pa-icon> 登录' +
        "      </button>" +
        "    </form>" +
        "  </div>" +
        "  <toast-host></toast-host>" +
        "</div>",
    });
    app.component("pa-icon", PanAdmin.Icon);
    app.component("toast-host", PanAdmin.ToastHost);
    app.mount("#login-app");
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", boot);
  } else {
    boot();
  }
})(window);
