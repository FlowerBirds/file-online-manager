<template>
    <div class="home-page">
        <el-container style="height: 100%;">
            <el-aside width="400px">
                <el-header height="50px">
                    <el-button type="primary" @click="uploadFile">上传</el-button>
                    <el-button type="primary" @click="createFolder">创建</el-button>
                    <div style="height: 30px; padding: 0px 0; line-height: 30px;"><span class="el-icon-coordinate">当前位置：
                        </span><span class="current-path">{{ currentPath }}</span></div>
                </el-header>
                <el-main>
                    <el-tree :data="treeData" @node-click="handleNodeClick" :props="defaultProps" :load="loadNode"
                        ref="directoryTree" node-key="id" lazy></el-tree>
                </el-main>
            </el-aside>
            <el-main style="">
              <div id="table-container">
                <el-table :data="tableData" style="width: 100%">
                    <el-table-column prop="name" label="名称" width="300"></el-table-column>
                    <el-table-column prop="size" label="大小"></el-table-column>
                    <el-table-column prop="type" label="类型">
                        <template slot-scope="scope">
                            <i class="el-icon-folder" v-if="scope.row.isDir"></i>
                            <i class="el-icon-document" v-if="!scope.row.isDir"></i>
                        </template>
                    </el-table-column>
                    <el-table-column prop="modTime" label="修改时间" ></el-table-column>
                    <el-table-column label="操作" width="350">
                        <template slot-scope="scope">
                            <el-button type="primary" size="small" icon="el-icon-delete" @click="deleteFile(scope.row)"
                                title="删除"></el-button>
                            <el-button type="primary" size="small" icon="el-icon-edit" @click="renameFile(scope.row)"
                                title="重命名"></el-button>
                            <el-button type="primary" size="small" icon="el-icon-document-copy"
                                @click="copyFile(scope.row)" title="复制"></el-button>
                            <el-button v-if="checkFileType(scope.row.name)" type="primary" size="small"
                                icon="el-icon-grape" @click="unzipFile(scope.row)" title="解压"></el-button>
                        </template>
                    </el-table-column>
                </el-table>
              </div>
            </el-main>
        </el-container>
        <el-dialog title="上传文件" :visible.sync="dialogVisible" width="930px" :before-close="handleClose">
            <LargeFileUpload :currentPath="currentPath"></LargeFileUpload>
            <span slot="footer" class="dialog-footer">
                <el-button @click="dialogVisible = false">取 消</el-button>
                <el-button type="primary" @click="dialogVisible = false">确 定</el-button>
            </span>
        </el-dialog>
    </div>
</template>

<style>
#table-container {
  height: 100%;
}
</style>

<script>
    import LargeFileUpload from './LargeFileUpload.vue';

    export default {
        name: 'HomePage',
        data() {
            return {
                treeData: [],
                tableData: [],
                defaultProps: {
                    label: 'name'
                },
                currentPath: '.',
                dialogVisible: false
            }
        },
        props: {

        },
        components: {
            LargeFileUpload
        },
        mounted() {
            this.listFile('')
        },
        methods: {
            handleClose(done) {
                this.$confirm('确认关闭？')
                    .then(() => {
                        done();
                    })
                    .catch(() => {});
            },
            loadNode(node, resolve) {
                let path = node.data.path
                if (!path) {
                    path = ''
                }
                this.$http.get('./api/manager/folder/list?path=' + path, {}).then(response => {
                    resolve(response.data.data)
                }, response => {
                    console.log(response.body)
                })
            },
            handleNodeClick(data) {
                this.currentPath = data.path
                this.listFile(data.path)
            },
            /**
             * 删除文件
             * @param row
             */
            deleteFile(row) {
                let $this = this;
                this.$confirm("是否删除文件：" + row.name, "确认").then(function() {
                    $this.$http.delete('./api/manager/file/delete?path=' + row.path, {
                        path: row.path
                    }).then(response => {
                        console.log(response.body)
                        $this.listFile($this.currentPath)
                    }, response => {
                        console.log(response.body)
                        $this.$alert(response.body.message, '错误', {
                            confirmButtonText: '确定',
                            type: 'error'
                        })
                    })
                })
            },
            renameFile(row) {
                this.$prompt('请确认文件名称', '提示', {
                    inputValue: row.name,
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                }).then(({
                    value
                }) => {
                    this.$http.post('./api/manager/file/rename', {
                        path: row.path,
                        name: value
                    }).then(response => {
                        console.log(response.body)
                        this.listFile(this.currentPath)
                    }, response => {
                        console.log(response.body)
                        this.$alert(response.body.message, '错误', {
                            confirmButtonText: '确定',
                            type: 'error'
                        })
                    })
                }).catch(() => {
                    this.$message({
                        type: 'info',
                        message: '取消重命名'
                    });
                });
            },
            copyFile(row) {
                let $this = this
                this.$prompt('请输入新的文件名称', '提示', {
                    inputValue: row.name,
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                }).then(({
                    value
                }) => {
                    $this.$http.post('./api/manager/file/copy', {
                        path: row.path,
                        name: value
                    }).then(response => {
                        console.log(response.body)
                        $this.listFile(this.currentPath)
                    }, response => {
                        $this.listFile(this.currentPath);
                        $this.$alert(response.body.message, '错误', {
                            confirmButtonText: '确定',
                            type: 'error'
                        })
                    })
                }).catch(() => {
                    this.$message({
                        type: 'info',
                        message: '取消重命名'
                    });
                });
            },
            uploadFile() {
                this.dialogVisible = true;
                // let input = document.createElement('input');
                // input.type = 'file';
                // input.onchange = () => {
                //     let file = input.files[0];
                //     let formData = new FormData();
                //     formData.append('file', file);
                //     formData.append('path', this.currentPath);
                //     let loadingInstance = this.$loading({
                //         lock: true,
                //         text: '文件上传中...',
                //         spinner: 'el-icon-loading',
                //         background: 'rgba(0, 0, 0, 0.7)',
                //         target: document.querySelector('#app')
                //     });
                //     this.$http.post('./api/manager/file/upload', formData).then(response => {
                //         console.log(response.body)
                //         loadingInstance.close();
                //         this.listFile(this.currentPath)
                //     }, response => {
                //         console.log(response.body)
                //         loadingInstance.close();
                //         this.listFile(this.currentPath)
                //     })
                // };
                // input.click();
            },
            /**
             * 检测文件类型
             * @param fileName {String} 文件名称
             * @param fileType {Array} 文件类型
             * @returns {boolean} true：是指定类型，否：不是指定类型
             */
            checkFileType(fileName, fileType) {
                if (!fileType) {
                    fileType = ['.zip', '.tar.gz', '.tar'];
                }
                const fileExtension = fileName.slice(fileName.lastIndexOf('.')).toLowerCase();
                return fileType.includes(fileExtension);
            },
            /**
             * 解压文件
             * @param row 文件信息
             */
            unzipFile(row) {
                let $this = this
                $this.$http.post('./api/manager/file/unzip', {
                    path: row.path,
                    name: row.name
                }).then(() => {
                    var currentNode = $this.$refs.directoryTree.getCurrentNode();
                    $this.loadNode({
                        data: currentNode
                    }, (data) => {
                        $this.$refs.directoryTree.updateKeyChildren(currentNode.id, data);
                    });
                    $this.listFile($this.currentPath)
                }).catch(error => {
                    $this.$alert(error.response.data || error.code, '错误', {
                        confirmButtonText: '确定',
                        type: 'error'
                    })
                })
            },
            /**
             * 删除文件夹，暂无使用
             */
            deleteFolder() {
                let $this = this;
                $this.$http.post('./api/manager/folder/delete', {}).then(() => {
                    var currentNode = $this.$refs.directoryTree.getCurrentNode();
                    $this.loadNode({
                        data: currentNode
                    }, (data) => {
                        $this.$refs.directoryTree.updateKeyChildren(currentNode.id, data);
                    });
                    $this.listFile($this.currentPath)
                }).catch(error => {
                    console.log(error.body)
                })
            },
            renameFolder() {
                this.$http.post('./api/manager/folder/rename', {}).then(response => {
                    console.log(response.body)
                }, response => {
                    console.log(response.body)
                })
            },
            createFolder() {
                let $this = this
                let postCreate = function(path) {
                    return new Promise(function(resolve, reject) {
                        $this.$http.post('./api/manager/folder/create?path=' + path, {}).then(response => {
                            console.log(response.body)
                            $this.listFolder()
                            $this.listFile(this.currentPath)
                            resolve(response.body)
                        }, response => {
                            console.log(response.body)
                            reject(new Error(response.body))
                        })
                    })
                }
                this.$prompt('请输入文件夾名称', '提示', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                }).then(({
                    value
                }) => {
                    let path = this.currentPath + "/" + value
                    return postCreate(path)
                }).catch(() => {
                    this.$message({
                        type: 'info',
                        message: '取消创建'
                    });
                });
            },
            listFolder() {
                let $this = this
                this.$http.get('./api/manager/folder/list', {}).then(response => {
                    $this.treeData = response.data.data
                }, response => {
                    console.log(response.body)
                })
            },
            listFile(path) {
                let $this = this
                this.$http.get('./api/manager/file/list?path=' + path, {}).then(response => {
                    $this.tableData = response.data.data
                }, response => {
                    console.log(response.body)
                })
            }
        }
    }
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
    .current-path {
        color: chocolate;
    }

    .el-main {
        margin-left: 20px;
    }
</style>