<template>
  <div style="height: 100%;">
    <div class="switch-container">
      <el-switch
          v-model="readonlyMode"
          active-text="只读模式"
          inactive-text="编辑模式">
      </el-switch>

      <el-select v-model="selectedValue" @change="onLanguageChange" style="margin-left: 10px;">
        <el-option v-for="option in options" :key="option.value" :label="option.label" :value="option.value">
        </el-option>
      </el-select>
      <el-button class="save-content-btn" v-if="!readonlyMode" @click="saveContent">保存</el-button>
    </div>

    <div ref="editorContainer" style="height: calc(100% - 60px);"></div>
  </div>
</template>
<script>
// 引入 monaco 命名空间
import * as monaco from 'monaco-editor';
export default {
  data() {
    return {
      editor: null,
      readonlyMode: true,
      content: '',
      selectedValue: 'plaintext', // 用于绑定选择的值
      options: [
        { label: 'text', value: 'plaintext' },
        { label: 'markdown', value: 'markdown' },
        { label: 'yaml', value: 'yaml' },
        { label: 'json', value: 'json' },
        { label: 'css', value: 'css' },
        { label: 'javascript', value: 'javascript' },
        { label: 'html', value: 'html' },
        { label: 'properties', value: 'properties' }
      ]
    };
  },
  props: {
    currentPath: {
      type: String,
      default: '.'
    }
  },
  mounted() {
    // 在 mounted 钩子函数中初始化 Monaco Editor
   if (!this.editor) {
     this.editor = monaco.editor.create(this.$refs.editorContainer, {
       value: '',
       language: 'plaintext',
       theme: "vs-dark",
       readOnly: true
     });
   }
    this.loadContent(this.currentPath)
  },
  methods: {
    onLanguageChange() {
      monaco.editor.setModelLanguage(this.editor.getModel(), this.selectedValue);
    },
    loadContent(currentPath) {
      this.$http.get('./api/manager/file/content?path=' + currentPath).then(response => {
        this.content = response.data.data;
        this.updateEditor()
      }).catch(error => {
        this.content = ""
        console.error('加载内容出错', error);
        this.$alert(error.response.data.message, '错误', {
          confirmButtonText: '确定',
          type: 'error'
        });
      });
    },
    updateEditor() {
      this.editor.setValue(this.content)
    },
    saveContent() {
      this.content = this.editor.getValue()
      const formData = new FormData();
      formData.append('content', this.content);
      formData.append('path', this.currentPath);
      this.$http.post('./api/manager/file/content', formData).then(response => {
        this.$message({
          message: "内容更新成功",
          type: 'success'
        })
      }).catch(error => {
        console.error('保存失败', error);
        this.$message.error("内容更新失败：" + error.response.data.message)
      });
    }
  },
  watch: {
    readonlyMode(value) {
      if (this.editor) {
        this.editor.updateOptions({readOnly: value});
      }
    }
  }
};
</script>

<style>
.switch-container {
  margin-bottom: 10px;
  line-height: 40px;
}
.language-container {
  margin-bottom: 5px;
  line-height: 20px;
}
.editor-container {
  width: 100%;
  height: 400px;
}
.save-content-btn {
  float: right;
}
</style>