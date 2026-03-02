import { createRouter, createWebHistory } from "vue-router";
import CashflowTransactionsPage from "../pages/CashflowTransactionsPage.vue";
import PortfolioPage from "../pages/PortfolioPage.vue";
import AdminListingsPage from "../pages/AdminListingsPage.vue";
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
      path: "/admin/listings",
      component: AdminListingsPage,
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
