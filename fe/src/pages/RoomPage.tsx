// RoomPage.tsx
// TODO refactor with a better data structure

import { FunctionComponent } from "react";
import * as Tabs from "@radix-ui/react-tabs";
import Room from "@/components/ui/Room";
import clsx from "clsx";
import { Data } from "../lib/types";
import { extractEmoticons } from "../lib/emoji"; // Import the function
import Search from "@/components/ui/Search"; // Import the Search component

interface RoomPageProps {
	data: Data;
	isStandalone: boolean;
	currentMinute: Date; // Add the currentMinute prop
	selectedRoom: string; // Add selectedRoom prop
	setSelectedRoom: (room: string) => void; // Add setSelectedRoom prop
}

const RoomPage: FunctionComponent<RoomPageProps> = ({
	data,
	isStandalone,
	currentMinute,
	selectedRoom,
	setSelectedRoom,
}) => {
	if (!data || !data.sets) {
		return (
			<div className="bg-[#222123] fixed inset-0 text-white">
				No data available
			</div>
		);
	}

	// Sort the days chronologically
	const rooms = data.sets.map((set) => set.room); // Map over the remaining sets
	const uniqueRooms = Array.from(new Set(rooms)).reverse();
	const iconsAreSmall = uniqueRooms.length > 10;
	const youarehere = data.meta.roomYouAreHereEmoticon;

	// Change this to expand the tabs vertically
	let vhForAllTabs = 79.0;
	if (isStandalone) {
		vhForAllTabs = 90.0;
	}

	const vhByTab = vhForAllTabs / uniqueRooms.length;

	return (
		<div className="w-full px-2">
			<div>
				<Tabs.Root
					className="flex flex-row w-full"
					value={selectedRoom} // Use selectedRoom as the current tab
					onValueChange={setSelectedRoom} // Update selectedRoom on tab change
				>
					<Tabs.List
						className="flex flex-col w-[40%] max-w-[100px] overflow-y-auto"
						style={{ maxHeight: `${vhForAllTabs}vh` }}
					>
						<Tabs.Trigger
							className="rounded-l-3xl text-[13px] bg-[#222123] px-5 py-2 flex-1 flex items-center justify-center text-mauve11 select-none hover:opacity-80 data-[state=active]:opacity-100 data-[state=active]:bg-[#353044] outline-none cursor-pointer opacity-50 leading-5 min-h-[48px]"
							style={{
								height: `${vhByTab}vh`,
							}}
							value="search"
						>
							<span className="text-3xl">🔍</span>
						</Tabs.Trigger>
						{uniqueRooms.map((room, i) => {
							const emoticon = extractEmoticons(room);
							return (
								<Tabs.Trigger
									key={i}
									className="rounded-l-3xl text-[13px] bg-[#222123] px-5 py-2 flex-1 flex items-center justify-center text-mauve11 select-none hover:opacity-80 data-[state=active]:opacity-100 data-[state=active]:bg-[#353044] outline-none cursor-pointer opacity-50 leading-5 min-h-[48px]"
									style={{
										height: `${vhByTab}vh`,
									}}
									value={room}
								>
									<span
										className={clsx(
											"flex items-center justify-center w-full h-full",
											{ "text-3xl": iconsAreSmall },
											{ "text-5xl": !iconsAreSmall }
										)}
									>
										{emoticon}
									</span>
								</Tabs.Trigger>
							);
						})}
					</Tabs.List>
					{uniqueRooms.map((room, i) => {
						return (
							<Tabs.Content
								className={clsx(
									"grow p-5 bg-[#353044] rounded-b-md outline-none w-[66%] relative overflow-y-scroll rounded-r-3xl",
									{ "rounded-tl-3xl": room != uniqueRooms[0] },
									{ "rounded-bl-3xl": room != uniqueRooms.slice(-1)[0] }
								)}
								style={{ height: `${vhForAllTabs}vh` }}
								value={room}
								key={i}
							>
								<div className="absolute top-5">
									<Room
										youarehere={youarehere}
										room={room}
										sets={data.sets.filter(
											(set) => set.room === room && set.dj !== "?"
										)}
										currentMinute={currentMinute}
									/>
									<div className="h-20"></div>
								</div>
							</Tabs.Content>
						);
					})}
					<Tabs.Content
						className="grow p-5 bg-[#353044] rounded-b-md outline-none w-[66%] relative overflow-y-scroll rounded-r-3xl"
						style={{ height: `${vhForAllTabs}vh` }}
						value="search"
					>
						<div className="absolute top-5">
							<Search sets={data.sets} />
						</div>
					</Tabs.Content>
				</Tabs.Root>
			</div>
		</div>
	);
};

export default RoomPage;
