// lib/time.ts

/**
 * Formats the date to include the full schedule information (weekday and date)
 * using "en-GB" locale.
 * @param date The date to format
 * @returns Formatted string with the full schedule (weekday and full date)
 */
export const formatDate = (date: Date): string => {
	return date.toLocaleDateString("en-GB", {
		day: "2-digit",
		month: "2-digit",
		year: "numeric",
	});
};

/**
 * Formats the date to show only the day and time (hour and minute)
 * using "en-GB" locale.
 * @param date The date to format
 * @returns Formatted string with the day and time (HH:mm)
 */
export const formatDayAndTime = (date: Date): string => {
	const day = date.toLocaleDateString("en-GB", {
		weekday: "short", // e.g., Mon, Tue
	});
	const time = date.toLocaleTimeString("en-GB", {
		hour: "2-digit",
		minute: "2-digit",
		hour12: false, // 24-hour format
	});
	return `${day} at ${time}`; // Day and time (e.g., Mon 14:00)
};
