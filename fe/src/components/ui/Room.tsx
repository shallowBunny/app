// Room.tsx
import {
	groupSetsByDayAndTime,
	shouldSkipYouAreHereInsertion,
} from "@/lib/setUtils";

import { Set, Like } from "@/lib/types"; // Import the Like type

interface RoomProps {
	sets: Set[];
	room: string;
	youarehere: string;
	currentMinute: Date;
	likedDJs: Like[]; // Use an array of Like objects instead of a map
}

// Utility function to check if a DJ is liked
const isDjLiked = (djName: string, likedDJs: Like[]): boolean => {
	return likedDJs.some((like) => like.dj === djName);
};

const shouldInsertBeforeNextDayMarker = (
	youAreHereInserted: boolean,
	currentTime: Date,
	groupedSets: Record<string, Set[]>,
	sortedDays: string[],
	dayIndex: number
): boolean => {
	if (youAreHereInserted || dayIndex >= sortedDays.length - 1) {
		return false;
	}

	const currentDay = sortedDays[dayIndex];
	const nextDay = sortedDays[dayIndex + 1];

	if (
		!currentDay ||
		!nextDay ||
		!groupedSets[currentDay] ||
		!groupedSets[nextDay] ||
		groupedSets[nextDay].length === 0
	) {
		return false;
	}

	const currentDaySets = groupedSets[currentDay];
	const lastSetOfCurrentDay = currentDaySets[currentDaySets.length - 1];

	if (!lastSetOfCurrentDay) {
		return false;
	}

	const nextDayFirstSet = groupedSets[nextDay][0];
	if (!nextDayFirstSet) {
		return false;
	}

	const lastSetEndTime = new Date(lastSetOfCurrentDay.start).getTime();
	const nextDayFirstSetTime = new Date(nextDayFirstSet.start).getTime();

	const currentDayStillValid =
		currentTime.getDate() === new Date(lastSetOfCurrentDay.start).getDate() &&
		currentTime.getMonth() === new Date(lastSetOfCurrentDay.start).getMonth() &&
		currentTime.getFullYear() ===
			new Date(lastSetOfCurrentDay.start).getFullYear();

	return (
		currentDayStillValid &&
		currentTime.getTime() > lastSetEndTime &&
		currentTime.getTime() < nextDayFirstSetTime
	);
};

export const isSetOngoingNow = (sets: Set[]): boolean => {
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
	const { sets, room, youarehere, currentMinute, likedDJs } = props; // Destructure likedDJs (array)
	const groupedSets = groupSetsByDayAndTime(sets);

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

	let youAreHereInserted = shouldSkipYouAreHereInsertion(sets, currentTime);

	// Find the last set of the last day
	const lastDay =
		sortedDays.length > 0 ? sortedDays[sortedDays.length - 1] : null;
	const lastSetOfLastDay =
		lastDay && groupedSets[lastDay]
			? groupedSets[lastDay][groupedSets[lastDay].length - 1]
			: null;

	const insertYouAreHereMarker = (setTime: Date): boolean => {
		if (!youAreHereInserted && setTime > currentTime) {
			if (!lastSetOfLastDay || currentTime < new Date(lastSetOfLastDay.end)) {
				return true;
			}
		}
		return false;
	};

	const renderClosingOrClosed = () => {
		if (lastSetOfLastDay) {
			return (
				<h2 className="text-[18px]">
					{new Date(lastSetOfLastDay.end).toLocaleTimeString("en-GB", {
						hour: "2-digit",
						minute: "2-digit",
					})}{" "}
					{new Date(lastSetOfLastDay.end) < currentTime ? "closed" : "closing"}
				</h2>
			);
		}
		return null;
	};

	const renderYouAreHere = () => (
		<h2 className="text-[18px]">
			{currentMinute.toLocaleTimeString("en-GB", {
				hour: "2-digit",
				minute: "2-digit",
			})}{" "}
			{youarehere} &larr; you are here
		</h2>
	);

	return (
		<>
			<div>
				<h2 className="text-[25px]">
					{room}
					{roomTag}
				</h2>
				{sortedDays.map((day, dayIndex) => {
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
							<h2 className="text-[20px]">{isToday ? "Today" : day}:</h2>
							<ul>
								{setsForDay.map((set, index) => {
									const setTime = new Date(set.start);
									const isCurrentDay =
										setTime.getDate() === currentTime.getDate() &&
										setTime.getMonth() === currentTime.getMonth() &&
										setTime.getFullYear() === currentTime.getFullYear();

									const shouldInsertYouAreHereMarker =
										insertYouAreHereMarker(setTime);

									return (
										<div key={`${set.dj}-${set.start}`}>
											{shouldInsertYouAreHereMarker && (
												<>
													<li key="you-are-here">
														<h2 className="text-[18px]">
															{currentMinute.toLocaleTimeString("en-GB", {
																hour: "2-digit",
																minute: "2-digit",
															})}{" "}
															{youarehere} &larr; you are here
														</h2>
													</li>
													{(youAreHereInserted = true)}
												</>
											)}
											<li
												key={`${set.dj}-${set.start}`}
												className="text-[18px]"
											>
												{set.start.toLocaleTimeString("en-GB", {
													hour: "2-digit",
													minute: "2-digit",
												})}{" "}
												{set.dj}
												{/* Show heart if the DJ is liked */}
												{isDjLiked(set.dj, likedDJs) && <span> ❤️</span>}
											</li>
											{index === setsForDay.length - 1 &&
												isCurrentDay &&
												!youAreHereInserted &&
												insertYouAreHereMarker(setTime) && (
													<>
														<li key="you-are-here-end">
															<h2 className="text-[18px]">
																{currentMinute.toLocaleTimeString("en-GB", {
																	hour: "2-digit",
																	minute: "2-digit",
																})}{" "}
																{youarehere} &larr; you are here
															</h2>
															{(youAreHereInserted = true)}
														</li>
													</>
												)}
										</div>
									);
								})}
								{shouldInsertBeforeNextDayMarker(
									youAreHereInserted,
									currentTime,
									groupedSets,
									sortedDays,
									dayIndex
								) && (
									<>
										<li key="you-are-here-before-next-day">
											<h2 className="text-[18px]">
												{currentMinute.toLocaleTimeString("en-GB", {
													hour: "2-digit",
													minute: "2-digit",
												})}{" "}
												{youarehere} &larr; you are here
											</h2>
											{(youAreHereInserted = true)}
										</li>
									</>
								)}
							</ul>
						</div>
					);
				})}
				{lastSetOfLastDay && (
					<div>
						{new Date(lastSetOfLastDay.end) < currentTime &&
							renderClosingOrClosed()}
						{!youAreHereInserted && renderYouAreHere()}
						{new Date(lastSetOfLastDay.end) >= currentTime &&
							renderClosingOrClosed()}
					</div>
				)}
			</div>
		</>
	);
};

export default Room;
