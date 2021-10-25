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
$('.icon-file-mdui').on('click', function(ev) {
    if(ev.target.tagName == "A" && (ev.target.text == "file_download" ||
        ev.target.text == "content_copy") || ev.target.title == "复制链接") return;
    var isFolder = $(this).attr("data-folder");
    var dURL = $(this).attr("data-url");
    if(isFolder == "true" ){
        window.location.href = dURL;
    }else{
        window.location.href = dURL+"?v";
    }


});
$(document).ready(function() {
    $('#theme-toggle').on('click', function(){
        $('body').removeClass('mdui-theme-layout-auto');
        if($('body').hasClass('mdui-theme-layout-dark')){
            $('body').removeClass('mdui-theme-layout-dark');
            $('#theme-toggle i').text('brightness_4');
            $.cookie("Theme", "mdui-light", {expires : 3650, path:"/"});
        }else{
            $('body').addClass('mdui-theme-layout-dark');
            $('#theme-toggle i').text('brightness_5');
            $.cookie("Theme", "mdui-dark", {expires : 3650, path:"/"});
        }
    });
    $("#image-preview-list").empty();
    $(".icon-file-mdui").each(function (i, item) {
        var mt = $(this).attr("data-media-type");
        var du = $(this).attr("data-url");
        var t = $(this).attr("data-title");
        if(mt == 1){
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
        $.cookie("SColumn", dColumn, {expires : 3650, path:"/"});
        $.cookie("SOrder", dOrder, {expires : 3650, path:"/"});
        location.reload();
    });
    $('.default-check').on('click', function () {
        $.cookie('SColumn', "default", {expires : 3650, path: '/'});
        $.cookie('SOrder', null, {expires : 3650, path: '/'});
        location.reload();
    });
    //初始化排序设置
    initSort();
    $('.icon-file').on('click', function(ev) {
        if(ev.target.tagName == "A" && (ev.target.text == "file_download" ||
            ev.target.text == "content_copy") || ev.target.title == "复制链接") return;
        var dURL = $(this).attr("data-url");
        var title = $(this).attr("data-title");
        var dmt = $(this).attr("data-media-type");
        var fileType = $(this).attr("data-file-type");
        if(dmt == 1){
            $(this).lightGallery({
                fullScreen: true,
                dynamic: true,
                thumbnail:false,
                animateThumb: false,
                showThumbByDefault: false,
                dynamicEl: findDynamicEl(this),
                share: false,
                actualSize: false,
                closable: true
            });
            return;
        }else if(dmt == 2){
            const ap = new APlayer({
                container: document.getElementById('aplayer'),
                fixed: true,
                lrcType: 3,
                autoplay: true,
                audio: [{
                    name: title,
                    artist: 'artist',
                    url: dURL,
                    cover: dURL.split('.')[0] + '.jpg',
                    lrc: dURL.split('.')[0] + '.lrc'
                }]
            });
            return;
        }else if(dmt == 3){
            $(this).lightGallery({
                dynamic: true,
                thumbnail:false,
                fullScreen: true,
                dynamicEl: findDynamicEl(this),
                share: false,
                actualSize: false,
                closable: false
            });
            return;
        }
        var fullUrl = window.location.protocol+"//"+window.location.host + dURL;
        if(fileType == "doc" || fileType == "docx" || fileType == "dotx"
            || fileType == "ppt" || fileType == "pptx" || fileType == "xls" || fileType == "xlsx"){
            window.open("https://view.officeapps.live.com/op/view.aspx?src="+fullUrl);
        }else{
            window.location.href = dURL;
        }
    });
    $('.folderDown').on('click', function() {
        var fileId = $(this).attr("data-file-id");
        var accountId = $(this).attr("data-account");
        var url =  "/api/public/downloadMultiFiles?fileId="+fileId+"&accountId="+accountId;
        if (fileId.startsWith("/")){
            window.location.href = url;
        }else{
            $.ajax({
                type: 'POST',
                url: url,
                async:false,
                success: function(data){
                    window.location.href = data.redirect_url;
                }
            });
        }
    });
    $('.table-head').on('click', function() {
        var orderColumn = $(this).text();
        var orderSeq = $(this).attr("data-order-seq");
        var orderType = $(this).attr("data-order-type");
        $('.table-head').each(function(){
            $(this).text($(this).text());
        });
        if(orderSeq == "" || orderSeq == "down"){
            //当前是升序排列，按照orderColumn降序
            sortTable("up", orderType);
            $(this).attr("data-order-seq", "up");
            $(this).html(orderColumn+" <i class=\"fa fa-angle-double-up\" aria-hidden=\"true\"></i>");
        }else if(orderSeq == "up"){
            sortTable("down", orderType);
            $(this).attr("data-order-seq", "down");
            $(this).html(orderColumn+" <i class=\"fa fa-angle-double-down\" aria-hidden=\"true\"></i>");
        }
    });
    if(document.getElementById('share-menu')){
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
    }
});
function sortTable(sort_order, data_type){
    $('table tbody > tr').not('.parent').sortElements(function (a, b) {
        let data_a = $(a).find("td[class='"+data_type+"']").text(), data_b = $(b).find("td[class='"+data_type+"']").text();
        let rt = data_a.localeCompare(data_b);
        return (sort_order === "down") ? 0-rt : rt;
    });
}
function findDynamicEl(obj) {
    var dynamicEls = [];
    var dataMediaType = $(obj).attr("data-media-type");
    var oDURL = $(obj).attr("data-url");
    var oTitle = $(obj).attr("data-title");
    if(dataMediaType == 1){
        dynamicEls.push({
            src: oDURL,
            thumb: oDURL,
            subHtml: '<h4>'+oTitle+'</h4>',
            downloadUrl:  oDURL
        });
    }else if(dataMediaType == 3){
        dynamicEls.push({
            html: '<video class="lg-video-object lg-html5" controls preload="none"><source src="'+oDURL+'">Your browser does not support HTML5 video</video>',
            subHtml: '<h4>'+oTitle+'</h4>',
            downloadUrl:  oDURL
        });
    }
   var ofs = $(obj).parent().parent().find(".icon-file");
    if($(obj).parent().get(0).tagName == "TD"){
        ofs = $(obj).parent().parent().parent().find(".icon-file");
    }
    ofs.each(function(i, d){
        var dURL = $(d).attr("data-url");
        var title = $(d).attr("data-title");
        var dmt = $(d).attr("data-media-type");
        if(dmt == dataMediaType && oTitle != title){
            if(dataMediaType == 1){
               dynamicEls.push({
                   src: dURL,
                   thumb: dURL,
                   subHtml: '<h4>'+title+'</h4>',
                   downloadUrl:  dURL
               });
            }/*else if(dataMediaType == 3){
                dynamicEls.push({
                    html: '<video class="lg-video-object lg-html5" controls preload="none"><source src="'+dURL+'">Your browser does not support HTML5 video</video>',
                    subHtml: '<h4>'+title+'</h4>',
                    downloadUrl:  dURL
                });
            }*/
        }
    });
   /* $(obj).parent().parent().parent().find(".icon-file").each(function(i, d){
        var dURL = $(d).attr("data-url");
        var title = $(d).attr("data-title");
        var dmt = $(d).attr("data-media-type");
        if(dmt == dataMediaType && oTitle != title){
            if(dataMediaType == 1){
                dynamicEls.push({
                    src: dURL,
                    thumb: dURL,
                    subHtml: '<h4>'+title+'</h4>',
                    downloadUrl:  dURL
                });
            }/!*else if(dataMediaType == 3){
                dynamicEls.push({
                    html: '<video class="lg-video-object lg-html5" controls preload="none"><source src="'+dURL+'">Your browser does not support HTML5 video</video>',
                    subHtml: '<h4>'+title+'</h4>',
                    downloadUrl:  dURL
                });
            }*!/
        }
    });*/
    return dynamicEls;
}
$.fn.extend({
    sortElements: function (comparator, getSortable) {
        getSortable = getSortable || function () { return this; };

        var placements = this.map(function () {
            var sortElement = getSortable.call(this),
                parentNode = sortElement.parentNode,
                nextSibling = parentNode.insertBefore(
                    document.createTextNode(''),
                    sortElement.nextSibling
                );

            return function () {
                parentNode.insertBefore(this, nextSibling);
                parentNode.removeChild(nextSibling);
            };
        });

        return [].sort.call(this, comparator).each(function (i) {
            placements[i].call(getSortable.call(this));
        });
    }
});
/*$(".mdui-textfield-input").keyup(function () {
    var keyword = $(this).val();
    var reg =  new RegExp(keyword);
    $(".mdui-list").find("li").each(function (i, item) {
        var title = $(this).find("div").attr("data-title");
        if("undefined" == typeof title || reg.test(title)){
            $(this).show();
        }else{
            $(this).hide();
        }
    });
});*/
/*$('.mdui-textfield-input').bind('keydown', function(event) {
    var dIndex = $(this).attr("data-index");
    var key = $(this).val();
    key = key.replace(/(^\s*)|(\s*$)/g,"")
    if (event.key === "Enter") {
        if( $(this).val() != ""){
            window.location.href = dIndex + "?search=" + key;
        }else{
            window.location.href = dIndex;
        }
    }
});*/
$(".search").bind('keydown', function(event) {
    var dIndex = $(this).attr("data-index");
    var key = $(this).val();
    key = key.replace(/(^\s*)|(\s*$)/g,"")
    if (event.key === "Enter") {
        if( $(this).val() != ""){
            window.location.href = "/?search=" + key;
        }else{
            window.location.href = dIndex;
        }
    }
});
/*$(".search").keyup(function () {
    var keyword = $(this).val();
    var reg =  new RegExp(keyword);
    $("tbody").find("tr").each(function (i, item) {
        var title = $(this).find(".file-name").text();
        if("undefined" == typeof title || reg.test(title)){
            $(this).show();
        }else{
            $(this).hide();
        }
    });
});*/
function initSort(){
    var sColumn = $.cookie("SColumn");
    var sOrder = $.cookie("SOrder");
    if (sColumn == "null" || sColumn == null || sColumn == "" || sColumn == "default"){
        $('.default-check').prepend('<i class="check mdui-menu-item-icon mdui-icon material-icons">check</i>');
    }else{
        $('a[data-column='+sColumn+']:not(.sort-order-check)').prepend('<i class="check mdui-menu-item-icon mdui-icon material-icons">check</i>');
        $('a[data-column='+sColumn+'][data-order='+sOrder+']').prepend('<i class="check mdui-menu-item-icon mdui-icon material-icons">check</i>');
    }
}

window.addEventListener('DOMContentLoaded', function () {
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
});
function promptPwd(dfi){
    if(dfi == ""){
        dfi = "all";
    }
    var result = $.cookie("dir_pwd");
    var ppwd = dfi + ":" + $("#input-password").val();
    if (result != null && result != "null" && result != ""){
        result = result + ","+ ppwd;
    }else{
        result = ppwd;
    }
    $.cookie("dir_pwd", result, {expires : 3650, path:"/"});
    location.reload();
}
$("#input-password").bind('keydown', function(event) {
    if (event.key === "Enter") {
        var dfi = $(this).attr("data-file-id");
        promptPwd(dfi);
    }
});
