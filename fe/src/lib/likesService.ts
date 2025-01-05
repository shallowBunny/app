import { Like } from "./types"; // Adjust the import path as needed
import { getLikesURL } from "./api"; // Adjust the import path as needed

interface LikesResponse {
	token: string;
	likes: Like[];
}

export const parseAndMigrateLikedDJs = (parsedLikedDJs: Like[]): Like[] => {
	//	const parsedLikedDJs = JSON.parse(storedLikedDJs) as Like[];

	const likedDJsWithDates = parsedLikedDJs.map((like) => {
		// Migrate if `links` exists and `meta` is empty
		if (like.links && like.links.length > 0 && !like.meta) {
			const [linkValue] = like.links; // Get the string value from links array
			if (like.dj !== "?" && linkValue !== "") {
				like.meta = [
					{
						key: `dj.link.sisyfan.${like.dj}`,
						value: linkValue,
					},
				];
			}
			console.log("Migrating " + like.meta);
		}

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

	//	console.log("Token: ", token);
	//	console.log("LikesResponse: ", likesResponse);

	try {
		const response = await fetch(getLikesURL("likes"), {
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
			console.log("Token received and stored:", data.token);
		}

		console.log("Received updated likes from server:", likes);
		return likes;
	} catch (error) {
		console.error("Error fetching data:", error);
		throw error; // Re-throw the error for the calling function to handle
	}
};
