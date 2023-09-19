<template>
  <div class="context-menu" v-show="visible" :style="{ top: y + 'px', left: x + 'px' }">
    <ul>
      <li v-for="(item, index) in items" :key="index" @click="handleClick(item)">
        <i :class="item.icon" style="color: #66b1ff;    margin-right: 5px; "></i>{{ item.label }}
      </li>
    </ul>
  </div>
</template>

<script>
export default {
  props: {
    items: Array, // 右键菜单项数组
    x: Number, // 鼠标点击位置横坐标
    y: Number, // 鼠标点击位置纵坐标
  },
  data() {
    return {
      visible: true, // 是否显示右键菜单
    };
  },
  methods: {
    handleClick(item) {
      this.visible = false;
      this.$emit('click', item);
    },
  },
};
</script>

<style scoped>
.context-menu {
  position: fixed;
  z-index: 1000;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  background-color: white;
}

.context-menu ul {
  list-style: none;
  margin: 0;
  padding: 0;
}

.context-menu li {
  padding: 8px 16px;
  cursor: pointer;
}

.context-menu li:hover {
  background-color: #f5f5f5;
}
</style>
