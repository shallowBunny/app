// src/lib/queryConfig.ts
import { UseQueryOptions } from "@tanstack/react-query";
import { fetchData } from "./fetchData";
import { Data } from "./types";

export const getQueryOptions = (
	festival?: string
): UseQueryOptions<Data, Error, Data> => ({
	queryKey: festival ? ["fetchData", festival] : ["fetchData"], // Ensure cache separation
	queryFn: () => fetchData(festival), // Pass the festival parameter
	refetchInterval: 60000,
});
