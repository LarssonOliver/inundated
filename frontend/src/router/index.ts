import { createRouter, createWebHistory } from "vue-router";
import MainView from "../views/MainView.vue";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "Home",
      component: MainView,
    },
    {
      path: "/projects",
      name: "Projects",
      // route level code-splitting
      // this generates a separate chunk (About.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: () => import("../views/ProjectListView.vue"),
    },
    {
      path: "/projects/new",
      name: "New Project",
      component: () => import("../views/ProjectView.vue"),
    },
    {
      path: "/projects/:id",
      name: "Project",
      component: () => import("../views/ProjectView.vue"),
    },
    {
      path: "/tags",
      name: "Tags",
      component: () => import("../views/TagListView.vue"),
    },
    {
      path: "/tags/new",
      name: "New Tag",
      component: () => import("../views/TagView.vue"),
    },
    {
      path: "/tags/:id",
      name: "Tag",
      component: () => import("../views/TagView.vue"),
    },
  ],
});

export default router;
