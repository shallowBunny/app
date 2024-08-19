// vite.config.ts
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import { TanStackRouterVite } from "@tanstack/router-vite-plugin";
import { VitePWA } from "vite-plugin-pwa";
import path from "path";
import history from "connect-history-api-fallback";

export default defineConfig({
	plugins: [
		react(),
		TanStackRouterVite(),
		VitePWA({
			registerType: "autoUpdate",
			includeAssets: [
				"favicon.svg",
				"favicon.ico",
				"robots.txt",
				"apple-touch-icon.png",
			],
			manifest: {
				name: "Demo app",
				short_name: "Demo app",
				description: "An app to display DJ sets",
				theme_color: "#222123",
				background_color: "#222123",
				icons: [
					{
						src: "demo-192x192.png",
						sizes: "192x192",
						type: "image/png",
						purpose: "any",
					},
					{
						src: "demo-180x180.png",
						sizes: "180x180",
						type: "image/png",
						purpose: "maskable",
					},
					{
						src: "demo-192x192.png",
						sizes: "192x192",
						type: "image/png",
						purpose: "maskable",
					},
				],
			},
			workbox: {
				sourcemap: false,
				globPatterns: ["**/*.{js,css,html,png,svg,jpeg,jpg}"],
				runtimeCaching: [
					{
						urlPattern: ({ request }) =>
							request.destination === "document" ||
							request.destination === "script" ||
							request.destination === "style" ||
							request.destination === "image" ||
							request.destination === "font",
						handler: "CacheFirst",
						options: {
							cacheName: "static-assets",
							expiration: {
								maxEntries: 100,
								maxAgeSeconds: 60 * 60 * 24 * 30, // 30 days
							},
						},
					},
					{
						urlPattern: /https:\/\/.*.shallowbunny.com\/api/,
						handler: "NetworkFirst",
						options: {
							cacheName: "api-cache",
							networkTimeoutSeconds: 10,
							expiration: {
								maxEntries: 50,
								maxAgeSeconds: 60 * 60 * 24 * 7, // 7 days
							},
							cacheableResponse: {
								statuses: [0, 200],
							},
						},
					},
					{
						// Add shallowbunny.com API endpoint caching
						urlPattern: /https:\/\/shallowbunny.com\/api/,
						handler: "NetworkFirst",
						options: {
							cacheName: "shallowbunny-api-cache",
							networkTimeoutSeconds: 10,
							expiration: {
								maxEntries: 50,
								maxAgeSeconds: 60 * 60 * 24 * 7, // 30 days
							},
							cacheableResponse: {
								statuses: [0, 200],
							},
						},
					},
					{
						// Add shallowbunny.com API endpoint caching
						urlPattern: /https:\/\/sisyduck.com\/api/,
						handler: "NetworkFirst",
						options: {
							cacheName: "sisyduck-api-cache",
							networkTimeoutSeconds: 10,
							expiration: {
								maxEntries: 50,
								maxAgeSeconds: 60 * 60 * 24 * 7, // 7 days
							},
							cacheableResponse: {
								statuses: [0, 200],
							},
						},
					},
					{
						urlPattern: /.*/,
						handler: "NetworkFirst",
						options: {
							cacheName: "default-cache",
							networkTimeoutSeconds: 10,
							expiration: {
								maxEntries: 200,
								maxAgeSeconds: 60 * 60 * 24 * 30, // 30 days
							},
						},
					},
				],
			},
			devOptions: {
				enabled: true,
				type: "module",
				navigateFallback: "index.html",
			},
		}),
	],
	resolve: {
		alias: {
			"@": path.resolve(__dirname, "./src"),
		},
	},

	server: {
		host: true,
		strictPort: true,
		port: 5173,
		// Add a configureServer function to handle history API fallback
		configureServer: (server) => {
			server.middlewares.use(
				history({
					index: "/index.html",
					disableDotRule: true,
					htmlAcceptHeaders: ["text/html", "application/xhtml+xml"],
				})
			);
		},
	},
	build: {
		sourcemap: false,
	},
	esbuild: {
		sourcemap: false,
	},
	css: {
		devSourcemap: false,
	},
	define: {
		__BUILD_DATE__: JSON.stringify(new Date().toISOString()),
	},
});
