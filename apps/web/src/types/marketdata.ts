export interface DailyRowDto {
  ID: string;
  ListingID: string;
  Symbol: string;
  Date: string;
  Open: number;
  Close: number;
  High: number;
  Low: number;
  Volume: number;
  CreatedAt: string;
  UpdatedAt: string;
}

export interface DailiesMetadataDto {
  Message: string;
  ResultCount: number;
  TotalCount: number;
}

export interface ListingDailiesResponseDto {
  Data: DailyRowDto[];
  Metadata: DailiesMetadataDto;
}

export interface ListingDaily {
  id: string;
  listingId: string;
  symbol: string;
  date: string;
  open: number;
  close: number;
  high: number;
  low: number;
  volume: number;
  createdAt: string;
  updatedAt: string;
}

export interface ListingDailiesResponse {
  data: ListingDaily[];
  metadata: {
    message: string;
    resultCount: number;
    totalCount: number;
  };
}

export interface FetchListingDailiesParams {
  listingId?: string;
  symbol: string;
  from?: string;
  to?: string;
  sortOrder?: "asc" | "desc";
  limit?: number;
  offset?: number;
}

export interface UploadListingDailiesAcceptedResponseDto {
  upload_id: string;
  status: string;
}

export interface UploadListingDailiesAcceptedResponse {
  uploadId: string;
  status: string;
}

