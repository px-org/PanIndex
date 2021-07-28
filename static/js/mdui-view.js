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
});
