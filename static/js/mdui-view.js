$(function () {
    $('#theme-toggle').on('click', function(){
        $('body').removeClass('mdui-theme-layout-auto');
        if($('body').hasClass('mdui-theme-layout-dark')){
            $('body').removeClass('mdui-theme-layout-dark');
            $('#theme-toggle i').text('brightness_4');
            $.cookie("Theme", "mdui-light", {expires : 3650, path:"/"});
            $(".aplayer-title").css("color", "");
            $(".aplayer-list-title").css("color", "");
        }else{
            $('body').addClass('mdui-theme-layout-dark');
            $('#theme-toggle i').text('brightness_5');
            $.cookie("Theme", "mdui-dark", {expires : 3650, path:"/"});
            $(".aplayer-title").css("color", "#666");
            $(".aplayer-list-title").css("color", "#666");
        }
    });
    var path = $("#file_link").attr("data-path");
    var mode = $("#file_link").attr("data-mode");
    var fullUrl = encodeURI(window.location.protocol + "//"+window.location.host + path);
    $("#file_link").attr("href", fullUrl);
    $("#file_link").text(fullUrl);
    if(mode == "native"){
        $("#view_down_link").attr("href", fullUrl);
    }
    var clipboard = new ClipboardJS('.copyBtn', {
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
    var inst = new mdui.Collapse('#info_panel');
    var si = $.cookie("Show-Info")
    if(si == "1"){
        inst.open('#item-1');
    }else{
        inst.close('#item-1');
    }
    document.getElementById('info-toggle').addEventListener('click', function () {
        inst.toggle('#item-1');
    });
    document.getElementById('item-1').addEventListener('open.mdui.collapse', function () {
        $.cookie("Show-Info", "1", {expires : 3650, path:"/"});
    });
    document.getElementById('item-1').addEventListener('close.mdui.collapse', function () {
        $.cookie("Show-Info", "0", {expires : 3650, path:"/"});
    });
    document.getElementById('share-menu').addEventListener('open.mdui.menu', function () {
        var formData = new FormData();
        var prefix = window.location.protocol + "//"+window.location.host + "/s/";
        formData.append("accountId", $(this).attr("data-aid"));
        formData.append("prefix", prefix);
        formData.append("path", $(this).attr("data-fp"));
        $.ajax({
            type: 'POST',
            url: '/api/public/shortInfo',
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
});
