// utils.ts
import { Set, RoomSets, Meta, Like, RoomSituation } from "./types";
import clsx from "clsx";

export function findCurrentAndNextSets(
	sets: Set[],
	now: Date = getOverriddenCurrentTime()
): Record<string, RoomSets> {
	const roomSets: Record<string, RoomSets> = {};

	sets.forEach((set) => {
		const currentRoomSets = roomSets[set.room] || {
			current: null,
			next: null,
			pauseDuration: null,
			closing: false,
		};

		if (set.start <= now && set.end >= now) {
			// If the set is happening now, update the current set
			currentRoomSets.current = set;
		} else if (set.start > now) {
			// If the set is in the future, check if it should be the next set
			if (!currentRoomSets.next || set.start < currentRoomSets.next.start) {
				currentRoomSets.next = set;
			}
		}

		if (currentRoomSets.current && currentRoomSets.next) {
			const currentEndTime = new Date(currentRoomSets.current.end).getTime();
			const nextStartTime = new Date(currentRoomSets.next.start).getTime();
			currentRoomSets.pauseDuration = Math.round(
				(nextStartTime - currentEndTime) / (1000 * 60)
			); // Convert milliseconds to minutes
			currentRoomSets.closing = currentRoomSets.pauseDuration > 120; // 2 hours
		} else {
			currentRoomSets.pauseDuration = null;
			currentRoomSets.closing = true;
		}

		roomSets[set.room] = currentRoomSets;
	});

	return roomSets;
}

export function getOverriddenCurrentTime(): Date {
	const isDevMode = import.meta.env.MODE === "development";
	const now = new Date();

	if (!isDevMode) {
		return now;
	}
	return now;
	/*	
	// Set the time to 17:01 on the upcoming Sunday
	const testTime = new Date();
	testTime.setDate(testTime.getDate() + ((7 - testTime.getDay()) % 7)); // Adjust to upcoming Sunday
	testTime.setHours(17, 1, 0, 0);
	return testTime;
	*/
}

const formatForClosedStages = (date: Date, dj: string): string => {
	const today = getOverriddenCurrentTime();
	const isToday =
		date.getDate() === today.getDate() &&
		date.getMonth() === today.getMonth() &&
		date.getFullYear() === today.getFullYear();

	const optionsDay: Intl.DateTimeFormatOptions = {
		weekday: "short",
	};
	const optionsTime: Intl.DateTimeFormatOptions = {
		hour: "2-digit",
		minute: "2-digit",
	};

	const time = date.toLocaleTimeString("en-GB", optionsTime);

	if (dj === "?") {
		if (isToday) {
			return `🚫 closed until ${time}`;
		} else {
			const day = date.toLocaleDateString("en-GB", optionsDay);
			return `🚫 closed until ${day} at ${time}`;
		}
	} else {
		if (isToday) {
			return `🚫 closed (${dj} at ${time})`;
		} else {
			const day = date.toLocaleDateString("en-GB", optionsDay);
			return `🚫 closed (${dj}, ${day} at ${time})`;
		}
	}
};

export function convertRoomSetsToRoomSituation(
	roomSets: Record<string, RoomSets>,
	meta: Meta // Use Meta instead of allRooms
): RoomSituation[] {
	// Extract the list of all rooms from meta
	const allRooms = meta.rooms;

	// Start with an array of RoomSituation with all rooms set to "⚠️ no data" and marked as closed
	const roomSituations: RoomSituation[] = allRooms.map((room: string) => {
		const sets = roomSets[room];
		return {
			room,
			situation: sets ? `${room} 🚫 closed` : `${room} ⚠️ no data`,
			closed: true, // Mark the room as closed initially
		};
	});

	// Iterate over the rooms in roomSets
	Object.keys(roomSets).forEach((room) => {
		const sets = roomSets[room];

		let situation = `${room} `;
		let like: Like | undefined = undefined; // Initialize like as undefined
		let closed = false;

		const formatTime = (date: Date, omitWeekday: boolean = false): string => {
			const today = new Date();
			const isToday =
				date.getDate() === today.getDate() &&
				date.getMonth() === today.getMonth() &&
				date.getFullYear() === today.getFullYear();

			const options: Intl.DateTimeFormatOptions =
				isToday || omitWeekday
					? { hour: "2-digit", minute: "2-digit" }
					: { weekday: "long" };

			const timeOptions: Intl.DateTimeFormatOptions = {
				hour: "2-digit",
				minute: "2-digit",
			};

			return isToday || omitWeekday
				? ` at ${date.toLocaleTimeString("en-GB", timeOptions)}`
				: `, ${date.toLocaleDateString("en-GB", options)} at ${date.toLocaleTimeString("en-GB", timeOptions)}`;
		};

		if (sets?.current) {
			situation += `✅ ${sets.current.dj}`;
			like = {
				dj: sets.current.dj,
				title: meta.title, // Get title from meta
				beginningSchedule: meta.beginningSchedule, // Get beginningSchedule from meta
				room: sets.current.room,
				started: sets.current.start,
				links: sets.current.links,
			};
			closed = false; // The room is open if there's a current DJ
		}

		if (sets?.closing) {
			if (sets.current) {
				const currentEndTime = formatTime(new Date(sets.current.end), true); // Omit weekday for closing time
				situation += ` (Closing${currentEndTime})`;
			} else if (sets.next) {
				const nextStartTime = new Date(sets.next.start);
				situation += formatForClosedStages(nextStartTime, sets.next.dj);
				closed = true;
			} else {
				situation += `🚫 closed`;
				closed = true;
			}
		} else if (sets?.next) {
			const nextStartTime = new Date(sets.next.start);
			const omitWeekday =
				sets.current && sets.current.end.getTime() === nextStartTime.getTime();
			if (sets.pauseDuration && sets.pauseDuration > 0) {
				situation += ` (${sets.next.dj}${formatTime(nextStartTime, omitWeekday || false)} after ${sets.pauseDuration} min of pause)`;
			} else {
				situation += ` (${sets.next.dj}${formatTime(nextStartTime, omitWeekday || false)})`;
			}
		}

		// Update room situation in the roomSituations array
		const situationIndex = roomSituations.findIndex((rs) => rs.room === room);
		if (situationIndex !== -1) {
			roomSituations[situationIndex] = {
				room,
				situation,
				like, // Assign the like structure with DJ's name, title, and beginningSchedule
				closed, // Whether the room is closed
			};
		}
	});

	return roomSituations;
}

export const cn = clsx;
