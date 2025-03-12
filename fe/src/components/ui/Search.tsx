// Search.tsx

import React, { useState, useEffect, useRef } from "react";
import type { Set, Like } from "../../lib/types";
import levenshtein from "fast-levenshtein";
import Likes from "./Likes"; // Import the Likes component

import MetaIcons from "./MetaIcons";

interface SearchProps {
	sets: Set[];
	likedDJs: Like[]; // Add this prop
	handleLikedDJsChange: (updateFn: (prevLikedDJs: Like[]) => Like[]) => void;
}

const Search: React.FC<SearchProps> = ({
	sets,
	likedDJs,
	handleLikedDJsChange,
}) => {
	const [query, setQuery] = useState("");
	const [filteredSets, setFilteredSets] = useState<Set[]>([]);
	const inputRef = useRef<HTMLInputElement>(null);

	useEffect(() => {
		if (inputRef.current) {
			setTimeout(() => {
				inputRef.current?.focus();
			}, 300); // Delay to ensure it focuses after the component fully mounts
		}
		handleLikedDJsChange(() => {
			return likedDJs;
		});
	}, []);

	const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		const inputValue = event.target.value;
		setQuery(inputValue);
		const words = inputValue
			.trim()
			.replace(/[^a-zA-Z]/g, " ")
			.split(" ")
			.map((word) => word.toLowerCase());

		//			.filter(Boolean);

		const processedSets = sets
			.filter((set) => set.dj && set.dj !== "?") // Ensure set.dj is defined and not "?"
			.flatMap((set) => {
				// Split the dj into words, removing non-alphabetical characters (e.g., "Kos:mo" becomes ["kos", "mo"])
				const djWords = (set.dj as string)
					.split(/[^a-zA-Z]/)
					.map((word) => word.toLowerCase());

				// Generate adjacent pairs of words (if applicable) and combinations
				const adjacentPairs: string[] = [];
				for (let i = 0; i < djWords.length - 1; i++) {
					const word1 = djWords[i] || ""; // Ensure djWords[i] is defined
					const word2 = djWords[i + 1] || ""; // Ensure djWords[i + 1] is defined
					adjacentPairs.push(word1 + word2); // Concatenate adjacent words
				}

				// Combine all individual words, adjacent pairs, and the entire concatenation of words
				const allCombinations = [
					...adjacentPairs,
					djWords.join(""), // Concatenate all words (e.g., "kosmo")
					...djWords, // Include individual words (e.g., "kos", "mo")
				];

				// Remove duplicates using a Set and convert back to an array
				const uniqueCombinations = Array.from(new Set(allCombinations));

				// Create new sets for each searchField combination
				return uniqueCombinations.map((combination) => ({
					...set, // Keep the original set fields
					searchField: combination, // Add the searchField for each combination (lowercased)
				}));
			});

		if (inputValue.length > 0) {
			const matchedSetsWithDistance = processedSets.flatMap((set) =>
				words.flatMap((word) => {
					// Use searchField directly since it's a single word and already in lowercase
					const cleanedDjWord = set.searchField; // No need to convert to lowercase

					// Truncate cleanedDjWord to use only the first len(word) characters
					const truncatedDjWord = cleanedDjWord.slice(0, word.length);

					// Calculate the Levenshtein distance
					const distance = levenshtein.get(
						truncatedDjWord,
						word // No need for toLowerCase() since both are already lowercase
					);

					// Return the set and distance
					return {
						set, // The matched set
						distance, // Levenshtein distance
					};
				})
			);

			// Sort the matched sets by their Levenshtein distance
			let sortedSets = matchedSetsWithDistance.sort(
				(a, b) => a.distance - b.distance
			);

			// Array to keep track of unique DJ names (simple comparison, no Levenshtein needed)
			const uniqueDJs: string[] = [];

			// Filter out exact duplicate results
			sortedSets = sortedSets.filter((item) => {
				// Check if this DJ name is already in the uniqueDJs list (case-insensitive check)
				const isDuplicate = uniqueDJs.includes(item.set.dj.toLowerCase());
				if (!isDuplicate) {
					// If no exact match is found, add this DJ to the unique list
					uniqueDJs.push(item.set.dj.toLowerCase());
					return true; // Keep this result
				}
				return false; // Discard this result if it's a duplicate
			});

			// Extract the sets from the sorted array and take only the first 6 results
			const sortedSetList = sortedSets
				.filter((item) => item.distance < 3)
				.map((item) => item.set);

			setFilteredSets(sortedSetList);
		} else {
			setFilteredSets([]);
		}
	};

	const isSetInFuture = (setTime: Date) => {
		const currentTime = new Date();
		return setTime > currentTime;
	};

	const isSetNow = (set: Set) => {
		const currentTime = new Date();
		return set.end > currentTime && set.start < currentTime;
	};

	return (
		<div className="search-container">
			<input
				type="text"
				value={query}
				onChange={handleInputChange}
				placeholder="Search for a set"
				className="search-input p-2 text-black bg-white border-2 border-[#715874] rounded-[4px] focus:outline-none focus:ring-0 mb-4"
				ref={inputRef} // Set inputRef to the input element
			/>
			{filteredSets.length > 0 && (
				<ul className="text-[18px] mb-4">
					{" "}
					{/* Added mb-4 here */}
					{filteredSets.map((set, index) => {
						const setTime = new Date(set.start);
						const setDay = setTime.toLocaleDateString("en-GB", {
							weekday: "short",
						});
						const setTimeStr = setTime.toLocaleTimeString("en-GB", {
							hour: "2-digit",
							minute: "2-digit",
						});

						let text =
							set.dj + ", " + setDay + " at " + setTimeStr + " in " + set.room;
						if (isSetNow(set)) {
							text =
								set.dj +
								" is playing now in " +
								set.room +
								" (started at " +
								setTimeStr +
								")";
						}
						return (
							<li key={index} className="search-result-item flex items-start">
								<span className="text-[20px] mr-2 flex-shrink-0">
									{isSetInFuture(set.end) ? "✅" : "🚫"}
								</span>
								<span className="text-[18px]">
									{text}
									<MetaIcons meta={set.meta} />
								</span>
							</li>
						);
					})}
				</ul>
			)}
			<div className="mt-2">
				{" "}
				{/* Changed from mt-8 to mt-2 */}
				<Likes likes={likedDJs} />
			</div>
		</div>
	);
};

export default Search;
