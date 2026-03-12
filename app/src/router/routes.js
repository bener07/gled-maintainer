import DashboardLayout from "@/layout/dashboard/DashboardLayout.vue";
import NotFound from "@/pages/NotFoundPage.vue";

const Dashboard = () =>
  import(/* webpackChunkName: "dashboard" */ "@/pages/Dashboard.vue");
const Clients = () =>
  import(/* webpackChunkName: "clients" */ "@/pages/Clients.vue");
const Updates = () =>
  import(/* webpackChunkName: "updates" */ "@/pages/Updates.vue");
const Profile = () =>
  import(/* webpackChunkName: "common" */ "@/pages/Profile.vue");
const Notifications = () =>
  import(/* webpackChunkName: "common" */ "@/pages/Notifications.vue");

const routes = [
  {
    path: "/",
    component: DashboardLayout,
    redirect: "/dashboard",
    children: [
      {
        path: "dashboard",
        name: "dashboard",
        component: Dashboard,
      },
      {
        path: "clients",
        name: "clients",
        component: Clients,
      },
      {
        path: "updates",
        name: "updates",
        component: Updates,
      },
      {
        path: "profile",
        name: "profile",
        component: Profile,
      },
      {
        path: "notifications",
        name: "notifications",
        component: Notifications,
      },
    ],
  },
  { path: "*", component: NotFound },
];

export default routes;
