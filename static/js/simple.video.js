var qa_info = {"FHD":"全高清","LD":"流畅","SD":"标清","HD":"高清"};
var qa_sort = {"全高清": 1080,"流畅":360,"标清":540,"高清":720};

function getTranscodeInfo(accountId, fileId){
    var qas = [];
    $.ajax({
        method: 'POST',
        async: false,
        url: $config.path_prefix+'/api/public/transcode?accountId='+accountId+'&fileId='+fileId,
        success: function (data) {
            var d = JSON.parse(data);
            if(d.video_preview_play_info){
                $.each(d.video_preview_play_info.live_transcoding_task_list, function(i, item){
                    if(item.status == "finished"){
                        var qa = {};
                        qa["url"] = item.url;
                        qa["html"] = qa_info[item.template_id];
                        qa["type"] = "hls";
                        qas.push(qa);
                    }
                });
                qas = qas.sort(function(s, t){
                    var a = qa_sort[s["html"]];
                    var b = qa_sort[t["html"]];
                    if (a < b) return -1;
                    if (a > b) return 1;
                    return 0;
                });
            }
        }
    });
    return qas;
}

function buildOriginalVideo(url, ft){
    var qas = [];
    var qa = {};
    qa["url"] = url;
    qa["html"] = "原画";
    if(ft == "m3u8"){
        qa["type"] = "hls";
    }else if (ft == "flv"){
        qa["type"] = "flv";
    }else{
        qa["type"] = "";
    }
    qas.push(qa);
    return qas;
}
function initVideo(container, qas, title, parentPath){
    var vname = title.split(".")[0];
    var subtitle = $("#video-modal").attr("data-config-subtitle");
    var subtitlePath = $("#video-modal").attr("data-config-subtitle-path");
    var danmuku = $("#video-modal").attr("data-config-danmuku");
    var danmukuPath = $("#video-modal").attr("data-config-danmuku-path");
    if(parentPath == "/"){
        parentPath = "";
    }
    var settings = [
        {
            html: '选择画质',
            width: 150,
            tooltip: qas[0].html,
            selector: qas,
            onSelect: function(item, $dom) {
                art.switchQuality(item.url, item.html);
                return item.html;
            },
        }
    ];
    if(subtitle != ""){
        var vpath =  subtitlePath + vname;
        var subtitlePlugin = {
            html: '选择字幕',
            width: 250,
            tooltip: '字幕',
            selector: [
                {
                    default: true,
                    html: '<span style="color:red">关闭</span>',
                    url: '',
                },
                {
                    default: false,
                    html: '<span style="color:yellow">字幕</span>',
                    url: parentPath + "/" + vpath + '.' + subtitle,
                }
            ],
            onSelect: function(item, $dom) {
                if(item.url == ''){
                    art.subtitle.show = false;
                    return "";
                }
                art.subtitle.show = true;
                art.subtitle.url = item.url;
                art.subtitle.encoding = "utf-8";
                art.subtitle.bilingual = true;
                art.subtitle.style({
                    'font-size': '30px',
                });
                return item.html;
            }
        };
        settings.push(subtitlePlugin);
    }
    var plugins = [];
    if(danmuku == "1"){
        danmukuPath = danmukuPath + vname + ".xml";
        plugins.push(artplayerPluginDanmuku({
            danmuku: parentPath + "/" + danmukuPath,
            speed: 5,
            maxlength: 50,
            margin: [10, 100],
            opacity: 1,
            fontSize: 25,
            synchronousPlayback: false,
            theme: "light",
            mount: document.querySelector('.danmuinput'),
        }));
    }
    if(qas.length > 0){
        $(".artplayer-app").css('height', $('.mdui-video-container').innerHeight());
        var art = new Artplayer({
            lang: 'zh-cn',
            title: title,
            container: container,
            url: qas[0].url,
            customType: {
                flv: function (video, url) {
                    const flvPlayer = flvjs.createPlayer({
                        type: 'flv',
                        url: url,
                    });
                    flvPlayer.attachMediaElement(video);
                    flvPlayer.load();
                },
                m3u8: function (video, url) {
                    var hls = new Hls();
                    hls.loadSource(url);
                    hls.attachMedia(video);
                },
            },
            //quality: qas,
            autoSize: true,
            fullscreen: true, //全屏
            //fullscreenWeb: true, //网页全屏
            //pip: true,
            autoplay: false, //自动播放
            autoPlayback: true,
            lock: true,
            isLock: true, //移动端锁屏操作
            fastForward: true, //移动端添加长按视频快进
            autoOrientation: true, //全屏自动翻转
            //autoSize: true,
            playbackRate: true,//显示视频播放速度
            aspectRatio: true,//显示视频长宽比
            //screenshot: true,
            setting: true,
            miniProgressBar: true,
            theme: '#23ade5',
            settings: settings,
            whitelist: ['*'],
            moreVideoAttr: {
               // crossOrigin: 'anonymous',
            },
            plugins: plugins
        });
        return art;
    }
}