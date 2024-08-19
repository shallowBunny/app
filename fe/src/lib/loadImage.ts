// src/lib/loadImage.ts

export const loadImageAsync = (src: string): Promise<string> => {
	return new Promise((resolve, reject) => {
		const img = new Image();
		img.src = src;
		img.onload = () => resolve(src);
		img.onerror = (err) => reject(err);
	});
};
