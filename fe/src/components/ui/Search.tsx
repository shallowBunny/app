// Search.tsx

import React, { useState, useEffect, useRef } from "react";
import { Set } from "../../lib/types";
import levenshtein from "fast-levenshtein";

interface SearchProps {
	sets: Set[];
}

const Search: React.FC<SearchProps> = ({ sets }) => {
	const [query, setQuery] = useState("");
	const [filteredSets, setFilteredSets] = useState<Set[]>([]);
	const inputRef = useRef<HTMLInputElement>(null);

	useEffect(() => {
		if (inputRef.current) {
			setTimeout(() => {
				inputRef.current?.focus();
			}, 300); // Delay to ensure it focuses after the component fully mounts
		}
	}, []);

	const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		const inputValue = event.target.value;
		setQuery(inputValue);
		const words = inputValue.trim().split(" ").filter(Boolean);

		if (inputValue.length > 0) {
			// Build a new data structure that contains the matched set and its Levenshtein distance
			const matchedSetsWithDistance = sets.flatMap((set) =>
				words.flatMap((word) =>
					set.dj
						.split(" ")
						.map((djWord) => djWord.replace(/[^a-zA-Z]/g, ""))
						.map((cleanedDjWord) => {
							// Truncate cleanedDjWord to use only the first len(word) characters
							const truncatedDjWord = cleanedDjWord.slice(0, word.length);
							// Calculate the Levenshtein distance
							const distance = levenshtein.get(
								truncatedDjWord.toLowerCase(),
								word.toLowerCase()
							);
							// Return the set and distance (no need for filtering)
							return {
								set, // The matched set
								distance, // Levenshtein distance
							};
						})
				)
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
			const sortedSetList = sortedSets.slice(0, 6).map((item) => item.set);

			setFilteredSets(sortedSetList);
		} else {
			setFilteredSets([]);
		}
	};

	const isSetInFuture = (setTime: Date) => {
		const currentTime = new Date();
		return setTime > currentTime;
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
				<ul className="search-results">
					{filteredSets.map((set, index) => {
						const setTime = new Date(set.start);
						const setDay = setTime.toLocaleDateString("en-GB", {
							weekday: "short",
						});
						const setTimeStr = setTime.toLocaleTimeString("en-GB", {
							hour: "2-digit",
							minute: "2-digit",
						});

						return (
							<li key={index} className="search-result-item">
								{isSetInFuture(setTime) ? "✅" : "🚫"} {set.dj} {set.room} -{" "}
								{setDay} at {setTimeStr}
							</li>
						);
					})}
				</ul>
			)}
		</div>
	);
};

export default Search;
