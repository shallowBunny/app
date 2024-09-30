// src/lib/fetchData.ts

import { Data } from "./types";
import { getApiURL } from "./api";

export async function fetchData(): Promise<Data> {
	const apiBaseURL = getApiURL("api");

	console.log("fetch on " + apiBaseURL);
	const response = await fetch(apiBaseURL);
	if (!response.ok) {
		throw new Error("Network response was not ok");
	}
	const data = (await response.json()) as Data;

	// Transform the data as required
	const transformedData = {
		sets: data.sets.map((item: any) => ({
			dj: item.dj,
			room: item.room,
			start: new Date(item.start),
			end: new Date(item.end),
		})),
		meta: {
			...data.meta,
			beginningSchedule: new Date(data.meta.beginningSchedule), // Ensure this is a Date object
		},
	} as Data;

	// Store the transformed data in localStorage

	return transformedData;
}
