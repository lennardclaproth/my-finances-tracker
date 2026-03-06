import { requestFormData, requestJson } from "./http";
import type {
  FetchListingDailiesParams,
  ListingDaily,
  ListingDailiesResponse,
  ListingDailiesResponseDto,
  UploadListingDailiesAcceptedResponse,
  UploadListingDailiesAcceptedResponseDto,
} from "../types/marketdata";

function toListingDaily(dto: ListingDailiesResponseDto["Data"][number]): ListingDaily {
  const toUnitPrice = (scaled: number): number => scaled / 1_000_000;
  return {
    id: dto.ID,
    listingId: dto.ListingID,
    symbol: dto.Symbol,
    date: dto.Date,
    open: toUnitPrice(dto.Open),
    close: toUnitPrice(dto.Close),
    high: toUnitPrice(dto.High),
    low: toUnitPrice(dto.Low),
    volume: dto.Volume,
    createdAt: dto.CreatedAt,
    updatedAt: dto.UpdatedAt,
  };
}

export async function fetchListingDailies(
  params: FetchListingDailiesParams,
): Promise<ListingDailiesResponse> {
  const payload = await requestJson<ListingDailiesResponseDto>("/marketdata/dailies", {
    method: "GET",
    query: {
      listing_id: params.listingId?.trim() || undefined,
      symbol: params.symbol.trim(),
      from: params.from,
      to: params.to,
      sort_order: params.sortOrder,
      limit: params.limit,
      offset: params.offset,
    },
  });

  return {
    data: payload.Data.map(toListingDaily),
    metadata: {
      message: payload.Metadata.Message,
      resultCount: payload.Metadata.ResultCount,
      totalCount: payload.Metadata.TotalCount,
    },
  };
}

export async function uploadListingDailies(input: {
  listingId: string;
  file: File;
}): Promise<UploadListingDailiesAcceptedResponse> {
  const formData = new FormData();
  formData.set("listing_id", input.listingId);
  formData.set("file", input.file);

  const payload = await requestFormData<UploadListingDailiesAcceptedResponseDto>(
    "/marketdata/dailies/upload",
    {
      method: "POST",
      body: formData,
    },
  );

  return {
    uploadId: payload.upload_id,
    status: payload.status,
  };
}

