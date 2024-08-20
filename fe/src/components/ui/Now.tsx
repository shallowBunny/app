// Now.tsx
import { useState, useEffect, FunctionComponent } from "react";
import {
	findCurrentAndNextSets,
	getOverriddenCurrentTime,
	convertRoomSetsToRoomSituation,
} from "../../lib/utils";
import { Data } from "../../lib/types";

import { loadImageAsync } from "../../lib/loadImage";

interface NowProps {
	data: Data;
	isStandalone: boolean;
	allSetsInPast: boolean;
	currentMinute: Date; // Add the currentMinute prop
}

export const Now: FunctionComponent<NowProps> = ({
	data,
	isStandalone,
	allSetsInPast,
}) => {
	const [mapImageSrc, setMapImageSrc] = useState<string | null>(null); // State to store the loaded image URL
	const overriddenNow = getOverriddenCurrentTime();
	const roomSets = findCurrentAndNextSets(data.sets, overriddenNow);
	const roomSituations = convertRoomSetsToRoomSituation(roomSets).reverse();

	useEffect(() => {
		if (data.meta.nowMapImage) {
			loadImageAsync(data.meta.nowMapImage)
				.then((src) => setMapImageSrc(src))
				.catch((err) => console.error("Failed to load map image", err));
		}
	}, [data.meta.nowMapImage]);

	const vh = isStandalone ? 89.5 : 78.0;
	const height = `${vh}vh`; // Subtracting 2.6vh as per your original code
	return (
		<div className="bg-[#222123] rounded-md px-4 pr-2 py-2 text-[22px] leading-7">
			<ul className="w-full overflow-y-scroll" style={{ height }}>
				{" "}
				{allSetsInPast && data.meta.nowTextWhenFinished && (
					<li key="next-message" className="mb-4 text-[26px]">
						{data.meta.nowTextWhenFinished}
					</li>
				)}
				{!allSetsInPast &&
					roomSituations.map((situation, index) => (
						<li key={index} className="mb-4">
							{situation.situation}
						</li>
					))}
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

					{data.meta.nowShowPleaseSendData && (
						<li key="infowpa" className="mb-4">
							Lineup data for this event is not available yet, when you have
							access to it, please share a picture in this{" "}
							<a
								href="https://t.me/shallowBunny"
								target="_blank"
								className="underline cursor-pointer"
							>
								Telegram group
							</a>
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
					{data.meta.nowBotUrl && (
						<li key="unique-key-bot-url" className="mb-4">
							🤖🤖 There's also a{" "}
							<a
								href={data.meta.nowBotUrl}
								target="_blank"
								className="underline cursor-pointer"
							>
								Telegram bot
							</a>{" "}
							🤖🤖
						</li>
					)}
				</div>
			</ul>
		</div>
	);
};
