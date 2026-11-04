/** Every reply from the API: the data, or an error it chose to return. */
export interface ApiResponse<T> {
  data?: T;
  error?: ApiErrorBody;
}

/** An error the API chose to return. */
export interface ApiErrorBody {
  code: string;
  message: string;
}
