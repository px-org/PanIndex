$(document).ready(function() {
    var idClipboard = new ClipboardJS('.copyIDBtn', {
        text: function(trigger) {
            var filePath = $(trigger).data("content");
            var fullUrl = encodeURI(window.location.protocol + "//"+window.location.host + filePath);
            return fullUrl;
        }
    });
    var clipboard = new ClipboardJS('.copyBtn', {
        text: function(trigger) {
            var path = $(trigger).data("path");
            var fullUrl = window.location.protocol + "//"+window.location.host + path;
            return encodeURI(fullUrl);
        }
    });
    var copyAllLinksBoard = new ClipboardJS('#copyAllLinks', {
        text: function(trigger) {
            var urls = [];
            $(".icon-file-mdui").each(function (i, item) {
                var folder = $(this).attr("data-folder");
                var path = $(this).attr("data-url");
                var fullUrl = window.location.protocol + "//"+window.location.host + path;
                if(folder == "false"){
                    urls.push(fullUrl);
                }
            });
            return urls.join("\n");
        }
    });
    copyAllLinksBoard.on('success', function(e) {
        mdui.snackbar({
            message: "链接已复制到剪切板"
        });
        e.clearSelection();
    });
    idClipboard.on('success', function(e) {
        mdui.snackbar({
            message: "链接已复制到剪切板"
        });
        e.clearSelection();
    });
    clipboard.on('success', function(e) {
        if(typeof(mdui) != "undefined"){
            mdui.snackbar({
                message: "链接已复制到剪切板"
            });
        }else if(typeof(M) != "undefined"){
            M.toast({html: '链接已复制到剪切板'});
        }else{
            if ($("#liveToast").length > 0) {
                $("#liveToast").toast("show");
            }
        }
        e.clearSelection();
    });
    $('#theme-toggle').on('click', function(){
        $('body').removeClass('mdui-theme-layout-auto');
        if($('body').hasClass('mdui-theme-layout-dark')){
            $('body').removeClass('mdui-theme-layout-dark');
            $('#theme-toggle i').text('brightness_4');
            Cookies.set("theme", "mdui-light", {expires : 3650, path:"/"});
            $(".aplayer-title").css("color", "");
            $(".aplayer-list-title").css("color", "");
            if(typeof art != 'undefined' && art){
                videoInit();
            }
        }else{
            $('body').addClass('mdui-theme-layout-dark');
            $('#theme-toggle i').text('brightness_5');
            Cookies.set("theme", "mdui-dark", {expires : 3650, path:"/"});
            $(".aplayer-title").css("color", "#666");
            $(".aplayer-list-title").css("color", "#666");
            if(typeof art != 'undefined' && art){
                videoInit();
            }
        }
    });
    $('.icon-file-mdui').on('click', function(ev) {
        if(ev.target.tagName == "A" && (ev.target.text == "file_download" ||
            ev.target.text == "content_copy") || ev.target.title == "复制链接") return;
        var isFolder = $(this).attr("data-folder");
        var dURL = $(this).attr("data-url");
        var preview = $(this).attr("data-preview");
        if(isFolder == "true" || preview == 0){
            window.location.href = dURL;
        }else{
            window.location.href = dURL+"?v";
        }
    });
    if(document.getElementById('share-menu')){
        document.getElementById('share-menu').addEventListener('open.mdui.menu', function () {
            $(".mdui-card").attr("style","min-height:403px");
            var formData = new FormData();
            var prefix = window.location.protocol + "//"+window.location.host + $config.path_prefix+"/s/";
            formData.append("prefix", prefix);
            formData.append("path", $(this).attr("data-fp"));
            formData.append("isFile", $(this).attr("data-file-type"));
            $.ajax({
                type: 'POST',
                url: $config.path_prefix+'/api/v3/public/shortInfo',
                data: formData,
                cache: false,
                contentType: false,
                processData: false,
                success: function(d){
                    $("#qrcode").attr("src", d.qr_code);
                    $("#copyShortUrl").attr("data-content", d.short_url);
                    $("#copyShortUrl").attr("data-clipboard-action", "copy");
                    var clipboard = new ClipboardJS('#copyShortUrl', {
                        text: function(trigger) {
                            var content = $(trigger).data("content");
                            return content;
                        }
                    });
                    clipboard.on('success', function(e) {
                        mdui.snackbar({
                            message: "已复制到剪切板"
                        });
                        e.clearSelection();
                    });
                }
            });
        });
        document.getElementById('share-menu').addEventListener('closed.mdui.menu', function () {
            $(".mdui-card").removeAttr('style');
        });
    }
    if(document.getElementById('sort-menu')){
        document.getElementById('sort-menu').addEventListener('open.mdui.menu', function () {
            $(".mdui-card").attr("style","min-height:322px");
        });
        document.getElementById('sort-menu').addEventListener('closed.mdui.menu', function () {
            $(".mdui-card").removeAttr('style');
        });
    }
    $("#image-preview-list").empty();
    $(".icon-file-mdui").each(function (i, item) {
        var vt = $(this).attr("data-view-type");
        var du = $(this).attr("data-url");
        var t = $(this).attr("data-title");
        if(vt == "img"){
            $("#image-preview-list").append("<img data-original=\""+du+"\" alt=\""+t+"\"></img>");
        }
    });
    $('#go-to-top').on('click',function () {
        $("html, body").animate({ scrollTop: 0 }, "slow");
        return false;
    });
    $('.sort-order-check').on('click', function () {
        var dOrder =  $(this).attr("data-order");
        var dColumn =  $(this).attr("data-column");
        Cookies.set("sort_column", dColumn, {expires : 3650, path:"/"});
        Cookies.set("sort_order", dOrder, {expires : 3650, path:"/"});
        location.reload();
    });
    $('.default-check').on('click', function () {
        Cookies.set('sort_column', "default", {expires : 3650, path: '/'});
        Cookies.set('sort_order', "null", {expires : 3650, path: '/'});
        location.reload();
    });
    initSort();
    if(document.getElementById('previewImages')){
        document.getElementById('previewImages').addEventListener('click', function () {
            var ipl = $('#image-preview-list').html();
            if(ipl.length != 0){
                var viewer = new Viewer(document.getElementById('image-preview-list'), {
                    url: 'data-original',
                    hidden: function () {
                        viewer.destroy();
                    },
                    title: function (image) {
                        return image.alt + ' (' + (this.index + 1) + '/' + this.length + ')';
                    }
                });
                viewer.show();
            }
        });
    }
    $('#layout-toggle').on('click', function () {
        var layout = $(this).find("i").text();
        if(layout == "view_comfy"){
            Cookies.set("layout", "view_list", {expires : 3650, path:"/"});
            $(this).find("i").text("view_list")
        }else{
            Cookies.set("layout", "view_comfy", {expires : 3650, path:"/"});
            $(this).find("i").text("view_comfy")
        }
        location.reload();
    });
    if(document.getElementById('info_panel')){
        var inst = new mdui.Collapse('#info_panel');
        var si = Cookies.get("show_info")
        if(si == "1"){
            inst.open('#item-1');
        }else{
            inst.close('#item-1');
        }
        document.getElementById('info-toggle').addEventListener('click', function () {
            inst.toggle('#item-1');
        });
        document.getElementById('item-1').addEventListener('open.mdui.collapse', function () {
            Cookies.set("show_info", "1", {expires : 3650, path:"/"});
        });
        document.getElementById('item-1').addEventListener('close.mdui.collapse', function () {
            Cookies.set("show_info", "0", {expires : 3650, path:"/"});
        });
    }
    var filePath = $("#file_link").attr("data-path");
    var fullUrl = encodeURI(window.location.protocol + "//"+window.location.host + filePath);
    $("#file_link").text(fullUrl);
    $('#file_link').on('click', function () {
        window.location.href = fullUrl;
    });
    $('#view_down_link').on('click', function () {
        window.location.href = fullUrl;
    });
    $(".search").bind('keydown', function(event) {
        var key = $(this).val();
        key = key.replace(/(^\s*)|(\s*$)/g,"")
        if(key.length < 30){
            if (event.key === "Enter") {
                if( $(this).val() != ""){
                    window.location.href = "/?search=" + key;
                }
            }
        }
    });
    document.getElementById('history_play_list_menu').addEventListener('open.mdui.menu', function () {
        $("#history_play_list_menu").attr("style", "width:22%;max-height: 200px;overflow:scroll");
    });
    initPlayHistoryList();
    $("#input-password").on('keydown', function(event) {
        if (event.key === "Enter") {
            var dfp = $(this).attr("data-file-path");
            promptPwd(dfp);
        }
    });
    if(document.getElementById('account-menu')){
        document.getElementById('account-menu').addEventListener('open.mdui.menu', function () {
            $("#account-menu").attr("style", "max-height: "+$(".mdui-card-content").height()+"px; transform-origin: 0px 50%; position: absolute; top: -4px; left: 21px;");
        });
    }
});
function promptPwd(){
    var pwd = $("#input-password").val();
    var fullPath = $("#input-password").attr("data-file-path");
    if(pwd && pwd != "" && pwd != null && pwd != "null" && pwd.length < 30){
        var result = Cookies.get("file_pwd");
        var ppwd = md5(fullPath) + ":" + pwd;
        if (result && result != null && result != "null" && result != ""){
            result = result + ","+ ppwd;
        }else{
            result = ppwd;
        }
        Cookies.set("file_pwd", result, { expires: 3650, path: '/' });
        location.reload();
    }
}
function removePwd(){
    var fullPath = $("#input-password").attr("data-file-path");
    var pathMd5 = md5(fullPath)
    var newFilePwd = [];
    var cookiePwd = Cookies.get("file_pwd");
    if(cookiePwd && cookiePwd != "" && cookiePwd != null && cookiePwd != "null"){
        $.each(cookiePwd.split(","), function (i, item){
            var pMd5 = item.split(":")[0]
            if(pMd5 != pathMd5){
                newFilePwd.push(item);
            }
        });
        if (newFilePwd.length > 0){
            Cookies.set("file_pwd", newFilePwd.join(","), {expires : 3650, path:"/"});
        }else{
            Cookies.remove('file_pwd', { path: '/' });
        }
    }
}

function initSort(){
    var sColumn = Cookies.get("sort_column");
    var sOrder = Cookies.get("sort_order");
    if (sColumn == "null" || sColumn == null || sColumn == "" || sColumn == "default"){
        $('.default-check').prepend('<i class="check mdui-menu-item-icon mdui-icon material-icons">check</i>');
    }else{
        $('a[data-column='+sColumn+']:not(.sort-order-check)').prepend('<i class="check mdui-menu-item-icon mdui-icon material-icons">check</i>');
        $('a[data-column='+sColumn+'][data-order='+sOrder+']').prepend('<i class="check mdui-menu-item-icon mdui-icon material-icons">check</i>');
    }
}

function initPlayHistoryList() {
    var play_history_list_arr_str = localStorage.getItem('play_history_list');
    if(play_history_list_arr_str && play_history_list_arr_str != null && play_history_list_arr_str != ""){
        var play_history_list_arr = JSON.parse(play_history_list_arr_str);
        if (play_history_list_arr.length > 0){
            $("#history_play_list_menu").empty();
            $.each(play_history_list_arr, function (i, item) {
                var name = item.title.substring(0, item.title.lastIndexOf("."));
                name = name + " 续播 " + formatSeconds(item.currentTime);
                var menu = '<li class="mdui-menu-item">'+
                '  <a href="'+item.url+'?v" class="mdui-ripple" mdui-tooltip="{content: \''+name+'\', position: \'right\'}">'+
                '    <i class="mdui-menu-item-icon mdui-icon material-icons">play_circle_outline</i>'+name+
                '  </a>'+
                '</li>';
                $("#history_play_list_menu").append(menu);
            });
        }
    }
}

function formatSeconds(value) {
    var theTime = parseInt(value);// 秒
    var theTime1 = 0;// 分
    var theTime2 = 0;// 小时
    if (theTime > 60) {
        theTime1 = parseInt(theTime / 60);
        theTime = parseInt(theTime % 60);
        if (theTime1 > 60) {
            theTime2 = parseInt(theTime1 / 60);
            theTime1 = parseInt(theTime1 % 60);
        }
    }
    var result = "" + parseInt(theTime);
    if(result < 10){
        result = '0' + result;
    }
    if (theTime1 > 0) {
        result = "" + parseInt(theTime1) + ":" + result;
        if(theTime1 < 10){
            result = '0' + result;
        }
    }else{
        result = '00:' + result;
    }
    if (theTime2 > 0) {
        result = "" + parseInt(theTime2) + ":" + result;
        if(theTime2 < 10){
            result = '0' + result;
        }
    }else{
        result = '00:' + result;
    }
    return result;
}
function mdContent(fullUrl, key, doc, isMark) {
    $.ajax({
        method: 'GET',
        url: fullUrl,
        success: function (data) {
            if(data && !data.status){
                localStorage.setItem(key, data);
                if(isMark){
                    $("#"+doc).append(marked.parse(data));
                    $("table").addClass("mdui-table");
                    $("#"+doc).toggle();
                }
            }else{
                localStorage.removeItem(key);
                $("#emptyList").attr("style", "height: 500px;");
            }
        }
    });
}