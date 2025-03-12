// src/lib/fetchData.ts

import { Data } from "./types";
import { getApiURL } from "./api";

export async function fetchData(festival?: string): Promise<Data> {
	const apiBaseURL = getApiURL(festival ? `api/lineup/${festival}` : "api");
	try {
		console.log("fetch on " + apiBaseURL);
		const response = await fetch(apiBaseURL);
		if (!response.ok) {
			throw new Error("Network response was not ok");
		}
		const data = (await response.json()) as Data;

		// Transform the data as required
		const transformedData = {
			sets: data.sets.map((item: any) => {
				if (!item.dj || !item.room || !item.start || !item.end) {
					throw new Error(`Invalid set data: ${JSON.stringify(item)}`);
				}
				const start = new Date(item.start);
				const end = new Date(item.end);
				if (isNaN(start.getTime()) || isNaN(end.getTime())) {
					throw new Error(`Invalid date for set: ${JSON.stringify(item)}`);
				}
				return {
					dj: item.dj,
					room: item.room,
					start,
					end,
					meta: item.meta,
				};
			}),
			meta: {
				...data.meta,
				beginningSchedule: new Date(data.meta.beginningSchedule),
			},
		} as Data;

		if (isNaN(transformedData.meta.beginningSchedule.getTime())) {
			throw new Error(
				`Invalid beginningSchedule: ${data.meta.beginningSchedule}`
			);
		}
		return transformedData;
	} catch (error) {
		//console.errr("Error transforming data:", error);
		if (error instanceof Error) {
			throw new Error(`Error while downloading data: ${error.message}`);
		} else {
			throw new Error(`Error while downloading data: ${String(error)}`);
		}
	}
}
