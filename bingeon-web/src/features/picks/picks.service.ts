import type { PicksView, Status } from "~/features/account/account.types";
import { invoke } from "~/core/lambda/lambda.client";

export const PicksService = {
  make: () => invoke<{ picks: PicksView; status: Status }>("picks.make"),
};
