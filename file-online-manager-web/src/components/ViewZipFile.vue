<template>
  <div class="zip-file-view">
    <el-input type="text" v-model="searchText" placeholder="搜索文件名"/>
    <el-table :data="filteredFileList" height="500px">
      <el-table-column prop="name" label="文件名" width="400px"></el-table-column>
      <el-table-column prop="size" label="压缩后大小"></el-table-column>
      <el-table-column prop="modTime" label="修改时间"></el-table-column>
      <el-table-column label="状态">
        <template slot-scope="scope">
          <i v-if="scope.row.status" class="el-icon-circle-check" style="color: green;"></i>
          <i v-if="!scope.row.status" class="el-icon-remove-outline" style="color: gray;"></i>
        </template>
      </el-table-column>
    </el-table>
    <div class="overlay" v-if="showOverlay">
      <div class="spinner"></div>
    </div>
  </div>
</template>

<style>
.overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 100;
}

.spinner {
  border: 16px solid #f3f3f3;
  border-top: 16px solid #3498db;
  border-radius: 50%;
  width: 60px;
  height: 60px;
  animation: spin 2s linear infinite;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}
</style>
<script>
export default {
  name: 'ZipFileView',
  data() {
    return {
      searchText: '',
      fileListTableData: [],
      showOverlay: false,
      isPatchZip: false
    }
  },
  props: {
    currentPath: {
      type: String,
      default: '.'
    }
  },
  computed: {
    filteredFileList() {
      const searchText = this.searchText.trim().toLowerCase();
      if (!searchText) {
        return this.fileListTableData;
      }
      return this.fileListTableData.filter(file => file.name.toLowerCase().includes(searchText));
    }
  },
  mounted() {
    let $this = this;
    $this.$http.get('./api/manager/file/zip/view?path=' + $this.currentPath)
        .then(response => {
          // console.log(response)
          this.fileListTableData = response.data.data.map(f => {
            f.status = false
            return f
          })
        }, response => {
          console.log(response)
          this.fileListTableData = []
          $this.$alert(response.message, '错误', {
            confirmButtonText: '确定',
            type: 'error'
          });

        })
  },
  methods: {
    releaseZipFile: function () {
      let $this = this;
      let files = this.fileListTableData.filter(file => file.name.split("/").length <= 2).map(x => x.name);
      if (files.indexOf("delete.conf") > -1 && files.indexOf("apps/") > -1) {
        this.showOverlay = true; // 显示遮罩层
        $this.$http.get('./api/manager/file/zip/release?path=' + $this.currentPath)
            .then(response => {
              // console.log(response.data)
              this.fileListTableData = this.fileListTableData.map(f => {
                f.status = true
                return f
              })
              console.log("delete file list: ")
              for (let i in response.data.data) {
                console.log(response.data.data[i].name)
              }
            }, data => {
              console.log(data)
              $this.$alert(data.response.data.message, '错误', {
                confirmButtonText: '确定',
                type: 'error'
              });

            }).finally(() => {
          this.showOverlay = false; // 显示遮罩层
        })
      } else {
        this.$alert("没有检测到deleted.conf和apps")
      }
    }
  }
}
</script>