$(function () {

    let obj = {
        apiData: null,
        valueData: null
    }
    let width = document.body.clientWidth;
    window.onresize = function () {
        let oldWidth = width;
        width = document.body.clientWidth;
        if ((oldWidth >= 768 && width <= 768) || (oldWidth <= 768 && width >= 768)) {
            createEel();
        }
    }
    let data = {
        chart_name: '',
        repo_name: '',
        repo_url: '',
        version: '',
        name: '',
        value: ''
    };
    let container, options, editor;
    $("body").on("click", "#autoinstall1", function () {
        autoinstall();
    })
    $("body").on("click", "#autoinstall2", function () {
        autoinstall();
    })
    $("body").on("click", "#submitBtn", function () {
        if (!editor) {
            bootoast({
                message: '配置参数还在加载中，请稍候！',
                type: 'warning',
                position: 'top-center',
                timeout: 3
            });
            return;
        }
        const updatedJson = editor.get();
        let value = $('#dataName')[0].value.trim();
        if (value) {
            data.value = JSON.stringify(updatedJson);
            data.name = value;
            data.chart_name = obj.apiData.name
            data.repo_url = obj.apiData.repository.url
            data.repo_name = obj.apiData.repository.name
            data.version = obj.apiData.version
            let createUrl = window.location.origin + '/api/v2/helm/release/create';
            jQuery.ajax({
                url: createUrl,
                type: "POST",
                data: JSON.stringify(data),
                dataType: "json",
                timeout : 30000,
                contentType: "application/json; charset=utf-8",
                success: function (res) {
                    bootoast({
                        message: '操作成功',
                        type: 'success',
                        position: 'top-center',
                        timeout: 3
                    });
                    $('#myModal').modal('hide');
                },
                error: function (error) {
                    bootoast({
                        message: error.msg || '接口请求失败，请联系管理员',
                        type: 'error',
                        position: 'top-center',
                        timeout: 3
                    });
                    $('#myModal').modal('hide');
                }
            });
        } else {
            bootoast({
                message: '名称必填',
                type: 'warning',
                position: 'top-center',
                timeout: 3
            });
        }
    })
    setInterval(function () {
        createEel();
    }, 100)
    setInterval(function () {
        getData()
    }, 500)

    function autoinstall() {
        $('#jsoneditor').html(null);
        $('#dataName')[0].value = null;

        function sleep(time) {
            return new Promise((resolve) => setTimeout(resolve, time));
        }

        sleep(10).then(() => {
            let res3 = obj.valueData
            container = document.getElementById("jsoneditor")
            options = {theme: 'bootstrap2'}
            editor = new JSONEditor(container, options)
            editor.set(res3.values)
            $('#myModal').modal('show');
        });
    }

    function createEel() {
        if (window.location.href.indexOf("/packages/helm") === -1) {
            return;
        }
        if (width >= 768) {
            let root = $(".PackageView_rightColumnWrapper__117TK .d-md-block")[0]
            if (root === undefined) {
                return
            }
            if ($(".PackageView_rightColumnWrapper__117TK .d-md-block #autoinstall1").length !== 0) {
                return
            }
            let btn = document.createElement("BUTTON")
            btn.innerText = "自动安装"
            btn.setAttribute("id", "autoinstall1");
            btn.setAttribute('class', 'installBtn');
            btn.setAttribute("data-toggle", "modal");
            btn.setAttribute('data-target', '#myModal');
            root.prepend(btn)
        } else {
            let root = $(".PackageView_jumbotron__2yiPH .position-relative .row")[0]
            if (root === undefined) {
                return
            }
            if ($(".PackageView_jumbotron__2yiPH .position-relative .row #autoinstall2").length !== 0) {
                return
            }
            let btn = document.createElement("BUTTON")
            btn.innerText = "自动安装"
            btn.setAttribute("id", "autoinstall2");
            btn.setAttribute('class', 'col mt-3 PackageView_btnMobileWrapper__2KdxQ installBtn');
            btn.setAttribute("data-toggle", "modal");
            btn.setAttribute('data-target', '#myModal');
            root.prepend(btn)
        }


        let dialog = document.createElement("div")
        dialog.setAttribute("class", "pri modal fade");
        dialog.setAttribute('id', 'myModal');
        dialog.setAttribute("tabindex", "-1");
        dialog.setAttribute('role', 'dialog');
        dialog.setAttribute("aria-labelledby", "myModalLabel");
        dialog.setAttribute('aria-hidden', 'true');
        $('body')[0].prepend(dialog);
        $('#myModal').html('<div class="modal-dialog" style="max-width: 800px">\n' +
            '        <div class="modal-content">\n' +
            '            <div class="modal-header">\n' +
            '                <h4 class="modal-title" id="myModalLabel">安装配置</h4>\n' +
            '                <button type="button" class="close" data-dismiss="modal" aria-hidden="true">&times;</button>\n' +
            '            </div>\n' +
            '            <div class="modal-body">' +
            '               <form class="form-horizontal">\n' +
            '                          <div class="form-group">\n' +
            '                              <label for="username" class="col-sm-2 control-label">名称</label>\n' +
            '                              <div class="col-sm-6">\n' +
            '                                  <em style="color: red;">*</em>\n' +
            '                                  <input type="text" class="form-control" id="dataName" name="dataName" placeholder="请输入名称">\n' +
            '                              </div>\n' +
            '                          </div>\n' +
            '                   </form>' +
            '               <p class="myTips">下面的是默认的values值，可以根据需要修改，不懂不要乱改</p>' +
            '               <div id="jsoneditor" style="width: 100%; height: 400px;"></div>' +
            '            </div>\n' +
            '            <div class="modal-footer">\n' +
            '                <button type="button" class="btn btn-default marginBotm" data-dismiss="modal">关闭</button>\n' +
            '                <button type="button" class="btn btn-primary" id="submitBtn">提交</button>\n' +
            '            </div>\n' +
            '        </div>\n' +
            '    </div>')
    }

    function getData() {
        let pathName = window.location.pathname
        let url = window.location.origin + '/api/v1' + pathName;
        if (pathName.indexOf("packages/helm") === -1) {
            return
        }
        if ((obj.apiData !== null && pathName.indexOf(obj.apiData.version) > -1) || isNewVersion()) {
            return
        }
        $.get(url, function (res) {
            obj.apiData = res
            let url2 = window.location.origin + '/api/v1/packages/' + res.package_id + '/' + res.version + '/templates'
            $.get(url2, function (res2) {
                obj.valueData = res2
            })
        })
    }

    function isNewVersion() {
        if (obj.apiData == null) {
            return false
        }
        let currentVersion = obj.apiData.version
        let version = ""
        let ts = 0
        for (let i = 0; i < obj.apiData["available_versions"].length; i++) {
            if (obj.apiData["available_versions"][i]["ts"] > ts) {
                version = obj.apiData["available_versions"][i].version
                ts = obj.apiData["available_versions"][i]["ts"]
            }
        }
        if (currentVersion === version && window.location.pathname.indexOf(obj.apiData.name) > -1 && window.location.pathname.indexOf(obj.apiData.version) > -1) {
            return true
        }
    }

});