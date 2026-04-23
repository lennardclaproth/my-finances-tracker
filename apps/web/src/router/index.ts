import { createRouter, createWebHistory } from "vue-router";
import CashflowTransactionsPage from "../pages/CashflowTransactionsPage.vue";
import PortfolioPage from "../pages/PortfolioPage.vue";
import AssetsPage from "../pages/AssetsPage.vue";
import AdminListingsPage from "../pages/AdminListingsPage.vue";
import AdminDailiesPage from "../pages/AdminDailiesPage.vue";
import { useAppSession } from "../composables/useAppSession";

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      redirect: "/cashflow",
    },
    {
      path: "/cashflow/transactions",
      redirect: "/cashflow",
    },
    {
      path: "/tagging",
      redirect: "/cashflow",
    },
    {
      path: "/cashflow",
      component: CashflowTransactionsPage,
    },
    {
      path: "/analyze",
      redirect: "/portfolio",
    },
    {
      path: "/portfolio",
      component: PortfolioPage,
    },
    {
      path: "/assets",
      component: AssetsPage,
    },
    {
      path: "/admin/listings",
      component: AdminListingsPage,
    },
    {
      path: "/admin/dailies",
      component: AdminDailiesPage,
    },
  ],
});

router.beforeEach((to) => {
  if (!to.path.startsWith("/admin")) {
    return true;
  }
  const { adminMode } = useAppSession();
  if (!adminMode.value) {
    return { path: "/cashflow" };
  }
  return true;
});
