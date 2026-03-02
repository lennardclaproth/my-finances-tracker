import { ref } from "vue";
import { APP_STORAGE_KEYS, BOOTSTRAP_ACCOUNT_ID } from "../config/app";

function readInitialAdminMode(): boolean {
  if (typeof window === "undefined") {
    return false;
  }
  const raw = window.localStorage.getItem(APP_STORAGE_KEYS.adminMode);
  if (raw === "true") {
    return true;
  }
  if (raw === "false") {
    return false;
  }
  return false;
}

const adminMode = ref<boolean>(readInitialAdminMode());
const activeAccountId = ref<string>(BOOTSTRAP_ACCOUNT_ID);

function persistAdminMode(value: boolean): void {
  if (typeof window === "undefined") {
    return;
  }
  window.localStorage.setItem(APP_STORAGE_KEYS.adminMode, value ? "true" : "false");
}

function setAdminMode(value: boolean): void {
  adminMode.value = value;
  persistAdminMode(value);
}

function toggleAdminMode(): void {
  setAdminMode(!adminMode.value);
}

export function useAppSession() {
  return {
    adminMode,
    activeAccountId,
    setAdminMode,
    toggleAdminMode,
  };
}
