// Layout.tsx
import React, { useEffect, useState, Suspense, lazy } from "react";
import { Helmet, HelmetProvider } from "react-helmet-async";
import {
	Drawer,
	DrawerClose,
	DrawerContent,
	DrawerDescription,
	DrawerFooter,
	DrawerHeader,
	DrawerTitle,
	DrawerTrigger,
} from "@/components/ui/drawer";
import BunnyIcon from "@/assets/icon-bunny.png";
import DuckIcon from "@/assets/icon-duck.png";
import TelegramIcon from "@/assets/icon-telegram.png";
import TelegramBotIcon from "@/assets/icon-telegram-bot.png";

import GithubIcon from "@/assets/icon-github.png";

import { Now } from "@/components/ui/Now";
import { useQuery } from "@tanstack/react-query";
import { queryOptions } from "../lib/queryConfig";
import { Data } from "../lib/types";
import { allSetsInPastAndFinishedMoreThanOneHourAgo } from "../lib/setUtils"; // Import the function
import useCurrentMinute from "../hooks/useCurrentMinute"; // Import the custom hook
import { loadImageAsync } from "../lib/loadImage";

// Lazy load RoomPage
const RoomPage = lazy(() => import("./RoomPage"));

const Layout: React.FC = () => {
	const [showRoomPage, setShowRoomPage] = useState(false); // State to toggle between Now and RoomPage
	const [selectedRoom, setSelectedRoom] = useState<string>(""); // State for selected room
	const [imageSrc, setImageSrc] = useState<string | null>(null);
	const { data, error, isLoading } = useQuery<Data, Error>(queryOptions);
	const [isContentLoaded, setIsContentLoaded] = useState<boolean>(false);
	const [pageTitle, setPageTitle] = useState("Lineup app"); // State for page title
	const [appleTouchIcon, setAppleTouchIcon] = useState<string | null>(null); // State for apple-touch-icon
	const [appleMobileWebAppTitle, setAppleMobileWebAppTitle] =
		useState<string>("Lineup app"); // State for apple-mobile-web-app-title
	const [isStandalone, setIsStandalone] = useState<boolean>(false); // State for isStandalone
	const [areAllSetsInPast, setAreAllSetsInPast] = useState<boolean>(false); // State to track if all sets are in the past
	const currentMinute = useCurrentMinute(); // Use the custom hook

	useEffect(() => {
		const checkStandalone = () => {
			let standalone =
				window.matchMedia("(display-mode: standalone)").matches ||
				(window.navigator as any).standalone === true; // Use any type assertion to bypass TypeScript error

			const userAgent =
				navigator.userAgent || navigator.vendor || (window as any).opera;
			if (
				/android/i.test(userAgent) ||
				(/iPad|iPhone|iPod/.test(userAgent) && !(window as any).MSStream)
			) {
				standalone = standalone;
			} else {
				standalone = true;
			}
			setIsStandalone(standalone);
		};
		checkStandalone();
	}, []);

	useEffect(() => {
		if (data && data.meta && data.meta.aboutBigIcon) {
			// Load the icon image asynchronously
			loadImageAsync(data.meta.aboutBigIcon)
				.then((src) => setImageSrc(src))
				.catch((err) => console.error("Failed to load about icon image", err));
		}
	}, [data]);

	useEffect(() => {
		// Preload Drawer content images
		const preloadImages = [
			TelegramIcon,
			TelegramBotIcon,
			BunnyIcon,
			DuckIcon,
			GithubIcon,
		];
		const preloadImagePromises = preloadImages.map((src) => {
			const img = new Image();
			img.src = src;

			return new Promise((resolve) => {
				img.onload = resolve;
				img.onerror = resolve; // Resolve even if the image fails to load
			});
		});

		Promise.all(preloadImagePromises).then(() => {
			setIsContentLoaded(true);
		});
	}, []);

	useEffect(() => {
		if (data) {
			const appleIcon = `${data.meta.prefix}`;
			setAppleTouchIcon(appleIcon);
			const mobileWebAppTitle = data.meta.mobileAppName || "Lineup app";
			setAppleMobileWebAppTitle(mobileWebAppTitle);

			// Check if all sets are in the past and the last one finished more than 3 hours ago
			const allSetsPast = allSetsInPastAndFinishedMoreThanOneHourAgo(data.sets);
			if (!allSetsPast) {
				const pageTitle = data.meta.title || "Lineup app";
				setPageTitle(pageTitle);
			} else {
				setShowRoomPage(false); // show NOW
			}
			setAreAllSetsInPast(allSetsPast);

			// Set the initial selected room
			if (selectedRoom === "" && data?.sets?.length > 0) {
				const lastSet = data.sets[data.sets.length - 1];
				if (lastSet) {
					setSelectedRoom(lastSet.room);
				}
			}
		}
	}, [data]);

	if (isLoading) return <div className="bg-[#222123] fixed inset-0"></div>;
	if (error)
		return (
			<div className="bg-[#222123] fixed inset-0 text-white">
				When you download the app to your homescreen, you need to run it once
				with internet FIRST. then it will be able to work offline. Error:{" "}
				{error.message}
			</div>
		);
	if (!data || !data.sets)
		return (
			<div className="bg-[#222123] fixed inset-0 text-white">
				No data available
			</div>
		);

	return (
		<HelmetProvider>
			<div className="bg-[#222123] w-screen h-screen flex flex-col justify-center items-center text-slate-100 overflow-hidden relative">
				<Helmet>
					<title>{pageTitle}</title>
					{appleTouchIcon && (
						<link
							rel="apple-touch-icon"
							sizes="180x180"
							href={`${appleTouchIcon}-180x180.png`}
						/>
					)}
					{appleTouchIcon && (
						<link
							rel="icon"
							sizes="16x16"
							href={`${appleTouchIcon}-16x16.png`}
						/>
					)}
					{appleTouchIcon && (
						<link
							rel="icon"
							sizes="32x32"
							href={`${appleTouchIcon}-32x32.png`}
						/>
					)}
					{appleTouchIcon && (
						<link
							rel="icon"
							sizes="96x96"
							href={`${appleTouchIcon}-96x96.png`}
						/>
					)}
					<meta
						name="apple-mobile-web-app-title"
						content={appleMobileWebAppTitle}
					/>
				</Helmet>
				<div
					style={{
						height: isStandalone ? "89vh" : "80vh",
						top: isStandalone ? "0vh" : "1vh",
					}}
					className="w-full absolute"
				>
					{showRoomPage ? (
						<Suspense
							fallback={<div className="bg-[#222123] fixed inset-0"></div>}
						>
							<RoomPage
								data={data}
								isStandalone={isStandalone}
								currentMinute={currentMinute}
								selectedRoom={selectedRoom}
								setSelectedRoom={setSelectedRoom}
							/>
						</Suspense>
					) : (
						<Now
							data={data}
							isStandalone={isStandalone}
							allSetsInPast={areAllSetsInPast}
							currentMinute={currentMinute}
						/>
					)}
				</div>
				<div
					className={`flex mb-1 mt-2 w-full absolute ${
						isStandalone ? "top-[90vh]" : "top-[80vh]"
					}`}
				>
					<Drawer>
						<div className="flex w-full gap-1 mx-2">
							{!areAllSetsInPast && (
								<button
									className="text-white text-center text-[18px] w-full p-2 rounded-[4px] flex-grow"
									style={{ background: "#715874" }}
									onClick={() => setShowRoomPage(false)} // Set to show Now component
								>
									Now
								</button>
							)}
							{!areAllSetsInPast && (
								<button
									className="text-white text-center text-[18px] w-full p-2 rounded-[4px] flex-grow"
									style={{ background: "#715874" }}
									onClick={() => setShowRoomPage(true)} // Set to show RoomPage component
								>
									Stages
								</button>
							)}
							<DrawerTrigger
								className="text-white text-center text-[18px] w-full p-2 rounded-[4px] flex-grow"
								style={{ background: "#715874" }}
							>
								About
							</DrawerTrigger>
						</div>
						{isContentLoaded && (
							<DrawerContent>
								<DrawerHeader>
									<DrawerTitle> </DrawerTitle>
									<DrawerDescription>
										<div className="w-full m-auto px-4 flex flex-col items-center">
											{imageSrc && (
												<img
													className="relative max-w-[300px] m-auto block overflow-hidden rounded-2xl mb-4"
													src={imageSrc}
													alt="Duck with good vibes"
												/>
											)}
											<div className="flex justify-center items-center gap-4">
												<a
													href="https://t.me/shallowBunny"
													target="_blank"
													className="relative max-w-[64px] block overflow-hidden"
												>
													<img
														className="w-full"
														src={TelegramIcon}
														alt="Telegram"
													/>
												</a>
												<a
													href="https://github.com/shallowBunny/app"
													target="_blank"
													className="relative max-w-[64px] block overflow-hidden"
												>
													<img
														className="w-full"
														src={GithubIcon}
														alt="Github"
													/>
												</a>
												{data.meta.aboutShowShallowBunnyIcon && (
													<a
														href="https://shallowbunny.com"
														target="_blank"
														className="relative max-w-[64px] block overflow-hidden"
													>
														<img
															className="w-full"
															src={BunnyIcon}
															alt="bunny"
														/>
													</a>
												)}
												{data.meta.aboutShowSisyDuckIcon && (
													<a
														href="https://sisyduck.com"
														target="_blank"
														className="relative max-w-[64px] block overflow-hidden"
													>
														<img className="w-full" src={DuckIcon} alt="duck" />
													</a>
												)}
												{data.meta.nowBotUrl && (
													<a
														href={data.meta.nowBotUrl}
														target="_blank"
														className="relative max-w-[64px] block overflow-hidden"
													>
														<img
															className="w-full"
															src={TelegramBotIcon}
															alt="TelegramBot"
														/>
													</a>
												)}
											</div>
										</div>
									</DrawerDescription>
								</DrawerHeader>
								<DrawerFooter>
									<DrawerClose className="p-6 block">
										<div
											className="text-white p-2 px-10 rounded-[4px] mt-1 max-w-[120px] m-auto text-[18px]"
											style={{ background: "#715874" }}
										>
											OK
										</div>
									</DrawerClose>
								</DrawerFooter>
							</DrawerContent>
						)}
					</Drawer>
				</div>
			</div>
		</HelmetProvider>
	);
};

export default Layout;
