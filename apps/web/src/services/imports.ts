import { requestFormData } from "./http";
import { useAppSession } from "../composables/useAppSession";

interface ImportCsvPayload {
  file: File;
  vendorId: string;
  accountId?: string;
}

export async function importCsv(payload: ImportCsvPayload): Promise<string> {
  const session = useAppSession();
  const formData = new FormData();
  formData.append("file", payload.file);
  formData.append("vendor_id", payload.vendorId);
  formData.append("account_id", payload.accountId ?? session.activeAccountId.value);

  return requestFormData<string>("/import/csv", {
    method: "POST",
    body: formData,
  });
}
