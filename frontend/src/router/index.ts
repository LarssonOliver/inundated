import { createRouter, createWebHistory } from "vue-router";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "Timesheet",
      // route level code-splitting
      // this generates a separate chunk (About.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: () => import("../views/TimesheetView.vue"),
    },
    {
      path: "/projects",
      name: "Projects",
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
