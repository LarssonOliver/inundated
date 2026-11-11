import { ResponseError } from "./generated";

/** Whether error is a response from the server with the given HTTP status. */
export function isResponseStatus(error: unknown, status: number): boolean {
  return error instanceof ResponseError && error.response.status === status;
}
