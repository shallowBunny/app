// lineup/$festival/$stage
import { createRoute } from "@tanstack/react-router";
import { useParams } from "@tanstack/react-router";
import Layout from "@/pages/Layout";
import { rootRoute } from "@/routes/__root"; // Ensure the correct parent route is imported

export default function LineupPage() {
	// Explicitly type the params as 'festival' and 'stage'
	//const params = useParams();

	const params = useParams({ from: "/lineup/$festival/$stage" });
	return <Layout festival={params.festival} stage={params.stage} />;
}

// Export the Route object
export const Route = createRoute({
	getParentRoute: () => rootRoute,
	path: "/lineup/$festival/$stage",
	component: LineupPage,
});
