// Now.tsx
import { useState, useEffect, FunctionComponent } from "react";
import {
	findCurrentAndNextSets,
	getOverriddenCurrentTime,
	convertRoomSetsToRoomSituation,
} from "../../lib/utils";
import { Data, Like } from "../../lib/types"; // Assuming `Like` is defined in types

import { loadImageAsync } from "../../lib/loadImage";

interface NowProps {
	data: Data;
	isStandalone: boolean;
	allSetsInPast: boolean;
	currentMinute: Date;
	likedDJs: Like[]; // Modify likedDJs to an array of Like objects
	setLikedDJs: React.Dispatch<React.SetStateAction<Like[]>>; // Modify setLikedDJs to update an array of Like objects
}

export const Now: FunctionComponent<NowProps> = ({
	data,
	isStandalone,
	allSetsInPast,
	likedDJs,
	setLikedDJs,
}) => {
	const [mapImageSrc, setMapImageSrc] = useState<string | null>(null); // State to store the loaded image URL
	const overriddenNow = getOverriddenCurrentTime();
	const roomSets = findCurrentAndNextSets(data.sets, overriddenNow);
	const roomSituations = convertRoomSetsToRoomSituation(
		roomSets,
		data.meta
	).reverse();

	useEffect(() => {
		if (data.meta.nowMapImage) {
			loadImageAsync(data.meta.nowMapImage)
				.then((src) => setMapImageSrc(src))
				.catch((err) => console.error("Failed to load map image", err));
		}
	}, [data.meta.nowMapImage]);

	const vh = isStandalone ? 89.5 : 78.0;
	const height = `${vh}vh`; // Subtracting 2.6vh as per your original code

	const toggleLike = (like: Like) => {
		setLikedDJs((prevLikedDJs) => {
			// Check if the Like already exists in the likedDJs array
			const isLiked = prevLikedDJs.some(
				(existingLike) =>
					existingLike.dj === like.dj &&
					existingLike.beginningSchedule.getTime() ===
						like.beginningSchedule.getTime()
			);

			if (isLiked) {
				// If the DJ is already liked, remove the Like
				return prevLikedDJs.filter(
					(existingLike) =>
						existingLike.dj !== like.dj ||
						existingLike.beginningSchedule.getTime() !==
							like.beginningSchedule.getTime()
				);
			} else {
				// If the DJ is not liked, add the new Like object
				return [like, ...prevLikedDJs];
			}
		});
	};

	// If the condition is met, render the error message
	if (allSetsInPast && data.meta.nowTextWhenFinished) {
		return (
			<div className="bg-[#222123] fixed inset-0 flex justify-center items-center">
				<div className="bg-[#2e2c2f] p-8 rounded-lg shadow-lg text-white max-w-lg text-center">
					<h1 className="text-2xl font-bold mb-4">
						{data.meta.roomYouAreHereEmoticon}
					</h1>
					<p className="text-white text-center text-[18px] w-full p-2">
						{data.meta.nowTextWhenFinished}
					</p>
				</div>
			</div>
		);
	}

	return (
		<div className="bg-[#222123] rounded-md px-4 pr-2 py-2 text-[22px] leading-7">
			<ul className="w-full overflow-y-scroll" style={{ height }}>
				{!allSetsInPast &&
					roomSituations.map((situation, index) => {
						const { like } = situation; // Extract the like object

						return (
							<li key={index} className="mb-4 flex items-center">
								<span>{situation.situation}</span>
								{/* Show heart and make it clickable */}
								{like && (
									<span
										className="ml-2 cursor-pointer"
										onClick={() => toggleLike(like)}
									>
										{likedDJs.some(
											(existingLike) =>
												existingLike.dj === like.dj &&
												existingLike.beginningSchedule.getTime() ===
													like.beginningSchedule.getTime()
										)
											? "❤️"
											: "🤍"}{" "}
										{/* Show filled or empty heart */}
									</span>
								)}
							</li>
						);
					})}
				<div className="text-[18px]">
					{!isStandalone && (
						<li key="infowpa" className="mb-4">
							This website should be able to work without internet if you keep
							its window open, but the best/safest solution is to add it to your
							homescreen:{" "}
							<a
								href="https://www.howtogeek.com/196087/how-to-add-websites-to-the-home-screen-on-any-smartphone-or-tablet/"
								target="_blank"
								className="underline cursor-pointer"
							>
								more info on how to do that on iOS and Android.
							</a>
						</li>
					)}
					{!allSetsInPast && data.meta.nowMapImage && mapImageSrc && (
						<li key="map-image" className="mb-4 text-[18px]">
							<img
								className="relative max-w-[344px] block overflow-hidden rounded-2xl mb-4"
								src={mapImageSrc}
								alt="map"
							/>
						</li>
					)}

					{data.meta.nowSubmitPR && (
						<li key="infowpa" className="mb-4">
							You can submit data{" "}
							<a
								href={data.meta.nowSubmitPR}
								target="_blank"
								className="underline cursor-pointer"
							>
								there
							</a>
							... 😘
						</li>
					)}

					{!allSetsInPast && data.meta.nowTextAfterMap && (
						<li key="next-message-little" className="mb-4">
							{data.meta.nowTextAfterMap}
						</li>
					)}

					{data.meta.nowShowSisyDuckAd && (
						<li key="info" className="mb-4">
							🦆 Webapp also running for{" "}
							<a
								href="http://sisyduck.com"
								target="_blank"
								className="underline cursor-pointer"
							>
								Sisyphos
							</a>{" "}
							🦆
						</li>
					)}
					{!allSetsInPast && data.meta.nowShowDataSourceAd && (
						<li key="info2" className="mb-4">
							🦆 Data source:{" "}
							<a
								href="https://t.me/sisyphosclub"
								target="_blank"
								className="underline cursor-pointer"
							>
								Sisy Telegram group
							</a>{" "}
							🦆
						</li>
					)}

					{data.meta.nowShowShallowBunnyAd && (
						<li key="festival" className="mb-4">
							🌲 Webapp also running for{" "}
							<a
								href="https://shallowbunny.com"
								target="_blank"
								className="underline cursor-pointer"
							>
								festivals
							</a>{" "}
							🌲
						</li>
					)}
				</div>
			</ul>
		</div>
	);
};
