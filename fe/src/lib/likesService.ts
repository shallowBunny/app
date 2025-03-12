import { Like } from "./types";
import { getApiURL } from "./api";

interface LikesResponse {
	token: string;
	likes: Like[];
}

export const parseAndMigrateLikedDJs = (parsedLikedDJs: Like[]): Like[] => {
	const likedDJsWithDates = parsedLikedDJs.map((like) => {
		return {
			...like,
			beginningSchedule: new Date(like.beginningSchedule),
			started: new Date(like.started),
			meta: like.meta || [],
		};
	});

	return likedDJsWithDates;
};

/**
 * Posts updated liked DJs to the server and handles token management.
 *
 * @param updatedLikedDJs - Array of updated Like objects
 * @returns Parsed and migrated liked DJs from the server
 * @throws Error if the network request fails
 */
export const postUpdatedLikes = async (
	updatedLikedDJs: Like[]
): Promise<Like[]> => {
	let token = localStorage.getItem("token") || "";
	const likesResponse = { token, likes: updatedLikedDJs };

	try {
		const response = await fetch(getApiURL("api/likes"), {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(likesResponse),
		});

		if (!response.ok) {
			throw new Error("Network response was not ok");
		}

		const data = (await response.json()) as LikesResponse;
		const likes = parseAndMigrateLikedDJs(data.likes);

		if (data.token) {
			localStorage.setItem("token", data.token);
		}

		return likes;
	} catch (error) {
		console.error("Error fetching data:", error);
		throw error; // Re-throw the error for the calling function to handle
	}
};
