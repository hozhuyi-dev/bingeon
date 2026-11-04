import type { AccountView } from "./account.types";
import { invoke } from "~/core/lambda/lambda.client";

export const AccountService = {
  get: () => invoke<AccountView>("account.get"),
  delete: () => invoke<null>("account.delete"),
};
