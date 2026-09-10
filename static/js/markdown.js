(function (w) {
  "use strict";

  var instance;

  function isExternalLink(href) {
    if (!href) return false;
    var link = String(href).trim();
    if (!link) return false;
    // Same-page / relative / site-root paths stay in the current tab.
    if (
      link.charAt(0) === "#" ||
      link.charAt(0) === "?" ||
      link.charAt(0) === "/" ||
      link.charAt(0) === "."
    ) {
      return false;
    }
    try {
      var url = new URL(link, w.location.href);
      if (url.protocol !== "http:" && url.protocol !== "https:") return false;
      return url.origin !== w.location.origin;
    } catch (e) {
      return false;
    }
  }

  function engine() {
    if (instance) return instance;
    if (typeof w.markdownit !== "function") return null;
    instance = w.markdownit({
      html: true,
      linkify: true,
      typographer: true,
      breaks: false,
    });

    var defaultLinkOpen =
      instance.renderer.rules.link_open ||
      function (tokens, idx, options, env, self) {
        return self.renderToken(tokens, idx, options);
      };

    instance.renderer.rules.link_open = function (tokens, idx, options, env, self) {
      var token = tokens[idx];
      var href = token.attrGet("href");
      if (isExternalLink(href)) {
        token.attrSet("target", "_blank");
        token.attrSet("rel", "noopener noreferrer");
      }
      return defaultLinkOpen(tokens, idx, options, env, self);
    };

    return instance;
  }

  w.parseMarkdown = function (src) {
    var text = src == null ? "" : String(src);
    var md = engine();
    return md ? md.render(text) : text;
  };
})(window);
