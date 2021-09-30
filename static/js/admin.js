var $ = mdui.$;
var bd = new mdui.Dialog("#account_dialog");
var pwdd = new mdui.Dialog("#pwd_dialog");
var hided = new mdui.Dialog("#hide_dialog");
var diskd = new mdui.Dialog("#disk_dialog", {modal:true});
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
$(".saveConfigBtn").on("click", function () {
    var config = $("#configForm").serializeObject();
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
    if(accountStatus == 1){
        return false;
    }
    accountStatus = 1;
    var btn = $(this);
    btn.toggleClass("running");
    var syncChild = $("#accountForm").find("input[name=sync_child]").prop("checked");
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
$("#accountForm").find("input[name=mode]").on('change', function () {
    var mode = $(this).val();
    dynamicChgMode(mode);
});
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
    $("#accountForm").find("input[name=mode][value=native]").prop("checked", true);
    $("#accountForm").find("input[name=sync_dir]").val("/");
    $("#accountForm").find("input[name=sync_child]").prop("checked",true);
    $("#accountForm").find("input[name=sync_child]").prop("checked",true);
    dynamicChgMode("native");
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
            method: 'POST',
            url: '/api/admin/getAccount?token='+ApiToken+"&id=" + id,
            contentType: 'application/json',
            success: function (data) {
                var account = JSON.parse(data);
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
                url: '/api/admin/sortAccounts?token='+ApiToken,
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
            method: 'POST',
            url: '/api/admin/deleteAccounts?token='+ApiToken,
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
//网盘挂载-end
//密码文件-start
$("#addPwdDirBtn").on('click', function (ev){
    $("#title").html("添加");
    $("#file_id").val("");
    $("#pwd").val("");
    pwdd.toggle();
});
function savePwddir(){
    var pwdDirId = $("#configForm").find("input[name=pwd_dir_id]").val();
    var fileId = $("#file_id").val();
    var pwd = $("#pwd").val();
    var result = [];
    var flag = false;
    $.each(pwdDirId.split(","), function(i, item){
        var fId = item.split(":")[0];
        if(fId == fileId){
            flag = true;
            result.push(fileId+":"+pwd);
        }else{
            result.push(item);
        }
    });
    if(!flag){
        //追加
        result.push(fileId+":"+pwd);
    }
    $("#configForm").find("input[name=pwd_dir_id]").val(result.join(","))
    configSave();
    pwdd.toggle();
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
        var fileId = selectRecords.find("td").eq(1).html();
        var pwd = selectRecords.find("td").eq(2).html();
        $("#file_id").val(fileId);
        $("#pwd").val(pwd);
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
        var pwdDirId = $("#configForm").find("input[name=pwd_dir_id]").val();
        var result = [];
        $.each(pwdDirId.split(","), function(i, item){
            var fId = item.split(":")[0];
            var flag = true;
            selectRecords.each(function (j, record) {
                var fileId = $(this).find("td").eq(1).html();
                if(fId == fileId){
                    flag = false;
                }
            });
            if(flag){
                result.push(item);
            }
        });
        $("#configForm").find("input[name=pwd_dir_id]").val(result.join(","))
        configSave();
    }
});
//密码文件-end
//隐藏文件-start
$("#addHideBtn").on('click', function (ev){
    $("#file_id").val("");
    hided.toggle();
});
function saveHide(){
    var hideId = $("#configForm").find("input[name=hide_file_id]").val();
    var fileId = $("#file_id").val();
    var result = [];
    var flag = false;
    $.each(hideId.split(","), function(i, item){
        var fId = item;
        if(fId == fileId){
            flag = true;
            result.push(fileId);
        }else{
            result.push(item);
        }
    });
    if(!flag){
        //追加
        result.push(fileId);
    }
    $("#configForm").find("input[name=hide_file_id]").val(result.join(","))
    configSave();
    hided.toggle();
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
        var hideId = $("#configForm").find("input[name=hide_file_id]").val();
        var result = [];
        $.each(hideId.split(","), function(i, item){
            var fId = item;
            var flag = true;
            selectRecords.each(function (j, record) {
                var fileId = $(this).find("td").eq(1).html();
                if(fId == fileId){
                    flag = false;
                }
            });
            if(flag){
                result.push(item);
            }
        });
        $("#configForm").find("input[name=hide_file_id]").val(result.join(","))
        configSave();
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
        $("#configForm").find("input[name=is_null_referrer]").val("1");
    }
    configSave();
});
//防盗链-end
function configSave() {
    var config = $("#configForm").serializeObject();
    $.ajax({
        method: 'POST',
        url: '/api/admin/save?token='+ApiToken,
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