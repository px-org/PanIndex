var qa_info = {"FHD":"全高清","LD":"流畅","SD":"标清","HD":"高清"};
var qa_sort = {"全高清": 1080,"流畅":360,"标清":540,"高清":720};

function getTranscodeInfo(accountId, fileId){
    var qas = [];
    $.ajax({
        method: 'POST',
        async: false,
        url: '/api/public/transcode?accountId='+accountId+'&fileId='+fileId,
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

function initVideo(container, qas, title){
    var vname = title.split(".")[0];
    if(qas.length > 0){
        var art = new Artplayer({
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
            quality: qas,
            fullscreen: true,
            fullscreenWeb: true,
            pip: true,
            autoplay: true,
            autoSize: true,
            playbackRate: true,
            aspectRatio: true,
            screenshot: true,
            setting: true,
            miniProgressBar: true,
            subtitle: {
                url: vname + '.vtt',
                type: 'srt',
                encoding: 'utf-8',
                bilingual: true,
                style: {
                    color: '#03A9F4',
                    'font-size': '30px',
                },
            },
        });
        return art;
    }
}