export type FunctionComponent = React.ReactElement | null;

export interface Set {
	dj: string;
	room: string;
	start: Date;
	end: Date;
}

export interface Meta {
	nowShowShallowBunnyAd: boolean;
	nowShowDataSourceAd: boolean;
	nowShowSisyDuckAd: boolean;
	nowTextAfterMap: string;
	nowTextWhenFinished: string;
	nowBotUrl: string;
	aboutBigIcon: string;
	aboutShowShallowBunnyIcon: boolean;
	aboutShowSisyDuckIcon: boolean;
	nowMapImage: string;
	roomYouAreHereEmoticon: string;
	mobileAppName: string;
	prefix: string;
	title: string;
}

export interface Data {
	sets: Set[];
	meta: Meta;
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
