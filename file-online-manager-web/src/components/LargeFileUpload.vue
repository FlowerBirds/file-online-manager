<template>
    <div class="large-file-upload">
        <uploader ref="uploader" :options="options" :autoStart="true" :file-status-text="fileStatusText"
            @file-added="onFileAdded" @file-success="onFileSuccess" @file-error="onFileError"
            @file-progress="onFileProgress" class="uploader-example">
            <uploader-unsupport></uploader-unsupport>
            <uploader-drop>
                <p>拖动文件到这里上传</p>
                <uploader-btn>选择文件</uploader-btn>
                <uploader-btn :directory="true">选择文件夹</uploader-btn>
            </uploader-drop>
            <!-- uploader-list可自定义样式 -->
            <!-- <uploader-list></uploader-list> -->
            <uploader-list>
                <div class="file-panel" :class="{ collapse: collapse }">
                    <div class="file-title">
                        <p class="file-list-title">文件列表</p>
                        <div class="operate">
                            <el-button type="text" @click="operate" :title="collapse ? '折叠' : '展开'">
                                <i class="icon" :class="collapse ? 'el-icon-caret-bottom' : 'el-icon-caret-top'"></i>
                            </el-button>
                        </div>
                    </div>

                    <ul class="file-list" :class="
              collapse ? 'uploader-list-ul-show' : 'uploader-list-ul-hidden'
            ">
                        <li v-for="file in uploadFileList" :key="file.id">
                            <uploader-file :class="'file_' + file.id" ref="files" :file="file"
                                :list="true"></uploader-file>
                        </li>
                        <div class="no-file" v-if="!uploadFileList.length">
                            <i class="icon icon-empty-file"></i> 暂无待上传文件
                        </div>
                    </ul>
                </div>
            </uploader-list>
        </uploader>
    </div>
</template>

<script>
    import { v4 as uuidv4 } from 'uuid';
    // 大文件 2G 会有提示
    const LARGE_FILE_SIZE = 2 * 1024 * 1024 * 1024;
    // 分片大小，30MB
    const CHUNK_SIZE = 30 * 1024 * 1024;
    export default {
        name: 'TestComponent',
        data() {
            return {
                options: {
                    // 上传地址
                    target: "http://localhost:8081/api/manager/file/upload1?path=" + this.currentPath,
                    // 是否开启服务器分片校验。默认为 true
                    testChunks: true,
                    // 真正上传的时候使用的 HTTP 方法,默认 POST
                    uploadMethod: "post",
                    // 分片大小
                    chunkSize: CHUNK_SIZE,
                    // 并发上传数，默认为 3
                    simultaneousUploads: 3,
                    /**
                     * 判断分片是否上传，秒传和断点续传基于此方法  暂时没有此功能，此功能需要数据库
                     * 这里根据实际业务来 用来判断哪些片已经上传过了 不用再重复上传了 [这里可以用来写断点续传！！！]
                     */
                    checkChunkUploadedByResponse: () => {
                        // console.log(chunk, message)
                        // message是后台返回
                        // let messageObj = JSON.parse(message);
                        // let dataObj = messageObj.data;
                        // if (dataObj.uploaded !== undefined) {
                        //     return dataObj.uploaded;
                        // }
                        // 判断文件或分片是否已上传，已上传返回 true
                        // 这里的 uploadedChunks 是后台返回]
                        // return (dataObj.uploadedChunks || []).indexOf(chunk.offset + 1) >= 0;
                        // 返回false代表需要上传
                        return false;
                    },
                    parseTimeRemaining: function(timeRemaining, parsedTimeRemaining) {
                        //格式化时间
                        return parsedTimeRemaining
                            .replace(/\syears?/, "年")
                            .replace(/\days?/, "天")
                            .replace(/\shours?/, "小时")
                            .replace(/\sminutes?/, "分钟")
                            .replace(/\sseconds?/, "秒");
                    },
                    generateUniqueIdentifier: function() {
                      return uuidv4();
                    }
                },
                // 修改上传状态
                fileStatusTextObj: {
                    success: "上传成功",
                    error: "上传错误",
                    uploading: "正在上传",
                    paused: "停止上传",
                    waiting: "等待中",
                },
                uploadIdInfo: null,
                uploadFileList: [],
                fileChunkList: [],
                collapse: true,
            };
        },
        props: {
            currentPath: {
                type: String,
                default: '.'
            }
        },
        methods: {
            onFileAdded(file) {
                this.uploadFileList.push(file);
                // 文件大小提示
                if (file.size > LARGE_FILE_SIZE) {
                    this.$message({
                        type: 'info',
                        message: '文件较大请耐心等待'
                    });
                }
                this.$nextTick(() => {
                  file.resume();
                });
            },
            onFileSuccess() {
                console.log("上传成功");
                // console.log(rootFile, file, response, chunk)
            },
            onFileError(rootFile, file, message, chunk) {
                console.log("上传出错：" + message);
                console.log(rootFile, file, message, chunk)
            },
            onFileProgress() {
                // console.log(rootFile, file, chunk)
                // console.log(`当前进度：${Math.ceil(file._prevProgress * 100)}%`);
            },
            fileStatusText(status) {
                // console.log(response)
                if (status === "md5") {
                    return "校验MD5";
                } else {
                    return this.fileStatusTextObj[status];
                }
            },
            /**
             * 折叠、展开面板动态切换
             */
            operate() {
                if (this.collapse === false) {
                    this.collapse = true;
                } else {
                    this.collapse = false;
                }
            }
        }
    };
</script>

<style scoped>
    .logo {
        font-family: "Avenir", Helvetica, Arial, sans-serif;
        -webkit-font-smoothing: antialiased;
        -moz-osx-font-smoothing: grayscale;
        text-align: center;
        color: #2c3e50;
        margin-top: 60px;
    }

    .uploader-example {
        width: 880px;
        padding: 0px 15px;
        margin: 0px auto;
        font-size: 12px;
    }

    .uploader-example .uploader-btn {
        margin-right: 4px;
    }

    .uploader-example .uploader-list {
        max-height: 440px;
        overflow: auto;
        overflow-x: hidden;
        overflow-y: auto;
    }

    #global-uploader {
        position: fixed;
        z-index: 20;
        right: 15px;
        bottom: 15px;
        width: 550px;
    }

    .file-panel {
        background-color: #fff;
        border: 1px solid #e2e2e2;
        border-radius: 7px 7px 0 0;
        box-shadow: 0 0 10px rgba(0, 0, 0, 0.2);
    }

    .file-title {
        display: flex;
        height: 60px;
        line-height: 30px;
        padding: 0 15px;
        border-bottom: 1px solid #ddd;
    }

    .file-title {
        background-color: #e7ecf2;
    }

    .uploader-file-meta {
        display: none !important;
    }

    .operate {
        flex: 1;
        text-align: right;
    }

    .file-list {
        position: relative;
        height: 240px;
        overflow-x: hidden;
        overflow-y: auto;
        background-color: #fff;
        padding: 0px;
        margin: 0 auto;
        transition: all 0.5s;
    }

    .uploader-file-size {
        width: 15% !important;
    }

    .uploader-file-status {
        width: 32.5% !important;
        text-align: center !important;
    }

    li {
        background-color: #fff;
        list-style-type: none;
    }

    .no-file {
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        font-size: 16px;
    }

    /* 隐藏上传按钮 */
    .global-uploader-btn {
        display: none !important;
        clip: rect(0, 0, 0, 0);
        /* width: 100px;
  height: 50px; */
    }

    .file-list-title {
        /*line-height: 10px;*/
        font-size: 16px;
    }

    .uploader-file-name {
        width: 36% !important;
    }

    .uploader-file-actions {
        float: right !important;
    }

    .uploader-list-ul-hidden {
        height: 0px;

    }
</style>