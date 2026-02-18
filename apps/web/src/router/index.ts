import { createRouter, createWebHistory } from "vue-router";
import CashflowTransactionsPage from "../pages/CashflowTransactionsPage.vue";
import AnalyzePage from "../pages/AnalyzePage.vue";

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      redirect: "/tagging",
    },
    {
      path: "/cashflow/transactions",
      redirect: "/tagging",
    },
    {
      path: "/tagging",
      component: CashflowTransactionsPage,
    },
    {
      path: "/analyze",
      component: AnalyzePage,
    },
  ],
});
