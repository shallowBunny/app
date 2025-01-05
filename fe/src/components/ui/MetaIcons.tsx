import React from "react";

import SoundcloudIcon from "@/assets/icon-soundcloud.png";
import SpotifyIcon from "@/assets/icon-spotify.png";
import InstagramIcon from "@/assets/icon-instagram.png";
import BandcampIcon from "@/assets/icon-bandcamp.png";
import WWWIcon from "@/assets/icon-www.png";

export interface SetMeta {
	key: string;
	value: string;
}

interface MetaIconsProps {
	meta: SetMeta[] | null;
	roomPage?: boolean;
}

// Helper function to get the icon for a given key
const getIconForKey = (key: string): string | undefined => {
	if (key.startsWith("dj.link.soundcloud")) return SoundcloudIcon;
	if (key.startsWith("dj.link.spotify")) return SpotifyIcon;
	if (key.startsWith("dj.link.instagram")) return InstagramIcon;
	if (key.startsWith("dj.link.bandcamp")) return BandcampIcon;
	if (key.startsWith("dj.link.www")) return WWWIcon;
	return undefined;
};

const MetaIcons: React.FC<MetaIconsProps> = ({ meta, roomPage = false }) => {
	const iconSizeClass = roomPage ? "max-w-[16px]" : "max-w-[64px]"; // Smaller size for no div

	if (!meta) return null;
	const validMetaItems = meta.filter(
		(metaItem) => metaItem.key && getIconForKey(metaItem.key)
	);
	if (validMetaItems.length === 0) return null;

	const icons = validMetaItems.map((metaItem) => {
		const icon = getIconForKey(metaItem.key!);

		return icon ? (
			<a
				key={metaItem.key}
				href={metaItem.value}
				target="_blank"
				rel="noopener noreferrer"
				className="ml-2 underline"
			>
				<img
					className={`relative block overflow-hidden ${iconSizeClass}`}
					src={icon}
					alt="icon"
				/>
			</a>
		) : null;
	});

	// Conditionally wrap in a <div>
	return roomPage ? (
		<div className="flex items-center">{icons}</div>
	) : (
		<div className="flex mt-2 mb-1 -ml-2">{icons}</div>
	);
};

export default MetaIcons;
