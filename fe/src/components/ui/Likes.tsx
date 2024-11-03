import React from "react";
import { Like } from "../../lib/types";

import { formatDate, formatDayAndTime } from "../../lib/time"; // Import the time functions

import InstagramIcon from "@/assets/icon-instagram.png";
import SoundcloudIcon from "@/assets/icon-soundcloud.png";
import SpotifyIcon from "@/assets/icon-spotify.png";

interface LikesProps {
	likes: Like[];
}

const Likes: React.FC<LikesProps> = ({ likes }) => {
	if (likes.length === 0) {
		return null; // Return nothing if there are no likes
	}

	/*
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
*/
	const renderDJName = (like: Like) => {
		if (like.meta === null || like.meta.length === 0) {
			return <span className="dj-name">{like.dj}</span>;
		}

		const elements: (JSX.Element | string)[] = [];
		let djText = like.dj; // Start with the full DJ text

		// Store links and their corresponding names in a map
		const linksMap: { [key: string]: string } = {};
		like.meta.forEach((metaItem) => {
			if (metaItem.key.startsWith("dj.link.sisyfan.")) {
				//const djName = metaItem.key.split(".").pop()!;
				const djName = metaItem.key.split("dj.link.sisyfan.")[1];
				const link = metaItem.value;
				// Ensure djName is defined before assigning to linksMap
				if (djName && link) {
					linksMap[djName] = link;
				} else {
					console.warn(`Failed to parse DJ name from key: ${metaItem.key}`);
				}
			}
		});

		// Now we replace the DJ names with links using the linksMap
		const djNames = Object.keys(linksMap);
		const regex = new RegExp(`\\b(${djNames.join("|")})\\b`, "g");

		// Split the text into parts
		const parts = djText.split(regex);

		parts.forEach((part, index) => {
			if (part) {
				// Check if the part is a DJ name that we have a link for
				if (linksMap[part]) {
					elements.push(
						<a
							key={`${part}-${index}`}
							href={linksMap[part]}
							target="_blank"
							rel="noopener noreferrer"
							className="underline"
						>
							{part}
						</a>
					);
				} else {
					elements.push(part);
				}
			}
		});

		return <span className="dj-name">{elements}</span>;
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
									<div className="flex mt-2">
										{like.meta &&
											like.meta.map((metaItem) => {
												let icon;
												if (metaItem.key && metaItem.key !== "") {
													if (metaItem.key.startsWith("dj.link.soundcloud")) {
														icon = SoundcloudIcon;
													} else if (
														metaItem.key.startsWith("dj.link.spotify")
													) {
														icon = SpotifyIcon;
													} else if (
														metaItem.key.startsWith("dj.link.instagram")
													) {
														icon = InstagramIcon;
													}
												}

												return icon ? (
													<a
														key={metaItem.key}
														href={metaItem.value}
														target="_blank"
														rel="noopener noreferrer"
														className="ml-2 underline"
													>
														<img
															className="relative max-w-[64px] block overflow-hidden"
															src={icon}
															alt="icon"
														/>
													</a>
												) : null;
											})}
									</div>
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
