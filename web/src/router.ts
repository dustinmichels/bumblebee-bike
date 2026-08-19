import { createRouter, createWebHistory } from "vue-router";
import FunctionHome from "./components/home/FunctionHome.vue";
import LightningMapFlow from "./components/flows/LightningMapFlow.vue";
import CompareFlow from "./components/flows/CompareFlow.vue";
import MapTest from "./components/MapTest.vue";
import UploadPage from "./components/uploads/UploadPage.vue";

const routes = [
  {
    path: "/",
    name: "home",
    component: FunctionHome,
  },
  {
    path: "/upload",
    name: "upload",
    component: UploadPage,
  },
  {
    path: "/lightning-map",
    name: "lightning-map",
    component: LightningMapFlow,
  },
  {
    path: "/lightning",
    redirect: "/lightning-map",
  },
  {
    path: "/compare",
    name: "compare",
    component: CompareFlow,
  },
  {
    path: "/map-test",
    name: "map-test",
    component: MapTest,
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
