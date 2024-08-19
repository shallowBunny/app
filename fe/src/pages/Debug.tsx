// src/pages/Debug.tsx
import { FunctionComponent } from "react";
import { useQuery } from "@tanstack/react-query";
import { fetchData } from "../lib/fetchData";

interface Data {
	sets: Set[];
}

interface Set {
	dj: string;
	room: string;
}

export const Debug: FunctionComponent = () => {
	const { data, error, isLoading } = useQuery<Data, Error>({
		queryKey: ["fetchData"],
		queryFn: fetchData,
	});

	if (isLoading) return <div>Loading...</div>;
	if (error) return <div>Error: {error.message}</div>;
	console.log(data);
	return (
		<div>
			<h1>Data List</h1>
			<ul>{data?.sets.map((item) => <li key={item.dj}>{item.room}</li>)}</ul>
		</div>
	);
};
