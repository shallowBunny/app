// useCurrentMinute.ts
import { useState, useEffect } from "react";

const useCurrentMinute = () => {
	const [currentMinute, setCurrentMinute] = useState<Date>(new Date());

	useEffect(() => {
		const interval = setInterval(() => {
			const now = new Date();
			if (now.getMinutes() !== currentMinute.getMinutes()) {
				setCurrentMinute(now);
			}
		}, 1000);

		return () => clearInterval(interval);
	}, [currentMinute]);

	return currentMinute;
};

export default useCurrentMinute;
