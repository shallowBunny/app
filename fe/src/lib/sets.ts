// src/lib/sets.ts

import { Set } from "./types";

export function groupSetsByDayAndTime(sets: Set[]): Record<string, Set[]> {
	const grouped: Record<string, Set[]> = {};

	sets.forEach((set) => {
		const day = set.start.toLocaleDateString([], {
			weekday: "long",
		});
		if (!grouped[day]) {
			grouped[day] = [];
		}
		grouped[day].push(set);
	});

	// Sort sets within each day
	Object.keys(grouped).forEach((day) => {
		if (grouped[day]) {
			grouped[day].sort(
				(a, b) => new Date(a.start).getTime() - new Date(b.start).getTime()
			);
		}
	});

	return grouped;
}

export const allSetsInPastAndFinishedMoreThanXHoursAgo = (
	sets: Set[],
	hoursAgo: number
): boolean => {
	if (!sets || sets.length === 0) return true;

	const timeAgo = new Date(Date.now() - hoursAgo * 60 * 60 * 1000);

	for (const set of sets) {
		const setEndTime = new Date(set.end);
		if (setEndTime >= timeAgo) {
			return false;
		}
	}

	return true;
};

export const shouldSkipYouAreHereInsertion = (
	sets: Set[],
	currentTime: Date
): boolean => {
	return !sets.some((set) => {
		const setStart = new Date(set.start);
		const setEnd = new Date(set.end);
		return (
			(setStart.getDate() === currentTime.getDate() &&
				setStart.getMonth() === currentTime.getMonth() &&
				setStart.getFullYear() === currentTime.getFullYear()) ||
			(setEnd.getDate() === currentTime.getDate() &&
				setEnd.getMonth() === currentTime.getMonth() &&
				setEnd.getFullYear() === currentTime.getFullYear())
		);
	});
};
