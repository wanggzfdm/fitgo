<template>
  <n-message-provider>
    <div id="app">
      <RunningRecord />
    </div>
  </n-message-provider>
</template>

<script>
import { NMessageProvider } from 'naive-ui'
import RunningRecord from './components/RunningRecord.vue'

export default {
  name: 'App',
  components: {
    NMessageProvider,
    RunningRecord
  }
}
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}

#nav {
  padding: 30px;
}

#nav a {
  font-weight: bold;
  color: #2c3e50;
}

#nav a.router-link-exact-active {
  color: #42b983;
}
</style>