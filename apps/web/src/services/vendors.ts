import type { Vendor } from "../types/vendors";
import { requestJson } from "./http";

interface VendorApiResponse {
  id: string;
  name: string;
  type: string;
  active: boolean;
  created_at: string;
  updated_at: string;
}

export async function fetchVendors(): Promise<Vendor[]> {
  const response = await requestJson<VendorApiResponse[]>("/vendors", {
    method: "GET",
  });

  return response.map((vendor) => ({
    id: vendor.id,
    name: vendor.name,
    type: vendor.type,
    active: vendor.active,
    createdAt: vendor.created_at,
    updatedAt: vendor.updated_at,
  }));
}
