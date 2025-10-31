<template>
  <div class="markdown-body" v-html="mdContent"></div>
</template>

<script>
import { ref, onMounted } from 'vue'
import MarkdownIt from 'markdown-it'
import apiMd from '../api.md'  // 更新路径，指向 src/api.md

export default {
  setup() {
    const md = new MarkdownIt()
    const mdContent = ref('')

    onMounted(() => {
      try {
        mdContent.value = md.render(apiMd)
      } catch (err) {
        console.error('渲染 markdown 失败:', err)
      }
    })

    return {
      mdContent
    }
  }
}
</script>

<style>
/* 添加 GitHub 风格的 markdown 样式 */
@import '~github-markdown-css/github-markdown.css';

.markdown-body {
  box-sizing: border-box;
  min-width: 200px;
  max-width: 980px;
  margin: 0 auto;
  padding: 45px;
}

@media (max-width: 767px) {
  .markdown-body {
    padding: 15px;
  }
}
</style>