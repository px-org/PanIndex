var $ = mdui.$;
var bd = new mdui.Dialog("#account_dialog");
$('#theme-toggle').on('click', function(){
    $('body').removeClass('mdui-theme-layout-auto');
    if($('body').hasClass('mdui-theme-layout-dark')){
        $('body').removeClass('mdui-theme-layout-dark');
        $('#theme-toggle i').text('brightness_4');
        Cookies.set("Theme", "mdui-light", {expires : 3650, path:"/"});
    }else{
        $('body').addClass('mdui-theme-layout-dark');
        $('#theme-toggle i').text('brightness_5');
        Cookies.set("Theme", "mdui-dark", {expires : 3650, path:"/"});
    }
});
function addAccount() {
    bd.open();
}
$("#accountForm").find("input[name=mode]").on('change', function () {
    var mode = $(this).val();
    dynamicChgMode(mode);
});
$("#resetBtn").on('click', function () {
    $("#accountForm").find("input[name=name]").val("");
    $("#accountForm").find("input[name=user]").val("");
    $("#accountForm").find("input[name=password]").val("");
    $("#accountForm").find("input[name=refresh_token]").val("");
    $("#accountForm").find("input[name=redirect_uri]").val("");
    $("#accountForm").find("input[name=root_id]").val("");
    $("#accountForm").find("input[name=mode][value=native]").prop("checked", true);
    $("#accountForm").find("input[name=sync_dir]").val("/");
    $("#accountForm").find("input[name=sync_child]").prop("checked",true);
    $("#accountForm").find("input[name=sync_child]").prop("checked",true);
});
//fillAccount(0);
function fillAccount(index) {
    if(accounts.length == 0){
        mdui.snackbar({
            message: '您还没有绑定网盘账号！'
        });
        dynamicChgMode("native");
        var inst = new mdui.Tab('.mdui-tab');
        inst.next();
        return;
    }
    $("ul").find("li").removeClass("mdui-list-item-active");
    $("ul").find("li").eq(index).addClass("mdui-list-item-active");
    var account = accounts[index];
    $("#accountForm").find("input[name=id]").val(account.id);
    $("#accountForm").find("input[name=name]").val(account.name);
    $("#accountForm").find("input[name=password]").val(account.password);
    $("#accountForm").find("input[name=mode][value="+account.mode+"]").prop("checked", true);
    $("#accountForm").find("input[name=user]").val(account.user);
    $("#accountForm").find("input[name=refresh_token]").val(account.refresh_token);
    $("#accountForm").find("input[name=redirect_uri]").val(account.redirect_uri);
    $("#accountForm").find("input[name=root_id]").val(account.root_id);
    if(account.sync_dir == ""){
        $("#accountForm").find("input[name=sync_dir]").val("/");
    }else{
        $("#accountForm").find("input[name=sync_dir]").val(account.sync_dir);
    }
    if(account.sync_child == "0"){
        $("#accountForm").find("input[name=sync_child]").prop("checked",true);
    }else{
        $("#accountForm").find("input[name=sync_child]").prop("checked",false);
    }
    fillCacheRecord(account)
    dynamicChgMode(account.mode);
}
function dynamicChgMode(mode){
    if(mode == "native"){
        $("#RedirectUriDiv").hide();
        $("#RefreshTokenDiv").hide();
        $("#UserDiv").hide();
        $("#PasswordDiv").hide();
        $("#recordDiv").hide();
        $(".sync-div").hide();
    }else if (mode == "cloud189"){
        $("#RedirectUriDiv").hide();
        $("#RefreshTokenDiv").hide();
        $("#UserDiv").show();
        $("#PasswordDiv").show();
        $("#recordDiv").show();
        $(".sync-div").show();
        $("#user_label").text("用户名");
        $("#password_label").text("密码");
    }else if (mode == "teambition"){
        $("#RedirectUriDiv").hide();
        $("#RefreshTokenDiv").hide();
        $("#UserDiv").show();
        $("#PasswordDiv").show();
        $("#recordDiv").show();
        $(".sync-div").show();
        $("#user_label").text("用户名");
        $("#password_label").text("密码");
    }else if (mode == "teambition-us"){
        $("#RedirectUriDiv").hide();
        $("#RefreshTokenDiv").hide();
        $("#UserDiv").show();
        $("#PasswordDiv").show();
        $("#recordDiv").show();
        $(".sync-div").show();
        $("#user_label").text("用户名");
        $("#password_label").text("密码");
    }else if (mode == "aliyundrive"){
        $("#RedirectUriDiv").hide();
        $("#RefreshTokenDiv").show();
        $("#UserDiv").hide();
        $("#PasswordDiv").hide();
        $(".sync-div").show();
        $("#recordDiv").show();
    }else if (mode == "onedrive"){
        $("#RedirectUriDiv").show();
        $("#RefreshTokenDiv").show();
        $("#UserDiv").show();
        $("#PasswordDiv").show();
        $("#recordDiv").show();
        $(".sync-div").show();
        $("#user_label").text("客户端ID（Client ID）");
        $("#password_label").text("客户端密码（Client Secret）");
    }
}
var accountStatus = 0;
$(".saveAccountBtn").on("click", function () {
    var account = $("#accountForm").serializeObject();
    var type = $(this).val();
    if(type == 0){
        account.id = "";
    }
    if(!account.root_id){
        mdui.snackbar({
            message: "请输入根目录ID",
            timeout: 2000
        });
        return false;
    }
    if(accountStatus == 1){
        return false;
    }
    accountStatus = 1;
    var btn = $(this);
    btn.toggleClass("running");
    var syncChild = $("#accountForm").find("input[name=sync_child]").prop("checked");
    var syncDir = $("#accountForm").find("input[name=sync_dir]").val();
    if(syncChild){
        account.sync_child = 0;
    }else{
        account.sync_child = 1;
    }
    var config = {accounts:[account]};
    $.ajax({
        method: 'POST',
        url: '/api/admin/save?token='+ApiToken,
        data: JSON.stringify(config),
        contentType: 'application/json',
        success: function (data) {
            var d = JSON.parse(data);
            mdui.snackbar({
                message: "账号保存成功，正在进行目录缓存，请稍后刷新页面查看缓存结果...在此期间请勿重启，以免造成数据重叠！",
                timeout: 2000,
                onClose: function(){
                    accountStatus = 0;
                    location.reload();
                }
            });
        }
    });
});
$("#deleteBtn").on("click", function (){
    var id = $("#accountForm").find("input[name=id]").val();
    if(id != ""){
        $.ajax({
            method: 'POST',
            url: '/api/admin/deleteAccount?token='+ApiToken+'&id='+id,
            success: function (data) {
                var d = JSON.parse(data);
                mdui.snackbar({
                    message: d.msg,
                    timeout: 2000,
                    onClose: function(){
                        location.reload();
                    }
                });
            }
        });
    }
});
$(".saveConfigBtn").on("click", function () {
    var config = $("#configForm").serializeObject();
   /* if(!config.host || !config.port || !config.admin_password){
        mdui.snackbar({
            message: "必填项不能为空",
            timeout: 2000
        });
        return false;
    }*/
    config.port = Number(config.port);
    $.ajax({
        method: 'POST',
        url: '/api/admin/save?token='+ApiToken,
        data: JSON.stringify(config),
        contentType: 'application/json',
        success: function (data) {
            var d = JSON.parse(data);
            mdui.snackbar({
                message: d.msg,
                timeout: 2000
            });
        }
    });
});
$(".saveCronBtn").on("click", function () {
    var config = $("#cronForm").serializeObject();
    if(!config.refresh_cookie){
        mdui.snackbar({
            message: "必填项不能为空",
            timeout: 2000
        });
        return false;
    }
    $.ajax({
        method: 'POST',
        url: '/api/admin/save?token='+ApiToken,
        data: JSON.stringify(config),
        contentType: 'application/json',
        success: function (data) {
            var d = JSON.parse(data);
            mdui.snackbar({
                message: d.msg,
                timeout: 2000
            });
        }
    });
});
$(".defaultAccount").on('click', function (ev){
    $(".defaultAccount").prop("checked", false);
    $(this).prop("checked", true);
    var id = $(this).attr("data-id");
    $.ajax({
        method: 'POST',
        url: '/api/admin/setDefaultAccount?token='+ApiToken+'&id='+id,
        success: function (data) {
            var d = JSON.parse(data);
            mdui.snackbar({
                message: d.msg,
                timeout: 2000
            });
        }
    });
});
function fillCacheRecord(account){
    var text = "";
    if(account.cookie_status == 1){
        text += "<p>cookie： 未刷新</p>";
    }else if (account.cookie_status == 2){
        text += "<p>cookie： 正常</p>";
    }else if (account.cookie_status == 3){
        text += "<p>cookie： 失效</p>";
    }else if (account.cookie_status == 4){
        text += "<p>cookie： 登录失败</p>";
    }else if (account.cookie_status == -1){
        text += "<p>cookie： 刷新中</p>";
    }
    if(account.status == 1){
        text += "<p>状态： 未缓存</p>";
    }else if (account.status == 2){
        text += "<p>状态： 缓存成功</p>";
    }else if (account.status == 3){
        text += "<p>状态： 缓存失败</p>";
    }else if (account.status == -1){
        text += "<p>状态： 缓存中</p>";
    }
    if(account.files_count > 0){
        text += "<p>文件数：" + account.files_count + "</p>";
    }else{
        text += "<p>文件数：-</p>";
    }
    if(account.time_span){
        text += "<p>耗时：" + account.time_span + "</p>";
    }else{
        text += "<p>耗时：-</p>";
    }
    if(account.last_op_time){
        text += "<p>最近一次缓存：" + account.last_op_time + "</p>";
    }else{
        text += "<p>最近一次缓存：-</p>";
    }
    $("#cacheRecord").html(text);
}
function updateCache(){
    var id = $("#accountForm").find("input[name=id]").val();
    $.ajax({
        method: 'POST',
        url: '/api/admin/updateCache?token='+ApiToken+'&id='+id,
        success: function (data) {
            var d = JSON.parse(data);
            mdui.snackbar({
                message: d.msg,
                timeout: 3000
            });
        }
    });
}
function updateCookie(){
    var id = $("#accountForm").find("input[name=id]").val();
    $.ajax({
        method: 'POST',
        url: '/api/admin/updateCookie?token='+ApiToken+'&id='+id,
        success: function (data) {
            var d = JSON.parse(data);
            mdui.snackbar({
                message: d.msg,
                timeout: 3000
            });
        }
    });
}
$(".acli").on('click', function(e){
    if(e.target.tagName == 'INPUT')return;
    fillAccount($(this).attr('data-index'));
});
var status = 0;
$('.uploadBtn').on('click', function () {
    var type = $(this).val();
    var fileObjs = document.getElementById('uploadFile').files;
    var formData = new FormData();
    formData.append("uploadAccount", $('#uploadAccount').val());
    formData.append("uploadPath", $('#uploadPath').val());
    formData.append("type", type);
    $.each(fileObjs, function (i, f) {
        formData.append("uploadFile", f);
    });
    if(status == 1){
        return;
    }
    status = 1;
    if(type == 0 || type == 2){
        mdui.snackbar({
            message: "开始上传，请耐心等待",
            timeout: 1000
        });
    }else{
        mdui.snackbar({
            message: "开始刷新，请耐心等待",
            timeout: 1000
        });
    }
    var btn = $(this);
    btn.toggleClass("running");
    $.ajax({
        method: 'POST',
        url: "/api/admin/upload?token="+ApiToken, //上传文件的请求路径必须是绝对路劲
        data: formData,
        cache: false,
        contentType: false,
        processData: false,
        success: function (data) {
            var d = JSON.parse(data);
            mdui.snackbar({
                message: d.msg,
                timeout: 3000
            });
            btn.toggleClass("running");
            status = 0
        }
    });
});
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