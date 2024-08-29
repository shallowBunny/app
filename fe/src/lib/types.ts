export type FunctionComponent = React.ReactElement | null;

export interface Set {
	dj: string;
	room: string;
	start: Date;
	end: Date;
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
}

export interface Data {
	meta: Meta;
	sets: Set[];
}
export interface RoomSituation {
	room: string;
	situation: string;
}

export interface RoomSets {
	current: Set | null;
	next: Set | null;
	pauseDuration: number | null;
	closing: boolean;
}
