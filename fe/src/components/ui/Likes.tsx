import React from "react";
import { Like } from "../../lib/types";

interface LikesProps {
	likes: Like[];
}

const Likes: React.FC<LikesProps> = ({ likes }) => {
	if (likes.length === 0) {
		return null; // Return nothing if there are no likes
	}

	return (
		<div className="likes-container">
			<h2 className="text-[25px] mb-3">
				<span className="text-[18px]">❤️</span> Likes
			</h2>
			<ul className="text-[18px] mb-4">
				{likes.map((like, index) => (
					<li key={index} className="likes-items">
						🤍 {like.dj} {like.room}
						{like.title !== "Sisyphos" && ` - ${like.title}`} -{" "}
						{formatDate(like.started)}
					</li>
				))}
			</ul>
		</div>
	);
};

const formatDate = (date: Date): string => {
	const days = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
	const dayName = days[date.getDay()];
	const day = date.getDate().toString().padStart(2, "0");
	const month = (date.getMonth() + 1).toString().padStart(2, "0");
	const year = date.getFullYear().toString().slice(-2); // Get last two digits of the year
	return `${dayName} ${day}/${month}/${year}`;
};

export default Likes;
