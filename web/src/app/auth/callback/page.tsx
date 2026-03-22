"use client";

import { useSearchParams } from "next/navigation";
import { useEffect } from "react";

export default function AuthCallback() {
	const searchParams = useSearchParams();

	useEffect(() => {
		const token = searchParams.get("access_token");
		if (token) {
			localStorage.setItem("access_token", token);
			window.location.href = "/";
		}
	}, [searchParams]);

	return (
		<div className="py-10 text-center">
			<p className="text-gray-500">Authenticating...</p>
		</div>
	);
}
