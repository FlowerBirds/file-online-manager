<template>
  <div style="height: 500px;">
    <div ref="editorContainer" style="height: 100%;"></div>
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
      selectedValue: 'log', // 用于绑定选择的值
    };
  },
  props: {
    content: {
      type: String,
      default: ''
    },
    mode: {
      type: String,
      default: 'yaml'
    }
  },
  mounted() {
    // 在 mounted 钩子函数中初始化 Monaco Editor
    if (!this.editor) {
      this.editor = monaco.editor.create(this.$refs.editorContainer, {
        value: this.content,
        language: this.mode,
        theme: "vs-dark",
        readOnly: true,
        wordWrap: true,
      });
    }
    this.loadContent(this.currentPath)
    this.updateEditor()
  },
  methods: {
    onLanguageChange() {
      monaco.editor.setModelLanguage(this.editor.getModel(), this.selectedValue);
    },
    loadContent(currentPath) {

    },
    updateEditor() {
      this.editor.setValue(this.content)
      debugger
      if (this.editor.getScrollHeight() > 900 && this.mode == 'log') {
        this.editor.setScrollTop(this.editor.getScrollHeight() - 900);
        // console.log(this.editor.getScrollHeight())
      }

    },
    saveContent() {

    }
  },
  watch: {
    readonlyMode(value) {
      if (this.editor) {
        this.editor.updateOptions({readOnly: value});
      }
    },
    content(value) {
      if (this.editor) {
        this.updateEditor()
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