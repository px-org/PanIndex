$(document).ready(function() {
    $('.icon-file').on('click', function() {
        var dURL = $(this).attr("data-url");
        var title = $(this).attr("data-title");
        var dmt = $(this).attr("data-media-type");
        var fileType = $(this).attr("data-file-type");
        if(dmt == 1){
            $(this).lightGallery({
                fullScreen: true,
                dynamic: true,
                dynamicEl: [{
                    "src": dURL,
                    "subHtml": "<h4>"+title+"</h4>"
                }]
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
                fullScreen: true,
                dynamicEl: [{
                    html: '<video class="lg-video-object lg-html5" controls preload="none"><source src="'+dURL+'">Your browser does not support HTML5 video</video>\'',
                    "subHtml": "<h4>"+title+"</h4>"
                }]
            })
            return;
        }
        var fullUrl = window.location.protocol+"//"+window.location.host + dURL;
        if(HerokuappUrl != ""){
             fullUrl = HerokuappUrl+dURL;
        }
        if(fileType == "doc" || fileType == "docx" || fileType == "dotx"
            || fileType == "ppt" || fileType == "pptx" || fileType == "xls" || fileType == "xlsx"){
            window.open("https://view.officeapps.live.com/op/view.aspx?src="+fullUrl);
        }else{
            window.location.href = dURL;
        }
    });
    $('.folderDown').on('click', function() {
        var fileId = $(this).attr("data-file-id");
        $.ajax({
            type: 'POST',
            url: "/api/downloadMultiFiles?fileId="+fileId,
            async:false,
            success: function(data){
                window.location.href = data.redirect_url;
            }
        });
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
});
function sortTable(sort_order, data_type){
    $('table tbody > tr').not('.parent').sortElements(function (a, b) {
        let data_a = $(a).find("td[class='"+data_type+"']").text(), data_b = $(b).find("td[class='"+data_type+"']").text();
        let rt = data_a.localeCompare(data_b);
        return (sort_order === "down") ? 0-rt : rt;
    });
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
