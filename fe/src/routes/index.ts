// index.ts

import { createFileRoute } from "@tanstack/react-router";
import Layout from "@/pages/Layout";

// Base layout route without a specific component
export const Route = createFileRoute("/")({
	component: Layout,
});
