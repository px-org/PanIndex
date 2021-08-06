$(function () {
    $('#theme-toggle').on('click', function(){
        $('body').removeClass('mdui-theme-layout-auto');
        if($('body').hasClass('mdui-theme-layout-dark')){
            $('body').removeClass('mdui-theme-layout-dark');
            $('#theme-toggle i').text('brightness_4');
            $.cookie("Theme", "mdui-light");
        }else{
            $('body').addClass('mdui-theme-layout-dark');
            $('#theme-toggle i').text('brightness_5');
            $.cookie("Theme", "mdui-dark");
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
    var inst = new mdui.Collapse('#collapse');
    document.getElementById('toggle-1').addEventListener('click', function () {
        inst.toggle('#item-1');
    });
});
