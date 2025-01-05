export function isLocalhost(): boolean {
	return window.location.hostname === "localhost";
}

export function getApiURL(endpoint: string): string {
	return isLocalhost()
		? `http://localhost:8082/${endpoint}`
		: `${window.location.origin}/${endpoint}`;
}

export function getLikesURL(endpoint: string): string {
	return isLocalhost()
		? `http://localhost:8897/${endpoint}`
		: `${window.location.origin}/${endpoint}`;
}
