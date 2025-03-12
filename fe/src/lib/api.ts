export function isLocalhost(): boolean {
	return window.location.hostname === "localhost";
}

export function getApiURL(endpoint: string): string {
	const localhostPort = endpoint === "api" ? 8082 : 8897;
	return isLocalhost()
		? `http://localhost:${localhostPort}/${endpoint}`
		: `${window.location.origin}/${endpoint}`;
}
