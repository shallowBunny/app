import React from "react";
import { Like } from "../../lib/types";
import { formatDate, formatDayAndTime } from "../../lib/time"; // Import the time functions

interface LikesProps {
	likes: Like[];
}

const Likes: React.FC<LikesProps> = ({ likes }) => {
	if (likes.length === 0) {
		return null; // Return nothing if there are no likes
	}

	const renderDJName = (like: Like) => {
		if (like.links && like.links.length > 0) {
			return (
				<a
					href={like.links[0]}
					target="_blank"
					rel="noopener noreferrer"
					className="underline"
				>
					{like.dj}
				</a>
			);
		}
		return <span className="dj-name">{like.dj}</span>;
	};

	let previousSchedule = ""; // Track the previous beginningSchedule

	return (
		<div className="likes-container">
			<ul className="text-[18px] mb-4">
				{likes.map((like, index) => {
					const currentSchedule = formatDate(like.beginningSchedule);

					// Compare current and previous schedule and render extra line when they differ
					const shouldRenderScheduleChange =
						previousSchedule !== currentSchedule;

					// Update the previousSchedule before rendering the extra line
					previousSchedule = currentSchedule;

					return (
						<React.Fragment key={index}>
							{/* Render the schedule change line when the schedule changes */}
							{shouldRenderScheduleChange && (
								<>
									<li className="separator-line" style={{ margin: "5px 0" }}>
										{" "}
									</li>
									<li className="likes-schedule-change  flex items-start">
										<span className="text-[20px] mr-2 flex-shrink-0">❤️</span>
										<span>
											{like.title} {currentSchedule}
										</span>
									</li>
									<li className="separator-line" style={{ margin: "5px 0" }}>
										{" "}
									</li>
								</>
							)}

							{/* Render the likes items */}
							<li className="likes-items flex items-start">
								<span className="text-[18px] mr-2 flex-shrink-0">🤍</span>
								<span>
									{renderDJName(like)}, {formatDayAndTime(like.started)} in{" "}
									{like.room}
								</span>
							</li>
						</React.Fragment>
					);
				})}
			</ul>
		</div>
	);
};

export default Likes;
