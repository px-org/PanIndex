var $ = mdui.$;
var pwdd = new mdui.Dialog("#pwd_dialog");
var hided = new mdui.Dialog("#hide_dialog");
var diskd = new mdui.Dialog("#disk_dialog", {modal:true});
var cached = new mdui.Dialog("#refresh_cache_dialog", {modal:true});
var ud = new mdui.Dialog("#upload_dialog", {modal:true});
var bypassd = new mdui.Dialog("#bypass_dialog");
var clearCached = new mdui.Dialog("#cache_dialog");
var cacheConfigd = new mdui.Dialog("#cache_config_dialog", {modal:true});
var modeSelect = new mdui.Select('#mode');
var cachePolicySelect = new mdui.Select('#cachePolicy');
function CommonRequest(urlPath, method, d) {
    $.ajax({
        method: method,
        url: AdminApiUrl + urlPath,
        data: JSON.stringify(d),
        contentType: 'application/json',
        success: function (data) {
            var joData = JSON.parse(data);
            mdui.snackbar({
                message: joData.msg,
                timeout: 1000,
                onClose: function(){
                    location.reload();
                }
            });
        }
    });
}
$('#theme-toggle').on('click', function(){
    $('body').removeClass('mdui-theme-layout-auto');
    if($('body').hasClass('mdui-theme-layout-dark')){
        $('body').removeClass('mdui-theme-layout-dark');
        $('#theme-toggle i').text('brightness_4');
        Cookies.set("theme", "mdui-light", {expires : 3650, path:"/"});
    }else{
        $('body').addClass('mdui-theme-layout-dark');
        $('#theme-toggle i').text('brightness_5');
        Cookies.set("theme", "mdui-dark", {expires : 3650, path:"/"});
    }
});
$(".saveConfigBtn").on("click", function () {
    var config = $("#configForm").serializeObject();
    if(config.port){
        config.port = Number(config.port);
    }
    var enable_preview = $("#configForm").find("input[name=enable_preview]");
    if(enable_preview && enable_preview.length > 0){
        if(enable_preview.prop('checked')){
            config["enable_preview"] = "1";
        }else{
            config["enable_preview"] = "0";
        }
    }
    var enable_lrc = $("#configForm").find("input[name=enable_lrc]");
    if(enable_lrc && enable_lrc.length > 0){
        if(enable_lrc.prop('checked')){
            config["enable_lrc"] = "1";
        }else{
            config["enable_lrc"] = "0";
        }
    }
    CommonRequest("/config", "POST", config);
});
//网盘挂载-start
//添加网盘
var accountStatus = 0;
$("#saveAccountBtn").on("click", function () {
    var account = $("#accountForm").serializeObject();
    if(!account.root_id){
        mdui.snackbar({
            message: "请输入根目录ID",
            timeout: 2000
        });
        return false;
    }
    var df = $("#accountForm").find("input[name=down_transfer]").prop('checked');
    if(accountStatus == 1){
        return false;
    }
    if(df){
        account.down_transfer = 1
    }else{
        account.down_transfer = 0
    }
    accountStatus = 1;
    $.ajax({
        method: 'POST',
        url: AdminApiUrl + '/account',
        data: JSON.stringify(account),
        contentType: 'application/json',
        success: function (data) {
            var d = JSON.parse(data);
            mdui.snackbar({
                message: d.msg,
                timeout: 2000,
                onClose: function(){
                    accountStatus = 0;
                    location.reload();
                }
            });
        }
    });
});
$("#accountForm").find("select[name=mode]").on('change', function () {
    var mode = $(this).val();
    dynamicChgMode(mode);
    diskd.handleUpdate();
});
function dynamicChgMode(mode){
    if(mode == "native"){
        $("#accountForm").find("input[name=password]").attr("type", "password");
        $("#RedirectUriDiv").hide();
        $("#RefreshTokenDiv").hide();
        $("#UserDiv").hide();
        $("#PasswordDiv").hide();
        $(".sync-div").hide();
        $("#ApiUrlDiv").hide();
        $("#SiteIdDiv").hide();
        $("#aliQrCodeBtn").hide();
        $("#accountForm").find("input[name=root_id]").val("/");
    }else if (mode == "cloud189"){
        $("#RedirectUriDiv").hide();
        $("#RefreshTokenDiv").hide();
        $("#ApiUrlDiv").hide();
        $("#UserDiv").show();
        $("#PasswordDiv").show();
        $(".sync-div").show();
        $("#SiteIdDiv").hide();
        $("#site_label").text("家庭ID（Family ID）");
        $("#user_label").text("用户名");
        $("#accountForm").find("input[name=password]").attr("type", "password");
        $("#password_label").text("密码");
        $("#aliQrCodeBtn").hide();
        $("#accountForm").find("input[name=root_id]").val("-11");
    }else if (mode == "teambition"){
        $("#RedirectUriDiv").hide();
        $("#RefreshTokenDiv").hide();
        $("#ApiUrlDiv").hide();
        $("#UserDiv").show();
        $("#PasswordDiv").show();
        $(".sync-div").show();
        $("#SiteIdDiv").show();
        $("#site_label").text("项目ID（Project ID）");
        $("#user_label").text("用户名");
        $("#accountForm").find("input[name=password]").attr("type", "password");
        $("#password_label").text("密码");
        $("#accountForm").find("input[name=root_id]").val("");
        $("#aliQrCodeBtn").hide();
    }else if (mode == "teambition-us"){
        $("#RedirectUriDiv").hide();
        $("#RefreshTokenDiv").hide();
        $("#ApiUrlDiv").hide();
        $("#UserDiv").show();
        $("#PasswordDiv").show();
        $(".sync-div").show();
        $("#SiteIdDiv").show();
        $("#site_label").text("项目ID（Project ID）");
        $("#user_label").text("用户名");
        $("#accountForm").find("input[name=password]").attr("type", "password");
        $("#password_label").text("密码");
        $("#accountForm").find("input[name=root_id]").val("");
        $("#aliQrCodeBtn").hide();
    }else if (mode == "aliyundrive"){
        $("#RedirectUriDiv").hide();
        $("#ApiUrlDiv").hide();
        $("#RefreshTokenDiv").show();
        $("#UserDiv").hide();
        $("#PasswordDiv").hide();
        $(".sync-div").show();
        $("#SiteIdDiv").hide();
        $("#aliQrCodeBtn").show();
        $("#accountForm").find("input[name=password]").attr("type", "password");
        $("#accountForm").find("input[name=root_id]").val("root");
    }else if (mode == "onedrive"){
        $("#RedirectUriDiv").show();
        $("#RefreshTokenDiv").show();
        $("#UserDiv").show();
        $("#PasswordDiv").show();
        $(".sync-div").show();
        $("#user_label").text("客户端ID（Client ID）");
        $("#password_label").text("客户端密码（Client Secret）");
        $("#accountForm").find("input[name=password]").attr("type", "password");
        $("#ApiUrlDiv").hide();
        $("#SiteIdDiv").hide();
        $("#aliQrCodeBtn").hide();
        $("#accountForm").find("input[name=root_id]").val("/");
    }else if (mode == "onedrive-cn"){
        $("#RedirectUriDiv").show();
        $("#RefreshTokenDiv").show();
        $("#UserDiv").show();
        $("#PasswordDiv").show();
        $(".sync-div").show();
        $("#user_label").text("客户端ID（Client ID）");
        $("#password_label").text("客户端密码（Client Secret）");
        $("#accountForm").find("input[name=password]").attr("type", "password");
        $("#ApiUrlDiv").hide();
        $("#SiteIdDiv").hide();
        $("#aliQrCodeBtn").hide();
        $("#accountForm").find("input[name=root_id]").val("/");
    }else if (mode == "ftp"){
        $("#RedirectUriDiv").hide();
        $("#RefreshTokenDiv").hide();
        $("#UserDiv").show();
        $("#PasswordDiv").show();
        $(".sync-div").show();
        $("#ApiUrlDiv").show();
        $("#user_label").text("用户名");
        $("#accountForm").find("input[name=password]").attr("type", "password");
        $("#password_label").text("密码");
        $("#api_url_label").text("FTP地址（FTP Addr）");
        $("#api_url").attr("placeholder", "192.168.1.1:21");
        $("#SiteIdDiv").hide();
        $("#accountForm").find("input[name=root_id]").val("/");
        $("#aliQrCodeBtn").hide();
    }else if (mode == "webdav"){
        $("#RedirectUriDiv").hide();
        $("#RefreshTokenDiv").hide();
        $("#UserDiv").show();
        $("#PasswordDiv").show();
        $(".sync-div").show();
        $("#ApiUrlDiv").show();
        $("#user_label").text("用户名");
        $("#accountForm").find("input[name=password]").attr("type", "password");
        $("#password_label").text("密码");
        $("#api_url_label").text("WebDav地址（WebDav Server）");
        $("#api_url").attr("placeholder", "https://webdav.mydomain.me");
        $("#SiteIdDiv").hide();
        $("#accountForm").find("input[name=root_id]").val("/");
        $("#aliQrCodeBtn").hide();
    }else if (mode == "yun139"){
        $("#RedirectUriDiv").hide();
        $("#RefreshTokenDiv").hide();
        $("#ApiUrlDiv").hide();
        $("#UserDiv").show();
        $("#PasswordDiv").show();
        $(".sync-div").show();
        $("#SiteIdDiv").hide();
        $("#user_label").text("手机号");
        $("#password_label").text("COOKIE");
        $("#accountForm").find("input[name=password]").attr("type", "text");
        $("#accountForm").find("input[name=root_id]").val("00019700101000000001");
        $("#aliQrCodeBtn").hide();
    }else if (mode == "googledrive"){
        $("#RedirectUriDiv").show();
        $("#RefreshTokenDiv").show();
        $("#UserDiv").show();
        $("#PasswordDiv").show();
        $(".sync-div").show();
        $("#SiteIdDiv").hide();
        $("#user_label").text("客户端ID（Client ID）");
        $("#password_label").text("客户端密码（Client Secret）");
        $("#accountForm").find("input[name=password]").attr("type", "password");
        $("#ApiUrlDiv").hide();
        $("#accountForm").find("input[name=root_id]").val("");
        $("#aliQrCodeBtn").hide();
    }
    diskd.handleUpdate();
}
$("#closeAccountBtn").on('click', function (ev){
    diskd.close();
});
$("#addDiskBtn").on('click', function (ev){
    $("#title").html("添加");
    $("#accountForm").find("input[name=name]").val("");
    $("#accountForm").find("input[name=user]").val("");
    $("#accountForm").find("input[name=password]").val("");
    $("#accountForm").find("input[name=refresh_token]").val("");
    $("#accountForm").find("input[name=redirect_uri]").val("");
    $("#accountForm").find("input[name=root_id]").val("");
    $("#accountForm").find("select[name=mode]").val("native");
    $("#accountForm").find("input[name=api_url]").val("");
    $("#accountForm").find("input[name=down_transfer]").prop("checked",false);
    $("#accountForm").find("input[name=transfer_domain]").val("");
    $("#accountForm").find("input[name=site_label]").val("");
    $("#accountForm").find("input[name=host]").val("");
    modeSelect.handleUpdate();
    dynamicChgMode("native");
    mdui.updateTextFields();
    diskd.toggle();
});

$("#updateDiskBtn").on('click', function (ev){
    $("#title").html("修改");
    var selectRecords = $('.mdui-table-row-selected');
    if(selectRecords.length != 1){
        mdui.snackbar({
            message: "请选择需要修改的挂载盘",
            timeout: 2000
        });
    }else{
        var id = selectRecords.eq(0).attr("data-id");
        $.ajax({
            method: 'GET',
            url: AdminApiUrl + '/account'+"?id=" + id,
            contentType: 'application/json',
            success: function (data) {
                var account = JSON.parse(data);
                dynamicChgMode(account.mode);
                $("#accountForm").find("input[name=id]").val(account.id);
                $("#accountForm").find("input[name=name]").val(account.name);
                $("#accountForm").find("input[name=password]").val(account.password);
                $("#accountForm").find("select[name=mode]").val(account.mode);
                $("#accountForm").find("input[name=user]").val(account.user);
                $("#accountForm").find("input[name=api_url]").val(account.api_url);
                $("#accountForm").find("input[name=refresh_token]").val(account.refresh_token);
                $("#accountForm").find("input[name=redirect_uri]").val(account.redirect_uri);
                $("#accountForm").find("input[name=root_id]").val(account.root_id);
                $("#accountForm").find("input[name=transfer_domain]").val(account.transfer_domain);
                $("#accountForm").find("input[name=site_label]").val(account.site_id);
                $("#accountForm").find("input[name=host]").val(account.host);
                modeSelect.handleUpdate();
                if(account.down_transfer == 1){
                    $("#accountForm").find("input[name=down_transfer]").prop("checked",true);
                }else{
                    $("#accountForm").find("input[name=down_transfer]").prop("checked",false);
                }
                mdui.updateTextFields();
                diskd.toggle();
            }
        });
    }
});
//拖动排序
var el = document.getElementById('items');
if (el){
    var sortable = new Sortable(el, {
        swap: true,
        swapClass: 'mdui-color-blue-100',
        animation: 150,
        handle: '.handle',
        onEnd: function (evt) {
            var sortIds = sortable.toArray();
            $.ajax({
                method: 'POST',
                url: AdminApiUrl + '/accounts/sort',
                data: JSON.stringify(sortIds),
                dataType: 'json',
                contentType: 'application/json',
                success: function (data) {
                }
            });
        },
    });
}
//删除网盘
$("#delDiskBtn").on("click", function (){
    var selectRecords = $('.mdui-table-row-selected');
    if(selectRecords.length == 0){
        mdui.snackbar({
            message: "请选择需要删除的挂载盘",
            timeout: 2000
        });
    }else{
        var delIds = [];
        selectRecords.each(function (j, record) {
            var id = $(this).attr("data-id");
            delIds.push(id);
        });
        $.ajax({
            method: 'DELETE',
            url: AdminApiUrl + '/accounts',
            data: JSON.stringify(delIds),
            dataType: 'json',
            contentType: 'application/json',
            success: function (data) {
                mdui.snackbar({
                    message: data.msg,
                    timeout: 2000,
                    onClose: function(){
                        location.reload();
                    }
                });
            }
        });
    }
});
$("#aliQrCodeBtn").on('click', function (ev) {
    $("#qrcodeDiv").toggle();
    var css = $("#qrcodeDiv").attr("style");
    if(!css.indexOf("display: none") != -1){
        genQrcode();
    }
    diskd.handleUpdate();
});
function genQrcode(){
    $("#qrInfo").show();
    $.ajax({url: AdminApiUrl + "/ali/qrcode", success:function(result){
            var d = JSON.parse(result);
            $("#qrcodeImg").attr("src", d.qr);
            param = JSON.parse(d.param);
            var interval =  setInterval(function(){
                timesRun += 1;
                if(timesRun === 60){
                    clearInterval(interval);
                }else{
                    queryStatus();
                }
            }, 2000);
        }});
}
let param = {};
var timesRun = 0;
$("#refreshQrcodeBtn").on('click', function (ev){
    genQrcode();
});
function queryStatus(){
    if(param != ""){
        $.ajax({
            method: "POST",
            url: AdminApiUrl + "/ali/qrcode/check",
            data: param,
            success:function(result){
                var d = JSON.parse(result);
                var qrc = d.qrCodeStatus;
                if(qrc == "NEW"){
                    //未扫码
                }else if(qrc === "SCANED"){
                    //已扫码
                }else if(qrc === 'CONFIRMED'){
                    timesRun = 59;
                    $("#accountForm").find("input[name=refresh_token]").val(d.refreshToken);
                    $("#qrcodeDiv").toggle();
                    snackbar("获取成功");
                }else if(d.qrCodeStatus === 'EXPIRED'){
                    //snackbar("二维码已过期，请刷新二维码！");
                    //二维码过期
                    timesRun = 59;
                }
            }
        });
    }
}
//网盘挂载-end
//密码文件-start
$("#addPwdDirBtn").on('click', function (ev){
    $("#title").html("添加");
    $("#configForm").find("input[name=file_path]").val("");
    $("#configForm").find("input[name=password]").val("");
    pwdd.toggle();
});
function savePwdFile(){
    var filePath = $("#configForm").find("input[name=file_path]").val();
    var password = $("#configForm").find("input[name=password]").val();
    var d = {};
    d["file_path"] = filePath;
    d["password"] = password;
    CommonRequest("/password/file", "POST", d)
    return false;
}
$("#updatePwdDirBtn").on('click', function (ev){
    $("#title").html("修改");
    var selectRecords = $('.mdui-table-row-selected');
    if(selectRecords.length > 1 || selectRecords.length == 0){
        mdui.snackbar({
            message: "请选择一条需要修改的记录",
            timeout: 2000
        });
    }else{
        var filePath = selectRecords.find("td").eq(1).html();
        var pwd = selectRecords.find("td").eq(2).html();
        $("#configForm").find("input[name=file_path]").val(filePath);
        $("#configForm").find("input[name=password]").val(pwd);
        pwdd.toggle();
    }
});
$("#delPwdDirBtn").on('click', function (ev){
    var selectRecords = $('.mdui-table-row-selected');
    if(selectRecords.length == 0){
        mdui.snackbar({
            message: "请选择需要删除的记录",
            timeout: 2000
        });
    }else{
        var delPaths = []
        selectRecords.each(function (j, record) {
            var filePath = $(this).find("td").eq(1).html();
            delPaths.push(filePath)
        });
        CommonRequest("/password/file", "DELETE", delPaths)
    }
});
//密码文件-end
//隐藏文件-start
$("#addHideBtn").on('click', function (ev){
    $("#configForm").find("input[name=file_path]").val("");
    hided.toggle();
});
function saveHide(){
    var hidePath = $("#configForm").find("input[name=file_path]").val();
    var d = {};
    d["hide_path"] = hidePath;
    CommonRequest("/hide/file", "POST", d)
    return false;
}
$("#delHideBtn").on('click', function (ev){
    var selectRecords = $('.mdui-table-row-selected');
    if(selectRecords.length == 0){
        mdui.snackbar({
            message: "请选择需要删除的记录",
            timeout: 2000
        });
    }else{
        var delPaths = []
        selectRecords.each(function (j, record) {
            var filePath = $(this).find("td").eq(1).html();
            delPaths.push(filePath)
        });
        CommonRequest("/hide/file", "DELETE", delPaths)
    }
});
//隐藏文件-end
//防盗链-start
$("#saveSafetyConfigBtn").on('click', function (ev){
    var enable_safety_link = $("#enable_safety_link").prop('checked');
    var is_null_referrer = $("#is_null_referrer").prop('checked');
    if(enable_safety_link){
        $("#configForm").find("input[name=enable_safety_link]").val("1");
    }else{
        $("#configForm").find("input[name=enable_safety_link]").val("0");
    }
    if(is_null_referrer){
        $("#configForm").find("input[name=is_null_referrer]").val("1");
    }else{
        $("#configForm").find("input[name=is_null_referrer]").val("0");
    }
    configSave();
});
//防盗链-end
function configSave() {
    var config = $("#configForm").serializeObject();
    $.ajax({
        method: 'POST',
        url: AdminApiUrl + '/config',
        data: JSON.stringify(config),
        contentType: 'application/json',
        success: function (data) {
            var d = JSON.parse(data);
            mdui.snackbar({
                message: d.msg,
                timeout: 1000,
                onClose: function(){
                    location.reload();
                }
            });
        }
    });
}
//手动刷新令牌-start
$("#refreshTokenBtn").on('click', function (ev){
    var selectRecords = $('.mdui-table-row-selected');
    if(selectRecords.length != 1){
        mdui.snackbar({
            message: "请选择需要刷新的挂载盘",
            timeout: 2000
        });
    }else{
        var id = selectRecords.eq(0).attr("data-id");
        $.ajax({
            method: 'POST',
            url: AdminApiUrl + '/refresh/login'+'?id='+id,
            success: function (data) {
                var d = JSON.parse(data);
                mdui.snackbar({
                    message: d.msg,
                    timeout: 3000
                });
            }
        });
    }
});
//手动刷新令牌-end
//手动刷新目录缓存-start
$("#refreshAllCacheBtn").on('click', function (ev){
    mdui.confirm('确认刷新全部挂载盘的缓存吗？', '', function(){
        $.ajax({
            method: 'GET',
            url: AdminApiUrl + '/cache/update/all',
            success: function (data) {
                var d = JSON.parse(data);
                mdui.snackbar({
                    message: d.msg,
                    timeout: 3000
                });
            }
        });
    },function(){
    }, {
        "confirmText": "确认",
        "cancelText": "取消",
    });
});
$("#refreshCacheBtn").on('click', function (ev){
    var selectRecords = $('.mdui-table-row-selected');
    if(selectRecords.length != 1){
        mdui.snackbar({
            message: "请选择需要刷新的挂载盘",
            timeout: 2000
        });
    }else{
        var id = selectRecords.eq(0).attr("data-id");
        var name = selectRecords.eq(0).attr("data-name");
        $("#cacheForm").find("input[name=account_id]").val(id);
        $.ajax({
            method: 'GET',
            url: AdminApiUrl + '/bypass'+'?account_id='+id,
            success: function (data) {
                var d = JSON.parse(data);
                if (d.data.name != ""){
                    $("#cacheForm").find("input[name=cache_folder]").val("/"+d.data.name);
                }else{
                    $("#cacheForm").find("input[name=cache_folder]").val("/"+name);
                }
            }
        });
        cached.toggle();
    }
});
$("#confirmRefreshCacheBtn").on('click', function (ev){
    var formData = new FormData();
    var id = $("#cacheForm").find("input[name=account_id]").val();
    var cf = $("#cacheForm").find("input[name=cache_folder]").val();
    if(cf == ""){
        cf = "/";
    }
    formData.append("accountId", id);
    formData.append("cachePath", cf);
    $.ajax({
        method: 'POST',
        url: AdminApiUrl + '/cache/update',
        data: formData,
        cache: false,
        contentType: false,
        processData: false,
        success: function (data) {
            var d = JSON.parse(data);
            mdui.snackbar({
                message: d.msg,
                timeout: 3000,
                onClose: function(){
                    cached.toggle();
                }
            });
        }
    });
});
$("#closeCacheBtn").on('click', function (ev){
    cached.close();
});
//手动刷新目录缓存-end
//手动上传文件-start
$("#openUploadDialog").on('click', function (ev){
    var selectRecords = $('.mdui-table-row-selected');
    if(selectRecords.length != 1){
        mdui.snackbar({
            message: "请选择需要上传的挂载盘",
            timeout: 2000
        });
    }else{
        var id = selectRecords.eq(0).attr("data-id");
        var name = selectRecords.eq(0).attr("data-name");
        $("#uploadForm").find("input[name=account_id]").val(id);
        $.ajax({
            method: 'GET',
            url: AdminApiUrl + '/bypass'+'?account_id='+id,
            success: function (data) {
                var d = JSON.parse(data);
                if (d.data.name != ""){
                    $("#uploadForm").find("input[name=upload_folder]").val("/"+d.data.name);
                }else{
                    $("#uploadForm  ").find("input[name=upload_folder]").val("/"+name);
                }
            }
        });
        ud.toggle();
    }
});
var status = 0;
$(".uploadBtn").on('click', function (ev){
    var type = $(this).val();
    var fileObjs = document.getElementById('uploadFile').files;
    var accountId = $("#uploadForm").find("input[name=account_id]").val();
    var uploadFolder = $("#uploadForm").find("input[name=upload_folder]").val();
    var formData = new FormData();
    formData.append("uploadAccount", accountId);
    formData.append("uploadPath", uploadFolder);
    formData.append("type", type);
    $.each(fileObjs, function (i, f) {
        formData.append("uploadFile", f);
    });
    if(status == 1){
        return;
    }
    status = 1;
    mdui.snackbar({
        message: "开始上传，请耐心等待",
        timeout: 1000
    });
    $.ajax({
        method: 'POST',
        url: AdminApiUrl + "/upload", //上传文件的请求路径必须是绝对路劲
        data: formData,
        cache: false,
        contentType: false,
        processData: false,
        success: function (data) {
            var d = JSON.parse(data);
            mdui.snackbar({
                message: d.msg,
                timeout: 3000,
                onClose: function(){
                    ud.toggle();
                }
            });
            status = 0
        }
    });
});
$("#closeUploadBtn").on('click', function (ev){
    ud.close();
});
//预览配置重置
$("#resetViewConfig").on('click', function (ev){
    $("#configForm").find("input[name=enable_preview]").prop("checked", true);
    $("#configForm").find("input[name=subtitle][value='']").prop('checked',true);
    $("#configForm").find("input[name=danmuku][value='']").prop('checked',true);
    $("#configForm").find("input[name=enable_lrc]").prop("checked", false);
    $("#configForm").find("input[name=image]").val("png,gif,jpg,bmp,jpeg,ico,svg");
    $("#configForm").find("input[name=video]").val("mp4,mkv,m3u8,ts,avi");
    $("#configForm").find("input[name=audio]").val("mp3,wav,ape,flac");
    $("#configForm").find("input[name=code]").val("txt,go,html,js,java,json,css,lua,sh,sql,py,cpp,xml,jsp,properties,yaml,ini");
    $("#configForm").find("input[name=other]").val("*");
    $("#configForm").find("input[name=lrc_path]").val("");
    $("#configForm").find("input[name=subtitle_path]").val("");
});
//手动上传文件-end
$.fn.serializeObject = function()
{
    var o = {};
    var a = this.serializeArray();
    $.each(a, function() {
        if (o[this.name]) {
            if (!o[this.name].push) {
                o[this.name] = [o[this.name]];
            }
            o[this.name].push(this.value || '');
        } else {
            o[this.name] = this.value || '';
        }
    });
    return o;
};
//分流下载-start
$("#addByPassBtn").on('click', function (ev){
    $("#title").html("添加");
    $("#configForm").find("input[name=name]").val("");
    $("#configForm").find("input[name=bind_account]").prop("checked", false);
    bypassd.toggle();
});
function saveBypass(){
    var name = $("#configForm").find("input[name=name]").val();
    var id = $("#configForm").find("input[name=id]").val();
    var accounts = [];
    $('input[name=bind_account]:checked').each(function(i){
        var account = {};
        account["id"] = $(this).val();
        accounts.push(account);
    });
    var d = {};
    d["id"] = id;
    d["name"] = name;
    d["accounts"] = accounts;
    if(accounts.length == 0){
        snackbar("请勾选需要分流的网盘");
    }else{
        CommonRequest("/bypass", "POST", d)
    }
    return false;
}
$("#updateByPassBtn").on('click', function (ev){
    $("#title").html("修改");
    var selectRecords = $('.mdui-table-row-selected');
    if(selectRecords.length > 1 || selectRecords.length == 0){
        snackbar("请选择一条需要修改的记录");
    }else{
        var id = selectRecords.find("td").eq(1).attr("data-id");
        var name = selectRecords.find("td").eq(1).html();
        var accounts = selectRecords.find("td").eq(2).attr("data-accounts");
        $("#configForm").find("input[name=name]").val(name);
        $("#configForm").find("input[name=id]").val(id);
        $("#configForm").find("input[name=bind_account]").prop("checked", false);
        $.each(accounts.split(","), function (i, item) {
            if(item != ""){
                $("#configForm").find("input[type='checkbox'][value='"+item+"']").prop('checked',true);
            }
        });
        bypassd.toggle();
    }
});
$("#delByPassBtn").on('click', function (ev){
    var selectRecords = $('.mdui-table-row-selected');
    if(selectRecords.length == 0){
        snackbar("请选择需要删除的记录");
    }else{
        var delIds = []
        selectRecords.each(function (j, record) {
            var id = $(this).find("td").eq(1).attr("data-id");
            delIds.push(id)
        });
        CommonRequest("/bypass", "DELETE", delIds)
    }
});
//分流下载-end
function snackbar(msg) {
    mdui.snackbar({
        message: msg,
        timeout: 2000
    });
}
//cache - start
$("#searchKey").on("keydown", function (event) {
    if (event.keyCode == 13) {
        var key = $("#searchKey").val();
        if(key != ''){
            window.location.href = AdminUrl+ "/cache?path="+encodeURI(key);
        }else{
            window.location.href = AdminUrl+ "/cache";
        }
    }
});
$(".clearCacheBtn").on('click', function (event) {
    var path = $(this).data("path");
    $("#configForm").find("input[name=path]").val(path);
    $("#configForm").find("input[type='checkbox'][value='1']").prop('checked',true);
    clearCached.toggle();
});
function saveClearCache(){
    var d = {};
    var path = $("#configForm").find("input[name=path]").val();
    var isLoopChildren= $("#configForm").find("input[type='radio']:checked").val();
    d["path"] = path;
    d["is_loop_children"] = isLoopChildren;
    CommonRequest("/cache/clear", "POST", d);
    return false;
}
//cache config
$("#cacheConfig").on('click', function (event) {
    var selectRecords = $('.mdui-table-row-selected');
    if(selectRecords.length != 1){
        snackbar("请选择需要操作的记录");
    }else{
        var data = selectRecords.eq(0).data();
        $("#cacheConfigForm").find("input[name=id]").val(data.id);
        $("#cacheConfigForm").find("select[name=cache_policy]").val(data.cachePolicy);
        $("#cacheConfigForm").find("input[name=expire_time_span]").val(data.expireTimespan);
        cachePolicySelect.handleUpdate();
        dynamicCachePolicy(data.cachePolicy)
        $("#cacheConfigForm").find("input[name=sync_cron]").val(data.syncCron);
        if(data.syncChild == 0){
            $("#cacheConfigForm").find("input[name=sync_child]").prop("checked", true);
        }
        var name = data.name;
        $.ajax({
            method: 'GET',
            url: AdminApiUrl + '/bypass'+'?account_id='+data.id,
            success: function (data) {
                var d = JSON.parse(data);
                if (d.data.name != ""){
                    $("#cacheConfigForm").find("input[name=sync_dir]").val("/"+d.data.name);
                }else{
                    $("#cacheConfigForm").find("input[name=sync_dir]").val("/"+name);
                }
            }
        });
        cacheConfigd.toggle();
    }
});
$("#cachePolicy").on('change', function () {
    var cachePolicy = $(this).val();
    dynamicCachePolicy(cachePolicy);
    cacheConfigd.handleUpdate();
});
function saveCacheConfig(){
    var formData = $("#cacheConfigForm").serializeArray();
    var d = parseFormData(formData);
    d["expire_time_span"] =  parseInt(d["expire_time_span"]);
    d["sync_child"] =  parseInt(d["sync_child"]);
    CommonRequest("/cache/config", "POST", d);
    cacheConfigd.toggle();
    return false;
}
function dynamicCachePolicy(cachePolicy){
    if(cachePolicy=="" || cachePolicy == "nc"){
        $(".memoryCacheConfigDiv").hide();
        $(".dbCacheConfigDiv").hide();
    }else if (cachePolicy == "dc"){
        $(".memoryCacheConfigDiv").hide();
        $(".dbCacheConfigDiv").show();
    }else if (cachePolicy == "mc"){
        $(".memoryCacheConfigDiv").show();
        $(".dbCacheConfigDiv").hide();
    }
}
//cache - end
//webdav - start
$("#saveDavConfigBtn").on('click', function (ev){
    var davMode= $("#configForm").find("input[name='dav_mode_group']:checked").val();
    var davDownMode= $("#configForm").find("input[name='dav_down_mode_group']:checked").val();
    $("#configForm").find("input[name=dav_mode]").val(davMode);
    $("#configForm").find("input[name=dav_down_mode]").val(davDownMode);
    var davPath = $("#configForm").find("input[name=dav_path]").val();
    if(davPath != ""){
        var enableDav = $("#enable_dav").prop('checked');
        if(enableDav){
            $("#configForm").find("input[name=enable_dav]").val("1");
        }else{
            $("#configForm").find("input[name=enable_dav]").val("0");
        }
        configSave();
    }else{
        snackbar("请输入WebDav请求路径!");
    }

});
//webdav - end
function parseFormData(formArray){
   var d = {};
   $.each(formArray, function (i, item) {
       d[item.name] = item.value;
   });
   return d;
}
