// src/declarations.d.ts

declare module "*.jpg" {
	const value: string;
	export default value;
}

declare module "*.png" {
	const value: string;
	export default value;
}

declare module "fast-levenshtein" {
	namespace levenshtein {
		function get(a: string, b: string): number;
	}
	export = levenshtein;
}
