export type FunctionComponent = React.ReactElement | null;

export interface SetMeta {
	key: string;
	value: string;
}

export interface Set {
	dj: string;
	room: string;
	start: Date;
	end: Date;
	meta: SetMeta[] | null;
}

export interface Meta {
	aboutBigIcon: string;
	aboutShowShallowBunnyIcon: boolean;
	aboutShowSisyDuckIcon: boolean;
	botUrl: string;
	nowMapImage: string;
	nowShowDataSourceAd: boolean;
	nowShowShallowBunnyAd: boolean;
	nowShowSisyDuckAd: boolean;
	nowSubmitPR: string;
	nowTextAfterMap: string;
	nowTextWhenFinished: string;
	mobileAppName: string;
	prefix: string;
	rooms: string[];
	roomYouAreHereEmoticon: string;
	title: string;
	beginningSchedule: Date;
}

export interface Data {
	meta: Meta;
	sets: Set[];
}

export interface Like {
	dj: string;
	title: string;
	beginningSchedule: Date;
	room: string;
	started: Date;
	meta: SetMeta[] | null;
	links: [string]; // to remove
}
export interface RoomSituation {
	room: string;
	situation: string;
	like?: Like; // Optional because it may not always be available
	closed: boolean; // Boolean indicating whether the room is closed
}

export interface RoomSets {
	current: Set | null;
	next: Set | null;
	pauseDuration: number | null;
	closing: boolean;
}
