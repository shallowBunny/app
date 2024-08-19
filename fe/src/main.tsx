import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import { createRouter } from "@tanstack/react-router";
import { routeTree } from "./routeTree.gen";
import "./styles/tailwind.css";
import { registerSW } from "virtual:pwa-register";
const router = createRouter({ routeTree });

declare module "@tanstack/react-router" {
	interface Register {
		router: typeof router;
	}
}

// Conditionally register the service worker for PWA
if (window.location.hostname !== "localhost") {
	registerSW({
		onNeedRefresh() {
			console.log("New content is available, please refresh.");
		},
		onOfflineReady() {
			console.log("App is ready to work offline.");
		},
	});
} else {
	console.log("Skipping service worker registration on localhost");
}

const rootElement = document.querySelector("#root") as Element;
if (!rootElement.innerHTML) {
	const root = ReactDOM.createRoot(rootElement);
	root.render(
		<React.StrictMode>
			<App router={router} />
		</React.StrictMode>
	);
}
