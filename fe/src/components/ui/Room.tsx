// Room.tsx
import { groupSetsByDayAndTime } from "@/lib/sets";

import { Set, Like } from "@/lib/types"; // Import the Like type

import MetaIcons from "./MetaIcons";

interface RoomProps {
	sets: Set[];
	room: string;
	youarehere: string;
	currentMinute: Date;
	likedDJs: Like[]; // Use an array of Like objects instead of a map
	isDesktop: boolean;
	partyName: string;
}

const mergeMissingDataSets = (sets: Set[]): Set[] => {
	const mergedSets: Set[] = [];
	let currentMergedSet: Set | null = null;

	sets.forEach((set) => {
		if (set.dj === "?") {
			if (currentMergedSet) {
				currentMergedSet.end = set.end; // Extend the current merged set
			} else {
				currentMergedSet = { ...set, dj: "⚠️ Missing Data ⚠️" }; // Start a new merged set
			}
		} else {
			if (currentMergedSet) {
				mergedSets.push(currentMergedSet); // Push merged set when a valid DJ set is found
				currentMergedSet = null;
			}
			mergedSets.push(set);
		}
	});

	if (currentMergedSet) {
		mergedSets.push(currentMergedSet); // Push the final merged set if it exists
	}

	return mergedSets;
};
// Utility function to check if a DJ is liked
const isDjLiked = (djName: string, likedDJs: Like[]): boolean => {
	return likedDJs.some((like) => like.dj === djName);
};

const addClosingAndClosedSets = (sets: Set[], currentTime: Date): Set[] => {
	const updatedSets: Set[] = [];

	for (let i = 0; i < sets.length; i++) {
		const currentSet = sets[i];
		const nextSet = sets[i + 1];

		// Add the current set to the updated list if it exists
		if (currentSet) {
			updatedSets.push(currentSet);
		}

		// If there's a gap between the current set's end and the next set's start
		if (nextSet && currentSet) {
			const currentSetEndTime = new Date(currentSet.end).getTime();
			const nextSetStartTime = new Date(nextSet.start).getTime();

			if (currentSetEndTime < nextSetStartTime) {
				const closingSet: Set = {
					room: currentSet.room,
					start: currentSet.end,
					end: nextSet.start,
					dj: currentSetEndTime < currentTime.getTime() ? "closed" : "closing",
					meta: null,
				};

				// Add the closing/closed set during the gap
				updatedSets.push(closingSet);
			}
		}
	}

	// If there's only one set or we're at the end of the set list, add a "closed" set
	const lastSet = sets[sets.length - 1];
	if (lastSet) {
		const lastSetEndTime = new Date(lastSet.end).getTime();

		// Add a "closed" set if the last set has ended and it's later than the current time
		const closedSet: Set = {
			room: lastSet.room,
			start: lastSet.end,
			end: lastSet.end, //new Date(currentTime.setHours(23, 59, 59)), // Set to end of the current day
			dj: lastSetEndTime < currentTime.getTime() ? "closed" : "closing",
			meta: null,
		};

		updatedSets.push(closedSet);
	}

	return updatedSets;
};

const addYouAreHereSet = (
	sets: Set[],
	currentTime: Date,
	youAreHere: string
): Set[] => {
	// Check if there's any set that starts or ends today
	const isSetToday = sets.some((set) => {
		const setStart = new Date(set.start);
		const setEnd = new Date(set.end);

		// Compare dates (ignoring time)
		return (
			setStart.toDateString() === currentTime.toDateString() ||
			setEnd.toDateString() === currentTime.toDateString()
		);
	});

	// If no set starts or finishes today, don't add "you are here"
	if (!isSetToday) {
		return sets;
	}

	// Otherwise, add the "you are here" set
	const updatedSets: Set[] = [...sets];
	const youAreHereSet: Set = {
		room: sets[0]?.room || "Unknown room", // Default to first room or "Unknown"
		start: currentTime,
		end: currentTime,
		dj: youAreHere,
		meta: null,
	};

	updatedSets.push(youAreHereSet);

	// The sorting logic will automatically place it in the correct spot
	return updatedSets;
};

const isSetOngoingNow = (sets: Set[]): boolean => {
	const currentTime = new Date();

	for (const set of sets) {
		const setStartTime = new Date(set.start);
		const setEndTime = new Date(set.end);

		if (currentTime >= setStartTime && currentTime <= setEndTime) {
			return true;
		}
	}

	return false;
};

const Room = (props: RoomProps) => {
	const {
		sets,
		room,
		youarehere,
		currentMinute,
		likedDJs,
		isDesktop,
		partyName,
	} = props; // Destructure likedDJs (array)

	const mergedSets = addClosingAndClosedSets(
		mergeMissingDataSets(sets),
		currentMinute
	);
	const finalSets = addYouAreHereSet(
		mergedSets,
		currentMinute,
		youarehere + " ← you are here"
	);

	const groupedSets = groupSetsByDayAndTime(finalSets);

	// Sort the days chronologically
	const sortedDays = Object.keys(groupedSets).sort((a, b) => {
		const dateA = groupedSets[a]?.[0]
			? new Date(groupedSets[a][0].start).getTime()
			: 0;
		const dateB = groupedSets[b]?.[0]
			? new Date(groupedSets[b][0].start).getTime()
			: 0;
		return dateA - dateB;
	});

	const roomTag = isSetOngoingNow(sets) ? " ✅" : " 🚫";

	const currentTime = new Date();

	return (
		<>
			<div>
				<h2 className="text-[25px]">
					{partyName && `${partyName} `}
					{room}
					{roomTag}
				</h2>
				{sortedDays.map((day) => {
					const setsForDay = groupedSets[day];
					if (!setsForDay) return null;

					const isToday = setsForDay.some((set) => {
						const setTime = new Date(set.start);
						return (
							setTime.getDate() === currentTime.getDate() &&
							setTime.getMonth() === currentTime.getMonth() &&
							setTime.getFullYear() === currentTime.getFullYear()
						);
					});

					return (
						<div key={day}>
							<br />
							<h2 className="text-[20px]">
								{isToday
									? "Today"
									: (() => {
											const firstSet = groupedSets[day]?.[0]; // Get the first set for the day
											if (!firstSet || !firstSet.start) return day; // Fallback if no valid date

											const dayDate = new Date(firstSet.start); // Ensure it's a Date
											const oneWeekAgo = new Date();
											oneWeekAgo.setDate(oneWeekAgo.getDate() - 7);

											function getOrdinalSuffix(day: number): string {
												if (day > 3 && day < 21) return "th"; // Covers 4th-20th (special case)
												switch (day % 10) {
													case 1:
														return "st";
													case 2:
														return "nd";
													case 3:
														return "rd";
													default:
														return "th";
												}
											}

											return dayDate < oneWeekAgo
												? `${dayDate.toLocaleDateString("en-US", { weekday: "long" })}, ` +
														`${dayDate.toLocaleDateString("en-US", { month: "long" })} ` +
														`${dayDate.getDate()}${getOrdinalSuffix(dayDate.getDate())}`
												: day;
										})()}
								:
							</h2>

							<ul>
								{setsForDay.map((set) => {
									return (
										<div key={`${set.dj}-${set.start}`}>
											{set.dj === "closed" || set.dj === "closing" ? (
												<li
													key={`${set.dj}-${set.start}`}
													className="text-[18px]" // Add italic styling
												>
													{set.start.toLocaleTimeString("en-GB", {
														hour: "2-digit",
														minute: "2-digit",
													})}{" "}
													<span className="italic">{set.dj}</span>
												</li>
											) : (
												<li
													key={`${set.dj}-${set.start}`}
													className={`text-[18px] ${!isDesktop ? "" : "flex items-center justify-between"}`}
												>
													{set.start.toLocaleTimeString("en-GB", {
														hour: "2-digit",
														minute: "2-digit",
													})}{" "}
													{set.dj}
													{isDjLiked(set.dj, likedDJs) && <span> ❤️</span>}
													{isDesktop && (
														<MetaIcons meta={set.meta} roomPage={true} />
													)}
												</li>
											)}
										</div>
									);
								})}
							</ul>
						</div>
					);
				})}
			</div>
		</>
	);
};

export default Room;
