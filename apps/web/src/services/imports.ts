import { requestFormData } from "./http";

interface ImportCsvPayload {
  file: File;
  vendorId: string;
}

export async function importCsv(payload: ImportCsvPayload): Promise<string> {
  const formData = new FormData();
  formData.append("file", payload.file);
  formData.append("vendor_id", payload.vendorId);

  return requestFormData<string>("/import/csv", {
    method: "POST",
    body: formData,
  });
}
