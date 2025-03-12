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
import BunnyIcon from "@/assets/icon-bunny.webp";
import DuckIcon from "@/assets/icon-duck.webp";
import TelegramIcon from "@/assets/icon-telegram.webp";
import TelegramBotIcon from "@/assets/icon-telegram-bot.webp";
import PatreonIcon from "@/assets/icon-patreon.webp";

import GithubIcon from "@/assets/icon-github.webp";
import { Now } from "@/components/ui/Now";
import { useQuery } from "@tanstack/react-query";
import { getQueryOptions } from "../lib/queryConfig";
import { Data, Like } from "../lib/types";
import useCurrentMinute from "../hooks/useCurrentMinute";
import { loadImageAsync } from "../lib/loadImage";
import { PacmanLoader } from "react-spinners";
import { postUpdatedLikes, parseAndMigrateLikedDJs } from "../lib/likesService";
import { isLocalhost } from "../lib/api";
import { useRouter } from "@tanstack/react-router";
import { getVhForAllTabs } from "../lib/utils";

// Lazy load RoomPage
const RoomPage = lazy(() => import("./RoomPage"));

type LayoutProps = {
	festival?: string;
	stage?: string;
};

//const Layout: React.FC = () => {
const Layout: React.FC<LayoutProps> = ({ festival, stage }) => {
	const router = useRouter();
	const [selectedRoom, setSelectedRoom] = useState<string>(""); // State for selected room

	useEffect(() => {
		console.log("Layout re-rendered");
	}, []); // Empty dependency array ensures it runs only on mount

	const getBaseUrl = () => {
		if (!festival || festival === "now") {
			return "/";
		}
		return `/lineup/${festival}/${selectedRoom}`;
	};

	//data?.meta?.part

	const [currentPage, setCurrentPage] = useState<"now" | "rooms">(
		!stage || stage === "now" ? "now" : "rooms"
	);

	const navigateTo = (page: "now" | "rooms") => {
		console.log("navigateTo" + page + " " + getBaseUrl());
		if (page === "now") {
			festival = undefined;
			setCurrentPage("now");
			router.history.replace(`${getBaseUrl()}`); // Navigate dynamically
		} else {
			setCurrentPage("rooms");
			router.history.replace(`${getBaseUrl()}`);
		}
	};

	useEffect(() => {
		if (selectedRoom !== "") {
			if (festival) {
				router.history.push(`/lineup/${festival}/${selectedRoom}`);
			} else {
				router.history.push("/"); // Navigate to home if no festival
			}
			console.log("set router:" + selectedRoom);
		} else {
			console.log("skip set router selectedRoom.");
		}
	}, [selectedRoom, festival, history]);

	const [imageSrc, setImageSrc] = useState<string | null>(null);
	const { data, error, isLoading } = useQuery<Data, Error>(
		getQueryOptions(festival)
	);

	// Update selectedRoom when stage or data.meta.rooms changes
	useEffect(() => {
		// Convert stage to a number if it's a string
		const stageIndex = stage ? parseInt(stage) : undefined;

		if (data?.meta?.rooms && stageIndex !== undefined && stageIndex >= 0) {
			if (stageIndex < data.meta.rooms.length) {
				setSelectedRoom(String(stageIndex)); // Set the selected room based on the converted index
			} else {
				setSelectedRoom("0");
			}
		} else {
			if (stage === "search") {
				setSelectedRoom("search");
			} else {
				if (stage === undefined) {
					setSelectedRoom("0");
				}
			}
		}
	}, [stage, data?.meta.rooms]); // Reacts to changes in stage or rooms

	const printPartyName = festival ? true : false;

	const [isContentLoaded, setIsContentLoaded] = useState<boolean>(false);
	const [pageTitle, setPageTitle] = useState("Lineup app"); // State for page title
	const [appleTouchIcon, setAppleTouchIcon] = useState<string | null>(null); // State for apple-touch-icon
	const [appleMobileWebAppTitle, setAppleMobileWebAppTitle] =
		useState<string>("Lineup app");
	const [isRunningAsWPA, setIsRunningAsWPA] = useState<boolean>(false); // State for isRunningAsWPA
	const [isDesktop, setIsDesktop] = useState<boolean>(false);

	const currentMinute = useCurrentMinute(); // Use the custom hook
	const [showLoadingModal, setShowLoadingModal] = useState(true); // New state for loading modal

	// Add likedDJs state
	const [likedDJs, setLikedDJs] = useState<Like[]>([]);

	const updateLikesWithServer = async (updatedLikedDJs: Like[]) => {
		try {
			const likes = await postUpdatedLikes(updatedLikedDJs);
			setLikedDJs(likes);
		} catch (error) {
			// Handle errors
			console.error("Error fetching data:", error);
		}
	};

	//	const handleLikedDJsChange = async (updatedLikedDJs: Like[]) => {
	const handleLikedDJsChange = async (
		updateFn: (prevLikedDJs: Like[]) => Like[]
	) => {
		// Get the updated liked DJs
		const updatedLikedDJs = updateFn(likedDJs);

		console.log("handleLikedDJsChange");
		// Update the state with the updated liked DJs
		setLikedDJs(updatedLikedDJs);
		updateLikesWithServer(updatedLikedDJs);
	};

	// Save likedDJs to localStorage whenever likedDJs changes
	useEffect(() => {
		if (Object.keys(likedDJs).length > 0) {
			localStorage.setItem("likedDJs", JSON.stringify(likedDJs));
		}
	}, [likedDJs]);

	useEffect(() => {
		const checkRunningAsWPA = () => {
			let standalone =
				window.matchMedia("(display-mode: standalone)").matches ||
				(window.navigator as any).standalone === true;

			// Load likedDJs from localStorage on mount
			const storedLikedDJs = localStorage.getItem("likedDJs");
			if (storedLikedDJs) {
				try {
					const likesNoDates = JSON.parse(storedLikedDJs) as Like[];
					const parsedLikedDJs = parseAndMigrateLikedDJs(likesNoDates);
					setLikedDJs(parsedLikedDJs);
					if (standalone || isLocalhost()) {
						updateLikesWithServer(parsedLikedDJs);
					}
				} catch (e) {
					console.error("Failed to parse likedDJs from localStorage:", e);
				}
			}
			// force standalone to true if not running on mobile
			const userAgent =
				navigator.userAgent || navigator.vendor || (window as any).opera;
			if (
				/android/i.test(userAgent) ||
				(/iPad|iPhone|iPod/.test(userAgent) && !(window as any).MSStream)
			) {
				standalone = standalone;
				setIsDesktop(false);
			} else {
				standalone = true;
				setIsDesktop(true);
			}
			setIsRunningAsWPA(standalone);
		};

		checkRunningAsWPA();
	}, []);

	useEffect(() => {
		if (data && data.meta && data.meta.aboutBigIcon) {
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
			PatreonIcon,
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
		if (!isLoading && data) {
			setShowLoadingModal(false); // Hide modal once data is loaded
		}
	}, [isLoading, data]);

	useEffect(() => {
		if (data) {
			const appleIcon = `${data.meta.prefix}`;
			setAppleTouchIcon(appleIcon);
			const mobileWebAppTitle = data.meta.mobileAppName || "Lineup app";
			setAppleMobileWebAppTitle(mobileWebAppTitle);

			const pageTitle = data.meta.title + " lineup";
			setPageTitle(pageTitle);
		}
	}, [data]);

	if (error) {
		setTimeout(() => {
			router.navigate({ to: "/" });
		}, 1000);
		return (
			<div className="bg-[#222123] fixed inset-0 flex justify-center items-center">
				<div className="bg-[#2e2c2f] p-8 rounded-lg shadow-lg text-white max-w-lg text-center">
					<h1 className="text-2xl font-bold mb-4">☠️</h1>
					<p className="mb-4">{error.message}</p>
				</div>
			</div>
		);
	}

	if (showLoadingModal) {
		return (
			<div className="fixed inset-0 flex items-center justify-center bg-opacity-80 z-50">
				<PacmanLoader color="yellow" size={20} />{" "}
			</div>
		);
	}

	if (!data || !data.sets) {
		return (
			<div className="bg-[#222123] fixed inset-0 text-white">
				No data available
			</div>
		);
	}
	const vhForAllTabs = getVhForAllTabs(isRunningAsWPA, isDesktop);
	return (
		<HelmetProvider>
			<div className="bg-[#222123] w-screen h-screen flex flex-col justify-center items-center text-slate-100 overflow-hidden relative">
				<Helmet>
					<title>{pageTitle}</title>
					{appleTouchIcon && (
						<link
							rel="apple-touch-icon"
							sizes="180x180"
							href={`${appleTouchIcon}-180x180.webp`}
						/>
					)}
					{appleTouchIcon && (
						<link
							rel="icon"
							sizes="16x16"
							href={`${appleTouchIcon}-16x16.webp`}
						/>
					)}
					{appleTouchIcon && (
						<link
							rel="icon"
							sizes="32x32"
							href={`${appleTouchIcon}-32x32.webp`}
						/>
					)}
					{appleTouchIcon && (
						<link
							rel="icon"
							sizes="96x96"
							href={`${appleTouchIcon}-96x96.webp`}
						/>
					)}
					<meta
						name="apple-mobile-web-app-title"
						content={appleMobileWebAppTitle}
					/>
					<meta
						name="description"
						content={`this website shows the lineup for ${data.meta.title}`}
					/>
				</Helmet>
				<div
					style={{
						height: isRunningAsWPA ? "89vh" : "80vh",
						top: isRunningAsWPA ? "0vh" : "1vh",
					}}
					className="w-full absolute"
				>
					{currentPage === "rooms" ? (
						<Suspense
							fallback={<div className="bg-[#222123] fixed inset-0"></div>}
						>
							<RoomPage
								data={data}
								isRunningAsWPA={isRunningAsWPA}
								isDesktop={isDesktop}
								currentMinute={currentMinute}
								selectedRoom={selectedRoom}
								setSelectedRoom={setSelectedRoom} // Pass setSelectedRoom to RoomPage
								likedDJs={likedDJs} // Pass likedDJs only to RoomPage
								handleLikedDJsChange={handleLikedDJsChange}
								printPartyName={printPartyName}
							/>
						</Suspense>
					) : (
						<Now
							data={data}
							isRunningAsWPA={isRunningAsWPA}
							isDesktop={isDesktop}
							currentMinute={currentMinute}
							likedDJs={likedDJs} // Pass likedDJs to Now
							handleLikedDJsChange={handleLikedDJsChange}
						/>
					)}
				</div>
				<div
					className="flex mb-1 mt-2 w-full absolute"
					style={{ top: `${vhForAllTabs}vh` }}
				>
					<Drawer>
						<div className="flex w-full gap-1 mx-2">
							<button
								className="text-white text-center text-[18px] w-full p-2 rounded-[4px] flex-grow"
								style={{ background: "#715874" }}
								onClick={() => navigateTo("now")} // Set to show Now component
							>
								Now
							</button>
							<button
								className="text-white text-center text-[18px] w-full p-2 rounded-[4px] flex-grow"
								style={{ background: "#715874" }}
								onClick={() => navigateTo("rooms")} // Set to show RoomPage component
							>
								Stages
							</button>
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
														title="Festivals lineup"
													>
														<img
															className="w-full"
															src={BunnyIcon}
															alt="Festivals lineup"
														/>
													</a>
												)}
												{data.meta.aboutShowSisyDuckIcon && (
													<a
														href="https://sisyduck.com"
														target="_blank"
														className="relative max-w-[64px] block overflow-hidden"
														title="Sisyphos lineup"
													>
														<img
															className="w-full"
															src={DuckIcon}
															alt="Sisyphos lineup"
														/>
													</a>
												)}
												{data.meta.botUrl && (
													<a
														href={data.meta.botUrl}
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
												{data.meta.aboutShowPatreonIcon && (
													<a
														href="https://www.patreon.com/shallowBunny"
														target="_blank"
														className="relative max-w-[64px] block overflow-hidden"
														title="Patreon"
													>
														<img
															className="w-full"
															src={PatreonIcon}
															alt="Patreon"
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
