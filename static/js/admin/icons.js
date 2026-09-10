(function (global) {
  var PanAdmin = global.PanAdmin || (global.PanAdmin = {});

  function toPascal(name) {
    return String(name || "")
      .split("-")
      .map(function (p) {
        return p.charAt(0).toUpperCase() + p.slice(1);
      })
      .join("");
  }

  PanAdmin.Icon = {
    name: "PaIcon",
    props: {
      name: { type: String, required: true },
      size: { type: [Number, String], default: 18 },
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
        var root = this.$refs.root || this.$el;
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
          lucide.createIcons({
            nodes: [root],
            attrs: {
              width: this.size,
              height: this.size,
              "stroke-width": this.stroke,
            },
          });
        }
      },
    },
    template: '<span ref="root" class="inline-flex items-center justify-center"></span>',
  };
})(window);
