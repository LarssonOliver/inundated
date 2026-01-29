import { createApp } from "vue";
import { createPinia } from "pinia";
import App from "./App.vue";
import router from "./router";
import { errorHandlingPlugin } from "./pinia/error-plugin";

const pinia = createPinia();
pinia.use(errorHandlingPlugin);

const app = createApp(App);

app.use(pinia);
app.use(router);
app.mount("#app");
