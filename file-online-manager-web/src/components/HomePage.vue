<template>
  <div class="home-page" @click="handlePageClick">
    <el-container style="height: 100%;">
      <el-aside width="400px">
        <el-header height="30px">
          <el-button type="primary" @click="uploadFile" class="fm-btn">上传</el-button>
          <el-button type="primary" @click="createFolder" class="fm-btn">创建</el-button>
<!--          <div style="height: 30px; padding: 0px 0; line-height: 30px;"><span class="el-icon-coordinate">当前位置：
                        </span><span class="current-path">{{ currentPath }}</span></div>-->
        </el-header>
        <el-main>
          <el-tree :data="treeData" @node-click="handleNodeClick" :props="defaultProps" :load="loadNode"
                   ref="directoryTree" node-key="id" lazy @node-contextmenu="handleContextMenu">
            <span class="custom-node" slot-scope="{ node }">{{ node.label }}</span>
          </el-tree>
        </el-main>
      </el-aside>
      <el-main style="">
        <div id="table-container">
          <el-input placeholder="请输入名称" prefix-icon="el-icon-search" v-model="searchKey"></el-input>
          <div class="full-path-nav" :title="currentPath"><span class="el-icon-coordinate">当前位置： </span> {{currentPath}}</div>
          <el-table
              :data="tableData.filter(data => !searchKey || data.name.toLowerCase().includes(searchKey.toLowerCase()))"
              style="width: 100%" resizable border>
            <el-table-column prop="name" label="名称" width="300"></el-table-column>
            <el-table-column prop="size" label="大小" width="160" :formatter="fileSizeFormat"></el-table-column>
            <el-table-column prop="type" label="类型" width="60">
              <template slot-scope="scope">
                <i class="el-icon-folder folder" v-if="scope.row.isDir"></i>
                <i class="el-icon-document" v-if="!scope.row.isDir"></i>
              </template>
            </el-table-column>
            <el-table-column sortable prop="modTime" label="修改时间" min-width="100" max-width="400"></el-table-column>
            <el-table-column sortable prop="mode" label="权限" width="160"></el-table-column>
            <el-table-column prop="gid" label="GID" width="60"></el-table-column>
            <el-table-column prop="gid" label="UID" width="60"></el-table-column>
            <el-table-column label="操作" align="left">
              <template slot-scope="scope">
                <el-button type="primary" size="small" icon="el-icon-edit"
                           @click="renameFile(scope.row)"
                           title="重命名"></el-button>
                <el-button v-if="checkFileType(scope.row.name)" size="small" type="primary"
                           @click="unzipFile(scope.row)" title="解压"
                           style="width: 44px; height: 32px">
                  <img src="@/assets/unzip.png" alt="编辑" style="height: 14px;width: 14px;vertical-align: middle;">
                </el-button>
                <el-dropdown @command="(command) => handleCommand(command, scope.row)" style="padding: 0px 8px">
                  <el-button type="primary" size="small">
                    更多<i class="el-icon-arrow-down el-icon--right"></i>
                  </el-button>
                  <el-dropdown-menu slot="dropdown">
                    <el-dropdown-item command="delete">删除</el-dropdown-item>
                    <el-dropdown-item command="copy">复制</el-dropdown-item>
                    <el-dropdown-item command="zip">压缩</el-dropdown-item>
                    <el-dropdown-item command="download">下载</el-dropdown-item>
                    <el-dropdown-item command="onlineEdit">在线编辑</el-dropdown-item>
                    <el-dropdown-item command="viewZipFile">查看ZIP文件</el-dropdown-item>
                  </el-dropdown-menu>
                </el-dropdown>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-main>
    </el-container>
    <el-dialog title="上传文件" :visible.sync="dialogVisible" v-if="dialogVisible" width="930px" :before-close="handleClose" destroy-on-close
               :close-on-click-modal="false">
      <large-file-upload :currentPath="currentPath"></large-file-upload>
      <span slot="footer" class="dialog-footer">
                 <el-button type="primary" @click="uploadOk">确 定</el-button>
            </span>
    </el-dialog>
    <el-dialog title="Zip文件查看" :visible.sync="zipViewDialogVisible" v-if="zipViewDialogVisible" width="930px" :before-close="handleClose"
               destroy-on-close :close-on-click-modal="false">
      <zip-file-view :currentPath="currentViewZipPath" ref="zipView"></zip-file-view>
      <span slot="footer" class="dialog-footer">
              <el-button type="primary" @click="releasePatch">释放补丁</el-button>
            </span>
    </el-dialog>
    <el-dialog ref="viewDialog" title="文本文件查看" :visible.sync="textViewDialogVisible" v-if="textViewDialogVisible" width="930px" :before-close="handleClose"
                 :close-on-click-modal="false" @open="beforeViewOpen" :fullscreen.sync="isMaximized" class="text-view-dialog">
      <text-editor :currentPath="currentViewZipPath" ref="textView"></text-editor>
    </el-dialog>
  </div>
</template>

<script>
import LargeFileUpload from './LargeFileUpload.vue';
import ZipFileView from './ViewZipFile.vue'
import TextEditor from "@/components/TextEditor";
import ContextMenu from '@/components/ContextMenu.vue';

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
      dialogVisible: false,
      zipViewDialogVisible: false,
      textViewDialogVisible: false,
      searchKey: "",
      loading: null,
      currentViewZipPath: '',
      isMaximized: true,
      menuItems: [
        { label: '变更用户权限', action: 'chown', icon: 'el-icon-user' },
        { label: '更改RWX权限', action: 'chmod', icon: 'el-icon-brush' },
      ],
      showMenu: false, // 是否显示右键菜单
      menuX: 0, // 右键菜单横坐标
      menuY: 0, // 右键菜单纵坐标
    }
  },
  props: {},
  components: {
    "large-file-upload": LargeFileUpload,
    "zip-file-view": ZipFileView,
    "text-editor": TextEditor,
  },
  mounted() {
    document.title = '文件管理工具';
    this.listFile('')
  },
  methods: {
    handleCommand(command, row) {
      switch (command) {
        case 'delete':
          this.deleteFile(row);
          break;
        case 'copy':
          this.copyFile(row);
          break;
        case 'zip':
          this.zipFolder(row);
          break;
        case 'download':
          this.downloadFiles(row);
          break
        case 'onlineEdit':
          this.onlineEdit(row);
          break
        case 'viewZipFile':
          this.viewZipFile(row);
          break
      }
    },
    loadNode(node, resolve) {
      let path = node.data.path
      if (!path) {
        path = ''
      }
      this.$http.get('./api/manager/folder/list?path=' + path, {}).then(response => {
        if (path == '') {
          let root = {id: '-1', name: 'root', path: '.', children: []};
          root.children = response.data.data;
          resolve([root])
        } else {
          resolve(response.data.data)
        }

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
      this.$confirm("是否删除文件(夹)：" + row.name, "确认").then(function () {
        $this.$http.delete('./api/manager/file/delete?path=' + row.path, {
          path: row.path
        }).then(response => {
          console.log(response)
          $this.listFile($this.currentPath)
        }, response => {
          console.log(response)
          $this.$alert(response.message, '错误', {
            confirmButtonText: '确定',
            type: 'error'
          })
        })
      }).catch(function () {
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
          $this.$alert(response.response.data.message, '错误', {
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
    },
    /**
     * 上传文件框关闭图标
     * @param done 关闭调用
     */
    handleClose(done) {
      this.listFile(this.currentPath);
      done();
    },
    beforeViewOpen() {
      if (this.$refs.textView) {
        this.$refs.textView.loadContent(this.currentViewZipPath)
      }
    },
    /**
     * 检测文件类型
     * @param fileName {String} 文件名称
     * @param fileType {Array} 文件类型
     * @returns {boolean} true：是指定类型，否：不是指定类型
     */
    checkFileType(fileName, fileType) {
      if (!fileType) {
        fileType = ['.zip', '.gz', '.tar'];
      }
      const fileExtension = fileName.slice(fileName.lastIndexOf('.')).toLowerCase();
      return fileType.includes(fileExtension);
    },
    /**
     *
     * @param row
     */
    downloadFiles(row) {
      window.open(window.location.href + "api/manager/file/download?filename=" + row.name + "&path=" + this.currentPath, "_blank")
    },
    /**
     * 上传文件框确定
     */
    uploadOk() {
      this.listFile(this.currentPath);
      this.dialogVisible = false;
    },
    releasePatch() {
      this.$refs.zipView.releaseZipFile()
    },
    /**
     * 解压文件
     * @param row 文件信息
     */
    unzipFile(row) {
      let $this = this;
      $this.loading = $this.$loading({
        lock: true,
        text: '解压中',
        spinner: 'el-icon-loading',
        background: 'rgba(0, 0, 0, 0.7)'
      });
      $this.$http.post('./api/manager/file/unzip', {
        path: row.path,
        name: row.name
      }).then((response) => {
        console.log(response)
        var currentNode = $this.$refs.directoryTree.getCurrentNode();
        $this.loadNode({
          data: currentNode
        }, (data) => {
          $this.$refs.directoryTree.updateKeyChildren(currentNode.id, data);
        });
        $this.listFile($this.currentPath);
        $this.$message({
          message: "解压成功",
          type: 'success'
        });
      }).catch(error => {
        console.log(error)
        $this.$alert(error.response.data.message, '错误', {
          confirmButtonText: '确定',
          type: 'error'
        })
      }).finally(() => {
        if ($this.loading) {
          $this.loading.close(); // 关闭并销毁 loading
          $this.loading = null; // 重置 loading 引用
        }
      })
    },
    /**
     * 压缩文件夹
     * @param row 文件信息
     */
    zipFolder(row) {
      let $this = this;
      $this.loading = $this.$loading({
        lock: true,
        text: '压缩中',
        spinner: 'el-icon-loading',
        background: 'rgba(0, 0, 0, 0.7)'
      });
      $this.$http.post('./api/manager/folder/zip', {
        path: row.path
      }).then(() => {
        $this.listFile($this.currentPath)
      }).catch(error => {
        $this.$alert(error.response.data.message, '错误', {
          confirmButtonText: '确定',
          type: 'error'
        })
        if ($this.loading) {
          $this.loading.close(); // 关闭并销毁 loading
          $this.loading = null; // 重置 loading 引用
        }
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
      let postCreate = function (path) {
        return new Promise(function (resolve, reject) {
          $this.$http.post('./api/manager/folder/create?path=' + path, {}).then(response => {
            console.log(response.body)
            $this.listFolder()
            $this.listFile($this.currentPath)
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
        let root = {id: '-1', name: 'root', path: '.'}
        root.children = response.data.data
        $this.treeData = [root]
      }, response => {
        console.log(response.body)
      })
    },
    listFile(path) {
      let $this = this
      this.$http.get('./api/manager/file/list?path=' + path, {}).then(response => {
        $this.tableData = response.data.data
        if ($this.loading) {
          $this.loading.close(); // 关闭并销毁 loading
          $this.loading = null; // 重置 loading 引用
        }
      }, response => {
        console.log(response.response)
        $this.tableData = []
        if ($this.loading) {
          $this.loading.close(); // 关闭并销毁 loading
          $this.loading = null; // 重置 loading 引用
        }
        this.$message.warning("加载失败，请刷新文件夹列表后重试")
      })
    },
    onlineEdit(row) {
      debugger
      // this.$message.warning("当前类型不支持：" + row.name)
      this.currentViewZipPath = row.path
      this.textViewDialogVisible = true
    },
    fileSizeFormat(row) {
      let size = row.size
      if (size == -1) {
        return "-";
      } else if (size < 1024) {
        return size + " B";
      } else if (size < 1024 * 1024) {
        return (size / 1024).toFixed(2) + " KB";
      } else if (size < 1024 * 1024 * 1024) {
        return (size / (1024 * 1024)).toFixed(2) + " MB";
      } else {
        return (size / (1024 * 1024 * 1024)).toFixed(2) + " GB";
      }
    },
    viewZipFile(row) {
      this.currentViewZipPath = row.path
      this.zipViewDialogVisible = true
    },
    handleContextMenu(event, data, node) {
      event.preventDefault(); // 阻止默认的右键菜单事件
      this.menuX = event.clientX;
      this.menuY = event.clientY;
      this.showMenu = true;
    },
    handleMenuClick(item) {
      this.showMenu = false
      switch (item.action) {
        case 'copy':
          // 复制操作
          break;
        case 'paste':
          // 粘贴操作
          break;
        case 'delete':
          // 删除操作
          break;
        default:
          break;
      }
    },
    handlePageClick(event) {
      // 点击页面其他地方时，隐藏右键菜单
      if (this.showMenu) {
        this.showMenu = false;
      }
    },
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style>
.current-path {
  color: chocolate;
}

.el-main {
  margin-left: 20px;
  padding-top: 10px;
}

.el-icon-folder {
  color: #0c23c9;
}

.el-icon-document {
  color: burlywood;
}

.text-view-dialog .el-dialog__body {
  height: calc(100% - 120px);
}
.full-path-nav {
  height: 28px;
  line-height: 28px;
  color: chocolate;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.fm-btn,.el-button {
  padding: 8px 12px;
  font-size: 12px;
}
</style>
