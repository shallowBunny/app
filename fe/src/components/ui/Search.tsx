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
		const words = inputValue.split(" ");
		if (inputValue.length > 0) {
			const matchedSets = sets.filter((set) =>
				words.some((word) =>
					set.dj
						.split(" ")
						.map((djWord) => djWord.replace(/[^a-zA-Z]/g, ""))
						.some(
							(cleanedDjWord) =>
								levenshtein.get(
									cleanedDjWord.toLowerCase(),
									word.toLowerCase()
								) <= 2
						)
				)
			);

			// Sort the results by Levenshtein distance
			const sortedSets = matchedSets.sort((a, b) => {
				const distanceA = levenshtein.get(
					a.dj.toLowerCase(),
					inputValue.toLowerCase()
				);
				const distanceB = levenshtein.get(
					b.dj.toLowerCase(),
					inputValue.toLowerCase()
				);
				return distanceA - distanceB;
			});
			setFilteredSets(sortedSets);
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
