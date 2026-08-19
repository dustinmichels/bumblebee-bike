import { createApp } from "vue";
import App from "./App.vue";
import MapTest from "./components/MapTest.vue";

const pathname = window.location.pathname.replace(/\/$/, "") || "/";
const root = pathname === "/map-test" ? MapTest : App;

createApp(root).mount("#app");
