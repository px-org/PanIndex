(function (global) {
  var PanAdmin = global.PanAdmin || (global.PanAdmin = {});
  PanAdmin.pages = PanAdmin.pages || {};

  function formatBytes(n) {
    n = Number(n) || 0;
    if (n < 1024) return n + " B";
    var u = ["KB", "MB", "GB", "TB", "PB"];
    var i = -1;
    do {
      n = n / 1024;
      i++;
    } while (n >= 1024 && i < u.length - 1);
    return n.toFixed(n >= 100 ? 0 : n >= 10 ? 1 : 2) + " " + u[i];
  }

  function formatNumber(n) {
    n = Number(n) || 0;
    return n.toLocaleString("zh-CN");
  }

  function formatUptime(sec) {
    sec = Math.max(0, Math.floor(Number(sec) || 0));
    var d = Math.floor(sec / 86400);
    var h = Math.floor((sec % 86400) / 3600);
    var m = Math.floor((sec % 3600) / 60);
    var s = sec % 60;
    if (d > 0) return d + "天 " + h + "时 " + m + "分";
    if (h > 0) return h + "时 " + m + "分 " + s + "秒";
    if (m > 0) return m + "分 " + s + "秒";
    return s + "秒";
  }

  function shortDate(d) {
    if (!d || d.length < 10) return d || "";
    return d.slice(5);
  }

  function cssVar(name, fallback) {
    var v = getComputedStyle(document.documentElement).getPropertyValue(name);
    return (v && v.trim()) || fallback;
  }

  PanAdmin.pages.Dashboard = {
    name: "PageDashboard",
    inject: ["store"],
    data: function () {
      return {
        loading: true,
        refreshing: false,
        autoTimer: null,
        chartInst: null,
        data: {
          overview: {},
          trend: [],
          top_files: [],
          system: {},
        },
      };
    },
    computed: {
      ov: function () {
        return this.data.overview || {};
      },
      sys: function () {
        return this.data.system || {};
      },
      cacheTotal: function () {
        return (
          (this.ov.cache_files || 0) +
          (this.ov.cache_file || 0) +
          (this.ov.cache_url || 0) +
          (this.ov.cache_db_nodes || 0)
        );
      },
    },
    watch: {
      "store.dark": function () {
        this.renderChart(true);
      },
      "data.trend": {
        deep: true,
        handler: function () {
          this.renderChart(false);
        },
      },
    },
    mounted: function () {
      this.load(true);
      var self = this;
      this.autoTimer = setInterval(function () {
        self.load(false);
      }, 30000);
    },
    beforeUnmount: function () {
      if (this.autoTimer) clearInterval(this.autoTimer);
      this.destroyChart();
    },
    methods: {
      formatBytes: formatBytes,
      formatNumber: formatNumber,
      formatUptime: formatUptime,
      shortDate: shortDate,
      destroyChart: function () {
        if (this.chartInst) {
          this.chartInst.destroy();
          this.chartInst = null;
        }
      },
      chartOptions: function (categories, series) {
        var primary = cssVar("--admin-primary", "#3949ab");
        var muted = cssVar("--admin-muted", "#6b7280");
        var border = cssVar("--admin-border", "rgba(15, 23, 42, 0.08)");
        var card = cssVar("--admin-card", "#ffffff");
        var text = cssVar("--admin-text", "#1f2937");
        var dark = !!(this.store && this.store.dark);

        return {
          chart: {
            type: "area",
            height: 260,
            fontFamily: '"Noto Sans SC", "Segoe UI", system-ui, sans-serif',
            toolbar: { show: false },
            zoom: { enabled: false },
            animations: {
              enabled: true,
              easing: "easeinout",
              speed: 650,
              animateGradually: { enabled: true, delay: 80 },
            },
            parentHeightOffset: 0,
            background: "transparent",
          },
          series: [{ name: "下载量", data: series }],
          colors: [primary],
          dataLabels: { enabled: false },
          stroke: {
            curve: "smooth",
            width: 3,
            lineCap: "round",
          },
          fill: {
            type: "gradient",
            gradient: {
              shadeIntensity: 1,
              opacityFrom: 0.42,
              opacityTo: 0.04,
              stops: [0, 90, 100],
            },
          },
          markers: {
            size: 0,
            colors: [card],
            strokeColors: primary,
            strokeWidth: 2,
            hover: { size: 6, sizeOffset: 0 },
          },
          grid: {
            borderColor: border,
            strokeDashArray: 4,
            padding: { left: 8, right: 8, top: 8, bottom: 0 },
            xaxis: { lines: { show: false } },
            yaxis: { lines: { show: true } },
          },
          xaxis: {
            categories: categories,
            tickAmount: Math.min(7, Math.max(categories.length - 1, 1)),
            axisBorder: { show: false },
            axisTicks: { show: false },
            labels: {
              style: { colors: muted, fontSize: "11px" },
              formatter: function (v) {
                return shortDate(String(v || ""));
              },
            },
            tooltip: { enabled: false },
            crosshairs: {
              show: true,
              width: 1,
              position: "front",
              opacity: 0.85,
              stroke: {
                color: primary,
                width: 1,
                dashArray: 4,
              },
              fill: {
                type: "gradient",
                gradient: {
                  colorFrom: primary,
                  colorTo: primary,
                  stops: [0, 100],
                  opacityFrom: 0.12,
                  opacityTo: 0.01,
                },
              },
            },
          },
          yaxis: {
            min: 0,
            forceNiceScale: true,
            labels: {
              style: { colors: muted, fontSize: "11px" },
              formatter: function (v) {
                return Math.round(v);
              },
            },
            crosshairs: { show: false },
          },
          tooltip: {
            theme: dark ? "dark" : "light",
            shared: true,
            intersect: false,
            followCursor: true,
            style: { fontSize: "12px" },
            x: {
              formatter: function (val, opts) {
                var i = opts && opts.dataPointIndex;
                return categories[i] || val;
              },
            },
            y: {
              formatter: function (v) {
                return formatNumber(v) + " 次";
              },
            },
            marker: { show: true },
          },
          legend: { show: false },
          noData: {
            text: "暂无趋势数据",
            style: { color: muted, fontSize: "13px" },
          },
          theme: { mode: dark ? "dark" : "light" },
        };
      },
      renderChart: function (forceRecreate) {
        var self = this;
        this.$nextTick(function () {
          var el = self.$refs.chartEl;
          if (!el || typeof ApexCharts === "undefined") return;

          var trend = self.data.trend || [];
          var categories = trend.map(function (p) {
            return p.date;
          });
          var series = trend.map(function (p) {
            return Number(p.count) || 0;
          });
          var options = self.chartOptions(categories, series);

          if (!self.chartInst || forceRecreate) {
            self.destroyChart();
            el.innerHTML = "";
            self.chartInst = new ApexCharts(el, options);
            self.chartInst.render();
            return;
          }

          self.chartInst.updateOptions(options, false, true);
        });
      },
      load: async function (first) {
        if (first) this.loading = true;
        else this.refreshing = true;
        try {
          var res = await PanAdmin.api.get("/dashboard");
          if (res && res.data) this.data = res.data;
          this.renderChart(first);
        } catch (e) {
          PanAdmin.toast(e.message || "加载仪表盘失败");
        } finally {
          this.loading = false;
          this.refreshing = false;
          this.renderChart(false);
        }
      },
      enableStats: async function () {
        try {
          await PanAdmin.api.saveConfig({ enable_download_statistics: "1" });
          this.store.config.enable_download_statistics = "1";
          if (this.data.overview) this.data.overview.stats_enabled = true;
          PanAdmin.toast("已开启下载统计");
          this.load(false);
        } catch (e) {
          PanAdmin.toast(e.message || "开启失败");
        }
      },
      barClass: function (pct) {
        pct = Number(pct) || 0;
        if (pct >= 90) return "is-danger";
        if (pct >= 70) return "is-warn";
        return "is-ok";
      },
    },
    template:
      '<div class="dash min-w-0">' +
      '  <div class="flex flex-wrap items-start justify-between gap-3 mb-5">' +
      "    <div>" +
      '      <h2 class="text-xl font-semibold tracking-tight">首页</h2>' +
      '      <p class="mt-1 text-sm" style="color: var(--admin-muted)">运行概览 · 每 30 秒自动刷新</p>' +
      "    </div>" +
      '    <button type="button" class="admin-btn admin-btn-ghost" :disabled="refreshing || loading" @click="load(false)">' +
      '      <pa-icon name="refresh-cw" :size="16" :class-name="refreshing ? \'dash-spin\' : \'\'"></pa-icon> 刷新' +
      "    </button>" +
      "  </div>" +

      '  <div v-if="!ov.stats_enabled" class="dash-banner mb-5">' +
      '    <div class="flex items-start gap-3">' +
      '      <pa-icon name="info" :size="18" class-name="mt-0.5 text-[var(--admin-primary)]"></pa-icon>' +
      "      <div class=\"flex-1 min-w-0\">" +
      '        <div class="font-medium text-sm">下载统计未开启</div>' +
      '        <div class="text-xs mt-1" style="color: var(--admin-muted)">开启后可统计今日下载量、趋势曲线与 Top 文件。也可在「基础配置」中随时关闭。</div>' +
      "      </div>" +
      '      <button type="button" class="admin-btn admin-btn-primary !py-1.5" @click="enableStats">立即开启</button>' +
      "    </div>" +
      "  </div>" +

      '  <div v-if="loading" class="admin-card p-10 text-center" style="color: var(--admin-muted)">加载中…</div>' +
      '  <template v-else>' +

      '    <div class="grid grid-cols-2 xl:grid-cols-4 gap-3 sm:gap-4 mb-5 min-w-0">' +
      '      <div class="dash-stat accent-indigo min-w-0">' +
      '        <div class="dash-stat-icon"><pa-icon name="download" :size="20"></pa-icon></div>' +
      '        <div class="dash-stat-label">今日下载</div>' +
      '        <div class="dash-stat-value">{{ formatNumber(ov.today_downloads) }}</div>' +
      '        <div class="dash-stat-sub">累计 {{ formatNumber(ov.total_downloads) }} 次</div>' +
      "      </div>" +
      '      <div class="dash-stat accent-teal min-w-0">' +
      '        <div class="dash-stat-icon"><pa-icon name="hard-drive" :size="20"></pa-icon></div>' +
      '        <div class="dash-stat-label">挂载网盘</div>' +
      '        <div class="dash-stat-value">{{ formatNumber(ov.account_count) }}</div>' +
      '        <div class="dash-stat-sub">状态正常 {{ formatNumber(ov.account_ok) }}</div>' +
      "      </div>" +
      '      <div class="dash-stat accent-amber min-w-0">' +
      '        <div class="dash-stat-icon"><pa-icon name="database" :size="20"></pa-icon></div>' +
      '        <div class="dash-stat-label">已缓存</div>' +
      '        <div class="dash-stat-value">{{ formatNumber(cacheTotal) }}</div>' +
      '        <div class="dash-stat-sub">内存 {{ formatNumber((ov.cache_files||0)+(ov.cache_file||0)+(ov.cache_url||0)) }} · DB {{ formatNumber(ov.cache_db_nodes) }}</div>' +
      "      </div>" +
      '      <div class="dash-stat accent-rose min-w-0">' +
      '        <div class="dash-stat-icon"><pa-icon name="share-2" :size="20"></pa-icon></div>' +
      '        <div class="dash-stat-label">分享 / 加密</div>' +
      '        <div class="dash-stat-value">{{ formatNumber(ov.share_count) }}</div>' +
      '        <div class="dash-stat-sub">密码夹 {{ formatNumber(ov.pwd_count) }} · 隐藏 {{ formatNumber(ov.hide_count) }}</div>' +
      "      </div>" +
      "    </div>" +

      '    <div class="grid grid-cols-1 xl:grid-cols-5 gap-4 mb-5 min-w-0">' +
      '      <div class="admin-card xl:col-span-3 overflow-hidden min-w-0">' +
      '        <div class="dash-panel-head">' +
      '          <div><div class="font-medium">下载量趋势</div><div class="text-xs mt-0.5" style="color: var(--admin-muted)">近 14 天 · 悬停查看详情</div></div>' +
      "        </div>" +
      '        <div class="px-2 pb-2 sm:px-3 sm:pb-3">' +
      '          <div ref="chartEl" class="dash-chart-el"></div>' +
      "        </div>" +
      "      </div>" +

      '      <div class="admin-card xl:col-span-2 overflow-hidden flex flex-col min-w-0">' +
      '        <div class="dash-panel-head">' +
      '          <div><div class="font-medium">下载量 Top 20</div><div class="text-xs mt-0.5" style="color: var(--admin-muted)">按累计下载次数</div></div>' +
      "        </div>" +
      '        <div class="flex-1 overflow-auto max-h-[340px] min-w-0">' +
      '          <table class="admin-table dash-top-table">' +
      "            <thead><tr><th style=\"width:2.5rem\">#</th><th>文件</th><th class=\"text-right whitespace-nowrap\">次数</th></tr></thead>" +
      "            <tbody>" +
      '              <tr v-for="(f, i) in data.top_files" :key="f.id || i">' +
      '                <td><span class="dash-rank" :class="{\'is-top\': i < 3}">{{ i + 1 }}</span></td>' +
      '                <td class="min-w-0 max-w-0">' +
      '                  <div class="font-medium truncate" :title="f.file_name">{{ f.file_name }}</div>' +
      '                  <div class="text-xs truncate" style="color: var(--admin-muted)" :title="f.path">{{ f.account_name }} · {{ f.path }}</div>' +
      "                </td>" +
      '                <td class="text-right font-medium whitespace-nowrap">{{ formatNumber(f.count) }}</td>' +
      "              </tr>" +
      '              <tr v-if="!(data.top_files && data.top_files.length)"><td colspan="3" class="text-center py-10" style="color: var(--admin-muted)">暂无下载记录</td></tr>' +
      "            </tbody>" +
      "          </table>" +
      "        </div>" +
      "      </div>" +
      "    </div>" +

      '    <div class="mb-3 flex items-center gap-2">' +
      '      <pa-icon name="activity" :size="16"></pa-icon>' +
      '      <span class="font-medium">系统信息</span>' +
      '      <span class="text-xs" style="color: var(--admin-muted)">运行 {{ formatUptime(ov.uptime_seconds) }}</span>' +
      "    </div>" +

      '    <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4 min-w-0">' +
      '      <div class="admin-card p-4 min-w-0 overflow-hidden">' +
      '        <div class="dash-sys-title"><pa-icon name="cpu" :size="16"></pa-icon> CPU</div>' +
      '        <div class="dash-sys-value">{{ (sys.cpu_percent || 0).toFixed(1) }}%</div>' +
      '        <div class="dash-meter"><div class="dash-meter-fill" :class="barClass(sys.cpu_percent)" :style="{width: Math.min(100, sys.cpu_percent||0) + \'%\'}"></div></div>' +
      '        <div class="dash-sys-meta">逻辑核心 {{ sys.cpu_cores || "-" }}</div>' +
      "      </div>" +

      '      <div class="admin-card p-4 min-w-0 overflow-hidden">' +
      '        <div class="dash-sys-title"><pa-icon name="memory-stick" :size="16"></pa-icon> 内存</div>' +
      '        <div class="dash-sys-value">{{ (sys.mem_percent || 0).toFixed(1) }}%</div>' +
      '        <div class="dash-meter"><div class="dash-meter-fill" :class="barClass(sys.mem_percent)" :style="{width: Math.min(100, sys.mem_percent||0) + \'%\'}"></div></div>' +
      '        <div class="dash-sys-meta">{{ formatBytes(sys.mem_used) }} / {{ formatBytes(sys.mem_total) }}</div>' +
      "      </div>" +

      '      <div class="admin-card p-4 min-w-0 overflow-hidden">' +
      '        <div class="dash-sys-title"><pa-icon name="hard-drive" :size="16"></pa-icon> 本地磁盘</div>' +
      '        <div class="dash-sys-value">{{ (sys.disk_percent || 0).toFixed(1) }}%</div>' +
      '        <div class="dash-meter"><div class="dash-meter-fill" :class="barClass(sys.disk_percent)" :style="{width: Math.min(100, sys.disk_percent||0) + \'%\'}"></div></div>' +
      '        <div class="dash-sys-meta truncate" :title="sys.disk_path">{{ formatBytes(sys.disk_used) }} / {{ formatBytes(sys.disk_total) }}</div>' +
      "      </div>" +

      '      <div class="admin-card p-4 min-w-0 overflow-hidden">' +
      '        <div class="dash-sys-title"><pa-icon name="layers" :size="16"></pa-icon> 运行时 / GC</div>' +
      '        <div class="grid grid-cols-2 gap-3 mt-3 text-sm">' +
      '          <div><div class="dash-kv-label">协程</div><div class="dash-kv-value">{{ formatNumber(sys.goroutines) }}</div></div>' +
      '          <div><div class="dash-kv-label">Heap</div><div class="dash-kv-value">{{ formatBytes(sys.heap_alloc) }}</div></div>' +
      '          <div><div class="dash-kv-label">GC 次数</div><div class="dash-kv-value">{{ formatNumber(sys.gc_num) }}</div></div>' +
      '          <div><div class="dash-kv-label">GC 暂停</div><div class="dash-kv-value">{{ (sys.gc_pause_total_ms || 0).toFixed(1) }} ms</div></div>' +
      "        </div>" +
      '        <div class="dash-sys-meta mt-3">最近 GC：{{ sys.last_gc || "-" }}</div>' +
      "      </div>" +

      '      <div class="admin-card p-4 min-w-0 overflow-hidden">' +
      '        <div class="dash-sys-title"><pa-icon name="list-ordered" :size="16"></pa-icon> 队列 / 任务</div>' +
      '        <div class="grid grid-cols-2 gap-3 mt-3 text-sm">' +
      '          <div><div class="dash-kv-label">缓存同步</div><div class="dash-kv-value"><span class="dash-dot" :class="ov.syncing ? \'is-busy\' : \'is-idle\'"></span>{{ ov.syncing ? "进行中" : "空闲" }}</div></div>' +
      '          <div><div class="dash-kv-label">Cron 任务</div><div class="dash-kv-value">{{ formatNumber(sys.cron_jobs) }}</div></div>' +
      '          <div><div class="dash-kv-label">目录缓存</div><div class="dash-kv-value">{{ formatNumber(ov.cache_files) }}</div></div>' +
      '          <div><div class="dash-kv-label">下载链缓存</div><div class="dash-kv-value">{{ formatNumber(ov.cache_url) }}</div></div>' +
      "        </div>" +
      "      </div>" +

      '      <div class="admin-card p-4 min-w-0 overflow-hidden">' +
      '        <div class="dash-sys-title"><pa-icon name="info" :size="16"></pa-icon> 版本信息</div>' +
      '        <div class="grid grid-cols-1 gap-2 mt-3 text-sm min-w-0">' +
      '          <div class="flex justify-between gap-3 min-w-0"><span class="dash-kv-label shrink-0">PanIndex</span><span class="dash-kv-value truncate">{{ ov.version || store.version || "-" }}</span></div>' +
      '          <div class="flex justify-between gap-3 min-w-0"><span class="dash-kv-label shrink-0">Go</span><span class="dash-kv-value truncate">{{ ov.go_version || "-" }}</span></div>' +
      '          <div class="flex justify-between gap-3 min-w-0"><span class="dash-kv-label shrink-0">构建时间</span><span class="dash-kv-value truncate">{{ ov.build_time || "-" }}</span></div>' +
      '          <div class="flex justify-between gap-3 min-w-0"><span class="dash-kv-label shrink-0">Commit</span><span class="dash-kv-value font-mono text-xs truncate">{{ ov.git_commit || "-" }}</span></div>' +
      "        </div>" +
      "      </div>" +
      "    </div>" +

      "  </template>" +
      "</div>",
  };
})(window);
