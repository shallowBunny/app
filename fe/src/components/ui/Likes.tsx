import React from "react";
import { Like } from "../../lib/types";

import { formatDate, formatDayAndTime } from "../../lib/time"; // Import the time functions

import MetaIcons from "./MetaIcons";

interface LikesProps {
	likes: Like[];
}

const Likes: React.FC<LikesProps> = ({ likes }) => {
	if (likes.length === 0) {
		return (
			<div className="likes-container">
				<ul className="text-[18px] mb-4">
					<li className="separator-line" style={{ margin: "5px 0" }}>
						{" "}
					</li>
					<li className="likes-schedule-change  flex items-start">
						<span className="text-[20px] mr-2 flex-shrink-0">ℹ️</span>
						<span>
							If you click the 🤍 buttons in the Now page, sets will be saved
							here!
						</span>
					</li>
					<li className="separator-line" style={{ margin: "5px 0" }}>
						{" "}
					</li>
				</ul>
			</div>
		);
	}

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
									{like.dj}, {formatDayAndTime(like.started)} in {like.room}
									<MetaIcons meta={like.meta} />
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
