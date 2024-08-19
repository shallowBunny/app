// src/lib/queryConfig.ts
import { UseQueryOptions } from "@tanstack/react-query";
import { fetchData } from "./fetchData";
import { Data } from "./types";

export const queryOptions: UseQueryOptions<Data, Error, Data> = {
	queryKey: ["fetchData"],
	queryFn: fetchData,
	refetchInterval: 60000,
};
