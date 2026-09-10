(function (global) {
  var PanAdmin = global.PanAdmin || (global.PanAdmin = {});

  function getBase() {
    return (global.__ADMIN__ && global.__ADMIN__.apiUrl) || "";
  }

  async function request(path, options) {
    options = options || {};
    var method = (options.method || "GET").toUpperCase();
    var headers = Object.assign({}, options.headers || {});
    var body = options.body;
    var isForm = options.formData === true;

    if (body != null && !isForm && typeof body !== "string") {
      headers["Content-Type"] = "application/json";
      body = JSON.stringify(body);
    }

    var res = await fetch(getBase() + path, {
      method: method,
      headers: headers,
      body: body,
      credentials: "same-origin",
    });

    var text = await res.text();
    var data = null;
    try {
      data = text ? JSON.parse(text) : null;
    } catch (e) {
      data = text;
    }

    if (!res.ok) {
      var msg = (data && data.msg) || res.statusText || "请求失败";
      throw new Error(msg);
    }
    return data;
  }

  PanAdmin.api = {
    get: function (path) {
      return request(path, { method: "GET" });
    },
    post: function (path, body) {
      return request(path, { method: "POST", body: body });
    },
    del: function (path, body) {
      return request(path, { method: "DELETE", body: body });
    },
    postForm: function (path, formData) {
      return request(path, { method: "POST", body: formData, formData: true });
    },
    saveConfig: function (config) {
      return request("/config", { method: "POST", body: config });
    },
    getConfig: function () {
      return request("/config", { method: "GET" });
    },
    uploadConfig: function (configJson) {
      // server expects raw Config JSON body
      var body =
        typeof configJson === "string" ? configJson : JSON.stringify(configJson);
      return request("/config/upload", {
        method: "POST",
        body: body,
        headers: { "Content-Type": "application/json" },
      });
    },
  };
})(window);
