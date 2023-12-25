<template>
<div>
  <div>
    <label>命名空间：</label>
    <el-select v-model="selectedNamespace" placeholder="请选择" @change="changeNamespace">
      <el-option
          v-for="option in options"
          :key="option.value"
          :label="option.label"
          :value="option.value"
      ></el-option>
    </el-select>
    <el-button type="primary" size="small" icon="el-icon-refresh" class="refresh-btn"
               @click="refresh()"
               title="刷新"></el-button>
  </div>
  <div>
    <el-table :data="tableData">
      <el-table-column prop="name" label="名称" width="440"></el-table-column>
      <el-table-column prop="ready" label="实例数" width="80"></el-table-column>
      <el-table-column prop="status" label="状态" ></el-table-column>
      <el-table-column prop="restarts" label="重启次数"></el-table-column>
      <el-table-column prop="age" label="运行时长"></el-table-column>
      <el-table-column prop="ip" label="IP地址"></el-table-column>
      <el-table-column prop="node" label="所在节点"></el-table-column>
      <el-table-column label="操作" align="left">
        <template slot-scope="scope">
          <el-button type="primary" size="small" icon="el-icon-refresh-right"
                     @click="restartPod(scope.row)"
                     title="重启"></el-button>
          <el-dropdown @command="(command) => handleCommand(command, scope.row)" style="padding: 0px 8px">
            <el-button type="primary" size="small">
              更多<i class="el-icon-arrow-down el-icon--right"></i>
            </el-button>
            <el-dropdown-menu slot="dropdown">
              <el-dropdown-item command="viewYaml">查看YAML</el-dropdown-item>
              <el-dropdown-item command="viewLogs">查看日期</el-dropdown-item>
            </el-dropdown-menu>
          </el-dropdown>
        </template>
      </el-table-column>
    </el-table>
  </div>
  <el-dialog ref="viewLogDialog" :title="logTitle" :visible.sync="logViewDialogVisible" v-if="logViewDialogVisible" width="930px" height="600px" :before-close="handleClose"
             :close-on-click-modal="false" @open="beforeLogViewOpen" :fullscreen.sync="isMaximized" class="text-view-dialog">
    <pre style="height: 600px; overflow-y: scroll; overflow-x: auto;  white-space: pre-line; word-wrap: break-word; word-break: break-all;">{{ logStreamData }}</pre>
  </el-dialog>
</div>
</template>

<script>
import axios from 'axios';

export default {
  name: "K8sServicePage",
  data() {
    return {
      logViewDialogVisible: false,
      isMaximized: false,
      selectedNamespace: 'default', // 默认选中的值
      logStreamData: "",
      logEventSource: null,
      logTitle: "查看日志",
      options: [
        { value: 'default', label: 'default' },
        { value: 'tempo611', label: 'tempo611' },
        { value: 'openfaas', label: 'openfaas' }
      ],
      tableData: [
        {
          "name": "file-manage-df4856c55-n2b2r",
          "ready": "1/1",
          "status": "Running",
          "restarts": "24 (3h34m ago)",
          "age": "84d",
          "ip": "10.42.0.248",
          "node": "laptop-tc4a0scv",
          "nominatedNode": "<none>",
          "readinessGates": "<none>"
        }
      ]
    };
  },
  mounted() {
    //
  },
  methods: {
    init() {
      this.listNamespace()
      this.listPods()
    },
    restartPod(row) {
      let $this = this;
      const formData = new FormData();
      formData.append('namespace', this.selectedNamespace);
      formData.append('name', row.name);
      this.$confirm("是否重启该pod：" + row.name, "确认").then(function () {
        $this.$http.postForm('./api/manager/k8s/restart-pod', formData, {
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
          }
        }).then(response => {
          console.log(response)
          $this.listPods()
        }, response => {
          console.log(response)
          $this.$alert(response.message, '错误', {
            confirmButtonText: '确定',
            type: 'error'
          })
        })
      });
    },
    handleClose(done) {
      if (this.logEventSource) {
        this.logEventSource.close()
      }
      done();
    },
    handleCommand(command, row) {
      switch (command) {
        case 'viewYaml':
          this.viewPodYaml(row);
          break;
        case 'viewLogs':
          this.viewPodLogs(row);
          break;
      }
    },
    viewPodYaml(row) {

    },
    beforeLogViewOpen() {

    },
    viewPodLogs(row) {
      this.logViewDialogVisible = true
      this.logStreamData = ""
      this.logTitle = row.name + " 日志"
      this.logEventSource = new EventSource('./api/manager/k8s/pod-stream-logs?name=' + row.name + "&namespace=" + this.selectedNamespace); // SSE 服务端的 URL

      this.logEventSource.addEventListener('message', this.handleLogMessage);
      this.logEventSource.addEventListener('error', this.handleLogError);

    },
    handleLogMessage(event) {
      this.logStreamData += event.data + "\n"
    },
    handleLogError(event) {
      console.error('Error occurred:', event);
    },
    listPods() {
      const formData = new FormData();
      formData.append('namespace', this.selectedNamespace);
      this.$http.post('./api/manager/k8s/list-pods', formData, {
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded'
        }
      }).then(response => {
        console.log(response)
        this.tableData = response.data.data
      }, response => {
        console.log(response)
        this.$alert(response.data.message, '错误', {
          confirmButtonText: '确定',
          type: 'error'
        })
      })
    },
    refresh() {
      this.listPods()
    },
    listNamespace() {
      this.$http.post('./api/manager/k8s/list-namespace', {}, {
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded'
        }
      }).then(response => {
        console.log(response)
        let namespace = []
        for (let i in response.data.data) {
          let ns = response.data.data[i]
          namespace.push({value: ns.name, label: ns.name})
        }
        this.options = namespace
      }, response => {
        console.log(response)
        this.$alert(response.data.message, '错误', {
          confirmButtonText: '确定',
          type: 'error'
        })
      })
    },
    changeNamespace(data) {
      console.log(data)
      this.listPods()
    }
  }
}
</script>

<style scoped>
.refresh-btn {
  margin-left: 5px;
}
</style>